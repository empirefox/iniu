package validate

import (
	"fmt"
	"time"
)

type TimeValidation struct {
	validators []Validator
}

func NewTimeValidation(opts []string) *TimeValidation {
	var vs []Validator

	for _, rawopt := range opts {
		opt := parseOption(rawopt)

		if v, ok := CreateParamValidator(opt, TimeParamTagRegexMap); ok {
			vs = append(vs, v)
		} else if opt.raw == "zero" {
			vs = append(vs, NewZeroTimeValidator(opt))
		} else {
			log.WithField("raw", opt.raw).Errorln("Failed to parse param validator")
		}
	}
	if len(vs) > 0 {
		return &TimeValidation{vs}
	}
	return nil
}

func (vs TimeValidation) Validate(data interface{}) error {
	switch d := data.(type) {
	case time.Time:
		return validateAll(vs.validators, &d)
	case *time.Time, nil:
		return validateAll(vs.validators, data)
	}
	return fmt.Errorf("Cannot convert %v to *time.Time", data)
}

type TimeRange struct {
	time    time.Time
	before  bool
	include bool
}
