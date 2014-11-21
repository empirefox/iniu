package base

import (
	"errors"

	"github.com/jinzhu/gorm"
	un "github.com/tobyhede/go-underscore"

	. "github.com/empirefox/iniu/gorm/db"
	. "github.com/empirefox/iniu/gorm/mod"
)

var (
	AnyIp func(func(ip IdPos, i int) bool, []IdPos) bool
	IpKey = "Ips"
)

func init() {
	un.MakeAny(&AnyIp)
}

type IdPos struct {
	Id     int64
	Pos    int64
	Orderd `json:"-" sql:"-"`
}

func (ip *IdPos) ToMap() map[string]interface{} {
	return map[string]interface{}{"id": ip.Id, "pos": ip.Pos}
}

func IpById(t string, id int64) (ip IdPos, err error) {
	err = DB.Table(t).Where("id=?", id).First(&ip).Error
	return
}

func IpByPos(t string, pos int64, fns ...func(*gorm.DB) *gorm.DB) (ip IdPos, err error) {
	err = DB.Table(t).Scopes(fns...).Where("pos=?", pos).First(&ip).Error
	return
}

func IpBeforeOrAfterAnd(t string, isBefore bool, pos int64, fns ...func(*gorm.DB) *gorm.DB) ([]IdPos, error) {
	return beforeOrAfterAnd(t, true, pos, fns...)
}

func IpDb(t string, fns []func(*gorm.DB) *gorm.DB) *gorm.DB {
	return DB.Table(t).Scopes(fns...).Select("id,pos")
}

func ToDb(t string, ips []IdPos) error {
	tx := DB.Begin()
	for _, ip := range ips {
		if err := tx.Table(t).UpdateColumn(ip.ToMap()).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}

func Rearr(t string, idBase int64, fns ...func(*gorm.DB) *gorm.DB) (int64, error) {
	reserved, _, err := RearrAndReurnMods(t, idBase, 0, 0, fns...)
	return reserved, err
}

func RearrAndReurnMods(t string, idBase, idBottom, idTop int64, fns ...func(*gorm.DB) *gorm.DB) (int64, []IdPos, error) {
	ips, mods := []IdPos{}, []IdPos{}

	if err := IpDb(t, fns).Find(&ips).Error; err != nil {
		return -1, nil, err
	}

	var reserved int64 = -2
	size := len(ips)
	offset := false
	isMod := false
	for i := range ips {
		ip := ips[size-1-i]
		ip.Pos = int64(i)*step + 1
		if offset {
			ip.Pos += step
		}
		if idBase == ip.Id {
			offset = true
			reserved = ip.Pos
		}

		if idBottom == ip.Id {
			isMod = true
		}
		if isMod && idBottom|idTop != 0 && len(mods) <= secureSliceLen {
			mods = append(mods, ip)
		}
		if idTop == ip.Id {
			isMod = false
		}
	}
	if err := ToDb(t, ips); err != nil {
		return -1, nil, err
	}
	return reserved, mods, nil
}

func ExchangeIp(t string, ips ...IdPos) error {
	ips[0].Pos, ips[1].Pos = ips[1].Pos, ips[0].Pos
	return ToDb(t, ips)
}

func Exchange(t string, id1, id2 int64) error {
	if id1 == id2 || id1|id2 == 0 {
		return errors.New("wrong ids")
	}
	ips := []IdPos{}
	err := DB.Table(t).Where("id in (?)", []int64{id1, id2}).Find(&ips).Error
	if err != nil || len(ips) != 2 {
		return err
	}
	return ExchangeIp(t, ips...)
}

func NewPosUp(t string, idBase, fns ...func(*gorm.DB) *gorm.DB) (int64, []IdPos, error) {
	return NewPosUpBetween(t, idBase, idBase, idBase, fns...)
}

func NewPosUpBetween(t string, idBase, idBottom, idTop int64, fns ...func(*gorm.DB) *gorm.DB) (int64, []IdPos, error) {
	var baseIp, bottomIp, topIp, maxIp IdPos

	maxIp, err = max(t, fns...)
	if err != nil {
		// the list is empty now
		if err == gorm.RecordNotFound {
			return 1, nil, nil
		}
		return -2, nil, err
	}

	// insert up to max
	if maxIp.Id == idBase {
		return maxIp.Pos + step, nil
	}

	err := DB.Table(t).Where("id=?", idBase).First(&baseIp).Error
	if err != nil {
		return -1, nil, err
	}
	err = DB.Table(t).Where("id=?", idBottom).First(&bottomIp).Error
	if err != nil {
		return -1, nil, err
	}
	err = DB.Table(t).Where("id=?", idTop).First(&topIp).Error
	if err != nil {
		return -1, nil, err
	}

	base, bottom, top := baseIp.Pos, bottomIp.Pos, topIp.Pos

	if base > top || base < bottom {
		return -2, nil, errors.New("wrong parents")
	}

	ips := []IdPos{}
	fixedTop := top + 2*step + 1
	err = IpDb(t, fns).Where("pos BETWEEN ? AND ?", bottom, fixedTop).Find(&ips).Error
	if err != nil {
		return -2, nil, err
	}

	index := indexWithId(ips, base)
	if index == -1 {
		return -2, nil, errors.New("base pos has been removed")
	}

	topPos := ips[0].Pos
	// there is a lot of space up here but it's not the max one
	if topPos == bottom {
		return topPos + step, nil, nil
	}
	// the length run out of the cap
	if int64(len(ips)) >= topPos-bottom {
		return RearrAndReurnMods(t, idBase, bottomIp.Id, topIp.Id, fns...)
	}

	// compute topSpace to use handleNewPos
	topSpace := fixedTop - topPos
	if topPos >= maxIp.Pos {
		topSpace = -1
	}

	newPos, mods, ok := handleNewPos(ips, index, topSpace)
	if !ok {
		// maybe could not been called here,just for assurance
		return RearrAndReurnMods(t, idBase, bottomIp.Id, topIp.Id, fns...)
	}

	err = ToDb(t, mods)
	if err != nil {
		return -2, nil, err
	}

	return newPos, mods, nil
}

func ToTop(t string, id int64, fns ...func(*gorm.DB) *gorm.DB) (IdPos, error) {
	var newPos int64
	maxIp, err = max(t, fns...)
	if err != nil {
		if err == gorm.RecordNotFound {
			newPos = 1
			goto SAVE
		}
		return nil, err
	}
	if id == maxIp.Id {
		return maxIp, nil
	}
	newPos = maxIp.Pos + step

SAVE:
	ip := IdPos{Id: id, Pos: newPos}
	err = ToDb(t, []IdPos{ip})
	if err != nil {
		return nil, err
	}

	return ip, nil
}

func ToBottom(t string, id int64, fns ...func(*gorm.DB) *gorm.DB) ([]IdPos, error) {
	minIp, err := min(t, fns...)
	if err != nil {
		return nil, err
	}
	if id == minIp.Id {
		return []IdPos{minIp}, nil
	}

	if minIp.Pos > 1 {
		return []IdPos{{Id: id, Pos: (minIp.Pos + 1) / 2}}
	}

	newPos, mods, err := NewPosUp(t, minIp.Id, fns...)
	if err != nil {
		return nil, err
	}

	minIp.Pos = newPos

	x := []IdPos{minIp, IdPos{Id: id, Pos: (newPos + 1) / 2}}
	err = ToDb(t, x)
	if err != nil {
		return nil, err
	}

	return append(mods, x...), nil
}
