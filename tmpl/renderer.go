//deprecated
package tmpl

import (
	bucketdb "github.com/empirefox/iniu/gorm"
	"github.com/golang/glog"
	"reflect"
)

var (
	//可写的bucketdb.Bucket字段名称
	names []interface{}
	//bucketdb.Bucket的全部字段
	fields []reflect.StructField
	//可写字段的map
	writableMap = make(map[string]reflect.StructField, 0)
)

type tplArgs struct {
	Fields  []reflect.StructField
	Buckets []bucketdb.Bucket
}

func initFieldsStruct() {
	VisitFields(reflect.TypeOf(bucketdb.Bucket{}), getNamesFunc(), getWritableMapFunc(), getFiledsFunc())
}

type FieldVisitor func(field reflect.StructField)

func VisitFields(typ reflect.Type, visit ...FieldVisitor) {
	if len(visit) < 1 {
		glog.Errorln("visit不存在")
	}

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	n := typ.NumField()
	for i := 0; i < n; i++ {
		for _, v := range visit {
			v(typ.Field(i))
		}
	}
}

var getWritableMapFunc = func() FieldVisitor {
	return func(field reflect.StructField) {
		if field.Tag.Get("form") != "-" {
			writableMap[field.Name] = field
		}
	}
}

var getFiledsFunc = func() FieldVisitor {
	return func(field reflect.StructField) {
		if field.Tag.Get("form") != "-" {
			fields = append(fields, field)
		}
	}
}

var getNamesFunc = func() FieldVisitor {
	return func(field reflect.StructField) {
		if field.Tag.Get("form") != "-" {
			names = append(names, field.Name)
		}
	}
}
