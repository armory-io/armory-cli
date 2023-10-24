package sourceControl

import (
	"context"
	"fmt"
	de "github.com/armory-io/deploy-engine/pkg/api"
	"github.com/fatih/color"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

type (
	Context interface {
		GetContext() (Context, error)
	}

	ServiceProvider interface {
		GetGithubService() GithubService
	}

	DefaultServiceProvider struct {
		Ctx context.Context
	}
)

func RetrieveContext(out io.Writer, provider ServiceProvider) (de.SCM, error) {
	var scmc de.SCM
	var err error

	scmc.GithubSCM, scmc.GithubData, err = GetGithubContext(provider.GetGithubService())

	ErrorLogger(out, err)

	return scmc, err
}

func GetContextFromFile(out io.Writer, path string) (de.GenericSCM, error) {
	var err error
	genericSCM := de.GenericSCM{}

	file, err := RetrieveFile(path)

	if ErrorLogger(out, err) {
		return genericSCM, err
	}

	genericSCM, err = Unmarshall(file)
	ErrorLogger(out, err)

	return genericSCM, err
}

func RetrieveFile(path string) ([]byte, error) {
	var file []byte
	var fullPath = path

	//in case this is running on a github instance
	gitWorkspace, present := os.LookupEnv("GITHUB_WORKSPACE")

	if present {
		fullPath = gitWorkspace + path
	}

	// read yaml file
	file, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func Unmarshall(file []byte) (de.GenericSCM, error) {
	var payload de.GenericSCM
	if err := yaml.Unmarshal(file, &payload); err != nil {
		return de.GenericSCM{}, err
	}
	return payload, nil
}

func (d DefaultServiceProvider) GetGithubService() GithubService {
	return DefaultGithubService{client: &DefaultGithubClient{ctx: context.Background()}}
}

func ErrorLogger(out io.Writer, err error) bool {
	if err != nil {
		msg := color.New(color.FgYellow, color.Bold).Sprintln("\nscm warning: ")
		fmt.Fprintf(out, "%s %s\n\n", msg, err)
		return true
	}
	return false
}
