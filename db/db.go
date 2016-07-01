//DB_DIALECT=mysql DB_URL="gorm:gorm@/gorm?charset=utf8&parseTime=True"
//DB_DIALECT=postgres DB_URL="user=gorm dbname=gorm sslmode=disable" PORT=8080 goconvey
//DB_DIALECT=sqlite3 DB_URL=/tmp/gorm.DB go test
package db

import (
	"fmt"
	"sync"

	"github.com/empirefox/gotool/paas"
	//	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	//	_ "github.com/lib/pq"
	//	_ "github.com/mattn/go-sqlite3"
)

var (
	DB *gorm.DB
	mu sync.Mutex
)

func ConnectIfNot() {
	mu.Lock()
	if DB == nil {
		Connect()
	}
	mu.Unlock()
}

func Connect() {
	var err error

	if paas.Gorm.Url == "" {
		panic("DB_URL must be set if not in paas")
	}

	DB, err = gorm.Open(paas.Gorm.Dialect, paas.Gorm.Url)

	if err != nil {
		panic(fmt.Sprintf("No error should happen when connect database, but got %+v", err))
	}

	DB.DB().SetMaxIdleConns(paas.Gorm.MaxIdle)
	DB.DB().SetMaxOpenConns(paas.Gorm.MaxOpen)
}
