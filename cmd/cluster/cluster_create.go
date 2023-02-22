package cluster

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/armory/armory-cli/cmd/login"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/armory/armory-cli/pkg/configuration"
	errorUtils "github.com/armory/armory-cli/pkg/errors"
	"github.com/armory/armory-cli/pkg/model"
	"github.com/armory/armory-cli/pkg/output"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"time"
)

const (
	createClusterShort = "Creates a temporary kubernetes cluster"
	createClusterLong  = "Creates a temporary kubernetes cluster for demo purposes. The created cluster is helpful for evaluating CD-as-a-Service and will be \n\n" +
		"automatically deleted within two hours. Only the CD-as-a-Service sample application (potato-facts) and remote network agent can be installed."
)

var (
	agentConnectedPollRate    = time.Minute * 10
	ErrOutputTypeNotSupported = errors.New("output type is not supported. Choose type 'text' to use this feature")
	ErrWritingSandboxSaveData = errors.New("unable to save sandbox data to file system")
)

type CreateOptions struct {
	Context context.Context

	ArmoryClient  *configuration.ConfigClient
	configuration *config.Configuration
	contextNames  []string
	credentials   *model.Credential
}

func NewCreateClusterCmd(configuration *config.Configuration) *cobra.Command {
	o := NewCreateOptions(configuration)

	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{},
		Short:   createClusterShort,
		Long:    createClusterLong,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Run(cmd); err != nil {
				return err
			}
			return nil
		},
		SilenceUsage: true,
	}
	return cmd
}

func NewCreateOptions(cfg *config.Configuration) *CreateOptions {
	o := &CreateOptions{}
	o.configuration = cfg

	ac := configuration.NewClient(cfg)
	o.ArmoryClient = ac
	o.Context = o.ArmoryClient.ArmoryCloudClient.Context
	return o
}

// Run performs the execution of 'cluster create' sub command and saves the info for later use
func (o *CreateOptions) Run(cmd *cobra.Command) error {
	if o.configuration.GetOutputType() != output.Text {
		return ErrOutputTypeNotSupported
	}

	ctx := context.Background()

	credentials, err := o.ArmoryClient.Credentials().Create(ctx, o.createNamedCredential())
	if err != nil {
		return err
	}

	sandboxResponse, err := o.ArmoryClient.Sandbox().Create(ctx, o.createSandboxRequest(credentials))
	if err != nil {
		return err
	}

	bar := progressbar.Default(100)
	for {
		cluster, err := o.ArmoryClient.Sandbox().Get(ctx, sandboxResponse.ClusterId)
		if err != nil {
			return err
		}
		err = o.writeToSandboxFile(cluster)
		if err != nil {
			return err
		}
		if cluster.Status == "READY" {
			_ = bar.Set(100)
			break
		}
		progress := float64(cluster.PercentComplete) - bar.State().CurrentPercent
		if progress > 0 {
			err := bar.Add(int(progress))
			if err != nil {
				fmt.Printf(".")
			}
		}
	}

	return nil
}

// createCredentials outputs a credentials object using the configured fields
func (o *CreateOptions) createNamedCredential() *model.Credential {
	c := 6
	prefix := make([]byte, c)
	_, err := rand.Read(prefix)
	if err != nil {
		prefix = []byte("a1b2bc")
	}

	return &model.Credential{
		Name: fmt.Sprintf("%s-temp-cluster-credentials", prefix),
	}
}

// createSandboxRequest outputs a sandboxRequest object using the configured fields
func (o *CreateOptions) createSandboxRequest(credential *model.Credential) *model.CreateSandboxRequest {
	return &model.CreateSandboxRequest{
		AgentIdentifier: "default-sandbox-rna",
		ClientId:        credential.ClientId,
		ClientSecret:    credential.ClientSecret,
	}
}

func (o *CreateOptions) writeToSandboxFile(cluster *model.SandboxCluster) error {
	fileLocation, err := o.getSandboxFileLocation()
	if err != nil {
		return err
	}
	err = cluster.SaveData(fileLocation)
	if err != nil {
		return errorUtils.NewWrappedError(ErrWritingSandboxSaveData, err)
	}
	return nil
}

func (o *CreateOptions) readSandboxFromFile() (*model.SandboxCluster, error) {
	fileLocation, err := o.getSandboxFileLocation()
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadFile(fileLocation)
	if err != nil {
		return nil, err
	}
	var cluster model.SandboxCluster
	err = json.Unmarshal(data, &cluster)
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}

func (o *CreateOptions) getSandboxFileLocation() (string, error) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return "", errorUtils.NewWrappedError(login.ErrGettingHomeDirectory, err)
	}
	return dirname + "/.armory/sandbox", nil
}
