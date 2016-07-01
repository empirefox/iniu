package base

import (
	"reflect"
	"regexp"
	"strconv"
	"time"

	"github.com/empirefox/iniu/base/bo"
	"github.com/empirefox/iniu/validate"
	"github.com/empirefox/shirolet"
	"github.com/jinzhu/gorm"
	"github.com/qor/qor/utils"
)

type FieldMeta struct {
	bo.FieldMetaOut
	IsEmbed bool
	IsPk    bool
	IsFk    bool
	//	Xmodify bool
	//	FkTable string
	Field  *gorm.Field
	Struct *StructMeta

	CreatePerm shirolet.Permit
	ReadPerm   shirolet.Permit
	// TODO MUST ignore any children table
	UpdatePerm shirolet.Permit

	Validation validate.Validation

	// shared with FormField
	FieldMetaData
}

func NewFieldMeta(gf *gorm.Field) *FieldMeta {
	field := &FieldMeta{
		IsEmbed:    !gf.IsNormal,
		IsPk:       gf.IsPrimaryKey,
		IsFk:       gf.IsForeignKey,
		Field:      gf,
		Validation: validate.NewValidation(gf.Struct.Type, gf.Tag.Get("valid")),
		//		Xmodify:    strings.TrimSpace(gf.Tag.Get("xmodify")) == "true",
	}
	field.Name = gf.Name
	field.Names = gf.Names

	fieldType := gf.Struct.Type
	for fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
	}

	// Set Meta Type
	if relationship := gf.Relationship; relationship != nil {
		if relationship.Kind == "has_one" {
			field.Type = "single_edit"
		} else if relationship.Kind == "has_many" {
			field.Type = "collection_edit"
		} else if relationship.Kind == "belongs_to" {
			field.Type = "select_one"
		} else if relationship.Kind == "many_to_many" {
			field.Type = "select_many"
		}
	} else {
		switch fieldType.Kind() {
		case reflect.String:
			var tag = gf.Tag
			if size, ok := utils.ParseTagOption(tag.Get("sql"))["SIZE"]; ok {
				if i, _ := strconv.Atoi(size); i > 255 {
					field.Type = "text"
				} else {
					field.Type = "string"
				}
			} else if text, ok := utils.ParseTagOption(tag.Get("sql"))["TYPE"]; ok && text == "text" {
				field.Type = "text"
			} else {
				field.Type = "string"
			}
		case reflect.Bool:
			field.Type = "checkbox"
		default:
			if regexp.MustCompile(`^(.*)?(u)?(int)(\d+)?`).MatchString(fieldType.Kind().String()) {
				field.Type = "number"
			} else if regexp.MustCompile(`^(.*)?(float)(\d+)?`).MatchString(fieldType.Kind().String()) {
				field.Type = "float"
			} else if _, ok := reflect.New(fieldType).Interface().(*time.Time); ok {
				field.Type = "datetime"
			} else {
				if fieldType.Kind() == reflect.Struct {
					field.Type = "single_edit"
				} else if fieldType.Kind() == reflect.Slice {
					refelectType := fieldType.Elem()
					for refelectType.Kind() == reflect.Ptr {
						refelectType = refelectType.Elem()
					}
					if refelectType.Kind() == reflect.Struct {
						field.Type = "collection_edit"
					}
				}
			}
		}
	}

	return field
}

func (fm *FieldMeta) GetPerm(permTyp PermType) (p shirolet.Permit, ok bool) {
	ok = true
	switch permTyp {
	case READ:
		p = fm.ReadPerm
	case CREATE:
		p = fm.CreatePerm
	case UPDATE:
		p = fm.UpdatePerm
	default:
		ok = false
	}
	return
}
