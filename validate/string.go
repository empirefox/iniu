package validate

import "fmt"

type StringValidation struct {
	validators []Validator
}

func NewStringValidation(opts []string) *StringValidation {
	var vs []Validator

	for _, rawopt := range opts {
		opt := parseOption(rawopt)

		if v, ok := CreateParamValidator(opt, StringParamTagRegexMap); ok {
			vs = append(vs, v)
		} else if v, ok := NewStringValidator(opt); ok {
			vs = append(vs, v)
		} else {
			log.WithField("raw", opt.raw).Errorln("Failed to parse param validator")
		}
	}
	if len(vs) > 0 {
		return &StringValidation{vs}
	}
	return nil
}

func (vs StringValidation) Validate(data interface{}) error {
	switch data.(type) {
	case string, *string, []byte:
		return validateAll(vs.validators, data)
	case nil:
		return validateAll(vs.validators, "")
	}
	return fmt.Errorf("Cannot convert to string")
}
