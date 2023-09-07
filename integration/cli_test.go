package integration

import (
	"github.com/stretchr/testify/assert"
	"os/exec"
	"testing"
)

const binaryName = "armory"
const version = "integration_test"

var binaryPath string

/*

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
	binaryPath = fmt.Sprintf("%s/build/bin/%s_%s/%s", dir, goos, goarch, binaryName)
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("could not make binary for %s: error=%s, %s", binaryName, err, string(output))
		os.Exit(1)
	}
	os.Exit(m.Run())
}
*/
func TestVersionCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	t.Skip("skipping - this test makes little or no sense")
	tests := []struct {
		name string
		args []string
	}{
		{"get version", []string{"version"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			output, err := cmd.Output()
			assert.NoError(t, err)
			actual := string(output)
			assert.Contains(t, actual, version, "must contain the test version")
		})
	}
}
