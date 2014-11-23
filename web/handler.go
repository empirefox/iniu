package web

import (
	"net/http"

	"github.com/go-martini/martini"
	"github.com/golang/glog"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"github.com/oleiade/reflections"

	. "github.com/empirefox/iniu/base"
	"github.com/empirefox/iniu/security"
)

var Handlers = ModelHandlers{

	ModelForm: ModelForm,

	DbForm: DbForm,

	// Get /google/form
	Form: Form,

	// Get /google/1?a=a
	One: One,

	// Get /google/names
	Names: Names,

	// Get /google/page?parent=10&size=10&num=2
	Page: Page,

	// Delete /google/recovery
	Recovery: Recovery,

	// Put /google/migrate
	AutoMigrate: AutoMigrate,

	Rearr: Rearr,

	// Put /google/binding/remove
	Remove: Remove,

	// Post /google/binding/update
	Update: Update,

	UpdateAll: UpdateAll,

	// Post /google/save/up?lower=BBB&upper=CCC&cp=Stu,Tech
	NewUpper: NewUpper,

	SinglePosUp: SinglePosUp,

	Xpos: Xpos,

	ModIps: ModIps,

	Pos2Top: Pos2Top,

	Pos2Bottom: Pos2Bottom,
}

type ModelHandlers struct {
	ModelForm martini.Handler

	DbForm martini.Handler

	// Get /google/form
	Form martini.Handler

	// Get /google/1?a=a
	One martini.Handler

	// Get /google/names
	Names martini.Handler

	// Get /google/page?parent=10&size=10&num=2
	Page martini.Handler

	// Delete /google/recovery
	Recovery martini.Handler

	// Put /google/migrate
	AutoMigrate martini.Handler

	Rearr martini.Handler

	// Put /google/binding/remove
	Remove martini.Handler

	// Post /google/binding/update
	Update martini.Handler

	UpdateAll martini.Handler

	// Post /google/save/up?lower=BBB&upper=CCC&cp=Stu,Tech
	NewUpper martini.Handler

	SinglePosUp martini.Handler

	Xpos martini.Handler

	ModIps martini.Handler

	Pos2Top martini.Handler

	Pos2Bottom martini.Handler
}

func Link(m martini.Router, model Model, h ModelHandlers) {
	Register(model)

	t := Tablename(model)
	f := Formname(model)
	glog.Infoln("加载web form ", f)
	m.Group("/"+f, func(r martini.Router) {

		r.Get("/mf", CheckWeb("ModelForm"), h.ModelForm)
		r.Get("/df", CheckWeb("DbForm"), h.DbForm)
		r.Get("/form", CheckWeb("Form"), h.Form)
		r.Get(`/1`, CheckWeb("One"), h.One)
		r.Get("/names", CheckWeb("Names"), h.Names)
		r.Get(`/page`, CheckWeb("Page"), h.Page)
		r.Post("/save", CheckWeb("Update"), binding.Bind(model, (*Model)(nil)), h.Update)
		r.Post("/saveall", CheckWeb("Update"), binding.Bind(IndirectSlice(t), (*Models)(nil)), h.UpdateAll)
		r.Put("/remove", CheckWeb("Remove"), binding.Bind(IdsData{}), binding.ErrorHandler, h.Remove)
		r.Delete("/recovery", CheckWeb("Recovery"), h.Recovery)
		r.Put("/migrate", CheckWeb("AutoMigrate"), h.AutoMigrate)
		if has, _ := reflections.HasField(model, "Pos"); has {
			r.Post("/saveup", CheckWeb("Update"), binding.Form(SaveUpData{}), binding.Bind(model, (*Model)(nil)), h.NewUpper)
			r.Put("/rearrange", CheckWeb("Update"), h.Rearr)
			r.Post("/modips", CheckWeb("Update"), binding.Bind([]IdPos{}), h.ModIps)
			r.Post("/xpos", CheckWeb("Update"), binding.Bind(Posx{}), h.Xpos)
			r.Post("/postop", CheckWeb("Update"), binding.Bind(PosParentsData{}), h.Pos2Top)
			r.Post("/posbottom", CheckWeb("Update"), binding.Bind([]PosParentsData{}), h.Pos2Bottom)
			r.Post("/singleposup", CheckWeb("Update"), binding.Bind(PosParentsData{}), h.SinglePosUp)
		}

	}, security.CheckLogin(), BindTable(t))
}

var CheckWeb = func(method string) martini.Handler {
	return func(t Table, r render.Render, a *security.Account, model Model) {
		p := security.WebPerm(t, method)
		if !a.Permitted(p) {
			r.JSON(http.StatusUnauthorized, "")
		}
	}
}

func LinkAll(m martini.Router, model Model) {
	Link(m, model, Handlers)
}
