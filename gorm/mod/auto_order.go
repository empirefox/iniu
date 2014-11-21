package mod

import (
	"reflect"

	"github.com/jinzhu/gorm"
)

var (
	descColumns = []string{"pos", "updated_at", "created_at"}
	ascColumns  = []string{"title", "name", "id"}
)

func init() {
	gorm.DefaultCallback.Query().Before("gorm:query").Register("gorm:auto_order", AutoOrderCallback)
}

func AutoOrderCallback(scope *gorm.Scope) {
	if _, ok := scope.InstanceGet("gorm:order_by_primary_key"); ok {
		return
	}

	s := scope.Search
	if len(s.Orders) > 0 {
		return
	}

	var dest = scope.Value
	if value, ok := scope.InstanceGet("gorm:query_destination"); ok {
		dest = value
	}

	destType := reflect.TypeOf(dest)
	if destType.Kind() == reflect.Ptr {
		destType = destType.Elem()
	}
	if destType.Kind() == reflect.Slice {
		destType = destType.Elem()
	}

	if destType.Kind() != reflect.Struct {
		return
	}

	field, ok := destType.FieldByName("Orderd")
	if !ok {
		return
	}

	if field.Type != reflect.TypeOf(Orderd{}) {
		return
	}

	if order := field.Tag.Get("order"); order != "" {
		s.Orders = []string{order}
		return
	}

	for _, column := range descColumns {
		if scope.HasColumn(column) {
			s.Orders = desc(column)
			return
		}
	}

	for _, column := range ascColumns {
		if scope.HasColumn(column) {
			s.Orders = asc(column)
			return
		}
	}
}

type Orderd struct {
}

func asc(name string) []string {
	return []string{name + " asc"}
}

func desc(name string) []string {
	return []string{name + " desc"}
}
