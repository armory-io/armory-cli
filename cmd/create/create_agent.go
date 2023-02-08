package create

import (
	"context"
	"errors"
	"fmt"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/armory/armory-cli/pkg/configuration"
	"github.com/armory/armory-cli/pkg/model"
	"github.com/armory/armory-cli/pkg/util"
	"github.com/manifoldco/promptui"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes/scheme"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubectl/pkg/cmd/apply"
	"k8s.io/kubectl/pkg/cmd/delete"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"time"
)

const (
	agentShort   = "Create an agent"
	agentLong    = "Create an agent"
	agentExample = `
	  # Create a new agent
	  armory create agent`

	defaultNamespaceName = "armory-rna"
	defaultSecretName    = "rna-client-credentials"
)

var agentConnectedPollRate = time.Second * 10

type AgentOptions struct {
	// Name of resource being created
	Name      string
	Namespace string

	Context context.Context

	ArmoryClient      *configuration.ConfigClient
	configuration     *config.Configuration
	contextNames      []string
	credentials       *model.Credential
	kubernetesFactory cmdutil.Factory
	configAccess      clientcmd.ConfigAccess
	KubernetesClient  corev1client.CoreV1Interface
}

// NewAgentOptions creates a new *AgentOptions with sane defaults
func NewAgentOptions() *AgentOptions {
	return &AgentOptions{}
}

func NewCmdCreateAgent(configuration *config.Configuration) *cobra.Command {

	o := NewAgentOptions()

	cmd := &cobra.Command{
		Use:     "agent",
		Aliases: []string{},
		Short:   agentShort,
		Long:    agentLong,
		Example: agentExample,
		RunE: func(cmd *cobra.Command, args []string) error {

			if err := o.Complete(configuration); err != nil {
				return err
			}

			if err := o.Validate(); err != nil {
				return err
			}

			if err := o.Run(); err != nil {
				return err
			}
			return nil
		},
	}
	return cmd
}

func (o *AgentOptions) Complete(cfg *config.Configuration) error {

	o.configAccess = clientcmd.NewDefaultPathOptions()
	o.configuration = cfg

	ac := configuration.NewClient(cfg)
	o.ArmoryClient = ac

	f := o.getKubernetesFactory()
	o.kubernetesFactory = f

	kc, err := o.getKubernetesClient()
	if err != nil {
		return err
	}
	o.KubernetesClient = kc

	contextNames, err := o.getContexts()
	if err != nil {
		return err
	}
	o.contextNames = contextNames
	o.Context = o.ArmoryClient.ArmoryCloudClient.Context
	return nil
}

// Run performs the execution of 'create agent' sub command
func (o *AgentOptions) Run() error {
	promptSelectAgent := promptui.Select{
		Label:  "Please select a context. Your agent will be deployed into the cluster you choose",
		Items:  o.contextNames,
		Stdout: &util.BellSkipper{},
	}

	_, requestedContext, err := promptSelectAgent.Run()
	if err != nil {
		return errors.New(fmt.Sprintf("failed to select a context to deploy to; %v\n", err))
	}

	err = o.useContext(requestedContext)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to set context %s; %v\n", requestedContext, err))
	}

	ctx, cancel := context.WithTimeout(o.Context, time.Minute)
	defer cancel()

	// fetch the list of agents
	existingAgents, err := o.ArmoryClient.Agents().List(ctx)
	if err != nil {
		return err
	}

	agentNameAlreadyExistFunc := func(name string) bool {
		_, exists := lo.Find(existingAgents, func(a model.Agent) bool {
			return name == a.AgentIdentifier
		})
		return exists
	}

	// set agent name
	promptSetAgentName := promptui.Prompt{
		Label: fmt.Sprintf("Provide an agent identifier %s", lo.Ternary(agentNameAlreadyExistFunc(requestedContext), "", fmt.Sprintf("[default=%s]", requestedContext))),
		Validate: func(name string) error {
			if agentNameAlreadyExistFunc(name) {
				return errors.New("sorry, there's already an agent with that name in your tenant")
			}
			return nil
		},
	}

	agentName, err := promptSetAgentName.Run()
	if err != nil {
		return errors.New(fmt.Sprintf("failed to set the agent name; %v\n", err))
	}

	if lo.IsEmpty(agentName) {
		agentName = requestedContext
	}

	o.Name = agentName

	// set namespace
	promptSetNamespace := promptui.Prompt{
		Label:  fmt.Sprintf("Provide a namespace where the agent will be installed [default=%s]", defaultNamespaceName),
		Stdout: &util.BellSkipper{},
	}

	namespaceName, err := promptSetNamespace.Run()
	if err != nil {
		return errors.New(fmt.Sprintf("failed to set the namespace; %v\n", err))
	}

	if lo.IsEmpty(namespaceName) {
		namespaceName = defaultNamespaceName
	}

	o.Namespace = namespaceName

	// fetch the list of credentials
	existingCredentials, err := o.ArmoryClient.Credentials().List(ctx)
	if err != nil {
		return err
	}

	// create new credentials
	credentials := o.createCredentials()
	existingCredential, credentialsExists := lo.Find(existingCredentials, func(c *model.Credential) bool {
		return credentials.Name == c.Name
	})

	if credentialsExists {
		// recreate credentials
		promptRecreateCredentials := promptui.Prompt{
			Label:     fmt.Sprintf("A Client Credential named %s already exists. Do you want to generate a new Client Credentials?", o.Name),
			IsConfirm: true,
			Default:   "Y",
			Stdout:    &util.BellSkipper{},
		}

		if _, err := promptRecreateCredentials.Run(); err != nil {
			return errors.New(fmt.Sprintf("Exiting %s\n", err))
		}

		err = o.ArmoryClient.Credentials().Delete(ctx, existingCredential)
		if err != nil {
			return errors.New(fmt.Sprintf("failed to delete credentials; %v\n", err))
		}
	}

	credentials, err = o.ArmoryClient.Credentials().Create(ctx, credentials)
	if err != nil {
		return err
	}

	existingRoles, err := o.ArmoryClient.Roles().ListForMachinePrincipals(ctx, o.configuration.GetCustomerEnvironmentId())
	if err != nil {
		return err
	}

	// add the RNA role to the newly created credentials
	rol, rolExists := lo.Find(existingRoles, func(c model.RoleConfig) bool {
		return "Remote Network Agent" == c.Name
	})

	if !rolExists {
		return errors.New("The default role Remote Network Agent role was missing, please ask your tenant admins to recreate it.")
	}

	_, err = o.ArmoryClient.Credentials().AddRoles(ctx, credentials, []string{rol.ID})
	if err != nil {
		return err
	}

	o.credentials = credentials

	// create new namespace if not exist
	if exist, _ := o.namespaceExist(); !exist {
		_, err := o.createNamespace()
		if err != nil {
			return errors.New(fmt.Sprintf("failed to create namespace; %v\n", err))
		}
	}

	// create new secret
	createSecretOptions := metav1.CreateOptions{}
	secret := o.createSecret()
	secret, err = o.KubernetesClient.Secrets(o.Namespace).Create(ctx, secret, createSecretOptions)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to create secret; %v\n", err))
	}

	// apply manifests
	err = o.apply(o.Namespace, fmt.Sprintf("%s/kubernetes/agent/manifest?agentIdentifier=%s&namespace=%s", o.configuration.GetArmoryCloudAddr(), o.Name, o.Namespace))
	if err != nil {
		return errors.New(fmt.Sprintf("failed to apply manifests; %v\n", err))
	}

	// poll
	o.waitForConnection()
	return nil
}

// getKubernetesFactory outputs the Kubernetes Factory
func (o *AgentOptions) getKubernetesFactory() cmdutil.Factory {
	var defaultConfigFlags = genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()

	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(defaultConfigFlags)
	return cmdutil.NewFactory(matchVersionKubeConfigFlags)
}

// getKubernetesClient outputs the Kubernetes Client
func (o *AgentOptions) getKubernetesClient() (corev1client.CoreV1Interface, error) {
	restConfig, err := o.kubernetesFactory.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	kubernetesClient, err := corev1client.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	return kubernetesClient, nil
}

// getContexts outputs the list of contexts contained in the kubeconfig file
func (o *AgentOptions) getContexts() ([]string, error) {
	var contexts []string
	kubeconfig, err := o.configAccess.GetStartingConfig()
	if err != nil {
		return contexts, err
	}
	for name := range kubeconfig.Contexts {
		contexts = append(contexts, name)
	}
	return contexts, nil
}

// useContext set the current-context in a kubeconfig file
func (o *AgentOptions) useContext(contextName string) error {
	kubeconfig, err := o.configAccess.GetStartingConfig()
	if err != nil {
		return err
	}

	kubeconfig.CurrentContext = contextName

	return clientcmd.ModifyConfig(o.configAccess, *kubeconfig, true)
}

// createCredentials outputs a credentials object using the configured fields
func (o *AgentOptions) createCredentials() *model.Credential {
	return &model.Credential{
		Name: fmt.Sprintf("%s-rna-credentials", o.Name),
	}
}

// createNamespace outputs a namespace object using the configured fields
func (o *AgentOptions) createNamespace() (*corev1.Namespace, error) {
	createNamespaceOptions := metav1.CreateOptions{}
	namespace := &corev1.Namespace{
		TypeMeta:   metav1.TypeMeta{APIVersion: corev1.SchemeGroupVersion.String(), Kind: "Namespace"},
		ObjectMeta: metav1.ObjectMeta{Name: o.Namespace},
	}
	return o.KubernetesClient.Namespaces().Create(o.Context, namespace, createNamespaceOptions)
}

// namespaceExist check if the provided namespace exists
func (o *AgentOptions) namespaceExist() (bool, error) {
	namespaceExistOptions := metav1.ListOptions{}
	namespaces, err := o.KubernetesClient.Namespaces().List(o.Context, namespaceExistOptions)
	if err != nil {
		return false, err
	}
	_, exists := lo.Find(namespaces.Items, func(n corev1.Namespace) bool {
		return o.Namespace == n.Name
	})
	return exists, nil
}

// createSecret outputs a secret object using the configured fields
func (o *AgentOptions) createSecret() *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultSecretName,
			Namespace: defaultNamespaceName,
		},
		Type: "string",
		StringData: map[string]string{
			"client-id":     o.credentials.ClientId,
			"client-secret": o.credentials.ClientSecret,
		},
	}
}

// apply apply a configuration to a resource
func (o *AgentOptions) apply(namespace, resourceFile string) error {
	mapper, err := o.kubernetesFactory.ToRESTMapper()
	if err != nil {
		return err
	}
	dynamicClient, err := o.kubernetesFactory.DynamicClient()
	if err != nil {
		return err
	}

	openAPISchema, _ := o.kubernetesFactory.OpenAPISchema()

	printFlags := genericclioptions.NewPrintFlags("created").WithTypeSetter(scheme.Scheme)

	// allow for a success message operation to be specified at print time
	noopPrinter := func(operation string) (printers.ResourcePrinter, error) {
		return printFlags.ToPrinter()
	}

	applyOptions := &apply.ApplyOptions{
		PrintFlags: printFlags,

		DeleteOptions: &delete.DeleteOptions{
			FilenameOptions: resource.FilenameOptions{
				Filenames: []string{
					resourceFile,
				},
			},
		},
		ServerSideApply:   true,
		FieldManager:      apply.FieldManagerClientSideApply,
		Recorder:          genericclioptions.NoopRecorder{},
		Namespace:         namespace,
		EnforceNamespace:  true,
		Builder:           o.kubernetesFactory.NewBuilder(),
		Mapper:            mapper,
		DynamicClient:     dynamicClient,
		OpenAPISchema:     openAPISchema,
		ToPrinter:         noopPrinter,
		IOStreams:         genericclioptions.NewTestIOStreamsDiscard(),
		VisitedUids:       sets.NewString(),
		VisitedNamespaces: sets.NewString(),
	}
	return applyOptions.Run()
}

// waitForConnection poll for agents to determine if the agent has connected.
func (o *AgentOptions) waitForConnection() {
	fmt.Print("Waiting for agent to connect.")
	for range time.Tick(agentConnectedPollRate) {
		_, _ = fmt.Print(".")
		//	TODO
	}
}

// Validate validates required fields are set to support structured generation
func (o *AgentOptions) Validate() error {
	return nil
}
