package errors

import (
	"fmt"
)

func WrapError(wrapper error, err error) error {
	if err == nil {
		return nil
	}
	if err == ErrBadRequest ||
		err == ErrUnauthorized ||
		err == ErrNotFound ||
		err == ErrPermissionDenied {
		return err
	}

	return fmt.Errorf("%w: %v", wrapper, err)
}

func WrapNotFound(err error) error {
	return WrapError(ErrNotFound, err)
}

func WrapPermissionDenied(err error) error {
	return WrapError(ErrPermissionDenied, err)
}

func WrapBadRequest(err error) error {
	return WrapError(ErrBadRequest, err)
}

func WrapUnauthorized(err error) error {
	return WrapError(ErrUnauthorized, err)
}
