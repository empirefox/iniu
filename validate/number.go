package validate

import (
	"fmt"
	"reflect"
	"regexp"
)

type NumberValidation struct {
	typ        reflect.Type
	validators []Validator
}

func NewNumberValidation(typ reflect.Type, opts []string) *NumberValidation {
	var regexMap map[string]*regexp.Regexp
	switch typ.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		regexMap = IntParamTagRegexMap
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		regexMap = UintParamTagRegexMap
	case reflect.Float32, reflect.Float64:
		regexMap = FloatParamTagRegexMap
	default:
		return nil
	}

	var vs []Validator
	for _, rawopt := range opts {
		opt := parseOption(rawopt)

		if v, ok := CreateParamValidator(opt, regexMap); ok {
			vs = append(vs, v)
		} else {
			log.WithField("raw", opt.raw).Errorln("Failed to parse param validator")
		}
	}
	if len(vs) > 0 {
		return &NumberValidation{typ: typ, validators: vs}
	}
	return nil
}

func (vs NumberValidation) Validate(data interface{}) error {
	switch data.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return validateAll(vs.validators, data)

	case []byte:
		switch vs.typ.Kind() {
		case reflect.Float32, reflect.Float64:
			return validateAll(vs.validators, data)
		default:
			return fmt.Errorf("Cannot convert []byte to %s", vs.typ.Name())
		}

	case nil:
		switch vs.typ.Kind() {
		case reflect.Float32:
			return validateAll(vs.validators, float32(0))
		case reflect.Float64:
			return validateAll(vs.validators, float64(0))
		default:
			return fmt.Errorf("%s only accept float nil value", vs.typ.Name())
		}
	}

	return fmt.Errorf("Cannot convert to %s", vs.typ.Name())
}

type FloatRange struct {
	min64 float64
	max64 float64
	min32 float32
	max32 float32
	inMin bool
	inMax bool
	prec  int
}
