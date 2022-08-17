package deploy

import (
	"errors"
	"fmt"
)

const (
	errManifestFileReadText               = "error trying to read manifest file"
	errInvalidTrafficManagementConfigText = "invalid traffic management config"
	errManifestFileNameReadText           = "error trying to read manifest file name"
)

var (
	ErrMissingMetricsProvider = errors.New("metric provider must be provided either in the analysis config, as defaultMetricProviderName, or in the query as metricProviderName")
	ErrTargetsNotSpecified    = errors.New("please omit targets to include the manifests for all targets or specify the targets")
	ErrMissingQueryConfig     = errors.New("query in step does not exist in top-level analysis config")
)

func newManifestFileReadError(fileName string, err error) error {
	return fmt.Errorf("%s '%s': %s", errManifestFileReadText, fileName, err)
}

func newInvalidTrafficManagementConfigError(err error) error {
	return fmt.Errorf("%s: %w", errInvalidTrafficManagementConfigText, err)
}

func newErrorReadingManifestsFromFile(err error) error {
	return fmt.Errorf("%s: %w", errManifestFileNameReadText, err)
}

func newMissingQueryConfigError(queryName string) error {
	return fmt.Errorf("%w: %q", ErrMissingQueryConfig, queryName)
}
