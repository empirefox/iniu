//GORM_DIALECT=mysql DB_URL="gorm:gorm@/gorm?charset=utf8&parseTime=True"
//GORM_DIALECT=postgres DB_URL="user=gorm dbname=gorm sslmode=disable"
//GORM_DIALECT=sqlite3 DB_URL=/tmp/gorm.DB go test
package security

import (
	"testing"

	"github.com/jinzhu/gorm"
	. "github.com/smartystreets/goconvey/convey"

	. "github.com/empirefox/iniu/gorm/db"
)

func recoveryForm() {
	DB.DropTableIfExists(&Field{})
	DB.DropTableIfExists(&Form{})
	DB.CreateTable(&Form{})
	DB.CreateTable(&Field{})
}

func TestFormAfterDelete(t *testing.T) {
	Convey("Form.AfterDelete", t, func() {
		simpleForm := Form{
			Name:  "simpleForm",
			Pos:   9,
			Newid: 1,
		}

		simpleFields := []Field{
			{
				Name: "field1",
				Pos:  1,
			},
			{
				Name: "field2",
				Pos:  2,
			},
		}

		simpleForm.Fields = simpleFields

		recoveryForm()

		Convey("should cascade delete fields", func() {
			err := superDB.Save(&simpleForm).Error
			So(err, ShouldBeNil)
			f := Form{}
			So(DB.Where("name=?", "simpleForm").Find(&f).RecordNotFound(), ShouldBeFalse)
			fs1 := []Field{}
			So(DB.Find(&fs1).Error, ShouldBeNil)
			So(len(fs1), ShouldEqual, 2)

			// must use &Form{}
			So(DB.Where("id=?", f.Id).Delete(&Form{}).Error, ShouldBeNil)

			f2 := Form{}
			So(DB.Where("name=?", "simpleForm").Find(&f2).RecordNotFound(), ShouldBeTrue)

			fs2 := []Field{}
			err = DB.Find(&fs2).Error
			So(err, ShouldEqual, gorm.RecordNotFound)
			So(len(fs2), ShouldEqual, 0)
		})
	})
}
