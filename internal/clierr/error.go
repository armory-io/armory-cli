package clierr

import (
	"fmt"
	"github.com/armory/armory-cli/internal/clierr/exitcodes"
	"github.com/fatih/color"
)

// Error is a generic CLI error type that was not necessarily caused by the API.
type Error struct {
	errorID  string
	message  string
	error    error
	exitCode exitcodes.ExitCode
}

func (e *Error) Error() string {
	return fmt.Sprintf("an error occurred, msg: %v", e.message)
}

// DetailedError returns a human readable multiline colored error message.
func (e *Error) DetailedError() string {
	msg := color.New(color.FgRed, color.Bold).Sprintln(e.message)
	if e.errorID != "" {
		msg += color.New(color.FgRed, color.Bold).Sprintf("Error Id: ")
		msg += fmt.Sprintf("%s\n", e.errorID)
	}
	if e.error != nil {
		msg += color.New(color.FgRed, color.Bold).Sprintf("Error Message: ")
		msg += fmt.Sprintf("%s\n", e.error)
	}
	return msg
}

func (e *Error) ExitCode() int {
	return int(e.exitCode)
}

func (e *Error) Unwrap() error {
	return e.error
}

func NewError(
	message string,
	errorID string,
	error error,
	exitCode exitcodes.ExitCode,
) error {
	return &Error{
		error:    error,
		errorID:  errorID,
		message:  message,
		exitCode: exitCode,
	}
}
