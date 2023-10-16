package deploy

import (
	scm "github.com/armory/armory-cli/cmd/sourceControl"
	errorUtils "github.com/armory/armory-cli/pkg/errors"
	"github.com/mitchellh/mapstructure"
	"github.com/samber/lo"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
)

type (
	StartPipelineOptions struct {
		// UnstructuredDeployment is the user's raw deployment (likely unmarshalled from their YAML file).
		UnstructuredDeployment  map[string]any
		ApplicationNameOverride string
		Context                 map[string]string
		SCMC                    scm.Context
		Headers                 map[string]string
		IsURL                   bool
	}

	structuredConfig struct {
		Kind             string            `yaml:"kind"`
		Application      string            `yaml:"application"`
		Manifests        []manifest        `yaml:"manifests"`
		DeploymentConfig *deploymentConfig `yaml:"deploymentConfig"`
	}

	manifest struct {
		Path    string   `yaml:"path"`
		Targets []string `yaml:"targets"`
		Inline  string   `yaml:"inline"`
	}

	deploymentConfig struct {
		IfDeploymentInProgress *ifDeploymentInProgress `yaml:"ifDeploymentInProgress"`
	}

	ifDeploymentInProgress struct {
		Strategy strategy `yaml:"strategy"`
	}

	strategy string
)

const (
	applicationKey        = "application"
	filesKey              = "files"
	contextKey            = "context"
	scmcKey               = "sourceControl"
	envVarGithubWorkspace = "GITHUB_WORKSPACE"

	enqueueOne strategy = "enqueueOne"
	reject     strategy = "reject"
)

func (s *StartPipelineOptions) structuredConfig() (*structuredConfig, error) {
	var structured structuredConfig
	return &structured, mapstructure.Decode(s.UnstructuredDeployment, &structured)
}

func convertPipelineOptionsToAPIRequest(options StartPipelineOptions) (map[string]any, error) {
	if options.IsURL {
		return options.UnstructuredDeployment, nil
	}
	deployment := options.UnstructuredDeployment

	structured, err := options.structuredConfig()
	if err != nil {
		return nil, err
	}

	application := options.ApplicationNameOverride
	if len(application) == 0 {
		if len(structured.Application) > 0 {
			application = structured.Application
		} else {
			return nil, ErrNoApplicationNameDefined
		}
	}

	manifestFiles, err := getManifestFiles(structured.Manifests)
	if err != nil {
		return nil, err
	}

	deployment[applicationKey] = application
	deployment[filesKey] = manifestFiles

	context := map[string]any{}
	if c, ok := deployment[contextKey].(map[string]any); ok {
		context = lo.Assign(context, c)
	}
	for key, value := range options.Context {
		context[key] = value
	}
	deployment[contextKey] = context

	deployment[scmcKey] = options.SCMC

	return deployment, nil
}

func getManifestFiles(manifests []manifest) (map[string][]string, error) {
	allManifests := make(map[string][]string)
	for _, m := range manifests {
		if IsURL(m.Path) {
			continue
		}
		fileNames, err := getFileNamesFromPath(m.Path)
		if err != nil {
			return nil, err
		}
		files, err := getFiles(fileNames)
		if err != nil {
			return nil, err
		}
		allManifests[m.Path] = files
	}
	return allManifests, nil
}

func getFileNamesFromPath(path string) ([]string, error) {
	var allFileNames []string
	gitWorkspace, present := os.LookupEnv(envVarGithubWorkspace)

	if path != "" {
		if present {
			path = gitWorkspace + "/" + path
		}
		fileNames, err := getFileNames(path)
		if err != nil {
			return nil, errorUtils.NewWrappedError(ErrManifestFileNameRead, err)
		}
		allFileNames = append(allFileNames, fileNames...)
	}
	return allFileNames, nil
}

func getFiles(dirFileNames []string) ([]string, error) {
	var files []string
	for _, fileName := range dirFileNames {
		file, err := os.ReadFile(fileName)
		if err != nil {
			return nil, errorUtils.NewWrappedErrorWithDynamicContext(ErrManifestFileRead, err, fileName)
		}
		files = append(files, string(file))
	}
	return files, nil
}

func IsURL(fileName string) bool {
	u, err := url.Parse(fileName)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func getFileNames(path string) ([]string, error) {
	var fileNames []string
	return fileNames, filepath.WalkDir(path, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml" {
			fileNames = append(fileNames, path)
		}

		return nil
	})
}
