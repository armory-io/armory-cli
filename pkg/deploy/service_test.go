package deploy

import (
	"fmt"
	"github.com/armory/armory-cli/pkg/model"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestCreateDeploymentRequestSuccess(t *testing.T){
	targets := map[string]model.DeploymentTarget{
		"test": model.DeploymentTarget{
			Account: "account1",
			Namespace: "dev",
			Strategy: "strategy1",
		},
	}
	strategies := map[string]model.Strategy{
		"strategy1": model.Strategy{
			Canary: &model.CanaryStrategy{
				Steps: &[]model.CanaryStep{
					model.CanaryStep {
						SetWeight: &model.WeightStep{
							Weight: 33,
						},
					},
					model.CanaryStep {
						Pause: &model.PauseStep{
							UntilApproved: true,
						},
					},
					model.CanaryStep {
						Pause: &model.PauseStep{
							Duration: 600,
							Unit: "SECONDS",
						},
					},
				},
			},
		},
	}

	tmpDir := t.TempDir()
	tempFile1 := tempAppFile(tmpDir, "app1*.yml",testAppYamlStr)
	if tempFile1 == nil {
		t.Fatal("TestGetManifestsFromFileSuccess failed with: Could not create temp app file.")
	}
	manifests := []model.ManifestPath{
		{
			Path: tempFile1.Name(),
		},
	}

	orchestration := model.OrchestrationConfig{
		Version: "v1",
		Kind: "kubernetes",
		Application: "app",
		Targets: &targets,
		Strategies: &strategies,
		Manifests: &manifests,
	}

	received, err := CreateDeploymentRequest(&orchestration)
	if err != nil {
		t.Fatalf("TestCreateDeploymentRequestSuccess failed with: %s", err)
	}
	receivedJson, err := received.MarshalJSON()
	if err != nil {
		t.Fatalf("TestCreateDeploymentRequestSuccess failed with: %s", err)
	}
	expected, err := ioutil.ReadFile("testdata/deploymentRequest.json")
	if err != nil {
		t.Fatalf("TestCreateDeploymentRequestSuccess failed with: Error loading tesdata file %s", err)
	}
	assert.JSONEq(t, string(receivedJson), string(expected), "json should be the same")
}

func TestGetManifestsFromPathSuccess(t *testing.T){
	tmpDir := t.TempDir()
	tempFile1 := tempAppFile(tmpDir, "app1*.yml",testAppYamlStr)
	if tempFile1 == nil {
		t.Fatal("TestGetManifestsFromFileSuccess failed with: Could not create temp app file.")
	}
	defer os.Remove(tempFile1.Name())
	tempFile2 := tempAppFile(tmpDir, "app2*.yml", testAppYamlStr)
	if tempFile2 == nil {
		t.Fatal("TestGetManifestsFromFileSuccess failed with: Could not create temp app file.")
	}
	defer os.Remove(tempFile2.Name())
	manifests := []model.ManifestPath{
		{
			Path: tempFile1.Name(),
		},
		{
			Path: tempFile2.Name(),
		},
	}
	files, err := GetManifestsFromFile(&manifests)
	if err != nil {
		t.Fatalf("TestGetManifestsFromPathSuccess failed with: %s", err)
	}
	assert.Equal(t, len(*files), 2)
}

func TestCreateDeploymentManifestsSuccess(t *testing.T){
	manifests := make([]string, 2)
	manifests[0] = testAppYamlStr
	manifests[1] = testAppYamlStr
	received := CreateDeploymentManifests(&manifests)
	assert.Equal(t, len(received), 2)
}

func TestCreateDeploymentCanaryStepSuccess(t *testing.T){
	weight := int32(33)
	untilApproved := true
	duration := int32(600)
	strategy := model.Strategy{
		Canary: &model.CanaryStrategy{
			Steps: &[]model.CanaryStep{
				model.CanaryStep {
					SetWeight: &model.WeightStep{
						Weight: weight,
					},
				},
				model.CanaryStep {
					Pause: &model.PauseStep{
						UntilApproved: untilApproved,
					},
				},
				model.CanaryStep {
					Pause: &model.PauseStep{
						Duration: duration,
						Unit: "SECONDS",
					},
				},
			},
		},
	}
	received, err := CreateDeploymentCanaryStep(strategy)
	if err != nil {
		t.Fatalf("TestCreateDeploymentCanaryStepSuccess failed with: %s", err)
	}
	assert.Equal(t, len(received), len(*strategy.Canary.Steps))
}

func tempAppFile(tmpDir, fileName, fileContent string) *os.File {
	tempFile, _ := ioutil.TempFile(tmpDir, fileName)
	bytes, err := tempFile.Write([]byte(fileContent))
	if err != nil || bytes == 0 {
		fmt.Println("Could not write temp file.")
		return nil
	}
	return tempFile
}

const testAppYamlStr = `
apiVersion: apps/v1
kind: Deployment
`