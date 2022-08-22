package errors

import "fmt"

func NewError(err error) error {
	return err
}

func NewWrappedError(err, thrownErr error) error {
	return fmt.Errorf("%w, thrown error: %v", err, thrownErr)
}

func NewErrorWithDynamicContext(err error, context string) error {
	return fmt.Errorf("%w, %s", err, context)
}

func NewWrappedErrorWithDynamicContext(err, thrownErr error, context string) error {
	return fmt.Errorf("%w, %s, thrownError: %v", err, context, thrownErr)
}
