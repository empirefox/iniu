package validate

import "fmt"

type ValidateErr struct {
	Field string
	Raw   string
	Show  string
}

func (err *ValidateErr) Error() string {
	if err.Show != "" {
		return fmt.Sprintf(`validate:{"field":"%s","error":"%s"}`, err.Field, err.Show)
	}
	return fmt.Sprintf(`validate:{"field":"%s","raw":"%s"}`, err.Field, err.Raw)
}

func (err *ValidateErr) WithField(f string) *ValidateErr {
	err.Field = f
	return err
}

func WrapErr(err error, f string) error {
	if verr, ok := err.(*ValidateErr); ok {
		return verr.WithField("Id")
	}
	return err
}
