//GORM_DIALECT=mysql DB_URL="gorm:gorm@/gorm?charset=utf8&parseTime=True"
//GORM_DIALECT=postgres DB_URL="user=gorm dbname=gorm sslmode=disable"
//GORM_DIALECT=sqlite3 DB_URL=/tmp/gorm.DB go test
package base

import (
	"errors"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/empirefox/iniu/db"
)

type Xchg struct {
	Id  int64
	Pos int64
}

type XchgDefault struct {
	Id  int64
	Pos int64 `default:"100"`
}

func init() {
	db.ConnectIfNot()
	Register(Xchg{})
}

var (
	xchgs        = "xchgs"
	xchgDefaults = "xchg_defaults"
)

// newTable(2, 3, 4, 5, 6, 7, 8, 9)
func newTable(ps ...int64) {
	db.DB.DropTableIfExists(Xchg{})
	db.DB.CreateTable(Xchg{})

	for _, p := range ps {
		db.DB.Save(&Xchg{Pos: p})
	}
}

func TestModel(t *testing.T) {
	Convey("Model", t, func() {
		Convey("should Register zero struct", func() {
			var s []Xchg
			smade := []Xchg{}
			formMap := map[string]interface{}{
				"Name":   "Xchg",
				"Fields": []NameOnlyField{{"Id"}, {"Pos"}},
			}

			So(New(xchgs), ShouldResemble, &Xchg{})
			So(Slice(xchgs), ShouldResemble, &s)
			So(MakeSlice(xchgs), ShouldResemble, &smade)
			So(IndirectSlice(xchgs), ShouldResemble, smade)
			mf, _ := ModelFormMetas(xchgs)
			So(mf, ShouldResemble, formMap)
			_, hasPos := GetSorting(xchgs)
			So(hasPos, ShouldBeTrue)
			So(Example(xchgs), ShouldBeNil)
		})

		Convey("should Register default value struct", func() {
			Register(XchgDefault{})
			var s []XchgDefault
			smade := []XchgDefault{}
			formMap := map[string]interface{}{
				"Name":   "XchgDefault",
				"Fields": []NameOnlyField{{"Id"}, {"Pos"}},
				"New":    &XchgDefault{Pos: 100},
			}

			So(New(xchgDefaults), ShouldResemble, &XchgDefault{})
			So(Slice(xchgDefaults), ShouldResemble, &s)
			So(MakeSlice(xchgDefaults), ShouldResemble, &smade)
			So(IndirectSlice(xchgDefaults), ShouldResemble, smade)
			mf, _ := ModelFormMetas(xchgDefaults)
			So(mf, ShouldResemble, formMap)
			_, hasPos := GetSorting(xchgDefaults)
			So(hasPos, ShouldBeTrue)
			So(Example(xchgDefaults), ShouldResemble, &XchgDefault{Pos: 100})
			Register(XchgDefault{})
		})

		Convey("SaveModel", func() {
			newTable(2)
			db := db.DB.Set("context:account", "*").Table(xchgs)
			param := map[string]interface{}{"Pos": int64(100)}
			r := &Xchg{}
			Convey("should save new model ok when no params", func() {
				var m Model = Xchg{Pos: 5}
				So(SaveModel(db, &m), ShouldBeNil)
				db.Where("Pos=?", 5).First(r)
				So(m.(Xchg), ShouldResemble, *r)
			})

			Convey("should save exist model ok when no params", func() {
				var m Model = Xchg{Id: 1, Pos: 5}
				So(SaveModel(db, &m), ShouldBeNil)
				db.Where("Pos=?", 5).First(r)
				So(m.(Xchg), ShouldResemble, *r)
			})

			Convey("should save new model ok with params", func() {
				var m Model = Xchg{Pos: 5}
				So(SaveModelWith(db, &m, param), ShouldBeNil)
				db.Where("Pos=?", 100).First(r)
				So(m.(Xchg), ShouldResemble, *r)
			})

			Convey("should save exist model ok with params", func() {
				var m Model = Xchg{Id: 1, Pos: 5}
				So(SaveModelWith(db, &m, param), ShouldBeNil)
				db.Where("Pos=?", 100).First(r)
				So(m.(Xchg), ShouldResemble, *r)
			})
		})
	})

	Convey("Models", t, func() {
		Convey("ForEach", func() {
			var ms Models = []Xchg{{Pos: 1}, {Pos: 2}, {Pos: 3}, {Pos: 5}, {Pos: 6}}
			Convey("should have nil error", func() {
				handleFunc := func(iPtr interface{}) error {
					if iPtr.(*Xchg).Pos == 4 {
						return errors.New("")
					}
					return nil
				}
				err := ForEach(ms, handleFunc)
				So(err, ShouldBeNil)
			})

			Convey("should have an error", func() {
				handleFunc := func(iPtr interface{}) error {
					if iPtr.(*Xchg).Pos == 5 {
						return errors.New("")
					}
					return nil
				}
				err := ForEach(ms, handleFunc)
				So(err, ShouldNotBeNil)
			})
		})
	})
}
