package webcore

import (
	"github.com/empirefox/iniu/base"
	"github.com/jinzhu/gorm"
)

func beforeQuery(scope *gorm.Scope) {
	if sorting, ok := base.TableSorting(scope.TableName()); ok {
		scope.Search.Order("pos" + sorting)
	}
}

// RegisterCallbacks register callbacks into gorm db instance
func RegisterCallbacks(db *gorm.DB) {
	db.Callback().Query().Before("gorm:query").Register("sorting:sort_by_pos", beforeQuery)
}
