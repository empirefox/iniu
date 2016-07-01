package base

import (
	"github.com/empirefox/iniu/base/bo"
	"github.com/jinzhu/gorm"
)

type Form struct {
	gorm.Model
	bo.FormOut
	Fields     []FormField
	Name       string `sql:"type:varchar(64);not null;unique"`
	DeletePerm string `sql:"type:varchar(64)"`
}

type FormField struct {
	gorm.Model
	bo.FieldOut
	Name       string `sql:"type:varchar(64);not null"`
	FormID     uint   `sql:"not null"`
	CreatePerm string `sql:"type:varchar(64)"`
	ReadPerm   string `sql:"type:varchar(64)"`
	// TODO MUST ignore any children table
	UpdatePerm string `sql:"type:varchar(64)"`

	FieldMetaData
}

// FieldMetaData share data between FormField/FieldMeta
type FieldMetaData struct {
	Xfilter bool
}
