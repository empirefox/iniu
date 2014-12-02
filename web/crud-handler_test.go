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
)

func TestModelForm(t *testing.T) {
	Convey("ModelForm", t, func() {
		newWebs("n1")
		req := func(m martini.Router) (*http.Request, error) {
			m.Get("/form", BindTable(webs), ModelForm)
			return http.NewRequest("GET", "/form", nil)
		}
		res := map[string]interface{}{
			"Name": "Web",
			"Fields": []NameOnlyField{
				{Name: "Id"},
				{Name: "Name"},
				{Name: "Childs"},
			},
		}
		So(req, ShouldResponseOk, res)
	})
}

func TestClientForm(t *testing.T) {
	Convey("ClientForm", t, func() {
		newWebs("n1")
		req := func(m martini.Router) (*http.Request, error) {
			m.Get("/form", BindTable(webs), ClientForm)
			return http.NewRequest("GET", "/form", nil)
		}
		res := map[string]interface{}{
			"Name": "Web",
			"Fields": []NameOnlyField{
				{Name: "Id"},
				{Name: "Name"},
				{Name: "Childs"},
			},
		}
		So(req, ShouldResponseOk, res)
	})
}

func TestOne(t *testing.T) {
	Convey("One", t, func() {
		newWebs("n1")
		req := func(m martini.Router) (*http.Request, error) {
			m.Get("/1", binding.Form(IdData{}), BindTable(webs), One)
			return http.NewRequest("GET", "/1?Id=1", nil)
		}
		res := Web{Id: 1, Name: "n1"}

		So(req, ShouldResponseOk, res)
	})
}

func TestNames(t *testing.T) {
	Convey("Names", t, func() {
		newWebs("n1")
		Convey("with parent", func() {
			newChilds(1, "c1")
			req := func(m martini.Router) (*http.Request, error) {
				m.Get("/page", ParseSearch, binding.Form(Pager{}), BindTable(child), Names)
				return http.NewRequest("GET", `/page?search={"web_id":1}&size=2&num=1`, nil)
			}
			res := []IdPosName{{Id: 1, Name: "c1"}}

			So(req, ShouldResponseOk, res)
		})

		Convey("without parent", func() {
			req := func(m martini.Router) (*http.Request, error) {
				m.Get("/page", ParseSearch, binding.Form(Pager{}), BindTable(webs), Names)
				return http.NewRequest("GET", "/page?size=2&num=1", nil)
			}
			res := []IdPosName{{Id: 1, Name: "n1"}}

			So(req, ShouldResponseOk, res)
		})
	})
}

func TestPage(t *testing.T) {
	Convey("Page", t, func() {
		newWebs("n1")
		Convey("with parent", func() {
			newChilds(1, "c1")
			req := func(m martini.Router) (*http.Request, error) {
				m.Get("/page", ParseSearch, binding.Form(Pager{}), BindTable(child), Page)
				return http.NewRequest("GET", `/page?search={"web_id":1}&size=2&num=1`, nil)
			}
			res := []Child{{Id: 1, WebId: 1, Name: "c1"}}

			So(req, ShouldResponseOk, res)
		})

		Convey("without parent", func() {
			req := func(m martini.Router) (*http.Request, error) {
				m.Get("/page", ParseSearch, binding.Form(Pager{}), BindTable(webs), Page)
				return http.NewRequest("GET", "/page?size=2&num=1", nil)
			}
			res := []Child{{Id: 1, Name: "n1"}}

			So(req, ShouldResponseOk, res)
		})
	})
}

func TestRemove(t *testing.T) {
	Convey("Remove", t, func() {
		newWebs("n1")
		So(DB.Where("Name=?", "n1").First(&Web{}).Error, ShouldBeNil)

		req := func(m martini.Router) (*http.Request, error) {
			m.Put("/remove", BindTable(webs), binding.Bind(IdsData{}), Remove)
			return http.NewRequest("PUT", "/remove", strings.NewReader(`{"Ids":[1]}`))
		}
		res := ""

		So(req, ShouldResponseOk, res)

		So(DB.Where("Name=?", "n1").First(&Web{}).RecordNotFound(), ShouldBeTrue)
	})
}

func TestUpdate(t *testing.T) {
	Convey("Update", t, func() {
		newWebs("n1")
		So(DB.Where("Name=?", "n1").First(&Web{}).Error, ShouldBeNil)
		Convey("should create new model", func() {

			req := func(m martini.Router) (*http.Request, error) {
				m.Post("/save", binding.Bind(Web{}, (*Model)(nil)), BindTable(webs), Update)
				return http.NewRequest("POST", "/save", strings.NewReader(`{"Name":"n2"}`))
			}
			res := Web{Id: 2, Name: "n2"}

			So(req, ShouldResponseOk, res)

			So(DB.Where("Name=?", "n2").First(&Web{}).RecordNotFound(), ShouldBeFalse)
		})
		Convey("should update the exist model", func() {
			req := func(m martini.Router) (*http.Request, error) {
				m.Post("/save", binding.Bind(Web{}, (*Model)(nil)), BindTable(webs), Update)
				return http.NewRequest("POST", "/save", strings.NewReader(`{"Id":1,"Name":"n2"}`))
			}
			res := Web{Id: 1, Name: "n2"}

			So(req, ShouldResponseOk, res)

			So(DB.Where("Name=?", "n1").First(&Web{}).RecordNotFound(), ShouldBeTrue)
			So(DB.Where("Name=?", "n2").First(&Web{}).RecordNotFound(), ShouldBeFalse)
		})
	})
}

func TestUpdateAll(t *testing.T) {
	Convey("UpdateAll", t, func() {
		newWebs("n1")
		So(DB.Where("Name=?", "n1").First(&Web{}).Error, ShouldBeNil)
		Convey("should update all models", func() {

			req := func(m martini.Router) (*http.Request, error) {
				m.Post("/saveall", binding.Bind([]Web{}, (*Models)(nil)), BindTable(webs), UpdateAll)
				return http.NewRequest("POST", "/saveall", strings.NewReader(`[{"Id":1,"Name":"n2"},{"Name":"n3"}]`))
			}
			res := []Web{{Id: 1, Name: "n2"}, {Id: 2, Name: "n3"}}

			So(req, ShouldResponseOk, res)

			web := Web{}
			So(DB.Where("Name=?", "n2").First(&web).RecordNotFound(), ShouldBeFalse)
			So(web, ShouldResemble, Web{Id: 1, Name: "n2"})

			web = Web{}
			So(DB.Where("Name=?", "n3").First(&web).RecordNotFound(), ShouldBeFalse)
			So(web, ShouldResemble, Web{Id: 2, Name: "n3"})
		})
	})
}

func TestRecovery(t *testing.T) {
	Convey("UpdateAll", t, func() {
		newWebs("n1")
		var total int64
		So(DB.Table(webs).Count(&total).Error, ShouldBeNil)
		So(total, ShouldEqual, 1)

		req := func(m martini.Router) (*http.Request, error) {
			m.Delete("/rec", BindTable(webs), Recovery)
			return http.NewRequest("DELETE", "/rec", nil)
		}

		So(req, ShouldResponseOk, "", http.StatusOK)
		So(DB.Table(webs).Count(&total).Error, ShouldBeNil)
		So(total, ShouldEqual, 0)
	})
}
