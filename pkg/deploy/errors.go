package deploy

import (
	"errors"
	"fmt"
)

const manifestFileReadErrorText = "error trying to read manifest file"
const invalidTrafficManagementConfigErrorText = "invalid traffic management config"
const manifestFileNameReadErrorText = "error trying to read manifest file name"

var ErrMissingMetricsProvider = errors.New("metric provider must be provided either in the analysis config, as defaultMetricProviderName, or in the query as metricProviderName")
var ErrTargetsNotSpecified = errors.New("please omit targets to include the manifests for all targets or specify the targets")
var ErrMissingQueryConfig = errors.New("query in step does not exist in top-level analysis config")

func newManifestFileReadError(fileName string, err error) error {
	return fmt.Errorf("%s '%s': %s", manifestFileReadErrorText, fileName, err)
}

func newInvalidTrafficManagementConfigError(err error) error {
	return fmt.Errorf("%s: %w", invalidTrafficManagementConfigErrorText, err)
}

func newErrorReadingManifestsFromFile(err error) error {
	return fmt.Errorf("%s: %w", manifestFileNameReadErrorText, err)
}

func newMissingQueryConfigError(queryName string) error {
	return fmt.Errorf("%w: %q", ErrMissingQueryConfig, queryName)
}
