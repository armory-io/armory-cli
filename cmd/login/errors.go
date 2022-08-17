package login

import "fmt"

const (
	errGettingDeviceCodeText      = "error at getting device code: %w"
	errPollingServerResponseText  = "error at polling auth server for response. Err: %w"
	errDecodingJwtText            = "error at decoding jwt. Err: %w"
	errGettingHomeDirectoryText   = "there was an error getting the home directory. Err: %w"
	errWritingCredentialsFileText = "there was an error writing the credentials file. Err: %w"
)

func newErrorGettingDeviceCode(err error) error {
	return fmt.Errorf(errGettingDeviceCodeText, err)
}

func newErrorPollingServerResponse(err error) error {
	return fmt.Errorf(errPollingServerResponseText, err)
}

func newErrorDecodingJwt(err error) error {
	return fmt.Errorf(errDecodingJwtText, err)
}

func newErrorGettingHomeDirectory(err error) error {
	return fmt.Errorf(errGettingHomeDirectoryText, err)
}

func newErrorWritingCredentialsFile(err error) error {
	return fmt.Errorf(errWritingCredentialsFileText, err)
}
