package base

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/empirefox/iniu/base/bo"
	"github.com/empirefox/iniu/validate"
	"github.com/empirefox/shirolet"
	"github.com/jinzhu/gorm"
	"github.com/mcuadros/go-defaults"
)

type PosInfo struct {
	HasPos  bool
	Sorting string
}

type StructMeta struct {
	bo.StructMetaOut
	PosInfo
	TableName string
	GroupBase string
	SoftDel   bool
	Fields    []*FieldMeta
	FieldMap  map[string]*FieldMeta

	// gen using field tag SA:"+/#"
	SearchAttrs []string

	DeletePerm shirolet.Permit

	PkShowFields []string
	GetPkShow    func(db *gorm.DB) (results [][]interface{})

	Mtype reflect.Type
	stype reflect.Type

	newer    func() interface{}
	slice    func() interface{}
	nilSlice func() interface{}

	defaultModel Model
}

func NewStructMeta(m Model) *StructMeta {
	scope := &gorm.Scope{Value: m}
	Mtype := scope.IndirectValue().Type()

	sm := &StructMeta{
		TableName: scope.TableName(),
		FieldMap:  make(map[string]*FieldMeta),
		Mtype:     Mtype,
		stype:     reflect.SliceOf(Mtype),
	}

	sm.Name = Mtype.Name()
	sm.NoDefault = reflect.DeepEqual(reflect.ValueOf(sm.Default()).Interface(), sm.New())

	for _, field := range scope.Fields() {
		if field.IsIgnored {
			continue
		}

		meta := NewFieldMeta(field)
		if field.Name == "Pos" && field.Struct.Type.Kind() == reflect.String {
			sm.HasPos = true
			ps := strings.Split(field.Tag.Get("pos"), ",")
			if strings.TrimSpace(ps[0]) == "desc" {
				sm.Sorting = " DESC"
			}
			//			if len(ps) > 1 {
			//				posBase := strings.TrimSpace(ps[1])
			//				if _, ok := scope.FieldByName(posBase); ok {
			//					sm.GroupBase = posBase
			//				}
			//			}
		} else if field.Name == "DeletedAt" {
			sm.SoftDel = true
		}
		if field.Tag.Get("GB") == "must" {
			if sm.GroupBase != "" {
				log.WithFields(logrus.Fields{
					"form":   sm.Name,
					"fields": []string{sm.GroupBase, field.Name},
				}).Fatalln("Duplicate GroupBase(GB) tag")
			}
			sm.GroupBase = field.Name
		}
		sm.Fields = append(sm.Fields, meta)
		sm.FieldMap[meta.Name] = meta
	}

	return sm
}

func (sm *StructMeta) LoadFromDb(db *gorm.DB) error {
	form := &Form{Name: sm.Name}
	if err := db.Preload("Fields").Where(form).Find(form).Error; err != nil {
		return nil
	}
	sm.LoadFromFields(form)
	return nil
}

func (sm *StructMeta) LoadFromFields(form *Form) {
	sm.Out = &form.FormOut
	if form.DeletePerm != "" {
		sm.DeletePerm = shirolet.NewPermitRaw(form.DeletePerm)
	}

	for _, ff := range form.Fields {
		field, ok := sm.FieldMap[ff.Name]
		if !ok {
			log.WithFields(logrus.Fields{
				"form":  sm.Name,
				"field": ff.Name,
			}).Errorln("LoadFromDb failed")
		}
		field.Out = &ff.FieldOut
		field.FieldMetaData = ff.FieldMetaData
		if ff.CreatePerm != "" {
			field.CreatePerm = shirolet.NewPermitRaw(ff.CreatePerm)
		}
		if ff.ReadPerm != "" {
			field.ReadPerm = shirolet.NewPermitRaw(ff.ReadPerm)
		}
		if ff.UpdatePerm != "" {
			field.UpdatePerm = shirolet.NewPermitRaw(ff.UpdatePerm)
		}
	}
}

func (sm *StructMeta) Validate(name string, data interface{}) error {
	field, ok := sm.FieldMap[name]
	if !ok {
		// Wrong field
		return fmt.Errorf("Field %s.%s not found", sm.Name, name)
	}
	if field.Validation == nil {
		// No validation, pass
		return nil
	}
	return validate.WrapErr(field.Validation.Validate(data), name)
}

type PermittedFields struct {
	Fields         map[string]bool
	StructedFields map[string]*StructMeta
	Structs        map[*StructMeta]bool
}

// GetPermittedFields return always non-nil
func (sm *StructMeta) GetPermittedFields(holds shirolet.Holds, permTyp PermType) *PermittedFields {
	pf := PermittedFields{
		Fields:         make(map[string]bool),
		StructedFields: make(map[string]*StructMeta),
		Structs:        make(map[*StructMeta]bool),
	}

	for _, field := range sm.Fields {
		perm, found := field.GetPerm(permTyp)
		if !found {
			// Wrong perm type
			panic("PermType not found")
		}
		if field.Struct == nil {
			if perm == nil || perm.SatisfiedBy(holds) {
				pf.Fields[field.Name] = true
			}
		} else {
			pf.StructedFields[field.Name] = field.Struct
			pf.Structs[field.Struct] = true
		}
	}
	pf.Structs[sm] = true
	return &pf
}

// GetPermittedMetas only basic field without struct based
func (sm *StructMeta) GetPermittedMetas(holds shirolet.Holds, permTyp PermType) *bo.MetasOutput {
	var fos []*bo.FieldMetaOut

	for name := range sm.GetPermittedFields(holds, permTyp).Fields {
		fos = append(fos, &sm.FieldMap[name].FieldMetaOut)
	}

	if len(fos) > 0 {
		return &bo.MetasOutput{
			StructMetaOut: &sm.StructMetaOut,
			Fields:        fos,
		}
	}
	return nil
}

func (sm *StructMeta) CanDelete(holds shirolet.Holds) bool {
	if sm.DeletePerm == nil {
		return true
	}
	return sm.DeletePerm.SatisfiedBy(holds)
}

func (sm *StructMeta) New() interface{} {
	if sm.newer != nil {
		return sm.newer()
	}
	return reflect.New(sm.Mtype).Interface()
}

func (sm *StructMeta) NilSlice() interface{} {
	if sm.nilSlice != nil {
		return sm.nilSlice()
	}
	return reflect.New(reflect.SliceOf(sm.Mtype)).Interface()
}

func (sm *StructMeta) Slice() interface{} {
	if sm.slice != nil {
		return sm.slice()
	}
	svalue := reflect.New(reflect.SliceOf(sm.Mtype))
	svalue.Elem().Set(reflect.MakeSlice(reflect.SliceOf(sm.Mtype), 0, 0))
	return svalue.Interface()
}

func (sm *StructMeta) Default() Model {
	if sm.defaultModel == nil {
		nm := sm.New()
		defaults.SetDefaults(nm)
		sm.defaultModel = nm
	}
	return sm.defaultModel
}

func (sm *StructMeta) GetSorting() (Sorting string, ok bool) {
	return sm.Sorting, sm.HasPos
}
