package quickStart

import (
	"archive/zip"
	"errors"
	"fmt"
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

type QuickStartProject interface {
}

type GithubQuickStartProject struct {
	ProjectName   string
	BranchName    string
	IsZipFile     bool
	DirName       string
	DeployYmlName string
}

func (p GithubQuickStartProject) GetUrlSuffix() string {
	if p.IsZipFile {
		return githubZipSuffix
	} else {
		return ""
	}
}

func (p GithubQuickStartProject) GetUrl() string {
	return fmt.Sprintf("%s%s%s", githubBaseUrl, p.ProjectName, p.GetUrlSuffix())
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

func (p GithubQuickStartProject) OverwritePrompt() error {
	if info, _ := os.Stat(p.DirName); info != nil {
		prompt := promptui.Prompt{
			Label:     fmt.Sprintf("`%s` directory will be overwritten. Proceed", p.DirName),
			IsConfirm: true,
			Default:   "Y",
			Stdout:    &util.BellSkipper{},
		}

		_, err := prompt.Run()
		if err != nil {
			return errors.New("Cancelled... ")
		}
	}

	return nil
}

func (p GithubQuickStartProject) Download() error {
	log.Info(fmt.Sprintf("Downloading sample application from `%s`...", p.GetUrl()))
	defaultErr := errors.New(fmt.Sprintf("Unable to download project from Github. Please download and unzip %s, then execute `%s`", p.GetUrl(), p.GetDeployCommand()))
	resp, err := http.Get(p.GetUrl())
	if err != nil {
		log.Debugln(err)
		return defaultErr
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(p.GetFileDownloadPath())
	if err != nil {
		log.Debugln(err)
		return defaultErr
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Debugln(err)
		return defaultErr
	}

	return nil
}

func (p GithubQuickStartProject) UpdateAgentAccount(selectedAgent string) error {
	deployFileName := fmt.Sprintf("%s%s%s", p.DirName, string(os.PathSeparator), p.DeployYmlName)
	log.Info(fmt.Sprintf("Replacing defaults in %s with Remote Network Agent '%s'", deployFileName, selectedAgent))
	yaml, err := ioutil.ReadFile(deployFileName)
	if err != nil {
		return err
	}

	lines := strings.Split(string(yaml), "\n")

	for i, line := range lines {
		lines[i] = strings.ReplaceAll(line, "my-first-cluster", selectedAgent)
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(deployFileName, []byte(output), 0644)
	if err != nil {
		return err
	}

	return nil
}
