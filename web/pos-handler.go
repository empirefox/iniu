package web

import (
	"errors"
	"reflect"

	"github.com/jinzhu/gorm"
	"github.com/martini-contrib/render"

	. "github.com/empirefox/iniu/base"
)

type SaveUpData struct {
	BottomId int64  `form:"b"`
	TopId    int64  `form:"t"`
	Cp       string `form:"cp"`
}

func SaveUp(t Table, db *gorm.DB, data Model, up SaveUpData, r render.Render) {
	if !HasPos(t) {
		Return(r, errors.New("have no pos here"))
		return
	}

	wFn, err := CpScopeFn(data, up.Cp)
	if err != nil {
		Return(r, err)
		return
	}

	// fix client request for the first time not saved
	var total int64
	total, err = Total(db, wFn)
	if err != nil {
		Return(r, err)
		return
	}
	if total == 0 {
		ReturnAnyway(r, SaveModelWith(db, &data, map[string]interface{}{"Pos": int64(1)}), J{"Newer": data})
		return
	}

	basePos := reflect.ValueOf(data).FieldByName("Pos").Int()

	var baseIp IdPos
	baseIp, err = IpByPosOrMax(t.(string), basePos, wFn)
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

	ReturnAnyway(r, SaveModelWith(db, &data, map[string]interface{}{"Pos": newPos}), J{IpKey: mods, "Newer": data})
}

type RearrData struct {
	Base   int64 `json:"baseid"`
	Bottom int64 `json:"bid"`
	Top    int64 `json:"tid"`
}

// PUT require base.ParseSearch ,bind RearrData
func Rearrange(t Table, data RearrData, r render.Render, searchFn func(db *gorm.DB) *gorm.DB) {
	_, mods, err := RearrAndReurnMods(t.(string), data.Base, data.Bottom, data.Top, searchFn)
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
// return the origin other model to take the Pos info to client
func PosUpSingle(t Table, db *gorm.DB, data IdPos, dir Direction, r render.Render, searchFn func(db *gorm.DB) *gorm.DB) {
	ips, err := IpBeforeOrAfterAnd(t.(string), !dir.Reverse, data.Pos, searchFn)
	if err != nil {
		Return(r, err)
		return
	}

	if len(ips) == 1 {
		if ips[0].Id == data.Id {
			Return(r, errors.New("already on top or bottom"))
			return
		}
		Return(r, errors.New("wrong request pos, need refresh"))
		return
	}

	var otherId int64 = -1
	for _, ip := range ips {
		if ip.Id != data.Id {
			otherId = ip.Id
			break
		}
	}
	if otherId == -1 {
		Return(r, errors.New("otherId should be found"))
		return
	}

	other := New(t)
	err = db.Where("id=?", otherId).First(other).Error
	if err != nil {
		Return(r, err)
		return
	}

	err = ExchangeIp(t.(string), ips...)
	Return(r, err, other)
}
