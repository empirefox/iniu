//GORM_DIALECT=mysql DB_URL="gorm:gorm@/gorm?charset=utf8&parseTime=True"
//GORM_DIALECT=postgres DB_URL="user=gorm dbname=gorm sslmode=disable" PORT=8080 goconvey
//GORM_DIALECT=sqlite3 DB_URL=/tmp/gorm.DB go test
package web

import (
	"net/http"
	"testing"

	"github.com/empirefox/shirolet"
	"github.com/empirefox/spy"
	"github.com/go-martini/martini"
	. "github.com/smartystreets/goconvey/convey"

	. "github.com/empirefox/iniu/base"
	. "github.com/empirefox/iniu/convey"
	"github.com/empirefox/iniu/security"
)

func TestTableForms(t *testing.T) {
	Convey("TableForms", t, func() {
		tableFormsMock := func() (tfs []TableForm, err error) {
			tfs = []TableForm{{Name: "Form", Title: "Form Title"}, {Name: "Field", Title: "Field Title"}}
			return
		}

		webFormMock := func(t Table, method string) shirolet.Permit {
			switch t.(string) + "." + method + "()" {
			case "Form.Form()":
				return shirolet.NewPermit("edit:form:Form")
			case "Field.Form()":
				return shirolet.NewPermit("edit:form:Field")
			}
			return nil
		}

		req := func(m martini.Router) (*http.Request, error) {
			m.Get("/forms", func(c martini.Context) {
				c.Map(&security.Account{HoldsPerm: "edit:form:Field"})
			}, TableForms)
			return http.NewRequest("GET", "/forms", nil)
		}
		res := []TableForm{{Name: "Field", Title: "Field Title"}}

		Convey("should get authorized tables", spy.On(&tableForms, tableFormsMock).On(&security.WebPerm, webFormMock).Spy(func() {
			So(req, ShouldResponseOk, res)
		}))
	})
}
