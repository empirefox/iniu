//DB_DIALECT=mysql DB_URL="gorm:gorm@/gorm?charset=utf8&parseTime=True"
//DB_DIALECT=postgres DB_URL="user=gorm dbname=gorm sslmode=disable"
//DB_DIALECT=sqlite3 DB_URL=/tmp/gorm.DB go test
package sorting

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPosBaseUtils(t *testing.T) {
	Convey("indexWithId", t, func() {
		ips := []IdPos{{ID: 1, Pos: 5}, {ID: 2, Pos: 6}, {ID: 3, Pos: 7}}
		Convey("should find index", func() {
			So(indexWithId(ips, IdPos{ID: 2}), ShouldEqual, 1)
		})
		Convey("should not find index", func() {
			So(indexWithId(ips, IdPos{ID: 5}), ShouldEqual, -1)
		})
	})

	// 0 >>> -1 >>> 1 >>> -2 >>> 2 >>> -3
	Convey("nextDir", t, func() {
		Convey("should return correct sequence", func() {
			So(nextDir(0), ShouldEqual, -1)
			So(nextDir(1), ShouldEqual, -2)
			So(nextDir(2), ShouldEqual, -3)
			So(nextDir(3), ShouldEqual, -4)
			So(nextDir(4), ShouldEqual, -5)
			So(nextDir(5), ShouldEqual, -6)
			So(nextDir(6), ShouldEqual, -7)

			So(nextDir(-1), ShouldEqual, 1)
			So(nextDir(-2), ShouldEqual, 2)
			So(nextDir(-3), ShouldEqual, 3)
			So(nextDir(-4), ShouldEqual, 4)
			So(nextDir(-5), ShouldEqual, 5)
		})
	})

	Convey("pointer", t, func() {
		Convey("when no more space", func() {
			//dir:    x -1  0  1  2  3
			//index:  0  1  2  3  4  5
			//              ^
			//            iBase
			p := pointer{iBase: 2, size: 6}
			p.dir = 0
			So(p.hasNext(), ShouldBeTrue)
			So(p.next(), ShouldEqual, 1)
			p.dir = -1
			So(p.hasNext(), ShouldBeTrue)
			So(p.next(), ShouldEqual, 3)
			p.dir = 1
			So(p.hasNext(), ShouldBeTrue)
			So(p.next(), ShouldEqual, 4)
			p.dir = 2
			So(p.hasNext(), ShouldBeTrue)
			So(p.next(), ShouldEqual, 5)
		})

		Convey("when with more space", func() {
			//dir:   -2 -1  0  1  2  3
			//index:  0  1  2  3  4  5
			//              ^
			//            iBase
			p := pointer{iBase: 2, size: 6, topSpace: 6}
			p.dir = 0
			So(p.hasNext(), ShouldBeTrue)
			So(p.next(), ShouldEqual, 1)
			p.dir = -1
			So(p.hasNext(), ShouldBeTrue)
			So(p.next(), ShouldEqual, 3)
			p.dir = 1
			So(p.hasNext(), ShouldBeTrue)
			So(p.next(), ShouldEqual, 0)
		})
	})

	Convey("average", t, func() {
		Convey("should average normal", func() {
			ips1 := []IdPos{{ID: 1, Pos: 11}, {ID: 2, Pos: 8}, {ID: 3, Pos: 6}, {ID: 4, Pos: 5}, {ID: 5, Pos: 3}, {ID: 6, Pos: 1}}
			ips2 := []IdPos{{ID: 1, Pos: 11}, {ID: 2, Pos: 10}, {ID: 3, Pos: 9}, {ID: 4, Pos: 8}, {ID: 5, Pos: 7}, {ID: 6, Pos: 1}}
			result := []IdPos{{ID: 1, Pos: 11}, {ID: 2, Pos: 9}, {ID: 3, Pos: 7}, {ID: 4, Pos: 5}, {ID: 5, Pos: 3}, {ID: 6, Pos: 1}}
			average(ips1)
			So(ips1, ShouldResemble, result)
			average(ips2)
			So(ips2, ShouldResemble, result)
		})

		Convey("should average with step", func() {
			ips1 := []IdPos{{ID: 1, Pos: -1}, {ID: 2, Pos: 8}, {ID: 3, Pos: 6}, {ID: 4, Pos: 5}, {ID: 5, Pos: 3}, {ID: 6, Pos: 2}}
			ips2 := []IdPos{{ID: 1, Pos: -6}, {ID: 2, Pos: 10}, {ID: 3, Pos: 9}, {ID: 4, Pos: 8}, {ID: 5, Pos: 7}, {ID: 6, Pos: 2}}
			result := []IdPos{{ID: 1, Pos: 42}, {ID: 2, Pos: 34}, {ID: 3, Pos: 26}, {ID: 4, Pos: 18}, {ID: 5, Pos: 10}, {ID: 6, Pos: 2}}
			average(ips1)
			So(ips1, ShouldResemble, result)
			average(ips2)
			So(ips2, ShouldResemble, result)
		})
	})

	Convey("findDestSpace", t, func() {
		Convey("should find directly", func() {
			//pos:   11 10  6  5  4  3
			//index:  0  1  2  3  4  5
			//              ^
			//            iBase
			p := &pointer{iBase: 2, size: 6}
			ips := []IdPos{{ID: 1, Pos: 11}, {ID: 2, Pos: 10}, {ID: 3, Pos: 6}, {ID: 4, Pos: 5}, {ID: 5, Pos: 4}, {ID: 6, Pos: 3}}
			start, end, space, ok := findDestSpace(ips, p)
			So(start, ShouldEqual, end)
			So(space, ShouldEqual, 4)
			So(ok, ShouldBeTrue)
		})

		Convey("should find up", func() {
			//move:         v
			//pos:   11 10  6  5  4  3
			//index:  0  1  2  3  4  5
			//                 ^
			//               iBase
			p := &pointer{iBase: 3, size: 6}
			ips := []IdPos{{ID: 1, Pos: 11}, {ID: 2, Pos: 10}, {ID: 3, Pos: 6}, {ID: 4, Pos: 5}, {ID: 5, Pos: 4}, {ID: 6, Pos: 3}}
			start, end, space, _ := findDestSpace(ips, p)
			So(start, ShouldEqual, 2)
			So(end, ShouldEqual, 3)
			So(space, ShouldEqual, 4)
		})

		Convey("should find down", func() {
			//move:      v  v  v
			//pos:   11 10  9  8  4  3
			//index:  0  1  2  3  4  5
			//           ^
			//         iBase
			p := &pointer{iBase: 1, size: 6}
			ips := []IdPos{{ID: 1, Pos: 11}, {ID: 2, Pos: 10}, {ID: 3, Pos: 9}, {ID: 4, Pos: 8}, {ID: 5, Pos: 4}, {ID: 6, Pos: 3}}
			start, end, space, _ := findDestSpace(ips, p)
			So(start, ShouldEqual, 1)
			So(end, ShouldEqual, 4)
			So(space, ShouldEqual, -4)
		})

		Convey("should find with space", func() {
			//move:     v
			//pos:  (6)11 10  9  8  4  3
			//index:-1  0  1  2  3  4  5
			//             ^
			//           iBase
			p := &pointer{iBase: 1, size: 6, topSpace: 6}
			ips := []IdPos{{ID: 1, Pos: 11}, {ID: 2, Pos: 10}, {ID: 3, Pos: 9}, {ID: 4, Pos: 8}, {ID: 5, Pos: 4}, {ID: 6, Pos: 3}}
			start, end, space, _ := findDestSpace(ips, p)
			So(start, ShouldEqual, -1)
			So(end, ShouldEqual, 1)
			So(space, ShouldEqual, 6)
		})
	})

	Convey("handleNewPos", t, func() {
		Convey("should handle directly up", func() {
			//pos:   11 10  6  5  4  3
			//index:  0  1  2  3  4  5
			//              ^
			//            iBase
			ips := []IdPos{{ID: 1, Pos: 11}, {ID: 2, Pos: 10}, {ID: 3, Pos: 6}, {ID: 4, Pos: 5}, {ID: 5, Pos: 4}, {ID: 6, Pos: 3}}
			newPos, mods, err := handleNewPos(ips, 2, 0)
			So(err, ShouldBeNil)
			So(newPos, ShouldEqual, 8)
			So(mods, ShouldBeEmpty)
		})

		Convey("should handle up", func() {
			//move:         v               (    v )
			//pos:   11 10  6  5  4  3  >>> 10 8 6 5
			//index:  0  1  2  3  4  5
			//id:     1  2  3  4  5  6
			//                 ^
			//               iBase
			ips := []IdPos{{ID: 1, Pos: 11}, {ID: 2, Pos: 10}, {ID: 3, Pos: 6}, {ID: 4, Pos: 5}, {ID: 5, Pos: 4}, {ID: 6, Pos: 3}}
			newPos, mods, _ := handleNewPos(ips, 3, 0)
			So(newPos, ShouldEqual, 6)
			So(mods, ShouldResemble, []IdPos{{ID: 3, Pos: 8}})
		})

		Convey("should handle down", func() {
			//move:      v  v  v           (  v       )
			//pos:   11 10  9  8  4  3 >>> 11 9 8 6 5 4
			//index:  0  1  2  3  4  5
			//id:     1  2  3  4  5  6      1   2 3 4 5
			//           ^
			//         iBase
			ips := []IdPos{{ID: 1, Pos: 11}, {ID: 2, Pos: 10}, {ID: 3, Pos: 9}, {ID: 4, Pos: 8}, {ID: 5, Pos: 4}, {ID: 6, Pos: 3}}
			newPos, mods, _ := handleNewPos(ips, 1, 0)
			So(newPos, ShouldEqual, 9)
			So(mods, ShouldResemble, []IdPos{{ID: 2, Pos: 8}, {ID: 3, Pos: 6}, {ID: 4, Pos: 5}})
		})

		Convey("should handle limited topSpace", func() {
			//move:     v                    (      v  )
			//pos:  (6)11 10  9  8  4  3 >>> 17 14 12 10
			//index:-1  0  1  2  3  4  5
			//             ^
			//           iBase
			ips := []IdPos{{ID: 1, Pos: 11}, {ID: 2, Pos: 10}, {ID: 3, Pos: 9}, {ID: 4, Pos: 8}, {ID: 5, Pos: 4}, {ID: 6, Pos: 3}}
			newPos, mods, err := handleNewPos(ips, 1, 6)
			So(err, ShouldBeNil)
			So(newPos, ShouldEqual, 12)
			So(mods, ShouldResemble, []IdPos{{ID: 1, Pos: 14}})
		})

		Convey("should handle unlimited topSpace", func() {
			//move:       v                    (      v  )
			//pos:  (-1) 11 10  9  8  4  3 >>> 34 26 18 10
			//index: -1   0  1  2  3  4  5
			//              ^
			//            iBase
			ips := []IdPos{{ID: 1, Pos: 11}, {ID: 2, Pos: 10}, {ID: 3, Pos: 9}, {ID: 4, Pos: 8}, {ID: 5, Pos: 4}, {ID: 6, Pos: 3}}
			newPos, mods, _ := handleNewPos(ips, 1, -1)
			So(newPos, ShouldEqual, 18)
			So(mods, ShouldResemble, []IdPos{{ID: 1, Pos: 26}})
		})
	})
}

func TestPosBaseWithDb(t *testing.T) {
	newTable(9, 6, 15, 2, 3, 5)
	Convey("max", t, func() {
		ip, _ := max(xchgs)
		So(ip.Pos, ShouldEqual, 15)
	})

	Convey("min", t, func() {
		ip, _ := min(xchgs)
		So(ip.Pos, ShouldEqual, 2)
	})
}
