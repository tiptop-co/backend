package errwrap

import "fmt"

func WrapWithMsg(prefixMsg string, err error) error {
	if err != nil {
		return fmt.Errorf("%s: %w", prefixMsg, err)
	}
	return nil
}

func WrapWithErr(prefixErr error, err error) error {
	if err != nil {
		return fmt.Errorf("%w: %w", prefixErr, err)
	}
	return nil
}
