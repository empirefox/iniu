//GORM_DIALECT=mysql DB_URL="gorm:gorm@/gorm?charset=utf8&parseTime=True"
//GORM_DIALECT=postgres DB_URL="user=gorm dbname=gorm sslmode=disable" PORT=8080 goconvey
//GORM_DIALECT=sqlite3 DB_URL=/tmp/gorm.DB go test
package web

import (
	"flag"
	"net/http"
	"testing"

	"github.com/go-martini/martini"
	. "github.com/smartystreets/goconvey/convey"

	. "github.com/empirefox/iniu/base"
	. "github.com/empirefox/iniu/convey"
	. "github.com/empirefox/iniu/security"
)

func TestLink(t *testing.T) {
	flag.Set("superdo", "true")
	flag.Parse()
	defer func() {
		flag.Set("superdo", "false")
		flag.Parse()
	}()
	Convey("Link", t, func() {
		newWebs("n1")
		req := func(m martini.Router) (*http.Request, error) {
			m.(*martini.ClassicMartini).Use(Prepare(Config{}, false))
			LinkAll(m, Web{})
			return http.NewRequest("GET", "/Web/mf", nil)
		}
		res := map[string]interface{}{
			"Name": "Web",
			"Fields": []NameOnlyField{
				{Name: "Id"},
				{Name: "Name"},
				{Name: "Childs"},
			},
		}
		So(req, ShouldResponseOk, res)
	})
}
