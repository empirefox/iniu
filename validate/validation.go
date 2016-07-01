package validate

import (
	"fmt"
	"reflect"
	"time"

	"github.com/Sirupsen/logrus"
)

var log = logrus.New()

type Validation interface {
	Validate(data interface{}) error
}

func NewValidation(typ reflect.Type, tag string) Validation {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	options := parseTag(typ, tag)
	switch typ.Kind() {
	case reflect.String:
		return NewStringValidation(options)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return NewNumberValidation(typ, options)
	case reflect.Bool:
		return NewBoolValidation(options)
	case reflect.Struct:
		if typ == reflect.TypeOf(time.Time{}) {
			return NewTimeValidation(options)
		}
	}
	fmt.Printf("Cannot create Validation from %s with type %s\n", tag, typ.Name())
	return nil
}

func logErrValidationType(typ reflect.Type, data interface{}) {
	log.WithFields(logrus.Fields{
		"type": typ.Name(),
		"data": data,
	}).Errorln("Type convert failed")
}
