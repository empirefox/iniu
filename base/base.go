package base

import (
	"errors"
	"reflect"

	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
)

var log = logrus.New()

type Model interface{}

type PermType int

const (
	READ PermType = iota + 1
	UPDATE
	CREATE
)

var (
	// key: structname
	StructMetaMap = make(map[string]*StructMeta)
	// key: tablename
	TableMetaMap = make(map[string]*StructMeta)
	// key: type
	TypeMetaMap = make(map[reflect.Type]*StructMeta)

	// exports
//	SaveModel     = saveModel
//	SaveModelWith = saveModelWith
)

func RegisterStruct(
	m Model,
	searchAttrs []string,
	newer func() interface{},
	slice func() interface{},
	nilSlice func() interface{},
	pkShowFields []string,
	getPkShow func(*gorm.DB) [][]interface{}) *StructMeta {

	sm := NewStructMeta(m)
	sm.SearchAttrs = searchAttrs
	sm.newer = newer
	sm.slice = slice
	sm.nilSlice = nilSlice
	sm.PkShowFields = pkShowFields
	sm.GetPkShow = getPkShow
	StructMetaMap[sm.Name] = sm
	TableMetaMap[sm.TableName] = sm
	TypeMetaMap[sm.Mtype] = sm
	//	TypeMetaMap[reflect.PtrTo(sm.Mtype)] = sm
	return sm
}

func TryLoadFieldStruct() {
	for _, sm := range TypeMetaMap {
		for _, field := range sm.Fields {
			ftyp := field.Field.Struct.Type
			for ftyp.Kind() == reflect.Slice || ftyp.Kind() == reflect.Ptr {
				ftyp = ftyp.Elem()
			}
			if fsm, ok := TypeMetaMap[ftyp]; ok {
				field.Struct = fsm
			}
		}
	}
}

func tablename(m Model) string {
	s := &gorm.Scope{Value: m}
	return s.GetModelStruct().TableName(nil)
}

func Formname(m Model) string {
	s := &gorm.Scope{Value: m}
	return s.IndirectValue().Type().Name()
}

func TableSorting(table string) (string, bool) {
	if sm, ok := StructMetaMap[table]; ok {
		return sm.GetSorting()
	}
	return "", false
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
