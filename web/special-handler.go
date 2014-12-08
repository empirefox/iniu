package web

import (
	"errors"
	"net/http"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"

	. "github.com/empirefox/iniu/base"
	"github.com/empirefox/iniu/comm"
	"github.com/empirefox/iniu/security"
)

type TableForm struct {
	Name  string `json:",omitempty"`
	Title string `json:",omitempty"`
}

var tableForms = func() (tfs []TableForm, err error) {
	err = SudoDb("forms").Select("name,title").Order("pos desc").Find(&tfs).Error
	return
}

// GET /forms
var TableForms = func(r render.Render, a *security.Account) {

	tfs, err := tableForms()
	if err != nil {
		panic(err)
	}

	result := []TableForm{}
	for _, tf := range tfs {
		p := security.WebPerm(tf.Name, "Form")
		if a.Permitted(p) {
			result = append(result, tf)
		}
	}
	r.JSON(http.StatusOK, result)
}

var ModelFields = modelFields

// GET /mfs/(?P<id>[\d]+)
func modelFields(r render.Render, params martini.Params) {
	t, ok := security.IdTables[comm.Atoi64(params["id"])]
	if !ok {
		Return(r, errors.New("wrong mfs id"))
		return
	}
	f, err := ModelFormMetas(t)
	if err != nil {
		Return(r, err)
		return
	}
	Return(r, f["Fields"])
}
