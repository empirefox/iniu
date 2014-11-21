package comm

import (
	"encoding/json"
	"reflect"
	"regexp"
	"strconv"
	"time"

	"github.com/golang/glog"
	"github.com/deckarep/golang-set"
)

func Atoi64(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

func I64toA(i int64) string {
	return strconv.FormatInt(i, 10)
}

func NewSet(s []string) mapset.Set {
	a := mapset.NewSet()
	for _, v := range s {
		a.Add(v)
	}
	return a
}

func UnescapeParam(input map[string]string) map[string]interface{} {
	l := len(input)
	output := make(map[string]interface{}, l)
	for k, v := range input {
		output[k] = unescapeParamValue(v)
	}
	return output
}

func unescapeParamValue(in string) interface{} {
	switch in {
	case "true":
		return true
	case "false":
		return false
	default:
		return in
	}
}

func ToJsonFunc(arg interface{}) string {
	typ := reflect.TypeOf(arg)
	val := reflect.ValueOf(arg)
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = val.Elem()
	}
	if reflect.DeepEqual(val.Interface(), reflect.Zero(typ).Interface()) {
		switch typ.Kind() {
		case reflect.String:
			return ""
		case reflect.Struct:
			return "{}"
		case reflect.Slice, reflect.Map, reflect.Array:
			return "[]"
		}
	}
	j, _ := json.Marshal(arg)
	return string(j)
}

// Copy from gorm scope.go
// TableName get table name
var pluralMapKeys = []*regexp.Regexp{regexp.MustCompile("ch$"), regexp.MustCompile("ss$"), regexp.MustCompile("sh$"), regexp.MustCompile("day$"), regexp.MustCompile("y$"), regexp.MustCompile("x$"), regexp.MustCompile("([^s])s?$")}
var pluralMapValues = []string{"ches", "sses", "shes", "days", "ies", "xes", "${1}s"}

func ToPlural(name string) string {
	for index, reg := range pluralMapKeys {
		if reg.MatchString(name) {
			return reg.ReplaceAllString(name, pluralMapValues[index])
		}
	}
	return name
}

func CurrYear() int {
	return time.Now().Year()
}

func CurrMonth() int {
	return int(time.Now().Month())
}

func Substr(str string, start, length int) string {
	runes := []rune(str)
	rl := len(runes)
	end := 0

	if start < 0 {
		start = rl - 1 + start
	}
	end = start + length

	if start > end {
		start, end = end, start
	}

	if start < 0 {
		start = 0
	}
	if start > rl {
		start = rl
	}
	if end < 0 {
		end = 0
	}
	if end > rl {
		end = rl
	}

	return string(runes[start:end])
}

func ErrLog(err error) error {
	if err != nil {
		glog.Errorln(err)
	}
	return err
}

func FatalLog(err error) error {
	if err != nil {
		glog.Fatalln(err)
	}
	return err
}

func PtrS(s string) *string {
	return &s
}

func PtrB(s bool) *bool {
	return &s
}

func PtrI64(s int64) *int64 {
	return &s
}

func PtrI(s int) *int {
	return &s
}

func PtrT(s time.Time) *time.Time {
	return &s
}

func FixNum2Int64(raw *map[string]interface{}) {
	m := *raw
	for k, v := range m {
		switch f := v.(type) {
		case float64:
			if i := int64(f); float64(i) == f {
				m[k] = i
			}
		}
	}
}
