package security

import (
	"errors"
	"reflect"

	"github.com/jinzhu/gorm"

	. "github.com/empirefox/iniu/base"
	"github.com/empirefox/shirolet"
)

var (
	AccountNotFound = errors.New("account not found")
)

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
	if updateAttrs, found := scope.Get("gorm:update_attrs"); !found {
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

func (c *Context) AfterSave(scope *gorm.Scope) {
	if !scope.HasError() {
		switch scope.TableName() {
		case "forms":
			var dest = scope.IndirectValue()
			switch dest.Kind() {
			case reflect.Slice:
				ForEach(dest.Interface(), func(mPtr interface{}) error {
					InitForm(*(mPtr.(*Form)))
					return nil
				})
			case reflect.Ptr:
				InitForm(dest.Elem().Interface().(Form))
			case reflect.Struct:
				InitForm(dest.Interface().(Form))
			}
		case "fields":
			var dest = scope.IndirectValue()
			var field = dest.Interface()
			if dest.Kind() == reflect.Slice {
				ForEach(dest.Interface(), func(mPtr interface{}) error {
					field = mPtr
					return errors.New("")
				})
			}

			f := Form{}
			e := scope.NewDB().Model(field).Related(&f).Error
			if e != nil {
				panic(e)
			}
			InitForm(f)
		}
	}
}
