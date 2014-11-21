//需设置环境变量
//GORM_DIALECT=mysql DB_URL="gorm:gorm@/gorm?charset=utf8&parseTime=True"
//GORM_DIALECT=postgres DB_URL="user=gorm dbname=gorm sslmode=disable" PORT=8080 goconvey
//GORM_DIALECT=sqlite3 DB_URL=/tmp/gorm.DB go test
package web

import (
	"net/http"
	"strings"
	"testing"

	"github.com/empirefox/shirolet"
	"github.com/erikstmartin/go-testdb"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	. "github.com/smartystreets/goconvey/convey"

	. "github.com/empirefox/iniu/base"
	. "github.com/empirefox/iniu/comm"
	. "github.com/empirefox/iniu/cv"
	. "github.com/empirefox/iniu/gorm/db"
	"github.com/empirefox/iniu/gorm/mock"
	. "github.com/empirefox/iniu/gorm/mod"
	"github.com/empirefox/iniu/security"
)

var (
	webs  = "webs"
	child = "childs"
)

type Web struct {
	Id     int64   `json:",omitempty"`
	Name   *string `json:",omitempty" binding:"required" sql:"not null;type:varchar(64);unique"`
	Childs []Child `json:",omitempty"`
	Orderd `json:"-" sql:"-"`
}

type Child struct {
	Id     int64   `json:",omitempty"`
	Name   *string `json:",omitempty" binding:"required" sql:"not null;type:varchar(64);unique"`
	WebId  int64   `json:",omitempty" binding:"required"`
	Orderd `json:"-" sql:"-"`
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
		DB.Save(&Web{Name: PtrS(n)})
	}
}

func newChilds(p int64, ns ...string) {
	DB.DropTableIfExists(Child{})
	DB.CreateTable(Child{})

	for _, n := range ns {
		DB.Save(&Child{WebId: p, Name: PtrS(n)})
	}
}

//func TestDF(t *testing.T) {
//	Convey("测试df", t, func() {
//		//db mock query Forms
//		db := mock.DB.Model(&security.Form{})
//		bdb := mock.NewBackend(db)

//		fs := []security.Form{}
//		columns := []string{"name", "pos"}
//		result := `Web,1`
//		bdb.StubQuery(&fs, testdb.RowsFromCSVString(columns, result))

//		//db mock DB.Model(f).Related(&fields)
//		db := mock.DB.Model(f)
//		bdb := mock.NewBackend(db)

//		tfs := []TableForm{}
//		columns := []string{"name", "title"}
//		result := `
//		Form,Form Title
//		Field,Field Title
//		`
//		bdb.StubQuery(&tfs, testdb.RowsFromCSVString(columns, result))

//		//spy security
//		security.Forms = map[string]*security.ComputedForm{
//			"Form": &security.ComputedForm{
//				WebPerms: map[string]shirolet.Permit{
//					"Form": shirolet.NewPermit("edit:form:Form"),
//				},
//			},
//			"Field": &security.ComputedForm{
//				WebPerms: map[string]shirolet.Permit{
//					"Form": shirolet.NewPermit("edit:form:Field"),
//				},
//			},
//		}
//		hp := "edit:form:Field"
//		a := &security.Account{HoldsPerm: &hp}

//		req := func(m martini.Router) (*http.Request, error) {
//			m.Get("/forms", func(c martini.Context) {
//				c.Map(a)
//			}, TableForms)
//			return http.NewRequest("GET", "/forms", nil)
//		}
//		res := []TableForm{{Name: "Field", Title: "Field Title"}}

//		So(r, ShouldResponseOk, res)
//	})
//}

func TestMF(t *testing.T) {
	Convey("测试mf", t, func() {
		newWebs("n1")
		req := func(m martini.Router) (*http.Request, error) {
			m.Get("/form", BindTable(webs), ModelForm)
			return http.NewRequest("GET", "/form", nil)
		}
		res := security.Form{
			Name: PtrS("Web"),
			Fields: []security.Field{
				{Name: PtrS("Id")},
				{Name: PtrS("Name")},
			},
		}
		So(req, ShouldResponseOk, res)
	})
}

func TestForm(t *testing.T) {
	Convey("测试默认Form", t, func() {
		newWebs("n1")
		req := func(m martini.Router) (*http.Request, error) {
			m.Get("/form", BindTable(webs), Form)
			return http.NewRequest("GET", "/form", nil)
		}
		res := security.Form{
			Name: PtrS("Web"),
			Fields: []security.Field{
				{Name: PtrS("Id")},
				{Name: PtrS("Name")},
			},
		}
		So(req, ShouldResponseOk, res)
	})
}

func TestOne(t *testing.T) {
	Convey("测试查找单一实例", t, func() {
		newWebs("n1")
		req := func(m martini.Router) (*http.Request, error) {
			m.Get("/1", BindTable(webs), One)
			return http.NewRequest("GET", "/1?name=n1", nil)
		}
		res := Web{Id: 1, Name: PtrS("n1")}

		So(req, ShouldResponseOk, res)
	})
}

// JsonContent(r, map[string]interface{}{"list": ms, "page_count": pageCount, "total": total})
func TestPageWithParent(t *testing.T) {
	Convey("有parent的page", t, func() {
		newWebs("n1")
		newChilds(1, "c1")
		req := func(m martini.Router) (*http.Request, error) {
			m.Get("/page", BindTable(child), Page)
			return http.NewRequest("GET", `/page?search={"web_id":1}&size=2&num=1`, nil)
		}
		res := []Child{{Id: 1, WebId: 1, Name: PtrS("c1")}}

		So(req, ShouldResponseOk, res)
	})
}

func TestPageWithoutParent(t *testing.T) {
	Convey("无parent的page", t, func() {
		newWebs("n1")
		req := func(m martini.Router) (*http.Request, error) {
			m.Get("/page", BindTable(webs), Page)
			return http.NewRequest("GET", "/page?size=2&num=1", nil)
		}
		res := []Child{{Id: 1, Name: PtrS("n1")}}

		So(req, ShouldResponseOk, res)
	})
}

func TestRemove(t *testing.T) {
	Convey("删除", t, func() {
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

func TestUpdateWithoutId(t *testing.T) {
	Convey("新建后保存", t, func() {
		newWebs("n1")
		So(DB.Where("Name=?", "n1").First(&Web{}).Error, ShouldBeNil)

		req := func(m martini.Router) (*http.Request, error) {
			m.Post("/save", binding.Bind(Web{}, (*Model)(nil)), BindTable(webs), Update)
			return http.NewRequest("POST", "/save", strings.NewReader(`{"Name":"n2"}`))
		}
		res := Web{Id: 2, Name: PtrS("n2")}

		So(req, ShouldResponseOk, res)

		So(DB.Where("Name=?", "n2").First(&Web{}).RecordNotFound(), ShouldBeFalse)
	})
}

func TestUpdateWithId(t *testing.T) {
	Convey("保存现有Model", t, func() {
		newWebs("n1")
		So(DB.Where("Name=?", "n1").First(&Web{}).Error, ShouldBeNil)

		req := func(m martini.Router) (*http.Request, error) {
			m.Post("/save", binding.Bind(Web{}, (*Model)(nil)), BindTable(webs), Update)
			return http.NewRequest("POST", "/save", strings.NewReader(`{"Id":1,"Name":"n2"}`))
		}
		res := Web{Id: 1, Name: PtrS("n2")}

		So(req, ShouldResponseOk, res)

		So(DB.Where("Name=?", "n1").First(&Web{}).RecordNotFound(), ShouldBeTrue)
		So(DB.Where("Name=?", "n2").First(&Web{}).RecordNotFound(), ShouldBeFalse)
	})
}

func TestSpecialTableForms(t *testing.T) {
	Convey("获取有相应权限的表单简单结构", t, func() {
		//prepare

		//db mock
		db := mock.DB.Table("forms").Select("name,title").Order("pos desc")
		bdb := mock.NewBackend(db)

		tfs := []TableForm{}
		columns := []string{"name", "title"}
		result := `
		Form,Form Title
		Field,Field Title
		`
		bdb.StubQuery(&tfs, testdb.RowsFromCSVString(columns, result))

		//spy security
		security.Forms = map[string]*security.ComputedForm{
			"Form": &security.ComputedForm{
				WebPerms: map[string]shirolet.Permit{
					"Form": shirolet.NewPermit("edit:form:Form"),
				},
			},
			"Field": &security.ComputedForm{
				WebPerms: map[string]shirolet.Permit{
					"Form": shirolet.NewPermit("edit:form:Field"),
				},
			},
		}
		hp := "edit:form:Field"
		a := &security.Account{HoldsPerm: &hp}

		req := func(m martini.Router) (*http.Request, error) {
			m.Get("/forms", func(c martini.Context) {
				c.Map(a)
			}, TableForms)
			return http.NewRequest("GET", "/forms", nil)
		}
		res := []TableForm{{Name: "Field", Title: "Field Title"}}

		So(req, ShouldResponseOk, res)
	})
}
