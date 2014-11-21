package security

import (
	. "github.com/empirefox/iniu/base"
	. "github.com/jinzhu/copier"
)

type JsonForm struct {
	Name        *string     `json:",omitempty" binding:"required" sql:"not null;type:varchar(64);unique"`
	Description *string     `json:",omitempty"                    sql:"type:varchar(128)"`
	Pos         *int64      `json:",omitempty" binding:"required"`
	Title       *string     `json:",omitempty"                    sql:"type:varchar(64)"`
	JsonFields  []JsonField `json:"Fields,omitempty"`
	New         interface{} `json:",omitempty"                    sql:"-"`
}

func (jform *JsonForm) Fields(fields []Field) {
	Copy(&jform.JsonFields, &fields)
}

type JsonField struct {
	Name        *string `json:",omitempty" binding:"required" sql:"not null;type:varchar(64)"`
	Description *string `json:",omitempty"                    sql:"type:varchar(128)"`
	Pos         *int64  `json:",omitempty" binding:"required"`
	Title       *string `json:",omitempty"                    sql:"type:varchar(64)"`
	Type        *string `json:",omitempty"                    sql:"type:varchar(16)"`
	Placeholder *string `json:",omitempty"                    sql:"type:varchar(32)"`
	Required    *bool   `json:",omitempty"`
	Readable    *bool   `json:",omitempty"`
	Readonly    *bool   `json:",omitempty"`
	Pattern     *string `json:",omitempty"                    sql:"type:varchar(64)"`
	Minlength   *int    `json:",omitempty"`
	Maxlength   *int    `json:",omitempty"`
	Min         *string `json:",omitempty"                    sql:"type:varchar(16)"`
	Max         *string `json:",omitempty"                    sql:"type:varchar(16)"`
	Help        *string `json:",omitempty"                    sql:"type:varchar(64)"`
	Ops         *string `json:",omitempty"                    sql:"type:text"`
}
