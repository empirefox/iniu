package security

import (
	"errors"
	"reflect"

	"github.com/jinzhu/gorm"

	. "github.com/empirefox/iniu/base"
	"github.com/empirefox/shirolet"
)

type Context struct {
	Account *Account
	Form    string
}

func (c *Context) BeforeSave(scope *gorm.Scope) {
	if updateAttrs, found := scope.Get("gorm:update_attrs"); !found {
		for _, field := range scope.Fields() {
			n := field.Name
			if !Skip && n != "CreatedAt" && n != "UpdatedAt" && !c.Account.Permitted(FieldPerm(c.Form, n)) {
				field.IsIgnored = true
			} else if permNameReg.MatchString(n) {

				if rv := reflect.Indirect(field.Field); rv.Kind() == reflect.String {
					v := rv.Interface().(string)
					field.Set(shirolet.Fmt(v))
				}
			}
		}
		c.Account = &Account{}
	} else {
		uas := updateAttrs.(map[string]interface{})
		for key, value := range uas {
			if permKeyReg.MatchString(key) {
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
			if dest.Kind() == reflect.Slice {
				ForEach(dest.Interface(), func(mPtr interface{}) error {
					InitForm(mPtr.(*Form))
					return nil
				})
			} else {
				InitForm(dest.Interface().(*Form))
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

			f := &Form{}
			e := scope.NewDB().Model(field).Related(f).Error
			if e != nil {
				panic(e)
			}
			InitForm(f)
		}
	}
}
