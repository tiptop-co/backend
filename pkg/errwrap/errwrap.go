package errwrap

import "fmt"

// WrapErr returns error like "wrapper: base"
func Wrap(wrapper error, base error) error {
	if base != nil {
		return fmt.Errorf("%w: %w", wrapper, base)
	}

	return nil
}

func WrapMsg(wrapper string, base error) error {
	if base != nil {
		return fmt.Errorf("%s: %w", wrapper, base)
	}

	return nil
}
