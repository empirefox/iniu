package base

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/empirefox/iniu/comm"
	"github.com/go-martini/martini"
	"github.com/jinzhu/gorm"
	"github.com/tobyhede/go-underscore"
)

var (
	IgnoresInSearch = strings.Split("size|page", "|")
)

func isIgnored(value string) bool {
	return un.AnyString(func(ignored string) bool {
		return ignored == value
	}, IgnoresInSearch)
}

func ParseSearch(c martini.Context, req *http.Request) {
	req.ParseForm()
	search := req.Form.Get("search")
	switch search {
	case "", "{}":
		c.Map(func(db *gorm.DB) *gorm.DB { return db })
		return
	}

	raw := map[string]interface{}{}
	err := json.Unmarshal([]byte(search), &raw)
	if err != nil {
		panic(err)
	}
	comm.FixNum2Int64(&raw)
	var searchFn = func(db *gorm.DB) *gorm.DB {
		return db.Where(raw)
	}
	c.Map(searchFn)
}
