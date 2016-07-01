package validate

import (
	"strconv"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
)

func validateAll(vs []Validator, data interface{}) error {
	for _, v := range vs {
		result := v.Validate(data)
		negate := v.Option().negate
		if !result && !negate || result && negate {
			return v.Err()
		}
	}
	return nil
}

type Validator interface {
	Validate(data interface{}) bool
	Err() error
	Option() *option
}

// StringValidator
type StringValidator struct {
	*option
	validator govalidator.Validator
}

func NewStringValidator(opt *option) (Validator, bool) {
	validator, ok := govalidator.TagMap[opt.raw]
	if !ok {
		return nil, false
	}
	return &StringValidator{
		option:    opt,
		validator: validator,
	}, true
}

func (v *StringValidator) Validate(data interface{}) bool {
	switch d := data.(type) {
	case []byte:
		return v.validator(string(d))
	case string:
		return v.validator(d)
	case *string:
		return v.validator(*d)
	}
	log.WithField("data", data).Errorln("Cannot convert to string")
	return false
}

// StrLenValidator
type StrLenValidator struct {
	IntValidator
}

func NewStrLenValidator(opt *option, min, max int64) (*StrLenValidator, bool) {
	intv, ok := NewIntValidator(opt, min, max)
	if !ok {
		return nil, false
	}
	return &StrLenValidator{IntValidator: *intv}, true
}

func (v StrLenValidator) Validate(data interface{}) bool {
	switch d := data.(type) {
	case []byte:
		return v.IntValidator.Validate(len(d))
	case string:
		return v.IntValidator.Validate(len(d))
	case *string:
		return v.IntValidator.Validate(len(*d))
	}
	log.WithField("data", data).Errorln("Cannot convert to string")
	return false
}

// IntValidator
type IntValidator struct {
	*option
	min int64
	max int64
}

func NewIntValidator(opt *option, min, max int64) (*IntValidator, bool) {
	if min > max {
		return nil, false
	}
	v := &IntValidator{
		option: opt,
		min:    min,
		max:    max,
	}
	return v, true
}

func (v IntValidator) Validate(data interface{}) bool {
	var d int64
	switch td := data.(type) {
	case int:
		d = int64(td)
	case int8:
		d = int64(td)
	case int16:
		d = int64(td)
	case int32:
		d = int64(td)
	case int64:
		d = int64(td)
	case nil:
		d = 0
	default:
		log.WithField("data", data).Errorln("Cannot convert to int")
		return false
	}
	return d >= v.min && d <= v.max
}

// UintValidator
type UintValidator struct {
	*option
	min uint64
	max uint64
}

func NewUintValidator(opt *option, min, max uint64) (*UintValidator, bool) {
	if min > max {
		return nil, false
	}
	v := &UintValidator{
		option: opt,
		min:    min,
		max:    max,
	}
	return v, true
}

func (v UintValidator) Validate(data interface{}) bool {
	var d uint64
	switch td := data.(type) {
	case uint:
		d = uint64(td)
	case uint8:
		d = uint64(td)
	case uint16:
		d = uint64(td)
	case uint32:
		d = uint64(td)
	case uint64:
		d = uint64(td)
	case nil:
		d = 0
	default:
		log.WithField("data", data).Errorln("Cannot convert to uint")
		return false
	}
	return d >= v.min && d <= v.max
}

// FloatValidator
type FloatValidator struct {
	*option
	FloatRange
}

func NewFloatValidator(opt *option, fr *FloatRange) *FloatValidator {
	return &FloatValidator{
		option:     opt,
		FloatRange: *fr,
	}
}

func (v FloatValidator) Validate(data interface{}) bool {
	var str string
	var d64 float64
	switch d := data.(type) {
	case float32:
		return v.precOk(data) && (v.min32 < d && d < v.max32 || v.inMin && v.min32 == d || v.inMax && v.max32 == d)
	case float64:
		d64 = d
	case []byte:
		str = string(d)
		var err error
		d64, err = strconv.ParseFloat(str, 64)
		if err != nil {
			return false
		}
	default:
		log.WithField("data", data).Errorln("Cannot convert to float")
		return false
	}

	return v.precOk(data) && (v.min64 < d64 && d64 < v.max64 || v.inMin && v.min64 == d64 || v.inMax && v.max64 == d64)
}

func (v FloatValidator) precOk(data interface{}) bool {
	ok := true
	if v.prec != 0 {
		var str string
		switch d := data.(type) {
		case float32:
			str = strconv.FormatFloat(float64(d), 'f', -1, 32)
		case float64:
			str = strconv.FormatFloat(d, 'f', -1, 64)
		case []byte:
			str = string(d)
		}
		ss := strings.SplitN(str, ".", 2)
		prec := 0
		if len(ss) == 2 {
			prec = len(ss[1])
		}
		ok = prec <= v.prec
	}
	return ok
}

// BoolValidator
type BoolValidator struct {
	*option
	value bool
}

func NewBoolValidator(opt *option, v bool) *BoolValidator {
	return &BoolValidator{
		option: opt,
		value:  v,
	}
}

func (v BoolValidator) Validate(data interface{}) bool {
	d, ok := data.(bool)
	if !ok {
		log.WithField("data", data).Errorln("Cannot convert to bool")
	}
	return d == v.value
}

// TimeRangeValidator
type TimeRangeValidator struct {
	*option
	TimeRange
}

func NewTimeRangeValidator(opt *option, tr *TimeRange) *TimeRangeValidator {
	return &TimeRangeValidator{
		option:    opt,
		TimeRange: *tr,
	}
}

func (v TimeRangeValidator) Validate(data interface{}) bool {
	var t *time.Time
	switch d := data.(type) {
	case time.Time:
		t = &d
	case *time.Time:
		t = d
	default:
		log.WithField("data", data).Errorln("Cannot convert to time")
		return false
	}

	if t == nil || t.IsZero() {
		return false
	}

	rangeOk := false
	if v.before {
		rangeOk = t.Before(v.time)
	} else {
		rangeOk = t.After(v.time)
	}

	return rangeOk || (v.include && t.Equal(v.time))
}

// ZeroTimeValidator
type ZeroTimeValidator struct {
	*option
}

func NewZeroTimeValidator(opt *option) *ZeroTimeValidator {
	return &ZeroTimeValidator{opt}
}

func (v ZeroTimeValidator) Validate(data interface{}) bool {
	var t *time.Time
	switch d := data.(type) {
	case time.Time:
		t = &d
	case *time.Time:
		t = d
	default:
		log.WithField("data", data).Errorln("Cannot convert to time")
		return false
	}

	return t == nil || t.IsZero()
}
