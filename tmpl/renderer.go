package tmpl

import (
	"fmt"
	"github.com/deckarep/golang-set"
	"github.com/empirefox/iniu/bucket"
	"github.com/empirefox/iniu/comm"
	bucketdb "github.com/empirefox/iniu/gorm"
	"github.com/go-martini/martini"
	"github.com/golang/glog"
	"net/http"
	"reflect"
	"text/template"
)

func init() {
	initFieldsStruct()
}

var (
	//可写的bucket.Bucket字段名称
	names []interface{}
	//bucketdb.Bucket的全部字段
	fields []reflect.StructField
	//可写字段的map
	writableMap = make(map[string]reflect.StructField, 0)
	formControl = `<div class="form-group">
				<label for="%s" class="col-sm-2 control-label">%s</label>
				<div class="col-sm-10">
					<input class="form-control" id="%s" name="%s" ng-model="bucket.%s" type="%s"%s>
				</div>
			</div>`
	staticControl = `<div class="form-group">
				<label class="col-sm-2 control-label">%s</label>
				<div class="col-sm-10">
					<p ng-bind="bucket.%s" class="form-control-static"></p>
				</div>
			</div>`
	hiddenControl = `<input name="%s" value="{{bucket.%s}}" type="hidden">`
)

type tplArgs struct {
	Fields  []reflect.StructField
	Buckets []bucketdb.Bucket
}

var BucketsHandler = func() martini.Handler {
	writableNameSet := mapset.NewSetFromSlice(names)
	bucketsTpl, err := template.New("buckets").Funcs(template.FuncMap{
		"toControl": ToControl(writableNameSet),
		"json":      comm.ToJsonFunc,
	}).Delims("[[", "]]").Parse(bucketsHtml)
	if err != nil {
		glog.Errorln(err)
		panic("初始化bucketsTpl错误")
	}
	return func(w http.ResponseWriter, r *http.Request) {
		tplArgs := &tplArgs{
			Fields:  fields,
			Buckets: bucketdb.All(),
		}
		bucketsTpl.Execute(w, tplArgs)
	}
}

var ToControl = func(writableNameSet mapset.Set) interface{} {
	return func(arg reflect.StructField) string {
		if wf, ok := writableMap[arg.Name]; ok {
			if wf.Tag.Get("hidden") == "true" {
				return ToHiddenControl(arg)
			}
			if wf.Tag.Get("static") == "true" {
				return ToStaticControl(arg)
			}
			return ToFormControl(wf)
		}
		return ToStaticControl(arg)
	}
}

func ToFormControl(arg reflect.StructField) string {
	a := arg.Name
	inputType := arg.Tag.Get("input-type")
	if inputType == "" {
		inputType = "text"
	}
	var status string
	if arg.Tag.Get("binding") == "required" {
		status = ` required="required"`
	}
	if arg.Tag.Get("disabled") == "true" {
		status += ` disabled`
	}
	return fmt.Sprintf(formControl, a, a, a, a, a, inputType, status)
}

func ToStaticControl(arg reflect.StructField) string {
	a := arg.Name
	return fmt.Sprintf(staticControl, a, a)
}

func ToHiddenControl(arg reflect.StructField) string {
	a := arg.Name
	return fmt.Sprintf(hiddenControl, a, a)
}

func initFieldsStruct() {
	visitFields(reflect.TypeOf(bucket.Bucket{}), getNamesFunc(), getWritableMapFunc())
	visitFields(reflect.TypeOf(bucketdb.Bucket{}), getFiledsFunc())
}

type FieldVisitor func(field reflect.StructField)

func visitFields(typ reflect.Type, visit ...FieldVisitor) {
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
