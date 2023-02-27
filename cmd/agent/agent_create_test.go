package agent

import (
	"github.com/stretchr/testify/assert"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"os"
	"testing"
)

func TestSetKubeContext(t *testing.T) {
	cases := []struct {
		name                string
		agentOptions        AgentOptions
		expectedContextName string
	}{
		{
			name: "will use default kube context",
			agentOptions: AgentOptions{
				UseCurrentContext: true,
				contextNames:      []string{"mock-context", "cheese-burger"},
				configAccess:      &MockConfigAccess{},
			},
			expectedContextName: "mock-context",
		},
		{
			name: "will use specified kube context name",
			agentOptions: AgentOptions{
				UseCurrentContext: false,
				contextNames:      []string{"my-test-context", "cheese-burger"},
				ContextName:       "cheese-burger",
				configAccess:      &MockConfigAccess{},
			},
			expectedContextName: "cheese-burger",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			requestedContext, err := c.agentOptions.setKubeContext()
			assert.NoError(t, err)
			assert.Equal(t, c.expectedContextName, requestedContext)
		})
	}
}

type MockConfigAccess struct {
}

// GetLoadingPrecedence returns the slice of files that should be used for loading and inspecting the config
func (m *MockConfigAccess) GetLoadingPrecedence() []string {
	return nil
}

// GetStartingConfig returns the config that subcommands should being operating against.  It may or may not be merged depending on loading rules
func (m *MockConfigAccess) GetStartingConfig() (*clientcmdapi.Config, error) {
	return &clientcmdapi.Config{
		CurrentContext: "mock-context",
	}, nil
}

// GetDefaultFilename returns the name of the file you should write into (create if necessary), if you're trying to create a new stanza as opposed to updating an existing one.
func (m *MockConfigAccess) GetDefaultFilename() string {
	file, err := os.CreateTemp("", "test-kubeconfig")
	if err != nil {
		panic(err)
	}
	return file.Name()
}

// IsExplicitFile indicates whether or not this command is interested in exactly one file.  This implementation only ever does that  via a flag, but implementations that handle local, global, and flags may have more
func (m *MockConfigAccess) IsExplicitFile() bool {
	return false
}

// GetExplicitFile returns the particular file this command is operating against.  This implementation only ever has one, but implementations that handle local, global, and flags may have more
func (m *MockConfigAccess) GetExplicitFile() string {
	return ""
}
