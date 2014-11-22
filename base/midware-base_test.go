//需设置环境变量
//GORM_DIALECT=mysql DB_URL="gorm:gorm@/gorm?charset=utf8&parseTime=True"
//GORM_DIALECT=postgres DB_URL="user=gorm dbname=gorm sslmode=disable"
//GORM_DIALECT=sqlite3 DB_URL=/tmp/gorm.DB go test
package base

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-martini/martini"
	"github.com/jinzhu/gorm"
	"github.com/martini-contrib/render"
	. "github.com/smartystreets/goconvey/convey"

	. "github.com/empirefox/iniu/gorm/db"
)

func TestParseSearch(t *testing.T) {
	Convey("测试ParseSearch", t, func() {
		Convey("should pass the origin db", func() {
			m := martini.Classic()
			m.Use(render.Renderer())
			odb := DB.Model(Xchg{})
			var rdb *gorm.DB
			m.Get("/parse", ParseSearch, func(searchFn func(db *gorm.DB) *gorm.DB) {
				rdb = odb.Scopes(searchFn)
			})
			res := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/parse", nil)
			m.ServeHTTP(res, req)

			So(rdb, ShouldEqual, odb)
			So(rdb.NewScope(Xchg{}).Search.WhereConditions, ShouldBeEmpty)
		})

		Convey("should handle and map searchFn", func() {
			m := martini.Classic()
			m.Use(render.Renderer())
			odb := DB.Model(Xchg{})
			var rdb *gorm.DB
			m.Get("/parse", ParseSearch, func(searchFn func(db *gorm.DB) *gorm.DB) {
				rdb = odb.Scopes(searchFn)
			})
			res := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", `/parse?search={"parent_id":2}`, nil)
			m.ServeHTTP(res, req)

			So(rdb, ShouldNotBeNil)
			scope := rdb.NewScope(Xchg{})
			So(scope, ShouldNotBeNil)
			search := scope.Search
			So(search, ShouldNotBeNil)
			cs := search.WhereConditions
			So(len(cs), ShouldEqual, 1)
		})
	})
}
