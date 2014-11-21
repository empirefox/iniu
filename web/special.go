package web

import (
	"net/http"

	"github.com/martini-contrib/render"

	. "github.com/empirefox/iniu/gorm/postgres"
	"github.com/empirefox/iniu/security"
)

type TableForm struct {
	Name  string `json:",omitempty"`
	Title string `json:",omitempty"`
}

var TableForms = func(r render.Render, a *security.Account) {

	tfs := []TableForm{}
	err := DB.Table("forms").Select("name,title").Order("pos desc").Find(&tfs).Error
	if err != nil {
		panic(err)
	}

	if security.Skip {
		r.JSON(http.StatusOK, tfs)
		return
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
