//GORM_DIALECT=mysql DB_URL="gorm:gorm@/gorm?charset=utf8&parseTime=True"
//GORM_DIALECT=postgres DB_URL="user=gorm dbname=gorm sslmode=disable" PORT=8080 goconvey
//GORM_DIALECT=sqlite3 DB_URL=/tmp/gorm.DB go test
package web

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-martini/martini"
	"github.com/jinzhu/gorm"
	"github.com/martini-contrib/render"
	. "github.com/smartystreets/goconvey/convey"

	. "github.com/empirefox/iniu/base"
	. "github.com/empirefox/iniu/comm"
	. "github.com/empirefox/iniu/gorm/db"
)

var (
	webs  = "webs"
	child = "childs"
)

type Web struct {
	Id     int64   `json:",omitempty" order:"auto"`
	Name   string  `json:",omitempty" binding:"required" sql:"not null;type:varchar(64);unique"`
	Childs []Child `json:",omitempty"`
}

type Child struct {
	Id    int64  `json:",omitempty" order:"auto"`
	Name  string `json:",omitempty" binding:"required" sql:"not null;type:varchar(64);unique"`
	WebId int64  `json:",omitempty" binding:"required"`
}

func init() {
	DB.LogMode(true)
	Register(Web{})
	Register(Child{})
}

func newWebs(ns ...string) {
	DB.DropTableIfExists(Web{})
	DB.CreateTable(Web{})

	for _, n := range ns {
		DB.Save(&Web{Name: n})
	}
}

func newChilds(p int64, ns ...string) {
	DB.DropTableIfExists(Child{})
	DB.CreateTable(Child{})

	for _, n := range ns {
		DB.Save(&Child{WebId: p, Name: n})
	}
}

func TestPager(t *testing.T) {
	Convey("Pager", t, func() {
		Convey("Offset", func() {
			Convey("should not offset", func() {
				pager := Pager{Num: 10, Size: -1}
				So(pager.Offset(), ShouldEqual, "")

				pager = Pager{Num: 0, Size: 11}
				So(pager.Offset(), ShouldEqual, "")
			})
			Convey("should get computed value", func() {
				pager := Pager{Num: 11, Size: 21}
				So(pager.Offset(), ShouldEqual, "210")
			})
		})
		Convey("Limit", func() {
			Convey("should get default value", func() {
				pager := Pager{Num: 10}
				So(pager.Limit(), ShouldEqual, "20")
			})
			Convey("should not limit", func() {
				pager := Pager{Num: 10, Size: -1}
				So(pager.Limit(), ShouldEqual, "")
			})
			Convey("should get computed value", func() {
				pager := Pager{Num: 10, Size: 2}
				So(pager.Limit(), ShouldEqual, "2")
			})
		})
	})
}

func TestCpScopeFn(t *testing.T) {
	Convey("CpScopeFn", t, func() {
		Convey("should return with empty", func() {
			c := Child{Name: "cname", WebId: 10}
			fn, err := CpScopeFn(c, "")
			So(err, ShouldBeNil)

			odb := DB.Model(&c)
			rdb := odb.fn(odb)
			So(rdb, ShouldEqual, odb)
			So(rdb.NewScope(&c).Search.WhereConditions, ShouldBeEmpty)
		})
		Convey("should return with single cp", func() {
			c := Child{Name: "cname", WebId: 10}
			fn, err := CpScopeFn(c, "WebId")
			So(err, ShouldBeNil)

			odb := DB.Model(&c)
			rdb := odb.fn(odb)
			So(rdb, ShouldNotEqual, odb)
			So(rdb.NewScope(&c).Search.WhereConditions, ShouldNotBeEmpty)
		})
		Convey("should return with multi cp", func() {
			c := Child{Name: "cname", WebId: 10}
			fn, err := CpScopeFn(c, "WebId|Name")
			So(err, ShouldBeNil)

			odb := DB.Model(&c)
			rdb := odb.fn(odb)
			So(rdb, ShouldNotEqual, odb)
			So(rdb.NewScope(&c).Search.WhereConditions, ShouldNotBeEmpty)
		})
	})
}

func TestWritePager(t *testing.T) {
	searchFn := func(db *gorm.DB) *gorm.DB {
		return db
	}
	w := httptest.NewRecorder()
	Convey("WritePager", t, func() {
		Convey("should return error", func() {
			newWebs()
			err := WritePager(webs, Pager{Num: 10, Size: 20}, w, searchFn)
			So(err, ShouldNotBeNil)
			So(w.Header().Get("X-Total-Items"), ShouldEqual, "")
			So(w.Header().Get("X-Page"), ShouldEqual, "")
			So(w.Header().Get("X-Page-Size"), ShouldEqual, "")
			So(w.Header().Get("Access-Control-Expose-Headers"), ShouldEqual, "")
		})

		Convey("should set pager to header", func() {
			newWebs("a", "b", "c")
			err := WritePager(webs, Pager{Num: 1, Size: 20}, w, searchFn)
			So(err, ShouldBeNil)
			So(w.Header().Get("X-Total-Items"), ShouldEqual, "3")
			So(w.Header().Get("X-Page"), ShouldEqual, "1")
			So(w.Header().Get("X-Page-Size"), ShouldEqual, "20")
			So(w.Header().Get("Access-Control-Expose-Headers"), ShouldEqual, "X-Total-Items, X-Page")
		})
	})
}

func TestReturnAnyway(t *testing.T) {
	Convey("ReturnAnyway", t, func() {
		var okOrNot interface{}
		data := "ok"

		m := martini.Classic()
		m.Use(render.Renderer())

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("GET", "/ReturnAnyway", nil)

		m.Get("/ReturnAnyway", func(r render.Render) {
			ReturnAnyway(r, okOrNot, data)
		})

		Convey("should know bool true", func() {
			okOrNot = true
			m.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.Body, ShouldEqual, data)
		})

		Convey("should know bool false", func() {
			okOrNot = false
			m.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
			So(w.Body, ShouldEqual, data)
		})

		Convey("should know nil", func() {
			okOrNot = nil
			m.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.Body, ShouldEqual, data)
		})

		Convey("should know common error", func() {
			okOrNot = errors.New("temp error")
			m.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
			So(w.Body, ShouldEqual, data)
		})

		Convey("should know not found error", func() {
			okOrNot = gorm.RecordNotFound
			m.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, http.StatusNotFound)
			So(w.Body, ShouldEqual, data)
		})
	})
}

func TestReturn(t *testing.T) {
	Convey("Return", t, func() {
		var okOrNot interface{}

		m := martini.Classic()
		m.Use(render.Renderer())

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("GET", "/Return", nil)

		Convey("no args", func() {
			data := ""
			m.Get("/Return", func(r render.Render) {
				Return(r)
			})
			Convey("should return ok status", func() {
				m.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusOK)
				So(w.Body, ShouldEqual, data)
			})
		})

		Convey("only conditions", func() {
			data := ""
			m.Get("/Return", func(r render.Render) {
				Return(r, okOrNot)
			})
			Convey("should know bool true", func() {
				okOrNot = true
				m.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusOK)
				So(w.Body, ShouldEqual, data)
			})
			Convey("should know bool false", func() {
				okOrNot = true
				m.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body, ShouldEqual, data)
			})
			Convey("should return ok status", func() {
				okOrNot = nil
				m.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusOK)
				So(w.Body, ShouldEqual, data)
			})
			Convey("should return common error status", func() {
				okOrNot = errors.New("err")
				m.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body, ShouldEqual, "err")
			})
			Convey("should return not found error status", func() {
				okOrNot = gorm.RecordNotFound
				m.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusNotFound)
				So(w.Body, ShouldEqual, data)
			})
			Convey("should return data directly", func() {
				okOrNot = "data"
				m.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusOK)
				So(w.Body, ShouldEqual, "data")
			})
		})

		Convey("with data", func() {
			data := "ok"
			m.Get("/Return", func(r render.Render) {
				Return(r, okOrNot, data)
			})
			Convey("should know bool true", func() {
				okOrNot = true
				m.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusOK)
				So(w.Body, ShouldEqual, data)
			})
			Convey("should know bool false", func() {
				okOrNot = true
				m.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body, ShouldEqual, "")
			})
			Convey("should return ok status", func() {
				okOrNot = nil
				m.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusOK)
				So(w.Body, ShouldEqual, data)
			})
			Convey("should return common error status", func() {
				okOrNot = errors.New("err")
				m.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body, ShouldEqual, "err")
			})
			Convey("should return not found error status", func() {
				okOrNot = gorm.RecordNotFound
				m.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusNotFound)
				So(w.Body, ShouldEqual, "")
			})
			Convey("should return data directly", func() {
				okOrNot = "data"
				m.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusOK)
				So(w.Body, ShouldEqual, data)
			})
		})
	})
}
