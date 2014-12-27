package web

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-martini/martini"
	"github.com/golang/glog"
	"github.com/jinzhu/gorm"
	"github.com/martini-contrib/render"

	. "github.com/empirefox/iniu/base"
	"github.com/empirefox/iniu/comm"
	. "github.com/empirefox/iniu/gorm/db"
	"github.com/empirefox/iniu/security"
)

type Pager struct {
	Size int64 `form:"size"`
	Num  int64 `form:"num"`
}

func (pager *Pager) Offset() string {
	os := pager.Size * (pager.Num - 1)
	if os > 0 {
		return comm.I64toA(os)
	}
	return ""
}

func (pager *Pager) Limit() string {
	switch {
	case pager.Size > 0:
		return comm.I64toA(pager.Size)
	case pager.Size == 0:
		return "20"
	}
	return ""
}

type IdPosName struct {
	Id   int64  `binding:"required" order:"auto"`
	Pos  int64  `binding:"required" json:",omitempty"`
	Name string `binding:"required"`
}

type IdsData struct {
	Ids []int64 `binding:"required"`
}

type IdData struct {
	Id int64 `form:"Id" binding:"required" order:"auto"`
}

func SudoDb(t Table) *gorm.DB {
	return DB.Set("context:account", "*").Table(t.(string))
}

func BindGorm(c martini.Context, a *security.Account, t Table) {
	c.Map(DB.Set("context:account", a).Table(t.(string)))
}

func CpScopeFn(data Model, cp string) (func(*gorm.DB) *gorm.DB, error) {
	var null = func(db *gorm.DB) *gorm.DB {
		return db
	}

	if cp == "" {
		// TODO need to sync with client
		return null, nil
	}

	query := map[string]interface{}{}
	scope := gorm.Scope{Value: data}

	done := comm.StrSlice(strings.Split(cp, "|")).All(func(p string) bool {
		field, found := scope.FieldByName(p)
		if found {
			query[field.DBName] = field.Field.Interface()
		}
		return found
	})

	if !done || len(query) == 0 {
		return nil, errors.New("cp not found")
	}

	return func(db *gorm.DB) *gorm.DB {
		return db.Where(query)
	}, nil
}

func BindTable(t string) martini.Handler {
	return func(c martini.Context) {
		c.MapTo(t, (*Table)(nil))
	}
}

// can be 0
func Total(db *gorm.DB, fns ...func(db *gorm.DB) *gorm.DB) (total int64, err error) {
	err = db.Scopes(fns...).Count(&total).Error
	return
}

func WritePager(db *gorm.DB, pager Pager, w http.ResponseWriter, searchFn func(db *gorm.DB) *gorm.DB) error {
	total, err := Total(db, searchFn)
	if err != nil {
		return err
	}

	if total == 0 {
		pager.Num = 1
	}

	w.Header().Set("X-Total-Items", strconv.FormatInt(total, 10))
	w.Header().Set("X-Page", strconv.FormatInt(pager.Num, 10))
	w.Header().Set("X-Page-Size", strconv.FormatInt(pager.Size, 10))
	w.Header().Set("Access-Control-Expose-Headers", "X-Total-Items, X-Page, X-Page-Size")
	return nil
}

func ReturnAnyway(r render.Render, okOrNot interface{}, data interface{}) {
	status := http.StatusInternalServerError
	switch ctype := okOrNot.(type) {
	case nil:
		status = http.StatusOK
	case bool:
		if ctype {
			status = http.StatusOK
		}
	case error:
		glog.Errorln(ctype)
		if ctype == gorm.RecordNotFound {
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
			r.JSON(http.StatusInternalServerError, d.Error())
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
			r.JSON(http.StatusInternalServerError, d.Error())
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
