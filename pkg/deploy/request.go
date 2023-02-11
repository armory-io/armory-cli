package deploy

import (
	errorUtils "github.com/armory/armory-cli/pkg/errors"
	"github.com/mitchellh/mapstructure"
	"io/fs"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
)

type (
	StartPipelineOptions struct {
		// UnstructuredDeployment is the user's raw deployment (likely unmarshalled from their YAML file).
		UnstructuredDeployment  map[string]any
		ApplicationNameOverride string
		ContextOverrides        map[string]string
		Headers                 map[string]string
		IsURL                   bool
	}

	structuredConfig struct {
		Application string     `yaml:"application"`
		Manifests   []manifest `yaml:"manifests"`
	}

	manifest struct {
		Path    string   `yaml:"path"`
		Targets []string `yaml:"targets"`
		Inline  string   `yaml:"inline"`
	}
)

const (
	applicationKey      = "application"
	filesKey            = "files"
	contextOverridesKey = "contextOverrides"

	envVarGithubWorkspace = "GITHUB_WORKSPACE"
)

func convertPipelineOptionsToAPIRequest(options StartPipelineOptions) (map[string]any, error) {
	if options.IsURL {
		return options.UnstructuredDeployment, nil
	}
	deployment := options.UnstructuredDeployment

	var structured structuredConfig
	if err := mapstructure.Decode(deployment, &structured); err != nil {
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
	deployment[contextOverridesKey] = options.ContextOverrides
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
		file, err := ioutil.ReadFile(fileName)
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
