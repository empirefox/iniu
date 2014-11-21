package web

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"

	"github.com/go-martini/martini"
	"github.com/golang/glog"
	"github.com/jinzhu/gorm"
	"github.com/martini-contrib/render"

	. "github.com/empirefox/iniu/base"
	"github.com/empirefox/iniu/comm"
	. "github.com/empirefox/iniu/gorm/db"
)

func Bind(m Model) martini.Handler {
	return func(c martini.Context) {
		c.MapTo(Tablename(m), (*Table)(nil))
	}
}

func BindTable(t string) martini.Handler {
	return func(c martini.Context) {
		c.MapTo(t, (*Table)(nil))
	}
}

// Get /google/mf
func ModelForm(t Table, r render.Render) {
	f, err := ModelFormMetas(t)
	Return(r, err, f)
}

// Get /google/df
func DbForm(t Table, r render.Render) {
	if sf, ok := JsonForms[t]; ok {
		if sf.New == nil {
			sf.New = Example(t)
		}
		r.JSON(http.StatusOK, sf)
		return
	}

	r.Error(http.StatusNoContent)
}

// Get /google/form
func Form(t Table, r render.Render) {
	if sf, ok := JsonForms[t]; ok {
		if sf.New == nil {
			sf.New = Example(t)
		}
		r.JSON(http.StatusOK, sf)
		return
	}

	f, err := ModelFormMetas(t)
	Return(r, err, f)
}

//TODO
// Get /google/search?q=
func Query(t Table, params martini.Params, r render.Render) {
	ms := Slice(t)
	err := DB.Where("name=?", params["name"]).Find(ms).Error
	Return(r, err, ms)
}

func One(t Table, req *http.Request, r render.Render) {
	m := New(t)

	req.ParseForm()
	q := map[string]interface{}{}
	for k, vs := range req.Form {
		if vs[0] != "" {
			q[k] = vs[0]
		}
	}

	err := DB.Where(q).First(m).Error
	Return(r, err, m)
}

func offset(os int64) string {
	if os > 0 {
		return comm.I64toA(os)
	}
	return ""
}

func limit(size int64) string {
	if size > 0 {
		return comm.I64toA(size)
	}
	if size == 0 {
		return "20"
	}
	return ""
}

type Pager struct {
	Size int64 `form:"size"`
	Num  int64 `form:"num"`
}

type IdPosName struct {
	Id   int64  `binding:"required"`
	Pos  int64  `binding:"required"`
	Name string `binding:"required"`
}

// Get /google/names?search={parent_id:10}&size=10&num=2
var Names = innerPage("id,pos,name", IdPosName{})

// Get /google/page?search={parent_id:10}&size=10&num=2
var Page = innerPage("*")

//columns: select sql
func innerPage(columns string, m ...interface{}) martini.Handler {
	if len(m) > 1 {
		panic("can only accept one m")
	}
	return func(t Table, params martini.Params, r render.Render, req *http.Request, w http.ResponseWriter) {
		req.ParseForm()
		var where = func(db *gorm.DB) *gorm.DB {
			search := req.Form.Get("search")
			switch search {
			case "", "{}":
				return db
			}

			raw := map[string]interface{}{}
			err := json.Unmarshal([]byte(search), &raw)
			if err != nil {
				panic(err)
			}
			comm.FixNum2Int64(&raw)

			return db.Where(raw)
		}

		size := comm.Atoi64(req.Form.Get("size"))
		num := comm.Atoi64(req.Form.Get("num"))

		os := offset(size * (num - 1))
		l := limit(size)

		var ms interface{}
		if len(m) == 0 {
			ms = Slice(t)
		} else {
			ms = reflect.New(reflect.SliceOf(reflect.TypeOf(m[0]))).Interface()
		}

		if err := DB.Table(t.(string)).Scopes(where).Select(columns).Offset(os).Limit(l).Find(ms).Error; err != nil {
			panic(err)
		}

		total, err := Total(t.(string), where)
		if err != nil {
			panic(err)
		}

		w.Header().Set("X-Total-Items", strconv.FormatInt(total, 10))
		w.Header().Set("X-Page", strconv.FormatInt(num, 10))
		w.Header().Set("Access-Control-Expose-Headers", "X-Total-Items, X-Page")
		r.JSON(http.StatusOK, ms)
	}
}

func Total(t Table, fns func(db *gorm.DB) *gorm.DB) (total int64, err error) {
	err = DB.Table(t).Scopes(fns...).Count(&total).Error
	return
}

func Recovery(t Table, r render.Render) {
	fact := New(t)
	DB.DropTableIfExists(fact)
	err := DB.CreateTable(fact).Error
	Return(r, err)
}

func AutoMigrate(t Table, r render.Render) {
	fact := New(t)
	err := DB.AutoMigrate(fact).Error
	Return(r, err)
}

type IdsData struct {
	Ids []int64 `binding:"required"`
}

// Put
func Remove(t Table, data IdsData, r render.Render) {
	err := DB.Where(data.Ids).Delete(New(t)).Error
	Return(r, err)
}

//will not set Pos
var Update = func(t Table, data Model, r render.Render) {
	Return(r, SaveModel(&data), data)
}

//will not affect Pos
//should be used when without Pos field or with Id and Pos already set
var UpdateAll = func(t Table, ms Models, r render.Render) {
	tx := DB.Begin()

	err := ForEach(ms, func(mPtr interface{}) error {
		return tx.Table(t.(string)).Save(mPtr).Error
	})

	if err != nil {
		tx.Rollback()
		Return(r, err)
		return
	}

	tx.Commit()
	Return(r, ms)
}

func ReturnAnyway(r render.Render, okOrNotI interface{}, data interface{}) {
	status := http.StatusInternalServerError
	switch okOrNot := okOrNot.(type) {
	case nil:
		status = http.StatusOK
	case bool:
		if okOrNot {
			status = http.StatusOK
		}
	case error:
		if okOrNot == gorm.RecordNotFound {
			status = http.StatusNotFound
		}
	}
	r.JSON(status, data)
}

func Return(r render.Render, data ...interface{}) {
	switch len(data) {
	case 0:
		r.JSON(http.StatusOK, "")
	case 1:
		switch d := data[0].(type) {
		case error:
			if d == gorm.RecordNotFound {
				r.JSON(http.StatusNotFound, "")
				return
			}
			glog.Errorln(d)
			r.JSON(http.StatusInternalServerError, d)
		case nil:
			r.JSON(http.StatusOK, "")
		case bool:
			if d {
				r.JSON(http.StatusOK, "")
				return
			}
			r.JSON(http.StatusInternalServerError, "")
		default:
			r.JSON(http.StatusOK, d)
		}
	default:
		switch d := data[0].(type) {
		case error:
			if d == gorm.RecordNotFound {
				r.JSON(http.StatusNotFound, "")
				return
			}
			glog.Errorln(d)
			r.JSON(http.StatusInternalServerError, d)
		case nil:
			r.JSON(http.StatusOK, data[1])
		case bool:
			if d {
				r.JSON(http.StatusOK, data[1])
				return
			}
			r.JSON(http.StatusInternalServerError, "")
		default:
			r.JSON(http.StatusOK, data[1])
		}
	}
}
