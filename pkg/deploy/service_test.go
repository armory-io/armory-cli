package deploy

import (
	"encoding/json"
	de "github.com/armory-io/deploy-engine/pkg"
	"github.com/armory/armory-cli/pkg/model"
	"github.com/r3labs/diff"
	"github.com/stretchr/testify/assert"
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
	received, err := createDeploymentForTests(suite, "testdata/happyPathDeploymentFile.yaml")
	if err != nil {
		suite.T().Fatal(err)
	}
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
	received, err := createDeploymentForTests(suite, "testdata/happyPathDeploymentFileNoDependsOn.yaml")
	if err != nil {
		suite.T().Fatal(err)
	}
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

func (suite *ServiceTestSuite) TestCreateDeploymentRequestWithBlueGreenSuccess() {
	received, err := createDeploymentForTests(suite, "testdata/happyPathDeploymentFileBlueGreen.yaml")
	if err != nil {
		suite.T().Fatal(err)
	}
	expectedJsonStr, err := ioutil.ReadFile("testdata/happyPathDeployEngineRequestBlueGreen.json")
	if err != nil {
		suite.T().Fatalf("TestCreateDeploymentRequestWithBlueGreenSuccess failed with: Error loading tesdata file %s", err)
	}
	expectedReq := de.PipelineStartPipelineRequest{}
	err = json.Unmarshal(expectedJsonStr, &expectedReq)
	if err != nil {
		suite.T().Fatalf("TestCreateDeploymentRequestWithBlueGreenSuccess failed with: Error Unmarshalling JSON string to Request obj %s", err)
	}
	suite.EqualValues(expectedReq, *received)
}

func (suite *ServiceTestSuite) TestCreateDeploymentRequestWithBadStrategyPath() {
	cases := []struct{
		file string
		expectErr string
	} {
		{
			"testdata/sadPathDeploymentFileBlueGreen1.yaml",
			"invalid blueGreen config: activeService is required",
		},
		{
			"testdata/sadPathDeploymentFileBadPause1.yaml",
			"pause is not valid: untilApproved cannot be set with both a unit and duration",
		},
		{
			"testdata/sadPathDeploymentFileBadPause2.yaml",
			"pause is not valid: unit must be set with a duration",
		},
		{
			"testdata/sadPathDeploymentFileBadPause3.yaml",
			"pause is not valid: duration must be set with a unit",
		},
	}

	for _, c := range cases {
		received, err := createDeploymentForTests(suite, c.file)
		suite.Nilf(received, "Expected deployment to not be created for an invalid pause step")
		suite.EqualErrorf(err, c.expectErr, "Error messages do not match. Want: '%s', got: '%s'", c.expectErr, err)
	}
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
	received, err := createDeploymentCanarySteps(strategy, &model.AnalysisConfig{})
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

func createDeploymentForTests(suite *ServiceTestSuite, pathToInput string) (*de.PipelineStartPipelineRequest, error) {
	inputYamlStr, err := ioutil.ReadFile(pathToInput)
	if err != nil {
		suite.T().Logf("TestCreateDeploymentRequestSuccess failed with: Error loading tesdata file %s", err)
		return nil, err
	}
	orchestration := model.OrchestrationConfig{}
	err = yaml.UnmarshalStrict(inputYamlStr, &orchestration)
	if err != nil {
		suite.T().Logf("TestCreateDeploymentRequestSuccess failed with: Error Unmarshalling YAML string to Request obj %s", err)
		return nil, err
	}

	received, err := CreateDeploymentRequest(orchestration.Application, &orchestration)
	if err != nil {
		suite.T().Logf("TestCreateDeploymentRequestSuccess failed with: %s", err)
		return nil, err
	}

	return received, nil
}

func (suite *ServiceTestSuite) TestCreateDeploymentAnalysisErrors() {
	cases := []struct{
		file string
		expectErr string
	} {
		{
			"testdata/sadPathAnalysisDeploymentFile.yaml",
			"analysis configuration block is present but default or explicit account is not set",
		},
		{
			"testdata/sadPathMissingTopLevelAnalysis.yaml",
			"analysis step is present but a top-level analysis config is not defined",
		},
		{
			"testdata/sadPathMissingTopLevelAnalysisQueries.yaml",
			"analysis step is present but a top-level analysis config is not defined",
		},
		{
			"testdata/sadPathAnalysisStepQueriesInvalid.yaml",
			"query in step does not exist in top-level analysis config: Who lives in a pineapple under the sea",
		},
	}

	for _, c := range cases {
		inputYamlStr, err := ioutil.ReadFile(c.file)
		if err != nil {
			suite.T().Fatalf("TestCreateDeploymentAnalysisErrors failed with: Error loading tesdata file %s", err)
		}
		orchestration := model.OrchestrationConfig{}
		err = yaml.UnmarshalStrict(inputYamlStr, &orchestration)
		if err != nil {
			suite.T().Fatalf("TestCreateDeploymentAnalysisErrors failed with: Error Unmarshalling YAML string to Request obj %s", err)
		}
		_, err = CreateDeploymentRequest(orchestration.Application, &orchestration)
		suite.Errorf(err, c.expectErr)
	}
}

func TestBuildStrategy(t *testing.T) {
	_, err := buildStrategy(model.OrchestrationConfig{
		Strategies: &map[string]model.Strategy{},
	}, "fakeStrategy")
	assert.Errorf(t, err, "fakeStrategy is not a valid strategy; define canary or blueGreen strategy")
}

const testAppYamlStr = `
apiVersion: apps/v1
kind: Deployment
`
