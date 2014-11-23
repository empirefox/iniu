//GORM_DIALECT=mysql DB_URL="gorm:gorm@/gorm?charset=utf8&parseTime=True"
//GORM_DIALECT=postgres DB_URL="user=gorm dbname=gorm sslmode=disable"
//GORM_DIALECT=sqlite3 DB_URL=/tmp/gorm.DB go test
package security

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	. "github.com/empirefox/iniu/gorm/db"
	"github.com/empirefox/shirolet"
)

func TestContext(t *testing.T) {
	Convey("Context", t, func() {

		Convey("BeforeSave", func() {
			ColumnPerm = func(t Table, column string) shirolet.Permit {
				switch column {
				case "name":
					return shirolet.NewPermit("save:name")
				case "protected":
					return shirolet.NewPermit("save:protected")
				}
				return nil
			}
			account := &Account{
				Name:      "empirefox",
				Enabled:   true,
				HoldsPerm: "save:name",
			}
			exist := &Callback{
				Id:        1,
				Name:      "empirefox",
				Protected: "secret",
			}
			scope := DB.Set("context:account", account).NewScope(exist)
			context := Context{}
			context.BeforeSave(scope)

			//check
			for _, field := range scope.Fields() {
				switch field.DBName {
				case "name":
					So(field.IsIgnored, ShouldBeFalse)
				case "protected":
					So(field.IsIgnored, ShouldBeTrue)
				}
			}
		})

		Convey("AfterSave", func() {
			recoveryCallback()
			So(superDB.Save(&a).Error, ShouldBeNil)

			Convey("should find account", func() {
				aPtr, oPtr := FindAccount("google", "empirefox@sina.com")
				So(aPtr.Name, ShouldEqual, "empirefox")
				So(oPtr.Name, ShouldEqual, "empirefox@sina")
			})

			Convey("should not find account", func() {
				aPtr, oPtr := FindAccount("google", "empirefox@sina")
				So(aPtr.Name, ShouldEqual, "")
				So(oPtr.Name, ShouldEqual, "")
			})
		})
	})
}
