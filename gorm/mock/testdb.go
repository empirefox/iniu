//IGNORE ARGS!
//mock like usually gorm:
//db := mock.DB.Table("forms").Select("name,title").Order("pos desc")
//bdb := mock.NewBackend(db)
//tfs := []TableForm{}
//bdb.StubQuery(&tfs, RowsFromCSVString(columns, result))
//.............
//the result should be the expected
package mock

import (
	"crypto/sha1"
	"database/sql/driver"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strings"

	"github.com/erikstmartin/go-testdb"
	"github.com/golang/glog"
	"github.com/jinzhu/gorm"
)

var DB gorm.DB
var scopes map[string]*gorm.Scope

func init() {
	var err error
	DB, err = gorm.Open("testdb", "")
	if err != nil {
		glog.Errorln(err)
	}
	scopes = make(map[string]*gorm.Scope)
	gorm.DefaultCallback.Query().Replace("gorm:query", Query)
}

type BDB struct {
	tdb *gorm.DB
	dir string
}

func NewBackend(tdb *gorm.DB) *BDB {
	bdb := &BDB{}
	bdb.tdb = tdb
	return bdb
}

func (bdb *BDB) First() *BDB {
	bdb.dir = "ASC"
	return bdb
}

func (bdb *BDB) Last() *BDB {
	bdb.dir = "DESC"
	return bdb
}

//just stub Parameterized Query string, WITHOUT ARGS
// out: the injected value is the mock returned value, so must equal with the actual type
func (bdb *BDB) StubQuery(out interface{}, rows driver.Rows) {
	scope := bdb.tdb.NewScope(out)
	if bdb.dir != "" {
		scope.InstanceSet("gorm:order_by_primary_key", bdb.dir)
	}
	prepareScope(scope)
	prepareQuerySql(scope)
	if !scope.HasError() {
		testdb.StubQuery(scope.Sql, rows)
		scopes[getQueryHash(scope.Sql)] = scope
	}
}

func prepareScope(scope *gorm.Scope) {
	var dest = reflect.Indirect(reflect.ValueOf(scope.Value))
	if value, ok := scope.Get("gorm:query_destination"); ok {
		dest = reflect.Indirect(reflect.ValueOf(value))
	}

	if dest.Kind() != reflect.Slice {
		scope.Search.Limit = "1"
	}
}

func selectSql(s *gorm.Scope) string {
	if len(s.Search.Select) == 0 {
		return "*"
	} else {
		return s.Search.Select
	}
}

func prepareQuerySql(scope *gorm.Scope) {
	if scope.Search.Raw {
		scope.Raw(strings.TrimLeft(scope.CombinedConditionSql(), "WHERE "))
	} else {
		scope.Raw(fmt.Sprintf("SELECT %v FROM %v %v", selectSql(scope), scope.QuotedTableName(), scope.CombinedConditionSql()))
	}
	return
}

//copy form github.com/erikstmartin/go-testdb
var whitespaceRegexp = regexp.MustCompile("\\s")

func getQueryHash(query string) string {
	// Remove whitespace and lowercase to make stubbing less brittle
	query = strings.ToLower(whitespaceRegexp.ReplaceAllString(query, ""))

	h := sha1.New()
	io.WriteString(h, query)

	return string(h.Sum(nil))
}
