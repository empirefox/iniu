package security

import (
	"time"

	"github.com/jinzhu/gorm"

	. "github.com/empirefox/iniu/base"
)

//New只输出不输入
type Form struct {
	Id              int64       `json:",omitempty" order:"auto"`
	Name            string      `json:",omitempty" binding:"required" sql:"type:varchar(64);not null;unique"`
	Description     string      `json:",omitempty"                    sql:"type:varchar(128);default:''"`
	Pos             int64       `json:",omitempty" binding:"required" sql:"not null"`
	Title           string      `json:",omitempty"                    sql:"type:varchar(64);default:''"`
	Fields          []Field     `json:",omitempty"`
	Newid           int64       `json:",omitempty"                    sql:"default:0"`
	New             interface{} `json:",omitempty"                    sql:"-"`
	AccessPerm      string      `json:",omitempty"                    sql:"type:varchar(64);default:''"`
	FormPerm        string      `json:",omitempty"                    sql:"type:varchar(64);default:''"`
	OnePerm         string      `json:",omitempty"                    sql:"type:varchar(64);default:''"`
	NamesPerm       string      `json:",omitempty"                    sql:"type:varchar(64);default:''"`
	PagePerm        string      `json:",omitempty"                    sql:"type:varchar(64);default:''"`
	QueryPerm       string      `json:",omitempty"                    sql:"type:varchar(64);default:''"`
	RecoveryPerm    string      `json:",omitempty"                    sql:"type:varchar(64);default:''"`
	AutoMigratePerm string      `json:",omitempty"                    sql:"type:varchar(64);default:''"`
	RemovePerm      string      `json:",omitempty"                    sql:"type:varchar(64);default:''"`
	UpdatePerm      string      `json:",omitempty"                    sql:"type:varchar(64);default:''"`
	CreatedAt       time.Time   `json:",omitempty"                    sql:"default:CURRENT_TIMESTAMP"`
	UpdatedAt       time.Time   `json:",omitempty"                    sql:"default:CURRENT_TIMESTAMP"`

	Context `json:"-" sql:"-"`
}

func (form *Form) AfterDelete(tx *gorm.DB) (err error) {
	err = tx.Where(Field{FormId: form.Id}).Delete(Field{}).Error
	return
}

//Ops由客户端使用时解析为json
type Field struct {
	Id          int64     `json:",omitempty" order:"form_id desc,pos desc"`
	FormId      int64     `json:",omitempty"                    sql:"not null"`
	Name        string    `json:",omitempty" binding:"required" sql:"type:varchar(64)";not null`
	Description string    `json:",omitempty"                    sql:"type:varchar(128);default:''"`
	Pos         int64     `json:",omitempty" binding:"required" sql:"not null"`
	Title       string    `json:",omitempty"                    sql:"type:varchar(64);default:''"`
	Type        string    `json:",omitempty"                    sql:"type:varchar(16);default:''"`
	Placeholder string    `json:",omitempty"                    sql:"type:varchar(32);default:''"`
	Required    bool      `json:",omitempty"                    sql:"default:false"`
	Readable    bool      `json:",omitempty"                    sql:"default:false"`
	Readonly    bool      `json:",omitempty"                    sql:"default:false"`
	Pattern     string    `json:",omitempty"                    sql:"type:varchar(64);default:''"`
	Minlength   int       `json:",omitempty"                    sql:"default:0"`
	Maxlength   int       `json:",omitempty"                    sql:"default:0"`
	Min         string    `json:",omitempty"                    sql:"type:varchar(16);default:''"`
	Max         string    `json:",omitempty"                    sql:"type:varchar(16);default:''"`
	Help        string    `json:",omitempty"                    sql:"type:varchar(64);default:''"`
	Ops         string    `json:",omitempty"                    sql:"type:text;default:''"`
	CreatedAt   time.Time `json:",omitempty"                    sql:"default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `json:",omitempty"                    sql:"default:CURRENT_TIMESTAMP"`
	Perm        string    `json:",omitempty"                    sql:"type:varchar(64);default:''"`

	Context `json:"-" sql:"-"`
}

func init() {
	Register(Form{})
	Register(Field{})
}
