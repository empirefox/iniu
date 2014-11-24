//GORM_DIALECT=mysql DB_URL="gorm:gorm@/gorm?charset=utf8&parseTime=True"
//GORM_DIALECT=postgres DB_URL="user=gorm dbname=gorm sslmode=disable"
//GORM_DIALECT=sqlite3 DB_URL=/tmp/gorm.DB go test
package security

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/empirefox/iniu/base"
	"github.com/empirefox/shirolet"
)

var simpleFormTable = "simple_forms"

func init() {
	base.FormTableMap["SimpleForm"] = simpleFormTable
}

func TestJsonForms(t *testing.T) {
	Convey("JsonForms", t, func() {
		simpleForm := Form{
			Name:       "SimpleForm",
			Pos:        9,
			Newid:      1,
			RemovePerm: "sys:remove:sf",
		}

		simpleFields := []Field{
			{
				Name: "field",
				Pos:  1,
				Perm: "sys:save:field1",
			},
			{
				Name: "field2",
				Pos:  2,
			},
		}

		simpleForm.Fields = simpleFields

		recoveryForm()

		err := superDB.Save(&simpleForm).Error
		So(err, ShouldBeNil)

		Convey("should init ok", func() {
			InitForms()

			jf, ok := JsonForms[simpleFormTable]
			So(ok, ShouldBeTrue)
			So(jf.Name, ShouldEqual, "SimpleForm")

			removePerm := WebPerm(simpleFormTable, "Remove")
			So(removePerm, ShouldResemble, shirolet.NewPermit("sys:remove:sf"))

			fieldPerm := ColumnPerm(simpleFormTable, "field")
			So(fieldPerm, ShouldResemble, shirolet.NewPermit("sys:save:field1"))
		})
	})
}
