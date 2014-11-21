//需设置环境变量
//GORM_DIALECT=mysql DB_URL="gorm:gorm@/gorm?charset=utf8&parseTime=True"
//GORM_DIALECT=postgres DB_URL="user=gorm dbname=gorm sslmode=disable"
//GORM_DIALECT=sqlite3 DB_URL=/tmp/gorm.DB go test
package base

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	. "github.com/empirefox/iniu/gorm/postgres"
)

type Xchg struct {
	Id  int64
	Pos int64
}

func newTable(ps ...int64) {
	DB.DropTableIfExists(Xchg{})
	DB.CreateTable(Xchg{})

	for _, p := range ps {
		DB.Save(&Xchg{Pos: p})
	}
}

func TestNewUpPosBetweenWithRearr(t *testing.T) {
	Convey("测试需要重排的插入", t, func() {

		newTable(2, 3, 4, 5, 6, 7, 8, 9)
		Convey("直接插入", func() {
			pos, ips := NewUpPosBetween("xchgs", 5, 3, 8)
			So(pos, ShouldEqual, 3*8+1+4)
			So(len(ips), ShouldEqual, 6)
			So(ips[5].Pos, ShouldEqual, 6*8+1)
		})
	})
}

func TestNewUpPos(t *testing.T) {
	Convey("测试生成新位置", t, func() {

		newTable(1, 3, 4, 5, 8, 9, 12, 14, 16, 17, 18)
		Convey("直接插入", func() {
			pos, ips := NewUpPos("xchgs", 1)
			So(pos, ShouldEqual, 2)
			So(ips, ShouldBeEmpty)
		})

		Convey("向下调整插入", func() {
			pos, ips := NewUpPos("xchgs", 3)
			So(pos, ShouldEqual, 3)
			So(len(ips), ShouldEqual, 1)
			So(ips[0].Pos, ShouldEqual, 2)
		})

		Convey("向上调整插入", func() {
			pos, ips := NewUpPos("xchgs", 4)
			So(pos, ShouldEqual, 5)
			So(len(ips), ShouldEqual, 1)
			So(ips[0].Pos, ShouldEqual, 6)
		})

		Convey("最顶上插入", func() {
			pos, ips := NewUpPos("xchgs", 18)
			So(pos, ShouldEqual, 26)
			So(ips, ShouldBeEmpty)
		})
	})
}

func TestRearrange(t *testing.T) {
	Convey("测试重排列", t, func() {
		newTable(2, 3, 4, 5, 8, 9)
		Convey("给定now id", func() {
			pos := Rearrange("xchgs", 3)
			So(pos, ShouldEqual, 2*8+1)
			So(ps(), ShouldResemble, []int64{41, 33, 25, 17, 9, 1})
		})
		Convey("不给定now id", func() {
			pos := Rearrange("xchgs", -3)
			So(pos, ShouldEqual, -1)
			So(ps(), ShouldResemble, []int64{41, 33, 25, 17, 9, 1})
		})
	})
}

func TestPageCount(t *testing.T) {
	newTable(0, 7, 10, 16, 17, 21, 22, 23, 30, 38)
	result := map[int64]int64{
		100: 1, 11: 1, 10: 1,
		9: 2, 8: 2, 7: 2, 6: 2, 5: 2,
		4: 3,
		3: 4,
		2: 5,
		1: 10,
	}

	Convey("测试分页数量", t, func() {
		for k, v := range result {
			func(k, v int64) {
				Convey(fmt.Sprintf("每页数量[%d]时，应该有[%d]页", k, v), func() {
					c, _ := PageCount("xchgs", k)
					So(c, ShouldEqual, v)
				})
			}(k, v)
		}
	})
}

//TestComputeDescTargetSpace
//type TargetSpace struct {
//	I      int
//	Adjust int64
//	Ok     bool
//}

//func newTargetSpage(i int, adjust int64, ok bool) TargetSpace {
//	return TargetSpace{i, adjust, ok}
//}

//func TestComputeDescTargetSpace(t *testing.T) {
//	newTable(0, 7, 10, 16, 17, 21, 22, 23, 30, 38)
//	success := map[int]TargetSpace{
//		1: {1, 4, true},
//		2: {2, 3, true},
//		3: {2, 3, true},
//		4: {5, -2, true},
//		5: {5, 2, true},
//		6: {5, 2, true},
//		7: {7, 3, true},
//		8: {8, 1, true},
//		9: {9, 3, true},
//	}
//	ms := []IdPos{}
//	DB.Table("xchgs").Order("pos desc").Find(&ms)
//	Convey("计算目标调整位置及大小", t, func() {
//		for k, v := range success {
//			func(k int, v TargetSpace) {
//				Convey(fmt.Sprintf("[%d]的计算结果应该是[%d]", k, v), func() {
//					result := newTargetSpage(computeDescTargetSpace(ms, k))
//					So(result, ShouldResemble, v)
//				})
//			}(k, v)
//		}
//	})
//}

func TestTop(t *testing.T) {
	newTable(1, 3, 4, 5, 8, 9, 12)
	Convey("置底", t, func() {
		Convey("not in bondry", func() {
			// place 8 to top
			pos, ok := Top("xchgs", 5)
			So(ok, ShouldBeTrue)
			So(pos, ShouldEqual, 20)
			e := Xchg{}
			DB.Table("xchgs").First(&e, 5)
			So(e.Pos, ShouldEqual, 20)
		})
	})
}

func TestBottom(t *testing.T) {
	newTable(1, 3, 4, 5, 8, 9, 12, 14, 16, 17, 18)
	Convey("置底", t, func() {
		Convey("not in bondry", func() {
			// place 8 to bottom
			ips, ok := Bottom("xchgs", 5)
			So(ok, ShouldBeTrue)
			So(ips[0].Id, ShouldEqual, 1)
			e := Xchg{}
			DB.Table("xchgs").Last(&e, 5)
			So(e.Pos, ShouldEqual, 1)
		})
	})
}

func TestExchange(t *testing.T) {
	//   id: 1  2  3  4  5  6  7   8   9   10  11
	newTable(1, 3, 4, 5, 8, 9, 12, 14, 16, 17, 18)
	Convey("交换位置测试", t, func() {
		Convey("直接交换", func() {
			//id: 3  2  1  4  5  6  7   8   9   10  11
			//   {1, 3, 4, 5, 8, 9, 12, 14, 16, 17, 18}
			So(Exchange("xchgs", 1, 3), ShouldBeTrue)
			e := Xchg{}
			DB.Table("xchgs").First(&e, 1)
			So(e.Pos, ShouldEqual, 4)

			e = Xchg{}
			DB.Table("xchgs").First(&e, 3)
			So(e.Pos, ShouldEqual, 1)
		})
	})
}

func ps() (s []int64) {
	DB.Table("xchgs").Order("pos desc").Pluck("pos", &s)
	return
}

func randoms(l int, maxStep int64) []int64 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	rs := make([]int64, l, l)
	for i := 0; i < l; i++ {
		step := r.Int63n(maxStep)
		if i != 0 && step == 0 {
			i--
			continue
		}
		if i == 0 {
			rs[i] = step
		} else {
			rs[i] = rs[i-1] + step
		}
	}
	return rs
}
