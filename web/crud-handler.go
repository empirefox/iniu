package web

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/martini-contrib/render"

	. "github.com/empirefox/iniu/base"
	. "github.com/empirefox/iniu/gorm/db"
	"github.com/empirefox/iniu/security"
)

// Get /google/mf
func ModelForm(t Table, r render.Render) {
	f, err := ModelFormMetas(t)
	Return(r, err, f)
}

// Get /google/df
func DbForm(t Table, r render.Render) {
	if sf, ok := security.JsonForms[t]; ok {
		if sf.New == nil {
			sf.New = Example(t)
		}
		r.JSON(http.StatusOK, sf)
		return
	}

	r.Error(http.StatusNoContent)
}

// Get /google/form
func ClientForm(t Table, r render.Render) {
	if sf, ok := security.JsonForms[t]; ok {
		if sf.New == nil {
			sf.New = Example(t)
		}
		r.JSON(http.StatusOK, sf)
		return
	}

	f, err := ModelFormMetas(t)
	Return(r, err, f)
}

// only can get by id?
// need form IdPosName
func One(t Table, data IdData, r render.Render) {
	m := New(t)
	err := DB.Table(t.(string)).Where("id=?", data.Id).First(m).Error
	Return(r, err, m)
}

// Get /google/names?search={parent_id:10}&size=10&num=2
// need ParseSearch midware and Pager
var Names = func(t Table, r render.Render, pager Pager, w http.ResponseWriter, searchFn func(db *gorm.DB) *gorm.DB) {
	ms := []IdPosName{}
	selects := "id,pos,name"
	if !HasPos(t) {
		selects = "id,name"
	}
	if err := DB.Table(t.(string)).Scopes(searchFn).Select(selects).Offset(pager.Offset()).Limit(pager.Limit()).Find(&ms).Error; err != nil {
		Return(r, err)
		return
	}
	Return(r, WritePager(t, pager, w, searchFn), ms)
}

// Get /google/page?search={parent_id:10}&size=10&num=2
// need ParseSearch midware and Pager
var Page = func(t Table, r render.Render, pager Pager, w http.ResponseWriter, searchFn func(db *gorm.DB) *gorm.DB) {
	ms := Slice(t)
	if err := DB.Table(t.(string)).Scopes(searchFn).Offset(pager.Offset()).Limit(pager.Limit()).Find(ms).Error; err != nil {
		Return(r, err)
		return
	}
	Return(r, WritePager(t, pager, w, searchFn), ms)
}

// Put
func Remove(t Table, data IdsData, r render.Render) {
	err := DB.Where(data.Ids).Delete(New(t)).Error
	Return(r, err)
}

//will not set Pos
func Update(t Table, data Model, r render.Render) {
	Return(r, SaveModel(&data), data)
}

//will not affect Pos
//should be used when without Pos field or with Id and Pos already set
func UpdateAll(t Table, ms Models, r render.Render) {
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

func Recovery(t Table, r render.Render) {
	fact := New(t)
	DB.DropTableIfExists(fact)
	err := DB.Table(t.(string)).CreateTable(fact).Error
	Return(r, err)
}

func AutoMigrate(t Table, r render.Render) {
	fact := New(t)
	err := DB.Table(t.(string)).AutoMigrate(fact).Error
	Return(r, err)
}
