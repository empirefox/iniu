//GORM_DIALECT=mysql DB_URL="gorm:gorm@/gorm?charset=utf8&parseTime=True"
//GORM_DIALECT=postgres DB_URL="user=gorm dbname=gorm sslmode=disable" PORT=8080 goconvey
//GORM_DIALECT=sqlite3 DB_URL=/tmp/gorm.DB go test
package web

import (
	"net/http"
	"strings"
	"testing"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	. "github.com/smartystreets/goconvey/convey"

	. "github.com/empirefox/iniu/base"
	. "github.com/empirefox/iniu/convey"
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

func TestSaveUp(t *testing.T) {
	Convey("SaveUp", t, func() {
		newTable(2, 3, 4, 5, 6, 7, 8, 9)
		Convey("should save new model up", func() {

			req := func(m martini.Router) (*http.Request, error) {
				m.Post("/saveup", binding.Form(SaveUpData{}), binding.Bind(Xchg{}, (*Model)(nil)), BindTable(xchgs), SaveUp)
				return http.NewRequest("POST", "/saveup?b=1&t=5", strings.NewReader(`{"Pos":3}`))
			}
			res := J{IpKey: []IdPos{{1, 1}, {2, 9}, {3, 25}, {4, 33}, {5, 41}}, "Newer": Xchg{9, 17}}

			So(req, ShouldResponseOk, res)
		})
	})
}

func TestRearrange(t *testing.T) {
	Convey("SaveUp", t, func() {
		newTable(2, 3, 4, 5, 6, 7, 8, 9)
		Convey("should arrange and return mods", func() {

			req := func(m martini.Router) (*http.Request, error) {
				m.Put("/Rearrange", ParseSearch, binding.Bind(RearrData{}), BindTable(xchgs), Rearrange)
				return http.NewRequest("PUT", "/Rearrange", strings.NewReader(`{"bid":1,"tid":5,"baseid":2}`))
			}
			res := []IdPos{{1, 1}, {2, 9}, {3, 25}, {4, 33}, {5, 41}}

			So(req, ShouldResponseOk, res)
		})
	})
}

// ignore ModIps Xpos PosTop PosBottom

func TestPosUpSingle(t *testing.T) {
	Convey("PosUpSingle", t, func() {
		newTable(2, 3, 4, 5, 6, 7, 8, 9)
		Convey("should move up usually", func() {

			req := func(m martini.Router) (*http.Request, error) {
				m.Post("/PosUpSingle", ParseSearch, binding.Form(Direction{}), binding.Bind(IdPos{}), BindTable(xchgs), PosUpSingle)
				return http.NewRequest("POST", "/PosUpSingle", strings.NewReader(`{"Id":2,"Pos":3}`))
			}
			res := Xchg{3, 4}

			So(req, ShouldResponseOk, res)
		})

		Convey("should move down usually", func() {

			req := func(m martini.Router) (*http.Request, error) {
				m.Post("/PosUpSingle", ParseSearch, binding.Form(Direction{}), binding.Bind(IdPos{}), BindTable(xchgs), PosUpSingle)
				return http.NewRequest("POST", "/PosUpSingle?reverse=true", strings.NewReader(`{"Id":2,"Pos":3}`))
			}
			res := Xchg{1, 2}

			So(req, ShouldResponseOk, res)
		})
	})
}
