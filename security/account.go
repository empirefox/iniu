package security

import (
	"time"

	"github.com/jinzhu/gorm"

	. "github.com/empirefox/iniu/base"
	. "github.com/empirefox/iniu/gorm/postgres"
	"github.com/empirefox/shirolet"
)

//Provider:Google,Github,Qq,Weibo,Baidu,Souhu,Netease,Douban
type Oauth struct {
	Id          int64      `json:",omitempty"`
	AccountId   int64      `json:",omitempty"`
	Oid         *string    `json:",omitempty" binding:"required" sql:"not null;type:varchar(128)"`
	Provider    *string    `json:",omitempty" binding:"required" sql:"not null;type:varchar(32)"`
	Name        *string    `json:",omitempty" binding:"required" sql:"not null;type:varchar(128)"`
	Description *string    `json:",omitempty"                    sql:"type:varchar(128)"`
	Validated   *bool      `json:",omitempty"`
	Enabled     *bool      `json:",omitempty"`
	LogedAt     *time.Time `json:",omitempty"`
	CreatedAt   *time.Time `json:",omitempty"`
	UpdatedAt   *time.Time `json:",omitempty"`

	Context `json:"-" sql:"-"`
}

type Account struct {
	Id          int64      `json:",omitempty"`
	Name        *string    `json:",omitempty" binding:"required" sql:"not null;type:varchar(128);unique"`
	Description *string    `json:",omitempty"                    sql:"type:varchar(128)"`
	Oauths      []Oauth    `json:",omitempty"`
	Enabled     *bool      `json:",omitempty"`
	CreatedAt   *time.Time `json:",omitempty"`
	UpdatedAt   *time.Time `json:",omitempty"`
	HoldsPerm   *string    `json:",omitempty"`

	Holds   shirolet.Holds `json:"-" sql:"-"`
	Context `json:"-" sql:"-"`
}

func (a *Account) AfterDelete(tx *gorm.DB) (err error) {
	err = tx.Where(Oauth{AccountId: a.Id}).Delete(Oauth{}).Error
	return
}

func (a *Account) Permitted(p shirolet.Permit) bool {
	if p == nil {
		return true
	}
	if a.Holds == nil {
		a.Holds = shirolet.NewHolds(*a.HoldsPerm)
	}
	return p.SatisfiedBy(a.Holds)
}

func FindAccount(provider, oid string) (a *Account, c *Oauth) {
	DB.Where(&Oauth{Provider: &provider, Oid: &oid}).First(c).Related(a)
	return
}

func LegalOauth(a *Account, c *Oauth) bool {
	return *a.Enabled && *c.Enabled && *c.Validated
}
func init() {
	Register(Oauth{})
	Register(Account{})
}
