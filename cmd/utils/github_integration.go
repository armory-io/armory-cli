package utils

import (
	"fmt"
	log "go.uber.org/zap"
	"os"
	"strings"
)

const GithubOutput = "GITHUB_OUTPUT"
const GithubSummary = "GITHUB_STEP_SUMMARY"

func TryWriteGitHubContext(kv ...string) {

	if len(kv)%2 == 1 {
		log.S().Errorf("expecting collection of {key, value} pairs")
		return
	}

	outLocation := os.Getenv(GithubOutput)
	if outLocation != "" {
		f, err := os.OpenFile(outLocation, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.S().Warnf("could not open file %s - %v", outLocation, err)
		}
		defer f.Close()

		for i := 0; i < len(kv); i += 2 {
			_, _ = fmt.Fprintf(f, "%s=%s\n", strings.ToUpper(kv[i]), kv[i+1])
		}

	}
}

func TryWriteGitHubStepSummary(summary string) {
	outLocation := os.Getenv(GithubSummary)
	if outLocation != "" {
		f, err := os.OpenFile(outLocation, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.S().Warnf("could not open file %s - %v", outLocation, err)
		}
		defer f.Close()

		_, _ = fmt.Fprintln(f, summary)

	}
}
