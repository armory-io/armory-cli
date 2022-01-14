package deploy

import (
	"encoding/json"
	de "github.com/armory-io/deploy-engine/pkg"
	"github.com/armory/armory-cli/pkg/model"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"testing"
)

const PathToTestManifest1 = "testdata/testManifest1.yaml"
const PathToTestManifest2 = "testdata/testManifest1.yaml"

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

func (suite *ServiceTestSuite) TestCreateDeploymentRequestSuccess() {
	received := createDeploymentForTests(suite, "testdata/happyPathDeploymentFile.yaml")

	expectedJsonStr, err := ioutil.ReadFile("testdata/happyPathDeployEngineRequest.json")
	if err != nil {
		suite.T().Fatalf("TestCreateDeploymentRequestSuccess failed with: Error loading tesdata file %s", err)
	}

	expectedReq := de.PipelineStartPipelineRequest{}
	err = json.Unmarshal(expectedJsonStr, &expectedReq)
	if err != nil {
		suite.T().Fatalf("TestCreateDeploymentRequestSuccess failed with: Error Unmarshalling JSON string to Request obj %s", err)
	}
	suite.EqualValues(expectedReq, *received)
}

func (suite *ServiceTestSuite) TestCreateDeploymentRequestWithoutDependsOnConstraintSuccess() {
	received := createDeploymentForTests(suite, "testdata/happyPathDeploymentFileNoDependsOn.yaml")

	expectedJsonStr, err := ioutil.ReadFile("testdata/happyPathDeployEngineRequestNoDependsOn.json")
	if err != nil {
		suite.T().Fatalf("TestCreateDeploymentRequestSuccess failed with: Error loading tesdata file %s", err)
	}
	expectedReq := de.PipelineStartPipelineRequest{}
	err = json.Unmarshal(expectedJsonStr, &expectedReq)
	if err != nil {
		suite.T().Fatalf("TestCreateDeploymentRequestSuccess failed with: Error Unmarshalling JSON string to Request obj %s", err)
	}
	suite.EqualValues(expectedReq, *received)
}

func (suite *ServiceTestSuite) TestCreateDeploymentRequestFailureDueToNoTargetsProvided() {
	createDeploymentWithErrorForTests(suite, "testdata/sadPathDeployFileNoTargets.yaml", "at least one target must be specified")
}

func (suite *ServiceTestSuite) TestCreateDeploymentRequestFailureDueToNonExistantManifest() {
	createDeploymentWithErrorForTests(suite, "testdata/sadPathDeployFileNonExistentManifest.yaml", "unable to read manifest(s) from file: lstat testdata/fake.yaml: no such file or directory")
}

func (suite *ServiceTestSuite) TestCreateDeploymentRequestFailureDueToInvalidTimeUnitInCanary() {
	createDeploymentWithErrorForTests(suite, "testdata/sadPathDeployFileInvalidTimeUnit.yaml", "invalid value 'FAKE_TIME!' for KubernetesV2CanaryPauseStepTimeUnit: valid values are [NONE SECONDS MINUTES HOURS]")
}

func (suite *ServiceTestSuite) TestCreateDeploymentRequestFailureDueToMissingCanaryStrategy() {
	createDeploymentWithErrorForTests(suite, "testdata/sadPathDeployFileMissingCanaryStrategy.yaml", "error converting steps for canary deployment strategy; canary strategy not provided and is required")
}
func (suite *ServiceTestSuite) TestCreateDeploymentRequestFailureDueToInvalidTimeUnitInBeforeConstraint() {
	createDeploymentWithErrorForTests(suite, "testdata/sadPathDeployFileInvalidTimeInBeforeConstraint.yaml", "invalid value 'FAKE_TIME!' for TimeTimeUnit: valid values are [NONE SECONDS MINUTES HOURS]")
}

func (suite *ServiceTestSuite) TestGetManifestsFromPathSuccess() {
	manifests := []model.ManifestPath{
		{
			Path: PathToTestManifest1,
			Targets: []string{
				"env-test",
			},
		},
		{
			Path: PathToTestManifest2,
			Targets: []string{
				"env-test",
			},
		},
		{
			Path: PathToTestManifest2,
			Targets: []string{
				"env-test2",
			},
		},
		{
			Inline: testAppYamlStr,
			Targets: []string{
				"env-test2",
			},
		},
		{
			Inline: testAppYamlStr,
		},
	}
	files, err := GetManifestsFromFile(&manifests, "env-test")
	if err != nil {
		suite.T().Fatalf("TestGetManifestsFromPathSuccess failed with: %s", err)
	}
	suite.Equal(len(*files), 3)
	files, err = GetManifestsFromFile(&manifests, "env-test2")
	if err != nil {
		suite.T().Fatalf("TestGetManifestsFromPathSuccess failed with: %s", err)
	}
	suite.Equal(len(*files), 3)
	for _, file := range *files {
		suite.Equal(testAppYamlStr, file, "TestGetManifestsFromPathSuccess expected files to match")
	}
}

func (suite *ServiceTestSuite) TestGetManifestsEmptyTargets() {
	manifests := []model.ManifestPath{
		{
			Inline:  testAppYamlStr,
			Targets: []string{},
		},
	}
	_, err := GetManifestsFromFile(&manifests, "env-test")
	if err == nil {
		suite.T().Fatalf("TestGetManifestsFromPathSuccess failed. Expected error")
	}

	suite.Equal("please omit targets to include the manifests for all targets or specify the targets", err.Error())
}

func (suite *ServiceTestSuite) TestCreateDeploymentManifestsSuccess() {
	manifests := make([]string, 2)
	manifests[0] = testAppYamlStr
	manifests[1] = testAppYamlStr
	received := CreateDeploymentManifests(&manifests)
	suite.Equal(len(*received), 2)
}

func (suite *ServiceTestSuite) TestCreateDeploymentCanaryStepSuccess() {
	weight := int32(33)
	untilApproved := true
	duration := int32(600)
	strategy := model.Strategy{
		Canary: &model.CanaryStrategy{
			Steps: &[]model.CanaryStep{
				model.CanaryStep{
					SetWeight: &model.WeightStep{
						Weight: weight,
					},
				},
				model.CanaryStep{
					Pause: &model.PauseStep{
						UntilApproved: untilApproved,
					},
				},
				model.CanaryStep{
					Pause: &model.PauseStep{
						Duration: duration,
						Unit:     "SECONDS",
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

func (suite *ServiceTestSuite) TestCreateBeforeDeploymentConstraintsSuccess() {
	untilApproved := true
	duration := int32(600)
	beforeDeployment := []model.BeforeDeployment{
		{
			Pause: &model.PauseStep{
				Duration: duration,
				Unit:     "SECONDS",
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

func createDeploymentForTests(suite *ServiceTestSuite, pathToInput string) *de.PipelineStartPipelineRequest {
	inputYamlStr, err := ioutil.ReadFile(pathToInput)
	if err != nil {
		suite.T().Fatalf("TestCreateDeploymentRequestSuccess failed with: Error loading tesdata file %s", err)
	}
	orchestration := model.OrchestrationConfig{}
	err = yaml.Unmarshal(inputYamlStr, &orchestration)
	if err != nil {
		suite.T().Fatalf("TestCreateDeploymentRequestSuccess failed with: Error Unmarshalling YAML string to Request obj %s", err)
	}

	received, err := CreateDeploymentRequest(orchestration.Application, &orchestration)
	if err != nil {
		suite.T().Fatalf("TestCreateDeploymentRequestSuccess failed with: %s", err)
	}

	return received

}

func createDeploymentWithErrorForTests(suite *ServiceTestSuite, pathToInput, expectedError string) {
	inputYamlStr, err := ioutil.ReadFile(pathToInput)
	if err != nil {
		suite.T().Fatalf("TestCreateDeploymentRequestSuccess failed with: Error loading tesdata file %s", err)
	}
	orchestration := model.OrchestrationConfig{}
	err = yaml.Unmarshal(inputYamlStr, &orchestration)
	if err != nil {
		suite.T().Fatalf("TestCreateDeploymentRequestSuccess failed with: Error Unmarshalling YAML string to Request obj %s", err)
	}

	_, err = CreateDeploymentRequest(orchestration.Application, &orchestration)
	if suite.Error(err) {
		suite.Equal(expectedError, err.Error())
	}

}

const testAppYamlStr = `
apiVersion: apps/v1
kind: Deployment
`
