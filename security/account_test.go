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

func recoveryAccount() {
	DB.DropTableIfExists(&Oauth{})
	DB.DropTableIfExists(&Account{})
	DB.CreateTable(&Account{})
	DB.CreateTable(&Oauth{})
}

func TestAccount(t *testing.T) {
	Convey("Account", t, func() {
		a := Account{
			Name:      "empirefox",
			Enabled:   true,
			HoldsPerm: "a:b",
		}

		os := []Oauth{
			{
				Oid:      "empirefox@sina.com",
				Provider: "google",
				Name:     "empirefox@sina",
			},
		}

		a.Oauths = os

		Convey("Permitted", func() {
			pfalse := shirolet.NewPermit("a")
			ptrue := shirolet.NewPermit("a:b")
			So(a.Permitted(pfalse), ShouldBeFalse)
			So(a.Permitted(ptrue), ShouldBeTrue)
		})

		Convey("FindAccount", func() {
			recoveryAccount()
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
