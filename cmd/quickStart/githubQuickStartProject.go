package quickStart

import (
	"archive/zip"
	"errors"
	"fmt"
	"github.com/ahmetb/go-linq/v3"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/armory/armory-cli/pkg/org"
	"github.com/armory/armory-cli/pkg/util"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type GithubQuickStartProject struct {
	ProjectName   string
	BranchName    string
	IsZipFile     bool
	DirName       string
	DeployYmlName string
}

func (p GithubQuickStartProject) GetSuffix() string {
	if p.IsZipFile {
		return githubZipSuffix
	} else {
		return ""
	}
}

func (p GithubQuickStartProject) GetUrl() string {
	return fmt.Sprintf("%s%s%s", githubBaseUrl, p.ProjectName, p.GetSuffix())
}

func (p GithubQuickStartProject) GetProjectFolder() string {
	return fmt.Sprintf("%s%s-%s", string(os.PathSeparator), p.ProjectName, p.BranchName)
}

func (p GithubQuickStartProject) GetFileDownloadPath() string {
	suffix := ""
	if p.IsZipFile {
		suffix = ".zip"
	}
	return fmt.Sprintf("%s%s", p.DirName, suffix)
}

func (p GithubQuickStartProject) GetDeployCommand() string {
	return fmt.Sprintf("armory deploy start -f %s", p.DeployYmlName)
}

func (p GithubQuickStartProject) Unzip() error {

	if !p.IsZipFile {
		return nil
	}
	log.Info("Unzipping project...")
	archive, err := zip.OpenReader(p.GetFileDownloadPath())

	if err != nil {
		return err
	}
	defer archive.Close()
	for _, f := range archive.File {
		if strings.Contains(f.Name, string(os.PathSeparator)+".") {
			log.Debugln(fmt.Sprintf("skipping hidden file %s", f.Name))
			continue
		}
		path := filepath.Join(p.DirName, f.Name)
		dstPath := strings.Replace(path, p.GetProjectFolder(), "", 1)

		if !strings.HasPrefix(path, filepath.Clean(p.DirName)+string(os.PathSeparator)) {
			return errors.New("found an invalid file path")
		}
		if f.FileInfo().IsDir() {
			log.Debugln(dstPath)
			os.MkdirAll(dstPath, os.ModePerm)
			continue
		}

		log.Debugln(dstPath)

		if err := os.MkdirAll(filepath.Dir(dstPath), os.ModePerm); err != nil {
			return err
		}

		dstFile, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		fileInArchive, err := f.Open()
		if err != nil {
			return err
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			return err
		}

		dstFile.Close()
		fileInArchive.Close()
	}

	return nil
}

/**
Returns isCancelled, error
*/
func (p GithubQuickStartProject) Download() (bool, error) {
	log.Info(fmt.Sprintf("Downloading demo project from `%s`...", p.GetUrl()))
	if info, _ := os.Stat(p.DirName); info != nil {
		prompt := promptui.Prompt{
			Label:     fmt.Sprintf("`%s` directory will be overwritten. Proceed", p.DirName),
			IsConfirm: true,
			Stdout:    &util.BellSkipper{},
		}

		_, err := prompt.Run()
		if err != nil {
			return true, nil
		}
	}

	resp, err := http.Get(p.GetUrl())
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(p.GetFileDownloadPath())
	if err != nil {
		return false, err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return false, err
	}

	return false, nil
}

func (p GithubQuickStartProject) UpdateAgentAccount(config *config.Configuration, selectedAgent string) error {
	deployFileName := fmt.Sprintf("%s%s%s", p.DirName, string(os.PathSeparator), p.DeployYmlName)
	log.Info(fmt.Sprintf("Replacing defaults in %s with agent %s", deployFileName, selectedAgent))
	yaml, err := ioutil.ReadFile(deployFileName)
	if err != nil {
		log.Fatalln(err)
	}

	lines := strings.Split(string(yaml), "\n")

	for i, line := range lines {
		lines[i] = strings.ReplaceAll(line, "my-first-cluster", selectedAgent)
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(deployFileName, []byte(output), 0644)
	if err != nil {
		log.Fatalln(err)
	}

	return nil
}

func (p GithubQuickStartProject) SelectAgent(config *config.Configuration, namedAgent string) (string, error) {
	log.Info("Fetching armory agents that are connected to your k8s cluster...")
	agents, err := org.GetAgents(config.GetArmoryCloudAddr(), config.GetAuthToken())
	if err != nil {
		return "", err
	}
	var agentIdentifiers []string
	linq.From(agents).Select(func(c interface{}) interface{} {
		log.Debugln(fmt.Sprintf("Found agent %s", c.(org.Agent).AgentIdentifier))
		return c.(org.Agent).AgentIdentifier
	}).ToSlice(&agentIdentifiers)

	if len(agentIdentifiers) < 1 {
		return "", errors.New(fmt.Sprintf("No agents were found. Please ensure you have a connected agent: %s%s", config.GetArmoryCloudEnvironmentConfiguration().CloudConsoleBaseUrl, "/configuration/agents"))
	}

	if len(namedAgent) > 0 && namedAgent != "" {
		requestedAgent := ""
		for _, agentName := range agentIdentifiers {
			if agentName == namedAgent {
				requestedAgent = agentName
			}
		}
		if requestedAgent == "" {
			return "", errors.New(fmt.Sprintf("Specified agent %s not found, please choose a known environment: [%s]", namedAgent, strings.Join(agentIdentifiers[:], ",")))
		}
	}

	selectedAgent := ""
	if len(agentIdentifiers) == 1 {
		selectedAgent = agentIdentifiers[0]
	} else {
		prompt := promptui.Select{
			Label:  "Select one of your connected agents",
			Items:  agentIdentifiers,
			Stdout: &util.BellSkipper{},
		}

		_, selectedAgent, err = prompt.Run()

		if err != nil || selectedAgent == "" {
			return "", errors.New(fmt.Sprintf("Failed to select an agent to deploy to; %v\n", err))
		}
	}

	log.Debugln(fmt.Sprintf("Selected agent %s", selectedAgent))
	return selectedAgent, nil
}
