package web

import (
	"net/http"

	"github.com/martini-contrib/render"

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
