package validate

import "fmt"

type BoolValidation struct {
	validator Validator
}

func NewBoolValidation(opts []string) *BoolValidation {
	var v *BoolValidator

	for _, rawopt := range opts {
		opt := parseOption(rawopt)

		switch opt.raw {
		case "true":
			v = NewBoolValidator(opt, true)
		case "false":
			v = NewBoolValidator(opt, false)
		default:
			log.WithField("raw", opt.raw).Errorln("Failed to parse param validator")
		}
	}
	if v == nil {
		return nil
	}
	return &BoolValidation{v}
}

func (vs BoolValidation) Validate(data interface{}) error {
	switch data.(type) {
	case bool:
		return validateAll([]Validator{vs.validator}, data)
	case nil:
		return validateAll([]Validator{vs.validator}, false)
	}
	return fmt.Errorf("Cannot convert to bool")
}
