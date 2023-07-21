package cluster

import (
	"encoding/json"
	"os"

	"github.com/armory/armory-cli/cmd/login"
	errorUtils "github.com/armory/armory-cli/pkg/errors"
	"github.com/armory/armory-cli/pkg/model"
)

type (
	SandboxClusterFileStore struct {
		saveData model.SandboxSaveData
	}

	SandboxStorage interface {
		writeToSandboxFile() error
		readSandboxFromFile() (*model.SandboxSaveData, error)
		getClusterId() string
		setClusterData(clusterData *model.SandboxCluster)
		setAgentIdentifier(agentIdentifier string)
		setCreateSandboxResponse(response model.CreateSandboxResponse)
	}
)

func (d *SandboxClusterFileStore) getClusterId() string {
	return d.saveData.CreateSandboxResponse.ClusterId
}

func (d *SandboxClusterFileStore) setClusterData(clusterData *model.SandboxCluster) {
	d.saveData.SandboxCluster = *clusterData
}

func (d *SandboxClusterFileStore) setAgentIdentifier(agentIdentifier string) {
	d.saveData.AgentIdentifier = agentIdentifier
}

func (d *SandboxClusterFileStore) setCreateSandboxResponse(response model.CreateSandboxResponse) {
	d.saveData.CreateSandboxResponse = response
}

// writeToSandboxFile stores the data for debugging info or future use by other commands
func (d *SandboxClusterFileStore) writeToSandboxFile() error {
	fileLocation, err := d.getSandboxFileLocation()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(d.saveData, "", " ")
	if err != nil {
		return err
	}
	if err = os.WriteFile(fileLocation, data, 0644); err != nil {
		return errorUtils.NewWrappedError(ErrWritingSandboxSaveData, err)
	}
	return nil
}

func (d *SandboxClusterFileStore) readSandboxFromFile() (*model.SandboxSaveData, error) {
	fileLocation, err := d.getSandboxFileLocation()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(fileLocation)
	if err != nil {
		return nil, err
	}
	var saveData model.SandboxSaveData
	return &saveData, json.Unmarshal(data, &saveData)
}

func (d *SandboxClusterFileStore) getSandboxFileLocation() (string, error) {
	_, isATest := os.LookupEnv("ARMORY_CLI_TEST")
	var dirname string
	var err error
	if isATest {
		return os.TempDir() + "/.armory_sandbox", nil
	}

	dirname, err = os.UserHomeDir()
	if err != nil {
		return "", errorUtils.NewWrappedError(login.ErrGettingHomeDirectory, err)
	}

	return dirname + sandboxFilePath, nil
}
