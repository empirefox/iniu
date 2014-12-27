package security

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/golang/glog"
	"github.com/jinzhu/gorm"

	. "github.com/empirefox/iniu/gorm/db"
	"github.com/empirefox/shirolet"
)

var (
	AccountNotFound   = errors.New("account not found")
	RefreshFormFailed = errors.New("refresh form failed")
)

func init() {
	gorm.DefaultCallback.Update().After("gorm:commit_or_rollback_transaction").Register("gorm:after_transaction", AfterUpdateTransaction)
	gorm.DefaultCallback.Create().After("gorm:commit_or_rollback_transaction").Register("gorm:after_transaction", AfterCreateTransaction)
}

func AfterUpdateTransaction(scope *gorm.Scope) {
	scope.CallMethod("AfterUpdateTransaction")
}

func AfterCreateTransaction(scope *gorm.Scope) {
	scope.CallMethod("AfterCreateTransaction")
}

type Context struct {
}

// add string perm to temp create an Account, specially useful in test case
func (c *Context) BeforeSave(scope *gorm.Scope) {
	accountI, hasAccount := scope.Get("context:account")
	if !hasAccount {
		scope.Err(AccountNotFound)
		return
	}

	var account *Account
	switch accountI.(type) {
	case *Account:
		account = accountI.(*Account)
	case string:
		account = &Account{
			Name:      "string_perm",
			Enabled:   true,
			HoldsPerm: accountI.(string),
		}
	default:
		scope.Err(AccountNotFound)
		return
	}

	table := scope.TableName()
	if updateAttrs, found := scope.InstanceGet("gorm:update_attrs"); !found {
		for key, field := range scope.Fields() {
			n := field.Name
			if n == "CreatedAt" || !account.Permitted(ColumnPerm(table, key)) {
				field.IsIgnored = true
			} else if permNameReg.MatchString(n) {
				if r, ok := reflect.Indirect(field.Field).Interface().(string); ok {
					field.Set(shirolet.Fmt(r))
				}
			}
		}
	} else {
		uas := updateAttrs.(map[string]interface{})
		for key, value := range uas {
			if !account.Permitted(ColumnPerm(table, key)) {
				delete(uas, key)
			} else if permNameReg.MatchString(key) {
				if v, ok := value.(string); ok {
					uas[key] = shirolet.Fmt(v)
				}
			}
		}
	}
}

func (c *Context) AfterCreateTransaction(scope *gorm.Scope) error {
	if !scope.HasError() {
		table := scope.TableName()
		if table == "forms" || table == "fields" {
			var where string
			switch table {
			case "forms":
				where = "id=?"
			case "fields":
				where = "id = (SELECT form_id FROM fields WHERE id = ?)"
			}

			form := Form{}
			err := DB.Where(where, scope.PrimaryKeyValue()).First(&form).Error
			if err != nil {
				glog.Errorln(err)
				return RefreshFormFailed
			}

			InitForm(form)
		}
	}
	return nil
}

// TODO maybe select record after every save?
func (c *Context) AfterUpdateTransaction(scope *gorm.Scope) error {
	if !scope.HasError() {
		table := scope.TableName()
		if table == "forms" || table == "fields" {
			skip := 0
			if updateAttrs, ok := scope.InstanceGet("gorm:update_attrs"); ok {
				skip = len(updateAttrs.(map[string]interface{}))
			} else {
				for _, field := range scope.Fields() {
					if !field.IsPrimaryKey && field.IsNormal && !field.IsIgnored {
						if field.DefaultValue != nil && field.IsBlank {
							continue
						}
						skip++
					}
				}
			}

			vars := scope.SqlVars[:]
			defer func() {
				scope.SqlVars = vars
			}()
			scope.SqlVars = []interface{}{}
			newVars := vars[skip:]

			var sql string
			switch table {
			case "forms":
				sql = fmt.Sprintf(
					"SELECT * FROM %v %v",
					scope.QuotedTableName(),
					scope.CombinedConditionSql(),
				)
			case "fields":
				sql = fmt.Sprintf(
					"SELECT * FROM forms WHERE id = (SELECT form_id FROM %v %v)",
					scope.QuotedTableName(),
					scope.CombinedConditionSql(),
				)
			}

			form := Form{}
			err := DB.Raw(strings.Replace(sql, "$$", "?", -1), newVars...).Scan(&form).Error
			if err != nil {
				glog.Errorln(err)
				return RefreshFormFailed
			}
			InitForm(form)
		}
	}
	return nil
}
