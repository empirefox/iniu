package security

import (
	"time"

	"github.com/golang/glog"
	"github.com/jinzhu/gorm"

	. "github.com/empirefox/iniu/base"
	. "github.com/empirefox/iniu/gorm/db"
	"github.com/empirefox/shirolet"
)

var (
	// exports
	FindAccount = findAccount
	LegalOauth  = legalOauth
)

//Provider:Google,Github,Qq,Weibo,Baidu,Souhu,Netease,Douban
type Oauth struct {
	Id          int64     `json:",omitempty"`
	AccountId   int64     `json:",omitempty"                    sql:"not null"`
	Oid         string    `json:",omitempty" binding:"required" sql:"type:varchar(128);not null"`
	Provider    string    `json:",omitempty" binding:"required" sql:"type:varchar(32);not null"`
	Name        string    `json:",omitempty" binding:"required" sql:"type:varchar(128);not null"`
	Description string    `json:",omitempty"                    sql:"type:varchar(128);default:''"`
	Validated   bool      `json:",omitempty"                    sql:"default:false"`
	Enabled     bool      `json:",omitempty"                    sql:"default:false"`
	LogedAt     time.Time `json:",omitempty"                    sql:"default:CURRENT_TIMESTAMP"`
	CreatedAt   time.Time `json:",omitempty"                    sql:"default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `json:",omitempty"                    sql:"default:CURRENT_TIMESTAMP"`

	Context `json:"-" sql:"-"`
}

type Account struct {
	Id          int64     `json:",omitempty"`
	Name        string    `json:",omitempty" binding:"required" sql:"not null;type:varchar(128);unique"`
	Description string    `json:",omitempty"                    sql:"type:varchar(128);default:''"`
	Oauths      []Oauth   `json:",omitempty"`
	Enabled     bool      `json:",omitempty"                    sql:"default:false"`
	CreatedAt   time.Time `json:",omitempty"                    sql:"default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `json:",omitempty"                    sql:"default:CURRENT_TIMESTAMP"`
	HoldsPerm   string    `json:",omitempty"                    sql:"type:varchar(128);default:''"`

	Holds   shirolet.Holds `json:"-" sql:"-"`
	Context `json:"-" sql:"-"`
}

// MUST pass *Account into delete method to triger the callback
func (a *Account) AfterDelete(tx *gorm.DB) (err error) {
	err = tx.Where(Oauth{AccountId: a.Id}).Delete(Oauth{}).Error
	return
}

func (a *Account) Permitted(p shirolet.Permit) bool {
	if p == nil {
		return true
	}
	if a.Holds == nil {
		a.Holds = shirolet.NewHolds(a.HoldsPerm)
	}
	return p.SatisfiedBy(a.Holds)
}

// a current Account, c current Oauth logged
func findAccount(provider, oid string) (*Account, *Oauth) {
	var a Account
	var c Oauth
	if err := DB.Where(Oauth{Provider: provider, Oid: oid}).First(&c).Error; err != nil {
		glog.Infoln(err)
		return &a, &c
	}
	if err := DB.Where(c.AccountId).First(&a).Error; err != nil {
		glog.Infoln(err)
	}
	return &a, &c
}

func legalOauth(a *Account, c *Oauth) bool {
	return a.Enabled && c.Enabled && c.Validated
}
func init() {
	Register(Oauth{})
	Register(Account{})
}
