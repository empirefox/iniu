package security

import (
	"testing"
	"time"

	. "github.com/jinzhu/copier"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFormCopy(t *testing.T) {
	Convey("FormCopy", t, func() {
		simpleForm := Form{
			Id:        2,
			Name:      "name",
			Pos:       9,
			Newid:     1,
			CreatedAt: time.Now(),
		}
		resultJForm := JsonForm{
			Name: "name",
			Pos:  9,
		}

		simpleFields := []Field{
			{
				Id:        1,
				Name:      "field1",
				Pos:       1,
				CreatedAt: time.Now(),
			},
			{
				Id:        2,
				Name:      "field2",
				Pos:       2,
				CreatedAt: time.Now(),
			},
		}
		resultJFields := []JsonField{
			{
				Name: "field1",
				Pos:  1,
			},
			{
				Name: "field2",
				Pos:  2,
			},
		}

		Convey("should copy without Fields", func() {
			var jf JsonForm
			err := Copy(&jf, &simpleForm)
			So(err, ShouldBeNil)
			So(jf, ShouldResemble, resultJForm)
		})

		Convey("should copy with Fields", func() {
			simpleForm.Fields = simpleFields
			resultJForm.JsonFields = resultJFields
			var jf JsonForm
			err := Copy(&jf, &simpleForm)
			So(err, ShouldBeNil)
			So(jf, ShouldResemble, resultJForm)
		})
	})
}
