package cluster

import (
	"context"
	"errors"
	"fmt"
	"github.com/armory/armory-cli/cmd/agent"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/armory/armory-cli/pkg/configuration"
	"github.com/armory/armory-cli/pkg/model"
	"github.com/armory/armory-cli/pkg/output"
	"github.com/samber/lo"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"io"
	"math/rand"
	"time"
)

const (
	sandboxFilePath    = "/.armory/sandbox"
	charset            = "abcdefghijklmnopqrstuvwxyz0123456789"
	createClusterShort = "Creates a temporary kubernetes cluster"
	createClusterLong  = "Creates a temporary kubernetes cluster for demo purposes. The created cluster is helpful for evaluating CD-as-a-Service and will be \n" +
		"automatically deleted within two hours. Only the CD-as-a-Service sample application (potato-facts) and Remote Network Agent can be installed."
)

var (
	ErrOutputTypeNotSupported = errors.New("output type is not supported. Choose type 'text' to use this feature")
	ErrWritingSandboxSaveData = errors.New("unable to save sandbox data to file system")
)

type CreateOptions struct {
	Context       context.Context
	ArmoryClient  *configuration.ConfigClient
	configuration *config.Configuration
	contextNames  []string
	credentials   *model.Credential
	progressbar   *progressbar.ProgressBar
	saveData      SandboxStorage
}

func NewCreateClusterCmd(configuration *config.Configuration, store SandboxStorage) *cobra.Command {
	o := NewCreateOptions(store)

	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{},
		Short:   createClusterShort,
		Long:    createClusterLong,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.InitializeConfiguration(configuration)
			return o.Run(cmd)
		},
		SilenceUsage: true,
	}
	return cmd
}

func NewCreateOptions(store SandboxStorage) *CreateOptions {
	return &CreateOptions{
		saveData: store,
	}
}

// InitializeConfiguration will fetch Auth info to configure the REST client, so initialization must be deferred to runtime
func (o *CreateOptions) InitializeConfiguration(cfg *config.Configuration) {
	o.configuration = cfg
	ac := configuration.NewClient(cfg)
	o.ArmoryClient = ac
	o.Context = o.ArmoryClient.ArmoryCloudClient.Context
}

// Run performs the execution of 'cluster create' sub command and saves the info for later use
func (o *CreateOptions) Run(cmd *cobra.Command) error {
	if o.configuration.GetOutputType() != output.Text {
		return ErrOutputTypeNotSupported
	}
	isTest := o.configuration.GetIsTest()
	ctx := cmd.Context()
	agentPrefix := randomString(6)
	credentials, err := o.ArmoryClient.Credentials().Create(ctx, o.createNamedCredential(agentPrefix))
	if err != nil {
		return err
	}
	environmentId := lo.If(lo.FromPtrOr[bool](isTest, false), "test-env").ElseF(o.configuration.GetCustomerEnvironmentId)
	err = AssignCredentialRNARole(ctx, credentials, o.ArmoryClient, environmentId)
	if err != nil {
		return err
	}
	createSandboxRequest := o.createSandboxRequest(agentPrefix, credentials)
	sandboxResponse, err := o.ArmoryClient.Sandbox().Create(ctx, createSandboxRequest)
	if err != nil {
		return err
	}
	o.InitializeProgressBar(cmd.OutOrStdout())
	o.saveData.setAgentIdentifier(createSandboxRequest.AgentIdentifier)
	o.saveData.setCreateSandboxResponse(*sandboxResponse)
	for {
		cluster, err := o.ArmoryClient.Sandbox().Get(ctx, o.saveData.getClusterId())
		if err != nil {
			return err
		}

		done, err := o.UpdateProgressBar(cluster)
		if err != nil {
			return err
		}

		if done {
			break
		}
	}

	cmd.Printf("\n\nTo use your temporary sandbox cluster, create a cluster preview. Run: `armory preview create --duration 2h --type cluster --agent %s`\n", createSandboxRequest.AgentIdentifier)
	return nil
}

func (o *CreateOptions) UpdateProgressBar(cluster *model.SandboxCluster) (bool, error) {
	o.saveData.setClusterData(cluster)
	o.progressbar.Describe(cluster.Status)
	err := o.saveData.writeToSandboxFile()
	if err != nil {
		return true, err
	}
	err = o.progressbar.Set(int(cluster.PercentComplete))
	if err != nil {
		return true, err
	}
	return cluster.PercentComplete == 100, nil

}

// InitializeProgressBar will create and display into StdOut the progress bar
func (o *CreateOptions) InitializeProgressBar(writer io.Writer) {
	o.progressbar = progressbar.NewOptions(100,
		progressbar.OptionSetWriter(writer),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetElapsedTime(true),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionShowBytes(false),
		progressbar.OptionSetWidth(15),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))
}

// createCredentials outputs a credentials object using the configured fields
func (o *CreateOptions) createNamedCredential(prefix string) *model.Credential {
	return &model.Credential{
		Name: fmt.Sprintf("%s-temp-cluster-credentials", prefix),
	}
}

func randomString(length int) string {
	seededRand := rand.New(
		rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// createSandboxRequest outputs a sandboxRequest object using the configured fields
func (o *CreateOptions) createSandboxRequest(prefix string, credential *model.Credential) *model.CreateSandboxRequest {
	return &model.CreateSandboxRequest{
		AgentIdentifier: fmt.Sprintf("%s-sandbox-rna", prefix),
		ClientId:        credential.ClientId,
		ClientSecret:    credential.ClientSecret,
	}
}

func AssignCredentialRNARole(ctx context.Context, credential *model.Credential, armoryClient *configuration.ConfigClient, envId string) error {
	existingRoles, err := armoryClient.Roles().ListForMachinePrincipals(ctx, envId)
	if err != nil {
		return err
	}

	// add the RNA role to the newly created credentials
	role, roleExists := lo.Find(existingRoles, func(c model.RoleConfig) bool {
		_, hasRightPermissions := lo.Find(c.Grants, func(g model.GrantConfig) bool {
			return g.Type == "api" && g.Resource == "agentHub" && g.Permission == "full"
		})
		return hasRightPermissions && c.SystemDefined
	})

	if !roleExists {
		return agent.ErrRoleMissing
	}

	_, err = armoryClient.Credentials().AddRoles(ctx, credential, []string{role.ID})
	return err
}
