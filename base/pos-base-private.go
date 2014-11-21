package base

import (
	"errors"

	"github.com/jinzhu/gorm"
	un "github.com/tobyhede/go-underscore"

	. "github.com/empirefox/iniu/gorm/db"
	. "github.com/empirefox/iniu/gorm/mod"
)

var (
	step           int64 = 8
	secureSliceLen int   = 100
)

func init() {
	un.MakeAny(&AnyIp)
}

func indexWithId(ips []IdPos, tar IdPos) int {
	index := -1

	indexOfFunc := func(ip IdPos, i int) bool {
		if ip.Id == tar.Id {
			index = i
			return true
		}
		return false
	}

	AnyIp(indexOfFunc, ips)
	return index
}

type pointer struct {
	iBase    int
	dir      int
	size     int
	topSpace int64
}

// 0 >>> -1 >>> 1 >>> -2 >>> 2 >>> -3
func nextDir(dir int) int {
	if dir < 0 {
		return ^dir + 1
	} else {
		return ^dir
	}
}

func (p *pointer) index() int {
	return p.iBase + p.dir
}

func (p *pointer) hasNext() bool {
	dir := nextDir(p.dir)
	return p.isInRange(dir) || p.isInRange(nextDir(dir))
}

// maybe return -1, repretation can get more space
func (p *pointer) next() int {
	p.dir = nextDir(p.dir)
	if p.isInRange(p.dir) {
		return p.index()
	}
	p.dir = nextDir(p.dir)
	return p.index()
}

func (p *pointer) isInRange(dir int) bool {
	index := p.iBase + dir
	if index == -1 && p.topSpace != 0 {
		return -1
	}
	return !(index < 0 || index > p.size-2)
}

// must assure the ips is legal
func average(ips []IdPos) {
	size := len(ips)
	top := ips[0].Pos
	bottom := ips[size-1].Pos

	var getStep = func(i int) int64 {
		if top > 0 {
			return i * (top - bottom) / (size - 1)
		}
		return i * step
	}
	for i := range ips {
		ips[size-1-i].Pos = bottom + getStep(i)
	}
}

func computeDestOffset(ips []IdPos, p *pointer) (int, int, int64, bool) {
	for p.hasNext() {
		index := p.next()

		if index == -1 {
			return -1, p.iBase, p.topSpace, true
		}

		space := ips[index].Pos - ips[index+1].Pos
		if space > 1 {
			if index < p.iBase {
				// go up
				return index + 1, p.iBase, space / 2, true
			}
			// go down
			return p.iBase, index + 1, -space / 2, true
		}
	}
	return -1, -1, -1, false
}

// topSpace: 0 >>> no more space, traded as normal
//           + >>> hase space to exceed
//           - >>> it is the first page, can set standard step pos without limit
func handleNewPos(ips []IdPos, iBase int, topSpace int64) (int64, []IdPos, error) {
	p := &pointer{iBase: iBase, topSpace: topSpace}
	start, end, os, ok := computeDestOffset(ips, p)
	if !ok {
		return -1, errors.New("need rearrange")
	}

	useTopSpace := false
	if start == -1 {
		useTopSpace = true
		start = 1
	}

	if end < start {
		return -1, nil, errors.New("end should be bigger than start")
	}

	var target, from, to int
	capacity := end - start + 4
	// mock a fixed slice to adjust pos from 1 to size-2
	mods := make([]IdPos, capacity-1, capacity)
	if os > 0 {
		// offset to up
		if useTopSpace {
			// exceed the origin top
			mods = append(mods, IdPos{Pos: p.topSpace})
		}
		from = 1
		mods = append(mods, ips[start-1:end]...)
		to = len(mods)
		target = to
		mods = append(mods, IdPos{}, ips[iBase])
	} else {
		// offset to down
		if iBase == 0 {
			mods = append(mods, IdPos{Pos: ips[iBase].Pos + 1})
		} else {
			mods = append(mods, ips[iBase-1])
		}
		target = len(mods)
		mods = append(mods, IdPos{})
		from = len(mods)
		mods = append(mods, ips[start:end+1]...)
		to = len(mods) - 1
	}
	average(mods)
	return mods[target].Pos, mods[from:to], true
}

func max(t string, fns ...func(*gorm.DB) *gorm.DB) (ip IdPos, err error) {
	err = IpDb(t, fns).First(&ip).Error
	return
}

func min(t string, fns ...func(*gorm.DB) *gorm.DB) (ip IdPos, err error) {
	err = IpDb(t, fns).Order("pos asc").First(&ip).Error
	return
}

func beforeOrAfterAnd(t string, isBefore bool, pos int64, fns ...func(*gorm.DB) *gorm.DB) ([]IdPos, error) {
	q := "pos >= ?"
	if !isBefore {
		q = "pos <= ?"
	}
	ips := []IdPos{}
	err := DB.Table(t).Scopes(fns...).Where(q, pos).Limit(2).Find(&ips).Error
	return ips, err
}
