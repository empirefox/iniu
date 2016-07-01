package validate

import (
	"regexp"
	"strconv"
	"time"
)

func CreateParamValidator(opt *option, rmap map[string]*regexp.Regexp) (Validator, bool) {
	// Check for param validators
	for key, reg := range rmap {
		ps := reg.FindStringSubmatch(opt.raw)
		if len(ps) > 0 {
			if validatorCreator, ok := ParamTagMap[key]; ok {
				return validatorCreator(opt, ps[1:]...)
			}
		}
	}
	return nil, false
}

type ParamValidatorCreator func(opt *option, params ...string) (Validator, bool)

var ParamTagMap = map[string]ParamValidatorCreator{
	"strlen": func(opt *option, params ...string) (Validator, bool) {
		min, max, ok := parseIntParam(params)
		if !ok {
			return nil, false
		}
		return NewStrLenValidator(opt, min, max)
	},
	"matches": func(opt *option, params ...string) (Validator, bool) {
		if len(params) != 1 {
			return nil, false
		}

		reg, err := regexp.Compile(params[0])
		if err != nil {
			return nil, false
		}

		v := &StringValidator{
			option: opt,
			validator: func(s string) bool {
				return reg.MatchString(s)
			},
		}
		return v, true
	},
	"intrange": func(opt *option, params ...string) (Validator, bool) {
		min, max, ok := parseIntParam(params)
		if !ok {
			return nil, false
		}
		return NewIntValidator(opt, min, max)
	},
	"uintrange": func(opt *option, params ...string) (Validator, bool) {
		min, max, ok := parseUintParam(params)
		if !ok {
			return nil, false
		}
		return NewUintValidator(opt, min, max)
	},
	"floatrange": func(opt *option, params ...string) (Validator, bool) {
		fr, ok := parseFloatParam(params)
		if !ok {
			return nil, false
		}
		return NewFloatValidator(opt, fr), true
	},
	"timerange": func(opt *option, params ...string) (Validator, bool) {
		tr, ok := parseTimeParam(params)
		if !ok {
			return nil, false
		}
		return NewTimeRangeValidator(opt, tr), true
	},
}

var StringParamTagRegexMap = map[string]*regexp.Regexp{
	"strlen":  regexp.MustCompile(`^strlen([\(\[])(\d+)\|(\d+)([\)\]])$`),
	"matches": regexp.MustCompile(`matches\(([^)]+)\)`),
}

var IntParamTagRegexMap = map[string]*regexp.Regexp{
	"intrange": regexp.MustCompile(`^range([\(\[])(-?\d+)\|(-?\d+)([\)\]])$`),
}

var UintParamTagRegexMap = map[string]*regexp.Regexp{
	"uintrange": regexp.MustCompile(`^range([\(\[])(\d+)\|(\d+)([\)\]])$`),
}

var FloatParamTagRegexMap = map[string]*regexp.Regexp{
	"floatrange": regexp.MustCompile(`^range([\(\[])(-?\d+\.?\d*)\|(-?\d+\.?\d*)\|(\d+)([\)\]])$`),
}

var TimeParamTagRegexMap = map[string]*regexp.Regexp{
	// const longForm = "2006-01-02 15:04:05"
	// before = 2017-01-01 02:02:03
	"timerange": regexp.MustCompile(`^(before|after)(=?)\((\d{4}-\d{2}-\d{2}\s\d{2}:\d{2}:\d{2})\)$`),
}

func parseTimeParam(params []string) (tr *TimeRange, ok bool) {
	if len(params) != 3 {
		return
	}

	t, err := time.Parse("2006-01-02 15:04:05", params[2])
	if err != nil {
		return
	}

	return &TimeRange{
		before:  params[0] == "before",
		include: params[1] == "=",
		time:    t,
	}, true
}

func parseIntParam(params []string) (min, max int64, ok bool) {
	if len(params) != 4 {
		return
	}

	min, _ = strconv.ParseInt(params[1], 0, 64)
	max, _ = strconv.ParseInt(params[2], 0, 64)

	if !isIncludeParam(params[0]) {
		min++
	}
	if !isIncludeParam(params[3]) {
		max--
	}
	if min > max {
		return
	}
	ok = true
	return
}

func parseUintParam(params []string) (min, max uint64, ok bool) {
	if len(params) != 4 {
		return
	}

	min, _ = strconv.ParseUint(params[1], 0, 64)
	max, _ = strconv.ParseUint(params[2], 0, 64)

	if !isIncludeParam(params[0]) {
		min++
	}
	if !isIncludeParam(params[3]) {
		max--
	}
	if min > max {
		return
	}
	ok = true
	return
}

func parseFloatParam(params []string) (fr *FloatRange, ok bool) {
	if len(params) != 5 {
		return
	}

	min64, _ := strconv.ParseFloat(params[1], 64)
	max64, _ := strconv.ParseFloat(params[2], 64)
	min32, _ := strconv.ParseFloat(params[1], 32)
	max32, _ := strconv.ParseFloat(params[2], 32)
	prec, _ := strconv.ParseInt(params[3], 0, 64)

	if min64 > max64 {
		return
	}

	return &FloatRange{
		min64: min64,
		max64: max64,
		min32: float32(min32),
		max32: float32(max32),
		inMin: isIncludeParam(params[0]),
		inMax: isIncludeParam(params[4]),
		prec:  int(prec),
	}, true
}

func isIncludeParam(s string) bool {
	switch s {
	case "[", "]":
		return true
	}
	return false
}
