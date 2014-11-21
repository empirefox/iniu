package mock

import (
	"testing"

	"github.com/erikstmartin/go-testdb"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	. "github.com/smartystreets/goconvey/convey"
)

var actualdb gorm.DB

func init() {
	actualdb, _ = gorm.Open("sqlite3", "/tmp/gorm.db")
}

type TableForm struct {
	Name  string `json:",omitempty"`
	Title string `json:",omitempty"`
}

func TestEnd(t *testing.T) {
	Convey("测试匹配", t, func() {
		db := DB.Table("t1").Select("name,title").Order("pos desc")
		bdb := NewBackend(db)

		// the injected value is the mock returned value, so must equal with the actual type
		tfs := []TableForm{}
		columns := []string{"name", "title"}
		result := `
		Bu,BU
		Cu,CU
		`

		bdb.StubQuery(&tfs, testdb.RowsFromCSVString(columns, result))

		tfs2 := []TableForm{}
		actualdb.Table("t1").Select("name,title").Order("pos desc").Find(&tfs2)

		So(tfs2[0].Name, ShouldEqual, "Bu")
		So(tfs2[1].Title, ShouldEqual, "CU")
	})
}
