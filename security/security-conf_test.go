//GORM_DIALECT=mysql DB_URL="gorm:gorm@/gorm?charset=utf8&parseTime=True"
//GORM_DIALECT=postgres DB_URL="user=gorm dbname=gorm sslmode=disable"
//GORM_DIALECT=sqlite3 DB_URL=/tmp/gorm.DB go test
package security

import (
	"flag"

	. "github.com/empirefox/iniu/gorm/db"
)

func init() {
	flag.Set("stderrthreshold", "INFO")
	flag.Parse()
	DB.LogMode(false)
}

var superDB = DB.Set("context:account", &Account{
	Name:      "empirefox",
	Enabled:   true,
	HoldsPerm: "*",
})
