package clierr

import (
	"encoding/json"
	"fmt"
	"github.com/armory/armory-cli/internal/clierr/exitcodes"
	"github.com/fatih/color"
)

type (
	// ApiErrorResponse the error wrapper that is returned from API errors with Armory APIs.
	ApiErrorResponse struct {
		// ErrorID this is a guid that is generated for each API error and can be used to look up the error in the log aggregator.
		ErrorID string `json:"error_id"`
		// Errors more often than not this is a single error, the notable exception is when request objects fail validation, in that case, there is an error for each validation violation.
		Errors []ApiErrorDTO `json:"errors"`
	}

	// ApiErrorDTO the struct that houses the actual error message.
	ApiErrorDTO struct {
		// Message the human-readable error message.
		Message string `json:"message"`
		// Metadata optionally used to add context around an error, such as JSON deserialization errors.
		Metadata map[string]any `json:"metadata,omitempty"`
		Code     any            `json:"code"`
	}

	// APIError an error that can be returned from a cobra command that wraps an API error from the Armory API.
	APIError struct {
		message          string
		httpStatusCode   int
		apiErrorResponse *ApiErrorResponse
		exitCode         exitcodes.ExitCode
	}
)

// Error the string representing the error.
func (a *APIError) Error() string {
	return fmt.Sprintf("an API error occurred, msg: %v, status code: %v", a.message, a.httpStatusCode)
}

// DetailedError returns a human readable multiline colored error message.
func (a *APIError) DetailedError() string {
	msg := color.New(color.FgRed, color.Bold).Sprintln(a.message)
	msg += color.New(color.FgRed, color.Bold).Sprintf("Error Id: ")
	msg += fmt.Sprintf("%v\n", a.apiErrorResponse.ErrorID)
	msg += color.New(color.FgRed, color.Bold).Sprintf("HTTP Status Code: ")
	msg += fmt.Sprintf("%v\n", a.httpStatusCode)
	e := a.apiErrorResponse.Errors[0]
	msg += color.New(color.FgRed, color.Bold).Sprintf("Error Message: ")
	msg += fmt.Sprintf("%v", e.Message)
	if len(e.Metadata) > 0 {
		msg += color.New(color.FgRed, color.Bold).Sprintf("\nError Metadata: ")
		for key := range e.Metadata {
			msg += fmt.Sprintf("\n  %v=%v", key, e.Metadata[key])
		}
	}
	return msg
}

// ExitCode returns the exit code to use with this error.
func (a *APIError) ExitCode() int {
	if a.exitCode == 0 {
		return int(exitcodes.Error)
	}
	return int(a.exitCode)
}

// NewAPIError returns an error that is an instance of APIError or error.
// Example:
//
//	 // get the enhanced api error details, if the error is an API error
//		var apiError *error.APIError
//		if errors.As(err, &apiError) {
//			console.Stderrln(apiError.DetailedError())
//			os.Exit(apiError.ExitCode())
//		}
//
//	 // else its a regular error
//	 console.Stderrln(err.Error())
//	 os.Exit(int(exitcodes.Error))
//
// If the body is not deserializable to an ApiErrorResponse struct a regular text based error will be returned.
func NewAPIError(
	msg string,
	httpStatusCode int,
	body []byte,
	exitCode exitcodes.ExitCode,
) error {
	if apiErrorResponse, ok := parseApiErrorFromBody(body); ok {
		return &APIError{
			message:          msg,
			httpStatusCode:   httpStatusCode,
			apiErrorResponse: apiErrorResponse,
			exitCode:         exitCode,
		}
	}
	return fmt.Errorf("an API error occurred, msg: %v, status code: %v, body: %v", msg, httpStatusCode, string(body))
}

// parseApiErrorFromBody makes an effort to see if the body is a well-formed api error response object.
// It returns a pointer to the ApiErrorResponse and a boolean indicated whether the body was successfully unmarshalled
func parseApiErrorFromBody(body []byte) (*ApiErrorResponse, bool) {
	apiError := &ApiErrorResponse{}
	if e := json.Unmarshal(body, apiError); e != nil {
		return nil, false
	}

	if apiError.ErrorID == "" || len(apiError.Errors) < 1 {
		return nil, false
	}

	return apiError, true
}
