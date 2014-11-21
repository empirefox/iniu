package security

import (
	"time"

	"github.com/jinzhu/gorm"

	. "github.com/empirefox/iniu/base"
	. "github.com/empirefox/iniu/gorm/mod"
)

//New只输出不输入
type Form struct {
	Id              int64       `json:",omitempty"`
	Name            *string     `json:",omitempty" binding:"required" sql:"not null;type:varchar(64);unique"`
	Description     *string     `json:",omitempty"                    sql:"type:varchar(128)"`
	Pos             *int64      `json:",omitempty" binding:"required"`
	Title           *string     `json:",omitempty"                    sql:"type:varchar(64)"`
	Fields          []Field     `json:",omitempty"`
	Newid           *int64      `json:",omitempty"`
	New             interface{} `json:",omitempty"                    sql:"-"`
	AccessPerm      *string     `json:",omitempty"                    sql:"type:varchar(64)"`
	FormPerm        *string     `json:",omitempty"                    sql:"type:varchar(64)"`
	OnePerm         *string     `json:",omitempty"                    sql:"type:varchar(64)"`
	NamesPerm       *string     `json:",omitempty"                    sql:"type:varchar(64)"`
	PagePerm        *string     `json:",omitempty"                    sql:"type:varchar(64)"`
	QueryPerm       *string     `json:",omitempty"                    sql:"type:varchar(64)"`
	RecoveryPerm    *string     `json:",omitempty"                    sql:"type:varchar(64)"`
	AutoMigratePerm *string     `json:",omitempty"                    sql:"type:varchar(64)"`
	RemovePerm      *string     `json:",omitempty"                    sql:"type:varchar(64)"`
	UpdatePerm      *string     `json:",omitempty"                    sql:"type:varchar(64)"`
	CreatedAt       *time.Time  `json:",omitempty"`
	UpdatedAt       *time.Time  `json:",omitempty"`

	Context `json:"-" sql:"-"`
	Orderd  `json:"-" sql:"-"`
}

func (form *Form) AfterDelete(tx *gorm.DB) (err error) {
	err = tx.Where(Field{FormId: form.Id}).Delete(Field{}).Error
	return
}

//Ops由客户端使用时解析为json
type Field struct {
	Id          int64      `json:",omitempty"`
	FormId      int64      `json:",omitempty"`
	Name        *string    `json:",omitempty" binding:"required" sql:"not null;type:varchar(64)"`
	Description *string    `json:",omitempty"                    sql:"type:varchar(128)"`
	Pos         *int64     `json:",omitempty" binding:"required"`
	Title       *string    `json:",omitempty"                    sql:"type:varchar(64)"`
	Type        *string    `json:",omitempty"                    sql:"type:varchar(16)"`
	Placeholder *string    `json:",omitempty"                    sql:"type:varchar(32)"`
	Required    *bool      `json:",omitempty"`
	Readable    *bool      `json:",omitempty"`
	Readonly    *bool      `json:",omitempty"`
	Pattern     *string    `json:",omitempty"                    sql:"type:varchar(64)"`
	Minlength   *int       `json:",omitempty"`
	Maxlength   *int       `json:",omitempty"`
	Min         *string    `json:",omitempty"                    sql:"type:varchar(16)"`
	Max         *string    `json:",omitempty"                    sql:"type:varchar(16)"`
	Help        *string    `json:",omitempty"                    sql:"type:varchar(64)"`
	Ops         *string    `json:",omitempty"                    sql:"type:text"`
	CreatedAt   *time.Time `json:",omitempty"`
	UpdatedAt   *time.Time `json:",omitempty"`
	Perm        *string    `json:",omitempty"                    sql:"type:varchar(64)"`

	Context `json:"-" sql:"-"`
	Orderd  `json:"-" sql:"-" order:"form_id desc,pos desc"`
}

func init() {
	Register(Form{})
	Register(Field{})
}
