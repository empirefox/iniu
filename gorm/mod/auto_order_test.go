package mod

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	. "github.com/empirefox/iniu/gorm/db"
)

var (
	autoOrderd    = "auto_orderds"
	nonAutoOrderd = "non_auto_orderds"
	autod         = "autods"
)

type Autod struct {
	Id  int `order:"auto"`
	Pos int
}

type AutoOrderd struct {
	Id  int `order:"pos desc"`
	Pos int
}

type NonAutoOrderd struct {
	Id  int
	Pos int
}

// newAutoOrderd(2, 3, 4, 5, 6, 7, 8, 9)
func newAutod(ps ...int) {
	DB.DropTableIfExists(Autod{})
	DB.CreateTable(Autod{})

	for _, p := range ps {
		DB.Save(&Autod{Pos: p})
	}
}

func newNonAutoOrderd(ps ...int) {
	DB.DropTableIfExists(NonAutoOrderd{})
	DB.CreateTable(NonAutoOrderd{})

	for _, p := range ps {
		DB.Save(&NonAutoOrderd{Pos: p})
	}
}

func newAutoOrderd(ps ...int) {
	DB.DropTableIfExists(AutoOrderd{})
	DB.CreateTable(AutoOrderd{})

	for _, p := range ps {
		DB.Save(&AutoOrderd{Pos: p})
	}
}

func TestAutoOrderCallback(t *testing.T) {
	Convey("OrderCallback", t, func() {
		Convey("Auto Orderd", func() {
			AutoOrder()
			//            1  2  3  4
			newAutoOrderd(8, 6, 7, 9)
			var aos []AutoOrderd
			err := DB.Table(autoOrderd).Find(&aos).Error
			So(err, ShouldBeNil)
			So(aos, ShouldResemble, []AutoOrderd{{Id: 4, Pos: 9}, {Id: 1, Pos: 8}, {Id: 3, Pos: 7}, {Id: 2, Pos: 6}})
			NoAutoOrder()
		})

		Convey("Not Auto Orderd", func() {
			//            1  2  3  4
			newAutoOrderd(8, 6, 7, 9)
			var aos []AutoOrderd
			err := DB.Table(autoOrderd).Find(&aos).Error
			So(err, ShouldBeNil)
			So(aos, ShouldResemble, []AutoOrderd{{Id: 1, Pos: 8}, {Id: 2, Pos: 6}, {Id: 3, Pos: 7}, {Id: 4, Pos: 9}})
		})

		Convey("Override Orderd", func() {
			AutoOrder()
			//            1  2  3  4
			newAutoOrderd(8, 6, 7, 9)
			var aos []AutoOrderd
			err := DB.Table(autoOrderd).Order("pos asc").Find(&aos).Error
			So(err, ShouldBeNil)
			So(aos, ShouldResemble, []AutoOrderd{{Id: 2, Pos: 6}, {Id: 3, Pos: 7}, {Id: 1, Pos: 8}, {Id: 4, Pos: 9}})
			NoAutoOrder()
		})

		Convey("No Orderd tag", func() {
			AutoOrder()
			//            1  2  3  4
			newNonAutoOrderd(8, 6, 7, 9)
			var aos []NonAutoOrderd
			err := DB.Table(nonAutoOrderd).Find(&aos).Error
			So(err, ShouldBeNil)
			So(aos, ShouldResemble, []NonAutoOrderd{{Id: 1, Pos: 8}, {Id: 2, Pos: 6}, {Id: 3, Pos: 7}, {Id: 4, Pos: 9}})
			NoAutoOrder()
		})

		Convey("Orderd tag = auto", func() {
			AutoOrder()
			//            1  2  3  4
			newAutod(8, 6, 7, 9)
			var aos []Autod
			err := DB.Table(autod).Find(&aos).Error
			So(err, ShouldBeNil)
			So(aos, ShouldResemble, []Autod{{Id: 4, Pos: 9}, {Id: 1, Pos: 8}, {Id: 3, Pos: 7}, {Id: 2, Pos: 6}})
			NoAutoOrder()
		})
	})
}
