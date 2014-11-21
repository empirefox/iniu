package web

import (
	"errors"
	"reflect"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/martini-contrib/render"
	un "github.com/tobyhede/go-underscore"

	. "github.com/empirefox/iniu/base"
	. "github.com/empirefox/iniu/gorm/db"
)

func ScopeFn(data Model, cp string) (func(*gorm.DB) *gorm.DB, error) {
	var null = func(db *gorm.DB) *gorm.DB {
		return db
	}

	if cp == "" {
		// TODO need to sync with client
		return null, nil
	}

	query := map[string]interface{}{}
	scope := gorm.Scope{Value: data}

	var proccess = func(p string) bool {
		field, found := scope.FieldByName(p)
		if found {
			query[field.DBName] = field.Field.Interface()
		}
		return found
	}

	done := un.EveryString(proccess, strings.Split(cp, "|"))
	if !done || len(query) == 0 {
		return nil, errors.New("cp not found")
	}

	return func(db *gorm.DB) *gorm.DB {
		return db.Where(query), nil
	}
}

type SaveUpData struct {
	BottomId int64  `form:"b"`
	TopId    int64  `form:"t"`
	Cp       string `form:"cp"`
}

func SaveUp(t Table, data Model, up SaveUpData, r render.Render) {
	if !HasPos(t) {
		Return(r, errors.New("have no pos here"))
		return
	}

	wFn, err := ScopeFn(data, up.Cp)
	if err != nil {
		Return(r, err)
		return
	}

	basePos := reflect.ValueOf(data).FieldByName("Pos").Int()

	var baseIp IdPos
	baseIp, err = IpByPos(t, basePos, wFn)
	if err != nil {
		Return(r, err)
		return
	}

	if up.BottomId|up.TopId == 0 {
		up.BottomId, up.TopId = baseIp.Id, baseIp.Id
	}
	newPos, mods, err := NewPosUpBetween(t.(string), baseIp.Id, up.BottomId, up.TopId, wFn)
	if err != nil {
		Return(r, err)
		return
	}

	ReturnAnyway(r, SaveModelWith(&data, map[string]interface{}{"Pos": newPos}), J{IpKey: mods, "Newer": data})
}

type RearrData struct {
	Base   int64 `json:"baseid"`
	Bottom int64 `json:"bid"`
	Top    int64 `json:"tid"`
}

// require base.ParseSearch ,bind RearrData
func Rearrange(t Table, data RearrData, r render.Render, searchFn func(db *gorm.DB) *gorm.DB) {
	_, mods, err := RearrAndReurnMods(t, data.Base, data.Bottom, data.Top, searchFn)
	ReturnAnyway(r, err, mods)
}

func ModIps(t Table, data []IdPos, r render.Render) {
	err := ToDb(t.(string), data)
	Return(r, err)
}

type Posx struct {
	Id1 int64 `json:",omitempty" binding:"required"`
	Id2 int64 `json:",omitempty" binding:"required"`
}

func Xpos(t Table, data Posx, r render.Render) {
	err := Exchange(t.(string), data.Id1, data.Id2)
	Return(r, err)
}

// required IpPos
func PosTop(t Table, data IdPos, r render.Render, searchFn func(db *gorm.DB) *gorm.DB) {
	ip, err := ToTop(t.(string), data.Id, searchFn)
	Return(r, err, ip)
}

// required IpPos
func PosBottom(t Table, data IdPos, r render.Render, searchFn func(db *gorm.DB) *gorm.DB) {
	mods, err := ToBottom(t.(string), data.Id, searchFn)
	Return(r, err, mods)
}

type Direction struct {
	Reverse bool `form:"reverse"`
}

// required IpPos Direction
func PosUpSingle(t Table, data IdPos, dir Direction, r render.Render, searchFn func(db *gorm.DB) *gorm.DB) {
	ips, err := IpBeforeOrAfterAnd(t.(string), !dir.Reverse, data.Pos, searchFn)
	if err != nil {
		Return(r, err)
		return
	}

	if len(ips) == 1 {
		if ips[0].Id == data.Id {
			Return(r)
			return
		}
		Return(r, errors.New("wrong request pos, need refresh"))
		return
	}

	Return(r, ExchangeIP(t.(string), ips...))
}
