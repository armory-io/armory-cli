package integration

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"os/exec"
	"runtime"
	"testing"
)

const binaryName = "armory"
const version = "integration_test"

var binaryPath string

func TestMain(m *testing.M) {
	err := os.Chdir("..")
	if err != nil {
		fmt.Printf("could not change dir: %v", err)
		os.Exit(1)
	}
	dir, err := os.Getwd()
	if err != nil {
		fmt.Printf("could not get pwd path for %s: %v", binaryName, err)
		os.Exit(1)
	}
	goarch := runtime.GOARCH
	goos := runtime.GOOS
	cmd := exec.Command("make", "build")
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "PWD="+dir, "GOARCH="+goarch, "GOOS="+goos, "VERSION="+version)
	binaryPath = fmt.Sprintf("%s/build/dist/%s_%s/%s", dir, goos, goarch, binaryName)
	if err := cmd.Run(); err != nil {
		fmt.Printf("could not make binary for %s: %v", binaryName, err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func TestVersionCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	tests := []struct {
		name string
		args []string
	}{
		{"get version", []string{"version"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("test run returned an error: %v\n%s", err, output)
			}
			actual := string(output)
			assert.Contains(t, actual, version, "must contains the test version")
		})
	}
}
