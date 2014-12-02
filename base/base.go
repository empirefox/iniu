package base

import (
	"errors"
	"reflect"

	"github.com/jinzhu/gorm"

	. "github.com/empirefox/iniu/gorm/db"
)

type J map[string]interface{}

type Creator interface {
	New() interface{}
	Slice() interface{}
	MakeSlice() interface{}
	IndirectSlice() interface{}
	//return nil if m is Zero
	Example() Model
	Form() map[string]interface{}
	HasPos() bool
}

type Model interface{}

type Table interface{}

var (
	Creators     = map[Table]Creator{}
	FormTableMap = map[string]string{}

	// exports
	SaveModel     = saveModel
	SaveModelWith = saveModelWith
)

type creator struct {
	mtype  reflect.Type
	stype  reflect.Type
	m      Model
	form   map[string]interface{}
	hasPos bool
}

func (c creator) New() interface{} {
	return reflect.New(c.mtype).Interface()
}

func (c creator) Slice() interface{} {
	return reflect.New(reflect.SliceOf(c.mtype)).Interface()
}

func (c creator) MakeSlice() interface{} {
	svalue := reflect.New(reflect.SliceOf(c.mtype))
	svalue.Elem().Set(reflect.MakeSlice(reflect.SliceOf(c.mtype), 0, 0))
	return svalue.Interface()
}

func (c creator) IndirectSlice() interface{} {
	return reflect.MakeSlice(reflect.SliceOf(c.mtype), 0, 0).Interface()
}

func (c creator) Example() Model {
	return c.m
}

func (c creator) Form() map[string]interface{} {
	return c.form
}

func (c creator) HasPos() bool {
	return c.hasPos
}

type NameOnlyField struct {
	Name string
}

func Register(m Model) {
	mtype := reflect.TypeOf(m)
	mname := mtype.Name()

	c := &creator{
		mtype: mtype,
		stype: reflect.SliceOf(mtype),
	}

	f := map[string]interface{}{}
	f["Name"] = mname

	fieldsCount := mtype.NumField()
	fields := make([]NameOnlyField, fieldsCount)
	for i := 0; i < fieldsCount; i++ {
		structField := mtype.Field(i)
		if tag := structField.Tag; tag.Get("json") != "-" || tag.Get("sql") != "-" {
			if structField.Name == "Pos" {
				c.hasPos = true
			}
			fields[i] = NameOnlyField{structField.Name}
		}
	}
	if fieldsCount > 0 {
		f["Fields"] = fields
	}

	if !reflect.DeepEqual(m, reflect.Zero(mtype).Interface()) {
		f["New"] = m
		c.m = m
	}

	c.form = f

	t := Tablename(m)
	Creators[t] = *c
	FormTableMap[mname] = t
}

func New(t Table) interface{} {
	if c, ok := Creators[t]; ok {
		return c.New()
	}
	return nil
}

func Slice(t Table) interface{} {
	if c, ok := Creators[t]; ok {
		return c.Slice()
	}
	return nil
}

func MakeSlice(t Table) interface{} {
	if c, ok := Creators[t]; ok {
		return c.MakeSlice()
	}
	return nil
}

func IndirectSlice(t Table) interface{} {
	if c, ok := Creators[t]; ok {
		return c.IndirectSlice()
	}
	return nil
}

func HasPos(t Table) bool {
	if c, ok := Creators[t]; ok {
		return c.HasPos()
	}
	return false
}

func Example(t Table) Model {
	if c, ok := Creators[t]; ok {
		return c.Example()
	}
	return nil
}

func ModelFormMetas(t Table) (map[string]interface{}, error) {
	if c, ok := Creators[t]; ok {
		return c.Form(), nil
	}
	return nil, errors.New("Form Metas Not Found: " + t.(string))
}

func ToTable(n string) string {
	return FormTableMap[n]
}

func ToFormname(t Table) string {
	if c, ok := Creators[t]; ok {
		f := c.Form()
		return f["Name"].(string)
	}
	return ""
}

func saveModel(iPtr *Model) error {
	return saveModelWith(iPtr, nil)
}

func saveModelWith(iPtr *Model, param map[string]interface{}) error {
	mPtr := reflect.New(reflect.TypeOf(*iPtr))
	m := mPtr.Elem()
	m.Set(reflect.ValueOf(*iPtr))

	t := ToTable(m.Type().Name())

	for k, v := range param {
		m.FieldByName(k).Set(reflect.ValueOf(v))
	}

	err := DB.Table(t).Save(mPtr.Interface()).Error
	if err != nil {
		return err
	}

	id := m.FieldByName("Id").Int()
	n := New(t)
	err = DB.Table(t).Where("id=?", id).First(n).Error
	if err != nil {
		return err
	}

	reflect.ValueOf(iPtr).Elem().Set(reflect.ValueOf(n).Elem())
	return nil
}

func Tablename(m Model) string {
	s := gorm.Scope{Value: m}
	return s.TableName()
}

func Formname(m Model) string {
	return reflect.TypeOf(m).Name()
}

//////////////////////////////
//         Models
//////////////////////////////

type Models interface{}

func ForEach(ms Models, handleFunc func(iPtr interface{}) error) error {
	switch reflect.TypeOf(ms).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(ms)
		for i := 0; i < s.Len(); i++ {
			err := handleFunc(s.Index(i).Addr().Interface())
			if err != nil {
				return err
			}
		}
		return nil
	}
	return errors.New("Models is Not a Slice")
}
