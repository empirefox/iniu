package security

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/jinzhu/gorm"

	. "github.com/empirefox/iniu/gorm/db"
	"github.com/empirefox/shirolet"
)

var (
	AccountNotFound   = errors.New("account not found")
	RefreshFormFailed = errors.New("refresh form failed")
)

func init() {
	gorm.DefaultCallback.Update().After("gorm:commit_or_rollback_transaction").Register("gorm:after_transaction", AfterTransaction)
}

func AfterTransaction(scope *gorm.Scope) {
	scope.CallMethod("AfterTransaction")
}

type Context struct {
}

func (c *Context) BeforeSave(scope *gorm.Scope) {
	accountI, hasAccount := scope.Get("context:account")
	if !hasAccount {
		scope.Err(AccountNotFound)
		return
	}

	account, isAccount := accountI.(*Account)
	if !isAccount {
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

// TODO maybe select record after every save?
func (c *Context) AfterTransaction(scope *gorm.Scope) error {
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

			var form = Form{}
			var err error

			vars := scope.SqlVars[:]
			defer func() {
				scope.SqlVars = vars
			}()
			scope.SqlVars = []interface{}{}
			newVars := vars[skip:]

			switch table {
			case "forms":
				sql := fmt.Sprintf(
					"SELECT * FROM %v %v",
					scope.QuotedTableName(),
					scope.CombinedConditionSql(),
				)
				err = DB.Raw(strings.Replace(sql, "$$", "?", -1), newVars...).Scan(&form).Error
			case "fields":
				s := scope.NewDB().Limit(1).NewScope(&Field{})
				s.Raw(fmt.Sprintf(
					"SELECT form_id FROM fields %v",
					scope.CombinedConditionSql(),
				))
				var formId int64
				err = s.DB().QueryRow(s.Sql, newVars...).Scan(&formId)
				if err != nil {
					return RefreshFormFailed
				}

				err = scope.NewDB().Where("id=?", formId).First(&form).Error
			}

			if err != nil {
				return RefreshFormFailed
			}
			InitForm(form)
		}
	}
	return nil
}
