//需设置环境变量
//GORM_DIALECT=mysql DB_URL="gorm:gorm@/gorm?charset=utf8&parseTime=True"
//GORM_DIALECT=postgres DB_URL="user=gorm dbname=gorm sslmode=disable"
//GORM_DIALECT=sqlite3 DB_URL=/tmp/gorm.DB go test
package base

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	. "github.com/empirefox/iniu/gorm/db"
)

func TestPosBasePublicUtils(t *testing.T) {
	newTable(1, 3, 4, 5, 8, 9, 12)
	// id:   1  2  3  4  5  6   7
	//                ^
	Convey("测试IpBeforeOrAfterAnd", t, func() {
		ips, _ := IpBeforeOrAfterAnd(xchgs, true, 5)
		So(ips, ShouldResemble, []IdPos{{Id: 4, Pos: 5}, {Id: 5, Pos: 8}})

		ips, _ = IpBeforeOrAfterAnd(xchgs, false, 5)
		So(ips, ShouldResemble, []IdPos{{Id: 4, Pos: 5}, {Id: 3, Pos: 4}})
	})
}

func TestToDb(t *testing.T) {
	newTable(1, 2, 3)
	Convey("测试ToDb", t, func() {
		ips := []IdPos{{Id: 3, Pos: 7}, {Id: 2, Pos: 3}, {Id: 1, Pos: 1}}
		So(ToDb(xchgs, ips), ShouldBeNil)

		result := []IdPos{}
		DB.Table(xchgs).Find(&result)
		So(result, ShouldResemble, ips)
	})
}

func TestRearr(t *testing.T) {
	Convey("测试重排列", t, func() {
		newTable(9, 8, 5, 4, 3, 2)
		// id    1  2  3  4  5  6
		//             ^
		Convey("给定now id", func() {
			pos, err := Rearr(xchgs, 3)
			So(err, ShouldBeNil)
			So(pos, ShouldEqual, 4*8+1)
			So(ps(), ShouldResemble, []int64{49, 41, 25, 17, 9, 1})
		})
		Convey("不给定now id", func() {
			pos, _ := Rearr(xchgs, -3)
			So(pos, ShouldEqual, -2)
			So(ps(), ShouldResemble, []int64{41, 33, 25, 17, 9, 1})
		})
		Convey("should return mods between", func() {
			pos, mods, _ := RearrAndReurnMods(xchgs, -3, 2, 5)
			So(pos, ShouldEqual, -2)
			So(mods, ShouldResemble, []IdPos{{Id: 5, Pos: 9}, {Id: 4, Pos: 17}, {Id: 3, Pos: 25}, {Id: 2, Pos: 33}})
		})
	})
}

func TestExchange(t *testing.T) {
	//       v     v
	//   id: 1  2  3
	newTable(1, 3, 4)
	Convey("交换位置测试", t, func() {
		Convey("直接交换", func() {
			//id: 3  2  1  4  5  6  7   8   9   10  11
			//   {1, 3, 4, 5, 8, 9, 12, 14, 16, 17, 18}
			So(Exchange("xchgs", 1, 3), ShouldBeNil)
			e := Xchg{}
			DB.Table("xchgs").First(&e, 1)
			So(e.Pos, ShouldEqual, 4)

			e = Xchg{}
			DB.Table("xchgs").First(&e, 3)
			So(e.Pos, ShouldEqual, 1)
		})
	})
}

func TestNewPosUpBetween(t *testing.T) {
	Convey("测试NewPosUpBetween", t, func() {
		Convey("normal move up", func() {
			newTable(19, 18, 17, 6, 5, 4, 3, 2)
			// range          v              v         <-
			//       19, 18, 17, 6, 5, 4, 3, 2  >>> 17 13 9 5
			// insert 1   2   3  4 ^5  6  7  8            ^
			pos, ips, err := NewPosUpBetween(xchgs, 5, 8, 3)
			So(err, ShouldBeNil)
			So(pos, ShouldEqual, 9)
			So(ips, ShouldResemble, []IdPos{{Id: 4, Pos: 13}})
		})

		Convey("normal move down", func() {
			newTable(19, 18, 17, 16, 5, 4, 3, 2)
			// range      v                v              ->
			//       19, 18, 17, 16, 5, 4, 3, 2  >>> 17 13 9 5
			// insert 1   2   3 ^ 4  5  6  7  8          ^
			pos, ips, _ := NewPosUpBetween(xchgs, 4, 7, 2)
			So(pos, ShouldEqual, 13)
			So(ips, ShouldResemble, []IdPos{{Id: 4, Pos: 9}})
		})

		Convey("with no limit and length full", func() {
			newTable(9, 8, 7, 6, 5, 4, 3, 2)
			// range      v                v
			//        9,  8,  7,  6, 5, 4, 3, 2
			// insert 1   2   3 ^ 4  5  6  7  8
			pos, ips, err := NewPosUpBetween(xchgs, 4, 7, 2)
			So(err, ShouldBeNil)
			So(pos, ShouldEqual, 41)
			So(ips, ShouldResemble, []IdPos{{Id: 7, Pos: 9}, {Id: 6, Pos: 17}, {Id: 5, Pos: 25}, {Id: 4, Pos: 33}, {Id: 3, Pos: 49}, {Id: 2, Pos: 57}})
		})

		Convey("with no limit and length not full", func() {
			newTable(10, 9, 8, 7, 6, 5, 2, 1)
			// range       v                v           <-    <-    <-
			//        10,  9,  8,  7, 6, 5, 2, 1  >>>   39    31    23   15   7
			// insert  1   2   3 ^ 4  5  6  7  8         1     2     3    ^
			pos, ips, err := NewPosUpBetween(xchgs, 4, 7, 2)
			So(err, ShouldBeNil)
			So(pos, ShouldEqual, 15)
			So(ips, ShouldResemble, []IdPos{{Id: 1, Pos: 39}, {Id: 2, Pos: 31}, {Id: 3, Pos: 23}})
		})

		Convey("with limit", func() {
			newTable(99, 9, 8, 7, 6, 5, 2, 1)
			// range (17) v                v             <-    <-
			//       99,  9,  8,  7, 6, 5, 2, 1  >>> 26  21    16    11   7
			// insert 1   2   3 ^ 4  5  6  7  8       s   2     3    ^
			pos, ips, err := NewPosUpBetween(xchgs, 4, 7, 2)
			So(err, ShouldBeNil)
			So(pos, ShouldEqual, 11)
			So(ips, ShouldResemble, []IdPos{{Id: 2, Pos: 21}, {Id: 3, Pos: 16}})
		})

		Convey("NewPosUp", func() {
			newTable(9, 8, 7, 6, 5, 4, 3, 2)
			// range
			//       19, 18, 17, 16, 5, 4, 3, 2  >>>  65  57   49   41   33
			// insert 1   2   3 ^ 4  5  6  7  8        1   2    3    ^    4
			pos, ips, err := NewPosUp(xchgs, 4)
			So(err, ShouldBeNil)
			So(pos, ShouldEqual, 41)
			So(ips, ShouldResemble, []IdPos{{Id: 4, Pos: 33}, {Id: 3, Pos: 49}, {Id: 2, Pos: 57}, {Id: 1, Pos: 65}})
		})
	})
}

func TestToTop(t *testing.T) {
	// to top            v
	newTable(1, 3, 4, 5, 8, 9, 12)
	// index 1  2  3  4  5  6   7
	Convey("置底", t, func() {
		Convey("not in bondry", func() {
			// place 8 to top
			ip, err := ToTop(xchgs, 5)
			So(err, ShouldBeNil)
			So(ip.Pos, ShouldEqual, 20)
			e := Xchg{}
			DB.Table(xchgs).First(&e, 5)
			So(e.Pos, ShouldEqual, 20)
		})
	})
}

func TestToBottom(t *testing.T) {
	newTable(1, 3, 4, 5, 8, 9, 12, 14, 16, 17, 18)
	// index 1  2  3  4  5  6   7   8   9  10  11
	//                   ^
	Convey("置底", t, func() {
		Convey("not in bondry", func() {
			// place 8 to bottom
			ips, err := ToBottom(xchgs, 5)
			So(err, ShouldBeNil)
			So(ips, ShouldResemble, []IdPos{{Id: 1, Pos: 2}, {Id: 5, Pos: 1}})
			e := Xchg{}
			DB.Table(xchgs).Last(&e, 5)
			So(e.Pos, ShouldEqual, 1)
		})
	})
}

func ps() (s []int64) {
	DB.Table(xchgs).Order("pos desc").Pluck("pos", &s)
	return
}
