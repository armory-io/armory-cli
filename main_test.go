package main

import (
	"bytes"
	"fmt"
	"github.com/armory/armory-cli/internal/clierr"
	"github.com/armory/armory-cli/internal/clierr/exitcodes"
	"github.com/armory/armory-cli/pkg/console"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net/http"
	"strings"
	"testing"
)

func TestMainTestSuite(t *testing.T) {
	suite.Run(t, new(MainTestSuite))
}

type MainTestSuite struct {
	suite.Suite
}

func (s *MainTestSuite) TestProcessCmdErrorsErrorPaths() {
	cases := []struct {
		name             string
		err              error
		expectedStdErr   string
		expectedStdOut   string
		expectedExitCode exitcodes.ExitCode
	}{
		{
			name:             "deals with nil errors as expected",
			expectedExitCode: exitcodes.Success,
		},
		{
			name:             "deals with generic errors as expected",
			expectedExitCode: exitcodes.Error,
			expectedStdErr:   "an error occurred\n",
			err:              fmt.Errorf("an error occurred"),
		},
		{
			name:             "deals with API errors as expected",
			expectedExitCode: exitcodes.Error,
			expectedStdErr:   "An error occurred when trying to do X\nError Id: some-id\nHTTP Status Code: 403\nError Message: access denied is not org admin\n",
			err:              clierr.NewAPIError("An error occurred when trying to do X", http.StatusForbidden, []byte(`{"error_id": "some-id", "errors": [{"message": "access denied is not org admin", "code": 42}]}`), exitcodes.Error),
		},
		{
			name:             "deals with an API error that has supplied a custom exit code as expected",
			expectedExitCode: exitcodes.Conflict,
			expectedStdErr:   "An error occurred when trying to do Y\nError Id: some-id\nHTTP Status Code: 409\nError Message: pipeline already in progress\n",
			err:              clierr.NewAPIError("An error occurred when trying to do Y", http.StatusConflict, []byte(`{"error_id": "some-id", "errors": [{"message": "pipeline already in progress", "code": 42}]}`), exitcodes.Conflict),
		},
		{
			name:             "deals with an API error that has metadata as expected",
			expectedExitCode: exitcodes.Error,
			expectedStdErr:   "Failed to submit tenant config\nError Id: some-id\nHTTP Status Code: 400\nError Message: invalid tenant config\nError Metadata: \n  foo=map[msg:embedded object]\n  bar=bam\n",
			err:              clierr.NewAPIError("Failed to submit tenant config", http.StatusBadRequest, []byte(`{"error_id": "some-id", "errors": [{"message": "invalid tenant config", "metadata":{"foo":{"msg":"embedded object"},"bar":"bam"}, "code": 42}]}`), exitcodes.Error),
		},
	}

	for _, c := range cases {
		stdOut := bytes.NewBufferString("")
		stdErr := bytes.NewBufferString("")
		console.Configure(&console.Options{
			Stdout: stdOut,
			Stderr: stdErr,
		})
		s.Run(c.name, func() {
			exitCode := processCmdResults(c.err)
			assert.Equal(s.T(), int(c.expectedExitCode), exitCode)
			assert.Equal(s.T(), c.expectedStdOut, stdOut.String())
			assertEqualIgnoringLineOrder(s.T(), c.expectedStdErr, stdErr.String())
		})
	}
}

func assertEqualIgnoringLineOrder(t *testing.T, expected string, actual string) {
	expectedLines := strings.Split(expected, "\n")
	actualLines := strings.Split(actual, "\n")
	assert.ElementsMatch(t, expectedLines, actualLines)
}
