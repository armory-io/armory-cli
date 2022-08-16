package login

import "fmt"

const errorGettingDeviceCodeText = "error at getting device code: %w"
const errorPollingServerResponseText = "error at polling auth server for response. Err: %w"
const errorDecodingJwtText = "error at decoding jwt. Err: %w"
const errorGettingHomeDirectoryText = "there was an error getting the home directory. Err: %w"
const errorWritingCredentialsFileText = "there was an error writing the credentials file. Err: %w"

func newErrorGettingDeviceCode(err error) error {
	return fmt.Errorf(errorGettingDeviceCodeText, err)
}

func newErrorPollingServerResponse(err error) error {
	return fmt.Errorf(errorPollingServerResponseText, err)
}

func newErrorDecodingJwt(err error) error {
	return fmt.Errorf(errorDecodingJwtText, err)
}

func newErrorGettingHomeDirectory(err error) error {
	return fmt.Errorf(errorGettingHomeDirectoryText, err)
}

func newErrorWritingCredentialsFile(err error) error {
	return fmt.Errorf(errorWritingCredentialsFileText, err)
}
