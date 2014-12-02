//GORM_DIALECT=mysql DB_URL="gorm:gorm@/gorm?charset=utf8&parseTime=True"
//GORM_DIALECT=postgres DB_URL="user=gorm dbname=gorm sslmode=disable"
//GORM_DIALECT=sqlite3 DB_URL=/tmp/gorm.DB go test
package security

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	. "github.com/empirefox/iniu/base"
	. "github.com/empirefox/iniu/gorm/db"
	"github.com/empirefox/shirolet"
)

type Callback struct {
	Id        int64
	Name      string
	Protected string
}

func TestContextBeforeSave(t *testing.T) {
	Convey("BeforeSave", t, func() {
		mockColumnPerm := func(t Table, column string) shirolet.Permit {
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

		Convey("should check struct", func() {
			ColumnPerm = mockColumnPerm
			defer func() {
				ColumnPerm = columnPerm
			}()
			exist := &Callback{
				Id:        1,
				Name:      "empirefox",
				Protected: "secret",
			}
			db := DB.Set("context:account", account)
			scope := db.NewScope(exist)
			context := Context{}
			context.BeforeSave(scope)

			So(db.Error, ShouldBeNil)

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

		Convey("should check map", func() {
			ColumnPerm = mockColumnPerm
			defer func() {
				ColumnPerm = columnPerm
			}()
			exist := map[string]interface{}{
				"name":      "empirefox",
				"protected": "secret",
			}
			scope := DB.Set("context:account", account).NewScope(&Callback{}).InstanceSet("gorm:update_attrs", exist)
			context := Context{}
			context.BeforeSave(scope)

			//check
			updateAttrs, found := scope.InstanceGet("gorm:update_attrs")
			So(found, ShouldBeTrue)
			filterd := false
			for key := range updateAttrs.(map[string]interface{}) {
				So(key, ShouldNotEqual, "protected")
				if key == "name" {
					filterd = true
				}
			}
			So(filterd, ShouldBeTrue)
		})
	})
}

func getSimpleForm() Form {
	simpleForm := Form{
		Name:       "SimpleForm",
		Pos:        9,
		Newid:      1,
		RemovePerm: "sys:remove:sf",
	}
	simpleFields := []Field{
		{
			Name: "field1",
			Pos:  1,
			Perm: "perm1",
		},
	}
	simpleForm.Fields = simpleFields
	return simpleForm
}

func TestContextAfterTransaction1(t *testing.T) {
	Convey("AfterTransaction", t, func() {
		simpleForm := getSimpleForm()
		Convey("should not affect when it is not form or field", func() {

			recoveryForm()
			So(superDB.Save(&simpleForm).Error, ShouldBeNil)
			InitForms()

			form := Form{}
			err := DB.Raw(`SELECT * FROM "forms" WHERE ("id" = $1)`, "1").Scan(&form).Error
			So(err, ShouldBeNil)
			a := Account{
				Name:      "empirefox",
				Enabled:   true,
				HoldsPerm: "a:b",
			}
			recoveryAccount()
			So(superDB.Save(&a).Error, ShouldBeNil)

			jf, ok := JsonForms[simpleFormTable]
			So(ok, ShouldBeTrue)
			So(jf.Name, ShouldEqual, "SimpleForm")
		})
	})
}

func TestContextAfterTransaction2(t *testing.T) {
	Convey("AfterTransaction", t, func() {
		simpleForm := getSimpleForm()
		Convey("should refresh form when it is form", func() {

			recoveryForm()
			So(superDB.Save(&simpleForm).Error, ShouldBeNil)
			InitForms()

			FormTableMap["FixedName"] = "FixedName"
			simpleForm.Name = "FixedName"
			So(superDB.Save(&simpleForm).Error, ShouldBeNil)

			jf, ok := JsonForms["FixedName"]
			So(ok, ShouldBeTrue)
			So(jf.Name, ShouldEqual, "FixedName")
		})
	})
}

func TestContextAfterTransaction3(t *testing.T) {
	Convey("AfterTransaction", t, func() {
		simpleForm := getSimpleForm()
		Convey("should refresh form when it is field", func() {

			recoveryForm()
			So(superDB.Save(&simpleForm).Error, ShouldBeNil)
			InitForms()

			field := Field{}
			So(superDB.First(&field).Error, ShouldBeNil)

			p := columnPerm(simpleFormTable, "field1")
			So(p, ShouldResemble, shirolet.NewPermit("perm1"))

			field.Perm = "perm2"
			So(superDB.Save(&field).Error, ShouldBeNil)

			p = columnPerm(simpleFormTable, "field1")
			So(p, ShouldResemble, shirolet.NewPermit("perm2"))
		})
	})
}
