package cluster

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/armory/armory-cli/cmd/agent"
	"github.com/armory/armory-cli/cmd/login"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/armory/armory-cli/pkg/configuration"
	errorUtils "github.com/armory/armory-cli/pkg/errors"
	"github.com/armory/armory-cli/pkg/model"
	"github.com/armory/armory-cli/pkg/output"
	"github.com/samber/lo"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"io/ioutil"
	"math/rand"
	"os"
	"time"
)

const (
	sandboxFilePath    = "/.armory/sandbox"
	charset            = "abcdefghijklmnopqrstuvwxyz0123456789"
	createClusterShort = "Creates a temporary kubernetes cluster"
	createClusterLong  = "Creates a temporary kubernetes cluster for demo purposes. The created cluster is helpful for evaluating CD-as-a-Service and will be \n\n" +
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
	saveData      *model.SandboxClusterSaveData
}

func NewCreateClusterCmd(configuration *config.Configuration) *cobra.Command {
	o := NewCreateOptions()

	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{},
		Short:   createClusterShort,
		Long:    createClusterLong,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.InitializeConfiguration(configuration)
			return o.Run(cmd)
				return err
			}
			return nil
		},
		SilenceUsage: true,
	}
	return cmd
}

func NewCreateOptions() *CreateOptions {
	return &CreateOptions{}
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
	ctx := cmd.Context()
	agentPrefix := randomString(6)
	credentials, err := o.ArmoryClient.Credentials().Create(ctx, o.createNamedCredential(agentPrefix))
	if err != nil {
		return err
	}
	environmentId := lo.If(lo.FromPtrOr[bool](o.configuration.GetIsTest(), false), "test-env").ElseF(o.configuration.GetCustomerEnvironmentId)
	err = AssignCredentialRNARole(ctx, credentials, o.ArmoryClient, environmentId)
	if err != nil {
		return err
	}
	createSandboxRequest := o.createSandboxRequest(agentPrefix, credentials)
	sandboxResponse, err := o.ArmoryClient.Sandbox().Create(ctx, createSandboxRequest)
	if err != nil {
		return err
	}
	o.InitializeProgressBar()
	o.saveData = &model.SandboxClusterSaveData{
		SandboxCluster:        model.SandboxCluster{},
		CreateSandboxRequest:  *createSandboxRequest,
		CreateSandboxResponse: *sandboxResponse,
	}
	for {
		cluster, err := o.ArmoryClient.Sandbox().Get(ctx, o.saveData.CreateSandboxResponse.ClusterId)
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

	return nil
}

func (o *CreateOptions) UpdateProgressBar(cluster *model.SandboxCluster) (bool, error) {
	o.saveData.SandboxCluster = *cluster
	o.progressbar.Describe(cluster.Status)
	err := o.writeToSaveDataToSandboxFile()
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
func (o *CreateOptions) InitializeProgressBar() {
	o.progressbar = progressbar.NewOptions(100,
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

func (o *CreateOptions) writeToSaveDataToSandboxFile() error {
	fileLocation, err := o.getSandboxFileLocation()
	if err != nil {
		return err
	}
	err = o.saveData.WriteToFile(fileLocation)
	if err != nil {
		return errorUtils.NewWrappedError(ErrWritingSandboxSaveData, err)
	}
	return nil
}

func (o *CreateOptions) readSandboxFromFile() (*model.SandboxClusterSaveData, error) {
	fileLocation, err := o.getSandboxFileLocation()
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadFile(fileLocation)
	if err != nil {
		return nil, err
	}
	var cluster model.SandboxClusterSaveData
	err = json.Unmarshal(data, &cluster)
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}

func (o *CreateOptions) getSandboxFileLocation() (string, error) {
	_, isATest := os.LookupEnv("ARMORY_CLI_TEST")
	var dirname string
	var err error
	if isATest {
		return os.TempDir() + "dotarmory_sandbox", nil
	}

	dirname, err = os.UserHomeDir()
	if err != nil {
		return "", errorUtils.NewWrappedError(login.ErrGettingHomeDirectory, err)
	}

	return dirname + sandboxFilePath, nil
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
	if err != nil {
		return err
	}

	return nil
}
