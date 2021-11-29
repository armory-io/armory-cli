package deploy

import (
	"github.com/armory/armory-cli/pkg/model"
	"github.com/armory/armory-cli/pkg/util"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"testing"
)

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}

type ServiceTestSuite struct {
	suite.Suite
}

func (suite *ServiceTestSuite) SetupSuite() {
	os.Setenv("ARMORY_CLI_TEST", "true")
}

func (suite *ServiceTestSuite) TearDownSuite() {
	os.Unsetenv("ARMORY_CLI_TEST")
}

func (suite *ServiceTestSuite) TestCreateDeploymentRequestSuccess(){
	targets := map[string]model.DeploymentTarget{
		"test1": model.DeploymentTarget{
			Account: "account1",
			Namespace: "dev",
			Strategy: "strategy1",
		},
		"test2": model.DeploymentTarget{
			Account: "account2",
			Namespace: "qa",
			Strategy: "strategy2",
			Constraints: &model.Constraints{
				DependsOn: &[]string{
					"test1",
				},
				BeforeDeployment: &[]model.BeforeDeployment{
					model.BeforeDeployment {
						Pause: &model.PauseStep {
							UntilApproved: true,
						},
					},
				},
			},
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
		"strategy2": model.Strategy{
			Canary: &model.CanaryStrategy{
				Steps: &[]model.CanaryStep{
					model.CanaryStep {
						SetWeight: &model.WeightStep{
							Weight: 50,
						},
					},
					model.CanaryStep {
						Pause: &model.PauseStep{
							Duration: 900,
							Unit: "SECONDS",
						},
					},
				},
			},
		},
	}

	tempFile1 := util.TempAppFile("", "app1*.yml",testAppYamlStr)
	if tempFile1 == nil {
		suite.T().Fatal("TestGetManifestsFromFileSuccess failed with: Could not create temp app file.")
	}
	suite.T().Cleanup(func() { os.Remove(tempFile1.Name()) })
	manifests := []model.ManifestPath{
		{
			Path: tempFile1.Name(),
			Targets: []string{
				"test1",
				"test2",
			},
		},
		{
			Path: tempFile1.Name(),
			Targets: []string{
				"test1",
			},
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
		suite.T().Fatalf("TestCreateDeploymentRequestSuccess failed with: %s", err)
	}
	receivedJson, err := received.MarshalJSON()
	if err != nil {
		suite.T().Fatalf("TestCreateDeploymentRequestSuccess failed with: %s", err)
	}
	expected, err := ioutil.ReadFile("testdata/deploymentRequest.json")
	if err != nil {
		suite.T().Fatalf("TestCreateDeploymentRequestSuccess failed with: Error loading tesdata file %s", err)
	}
	suite.JSONEq(string(receivedJson), string(expected), "json should be the same")
}

func (suite *ServiceTestSuite) TestGetManifestsFromPathSuccess(){
	tempFile1 := util.TempAppFile("", "app1*.yml",testAppYamlStr)
	if tempFile1 == nil {
		suite.T().Fatal("TestGetManifestsFromFileSuccess failed with: Could not create temp app file.")
	}
	tempFile2 := util.TempAppFile("", "app2*.yml", testAppYamlStr)
	if tempFile2 == nil {
		suite.T().Fatal("TestGetManifestsFromFileSuccess failed with: Could not create temp app file.")
	}
	suite.T().Cleanup(func() {
		os.Remove(tempFile1.Name())
		os.Remove(tempFile2.Name())
	})
	manifests := []model.ManifestPath{
		{
			Path: tempFile1.Name(),
			Targets: []string{
				"env-test",
			},
		},
		{
			Path: tempFile2.Name(),
			Targets: []string{
				"env-test",
			},
		},
		{
			Path: tempFile2.Name(),
			Targets: []string{
				"env-test2",
			},
		},
	}
	files, err := GetManifestsFromFile(&manifests, "env-test")
	if err != nil {
		suite.T().Fatalf("TestGetManifestsFromPathSuccess failed with: %s", err)
	}
	suite.Equal(len(*files), 2)
}

func (suite *ServiceTestSuite) TestCreateDeploymentManifestsSuccess(){
	manifests := make([]string, 2)
	manifests[0] = testAppYamlStr
	manifests[1] = testAppYamlStr
	received := CreateDeploymentManifests(&manifests)
	suite.Equal(len(*received), 2)
}

func (suite *ServiceTestSuite) TestCreateDeploymentCanaryStepSuccess(){
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
		suite.T().Fatalf("TestCreateDeploymentCanaryStepSuccess failed with: %s", err)
	}
	suite.Equal(len(received), len(*strategy.Canary.Steps))
}

func (suite *ServiceTestSuite) TestCreateBeforeDeploymentConstraintsSuccess(){
	untilApproved := true
	duration := int32(600)
	beforeDeployment := []model.BeforeDeployment{
		{
			Pause:  &model.PauseStep{
				Duration: duration,
				Unit: "SECONDS",
			},
		},
		{
			Pause: &model.PauseStep{
				UntilApproved: untilApproved,
			},
		},
	}
	received, err := CreateBeforeDeploymentConstraints(&beforeDeployment)
	if err != nil {
		suite.T().Fatalf("TestCreateBeforeDeploymentConstraintsSuccess failed with: %s", err)
	}
	suite.Equal(len(received), len(beforeDeployment))
}


const testAppYamlStr = `
apiVersion: apps/v1
kind: Deployment
`