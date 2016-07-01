package sorting

import (
	"errors"

	"github.com/jinzhu/gorm"
)

var (
	step           int64 = 8
	secureSliceLen int   = 100
)

func indexWithId(ips []IdPos, tar IdPos) int {
	for i, ip := range ips {
		if ip.ID == tar.ID {
			return i
		}
	}
	return -1
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
	if index == 0 && p.topSpace != 0 {
		return true
	}
	return !(index < 1 || index > p.size-1)
}

// must assure the ips is legal
func average(ips []IdPos) {
	size := len(ips)
	top := ips[0].Pos
	bottom := ips[size-1].Pos

	var getStep = func(i_in int) int64 {
		i := int64(i_in)
		if top > 0 {
			return i * (top - bottom) / (int64(size) - 1)
		}
		return i * step
	}
	for i := range ips {
		ips[size-1-i].Pos = bottom + getStep(i)
	}
}

// return moved_first_index moved_last_index(be care that the last one is not included) space_with_direction success?
func findDestSpace(ips []IdPos, p *pointer) (int, int, int64, bool) {
	for index := p.iBase; p.hasNext(); index = p.next() {
		if index == 0 {
			return -1, p.iBase, p.topSpace, true
		}

		space := ips[index-1].Pos - ips[index].Pos
		if space > 1 {
			if index <= p.iBase {
				// go up
				return index, p.iBase, space, true
			}
			// go down
			return p.iBase, index, -space, true
		}
	}
	return -1, -1, -1, false
}

// topSpace: 0 >>> no more space, traded as normal
//           + >>> hase space to exceed
//           - >>> it is the first page, can set standard step pos without limit
func handleNewPos(ips []IdPos, iBase int, topSpace int64) (int64, []IdPos, error) {
	p := &pointer{iBase: iBase, size: len(ips), topSpace: topSpace}
	start, end, space, ok := findDestSpace(ips, p)
	if !ok {
		return -1, nil, errors.New("need rearrange")
	}

	if end < start {
		return -1, nil, errors.New("end should be bigger than start")
	}

	if start == end {
		return (ips[iBase].Pos + ips[iBase-1].Pos) / 2, nil, nil
	}

	useTopSpace := false
	if start == -1 {
		useTopSpace = true
		start = 1
	}

	var target, from, to int
	capacity := end - start + 3
	// mock a fixed slice to adjust pos from 1 to size-2
	mods := make([]IdPos, 0, capacity)
	if space > 0 || useTopSpace {
		// offset to up
		if useTopSpace {
			// exceed the origin top
			exceedPos := p.topSpace
			if space > 0 {
				exceedPos = ips[0].Pos + p.topSpace
			}
			mods = append(mods, IdPos{Pos: exceedPos})
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
	return mods[target].Pos, mods[from:to], nil
}

func max(t string, fns ...func(*gorm.DB) *gorm.DB) (ip IdPos, err error) {
	err = OrderIpDb(t, true, fns).First(&ip).Error
	return
}

func min(t string, fns ...func(*gorm.DB) *gorm.DB) (ip IdPos, err error) {
	err = OrderIpDb(t, false, fns).First(&ip).Error
	return
}
