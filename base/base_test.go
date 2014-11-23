//需设置环境变量
//GORM_DIALECT=mysql DB_URL="gorm:gorm@/gorm?charset=utf8&parseTime=True"
//GORM_DIALECT=postgres DB_URL="user=gorm dbname=gorm sslmode=disable"
//GORM_DIALECT=sqlite3 DB_URL=/tmp/gorm.DB go test
package base

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	. "github.com/empirefox/iniu/gorm/db"
	. "github.com/empirefox/iniu/gorm/mod"
)

type Xchg struct {
	Id  int64 `order:"auto"`
	Pos int64
}

func init() {
	AutoOrder()
	Register(Xchg{})
}

var (
	xchgs = "xchgs"
)

// newTable(2, 3, 4, 5, 6, 7, 8, 9)
func newTable(ps ...int64) {
	DB.DropTableIfExists(Xchg{})
	DB.CreateTable(Xchg{})

	for _, p := range ps {
		DB.Save(&Xchg{Pos: p})
	}
}

func TestModel(t *testing.T) {
	Convey("测试Model", t, func() {
		Convey("should Register zero struct", func() {
			var s []Xchg
			smade := []Xchg{}
			formMap := map[string]interface{}{
				"Name":   "Xchg",
				"Fields": []nameOnlyField{{"Id"}, {"Pos"}},
			}

			So(New(xchgs), ShouldResemble, &Xchg{})
			So(Slice(xchgs), ShouldResemble, &s)
			So(MakeSlice(xchgs), ShouldResemble, &smade)
			So(IndirectSlice(xchgs), ShouldResemble, smade)
			mf, _ := ModelFormMetas(xchgs)
			So(mf, ShouldResemble, formMap)
			So(HasPos(xchgs), ShouldBeTrue)
			So(Example(xchgs), ShouldBeNil)
		})

		Convey("should Register non-zero struct", func() {
			Register(Xchg{Pos: 100})
			var s []Xchg
			smade := []Xchg{}
			formMap := map[string]interface{}{
				"Name":   "Xchg",
				"Fields": []nameOnlyField{{"Id"}, {"Pos"}},
				"New":    Xchg{Pos: 100},
			}

			So(New(xchgs), ShouldResemble, &Xchg{})
			So(Slice(xchgs), ShouldResemble, &s)
			So(MakeSlice(xchgs), ShouldResemble, &smade)
			So(IndirectSlice(xchgs), ShouldResemble, smade)
			mf, _ := ModelFormMetas(xchgs)
			So(mf, ShouldResemble, formMap)
			So(HasPos(xchgs), ShouldBeTrue)
			So(Example(xchgs), ShouldResemble, Xchg{Pos: 100})
			Register(Xchg{})
		})

		Convey("SaveModel", func() {
			newTable(2)
			param := map[string]interface{}{"Pos": int64(100)}
			r := &Xchg{}
			Convey("should save new model ok when no params", func() {
				var m Model = Xchg{Pos: 5}
				So(SaveModel(&m), ShouldBeNil)
				DB.Table(xchgs).Where("Pos=?", 5).First(r)
				So(m.(Xchg), ShouldResemble, *r)
			})

			Convey("should save exist model ok when no params", func() {
				var m Model = Xchg{Id: 1, Pos: 5}
				So(SaveModel(&m), ShouldBeNil)
				DB.Table(xchgs).Where("Pos=?", 5).First(r)
				So(m.(Xchg), ShouldResemble, *r)
			})

			Convey("should save new model ok with params", func() {
				var m Model = Xchg{Pos: 5}
				So(SaveModelWith(&m, param), ShouldBeNil)
				DB.Table(xchgs).Where("Pos=?", 100).First(r)
				So(m.(Xchg), ShouldResemble, *r)
			})

			Convey("should save exist model ok with params", func() {
				var m Model = Xchg{Id: 1, Pos: 5}
				So(SaveModelWith(&m, param), ShouldBeNil)
				DB.Table(xchgs).Where("Pos=?", 100).First(r)
				So(m.(Xchg), ShouldResemble, *r)
			})
		})
	})

	Convey("测试Models", t, func() {
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
