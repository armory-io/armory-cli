package deploy

import (
	"encoding/json"
	de "github.com/armory-io/deploy-engine/pkg"
	"github.com/armory/armory-cli/pkg/model"
	"github.com/r3labs/diff"
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
	diffOfExpectedAndRecieved, err := diff.Diff(expectedReq, *received)
	suite.NoError(err)
	suite.Len(diffOfExpectedAndRecieved, 0)
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
	diffOfExpectedAndRecieved, err := diff.Diff(expectedReq, *received)
	suite.NoError(err)
	suite.Len(diffOfExpectedAndRecieved, 0)
}

func (suite *ServiceTestSuite) TestCreateDeploymentRequestInvalidYaml() {
	inputYamlStr, err := ioutil.ReadFile("testdata/sadPathDeploymentFile.yaml")
	if err != nil {
		suite.T().Fatalf("TestCreateDeploymentRequestSuccess failed with: Error loading tesdata file %s", err)
	}
	orchestration := model.OrchestrationConfig{}
	err = yaml.UnmarshalStrict(inputYamlStr, &orchestration)
	suite.Error(err)
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
				{
					SetWeight: &model.WeightStep{
						Weight: weight,
					},
				},
				{
					Pause: &model.PauseStep{
						UntilApproved: untilApproved,
					},
				},
				{
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
	err = yaml.UnmarshalStrict(inputYamlStr, &orchestration)
	if err != nil {
		suite.T().Fatalf("TestCreateDeploymentRequestSuccess failed with: Error Unmarshalling YAML string to Request obj %s", err)
	}

	received, err := CreateDeploymentRequest(orchestration.Application, &orchestration)
	if err != nil {
		suite.T().Fatalf("TestCreateDeploymentRequestSuccess failed with: %s", err)
	}

	return received
}

const testAppYamlStr = `
apiVersion: apps/v1
kind: Deployment
`
