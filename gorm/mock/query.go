package mock

import "github.com/jinzhu/gorm"

func Query(scope *gorm.Scope) {
	prepareScope(scope)
	prepareQuerySql(scope)
	if !scope.HasError() {
		if mscope, ok := scopes[getQueryHash(scope.Sql)]; ok {
			gorm.Query(mscope)
			scope.IndirectValue().Set(mscope.IndirectValue())
			return
		}
		gorm.Query(scope)
	}
}
