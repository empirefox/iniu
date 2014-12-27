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
func One(t Table, db *gorm.DB, data IdData, r render.Render) {
	m := New(t)
	err := db.Where("id=?", data.Id).First(m).Error
	Return(r, err, m)
}

// Get /google/names?search={parent_id:10}&size=10&num=2
// need ParseSearch midware and Pager
var Names = func(t Table, db *gorm.DB, r render.Render, pager Pager, w http.ResponseWriter, searchFn func(db *gorm.DB) *gorm.DB) {
	ms := []IdPosName{}
	selects := "id,pos,name"
	if !HasPos(t) {
		selects = "id,name"
	}
	if err := db.Scopes(searchFn).Select(selects).Offset(pager.Offset()).Limit(pager.Limit()).Find(&ms).Error; err != nil {
		Return(r, err)
		return
	}
	Return(r, WritePager(db, pager, w, searchFn), ms)
}

// Get /google/page?search={parent_id:10}&size=10&num=2
// need ParseSearch midware and Pager
var Page = func(t Table, db *gorm.DB, r render.Render, pager Pager, w http.ResponseWriter, searchFn func(db *gorm.DB) *gorm.DB) {
	ms := Slice(t)
	if err := db.Scopes(searchFn).Offset(pager.Offset()).Limit(pager.Limit()).Find(ms).Error; err != nil {
		Return(r, err)
		return
	}
	Return(r, WritePager(db, pager, w, searchFn), ms)
}

// Put
func Remove(t Table, db *gorm.DB, data IdsData, r render.Render) {
	err := db.Set("context:delete_ids", data.Ids).Where("id in (?)", data.Ids).Delete(New(t)).Error
	Return(r, err)
}

//will not set Pos
func Update(db *gorm.DB, data Model, r render.Render) {
	Return(r, SaveModel(db, &data), data)
}

//will not affect Pos
//should be used when without Pos field or with Id and Pos already set
func UpdateAll(db *gorm.DB, ms Models, r render.Render) {
	tx := db.Begin()

	err := ForEach(ms, func(mPtr interface{}) error {
		return tx.Save(mPtr).Error
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
