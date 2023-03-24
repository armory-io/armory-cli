package deploy

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

const (
	pathToTestManifest1 = "testdata/testManifest1.yaml"
	pathToTestManifest2 = "testdata/testManifest2.yaml"
	pathToNestedDir     = "testdata/nested"
	httpPath            = "http://my.hosted.yamls.armory.io/test.yaml"
	httpsPath           = "https://my.hosted.yamls.armory.io/test.yaml"
	s3Path              = "s3://th3morg-public/potato-facts-service.yaml"
)

func TestConvertRequestTestSuite(t *testing.T) {
	suite.Run(t, new(ConvertRequestTestSuite))
}

type ConvertRequestTestSuite struct {
	suite.Suite
	originalWorkspace string
}

func (s *ConvertRequestTestSuite) SetupSuite() {
	if workspace, ok := os.LookupEnv(envVarGithubWorkspace); ok {
		s.originalWorkspace = workspace
	}

	dir, err := os.Getwd()
	s.NoError(err)
	s.NoError(os.Setenv(envVarGithubWorkspace, dir))
}

func (s *ConvertRequestTestSuite) TearDownSuite() {
	s.NoError(os.Setenv(envVarGithubWorkspace, s.originalWorkspace))
}

func (s *ConvertRequestTestSuite) TestGetManifestsFromPaths() {
	manifests := []manifest{
		{
			Path: pathToTestManifest1,
		},
		{
			Path: pathToTestManifest2,
		},
		{
			Path: pathToNestedDir,
		},
		{
			Path: httpPath,
		},
		{
			Path: httpsPath,
		},
		{
			Path: s3Path,
		},
	}
	files, err := getManifestFiles(manifests)
	s.NoError(err)

	s.Len(files[pathToTestManifest1], 1)
	s.Len(files[pathToTestManifest2], 1)
	s.Len(files[pathToNestedDir], 2)
}

func (s *ConvertRequestTestSuite) TestGetManifestsFromGithubPath() {
	manifests := []manifest{
		{
			Path: "/" + pathToTestManifest1,
		},
		{
			Path: pathToTestManifest2,
		},
		{
			Path: pathToNestedDir,
		},
		{
			Path: httpPath,
		},
		{
			Path: httpsPath,
		},
	}
	files, err := getManifestFiles(manifests)
	s.NoError(err)

	s.Len(files["/"+pathToTestManifest1], 1)
	s.Len(files[pathToTestManifest2], 1)
	s.Len(files[pathToNestedDir], 2)
}

func (s *ConvertRequestTestSuite) TestConvertPipelineOptionsToAPIRequest() {
	cases := []struct {
		options   StartPipelineOptions
		assertion func(t *testing.T, request map[string]any)
		expectErr bool
	}{
		{
			options: StartPipelineOptions{
				UnstructuredDeployment: map[string]any{
					"application": "dont-override-me",
					"manifests": []map[string]any{
						{
							"path": pathToTestManifest1,
						},
					},
				},
				ApplicationNameOverride: "",
				Context: map[string]string{
					"choo": "choo",
				},
			},
			assertion: func(t *testing.T, request map[string]any) {
				assert.Equal(t, "dont-override-me", request[applicationKey])
				assert.Len(t, request[filesKey].(map[string][]string)[pathToTestManifest1], 1)
				assert.Equal(t, map[string]string{"choo": "choo"}, request[contextKey])
			},
		},
		{
			options: StartPipelineOptions{
				UnstructuredDeployment: map[string]any{
					"application": "please-override-me",
					"manifests": []map[string]any{
						{
							"path": pathToTestManifest1,
						},
					},
				},
				ApplicationNameOverride: "i-am-an-override!",
				Context: map[string]string{
					"choo": "choo",
				},
			},
			assertion: func(t *testing.T, request map[string]any) {
				assert.Equal(t, "i-am-an-override!", request[applicationKey])
			},
		},
		{
			options: StartPipelineOptions{
				UnstructuredDeployment: map[string]any{
					"application": "please-override-me",
					"context": map[string]any{
						"foo": "bar",
					},
					"manifests": []map[string]any{
						{
							"path": pathToTestManifest1,
						},
					},
				},
				ApplicationNameOverride: "i-am-an-override!",
				Context: map[string]string{
					"choo": "choo",
				},
			},
			assertion: func(t *testing.T, request map[string]any) {
				expected := map[string]any{
					"choo": "choo",
					"foo":  "bar",
				}
				assert.Equal(t, expected, request[contextKey])
			},
		},
		{
			options: StartPipelineOptions{
				UnstructuredDeployment: map[string]any{
					"manifests": []map[string]any{
						{
							"path": pathToTestManifest1,
						},
					},
				},
				Context: map[string]string{
					"choo": "choo",
				},
			},
			expectErr: true,
		},
	}

	for i, c := range cases {
		s.T().Run(fmt.Sprintf("case-%d", i), func(t *testing.T) {
			request, err := convertPipelineOptionsToAPIRequest(c.options)
			if c.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				c.assertion(t, request)
			}
		})
	}
}
