package webcore

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jinzhu/gorm"
)

type Filter struct {
	Name    string
	Field   string
	Handler FilterHandler
}

type FilterHandler func(column string, params []string, context *Context) *gorm.DB

// HandleQueryFilters handle filters from query string:
// &ft(filter)=Price:gteq:10+Price:lteq:20+Discount:true
func (c *Context) HandleQueryFilters() {
	filters := c.GetFilters()
	for _, ft := range strings.Split(c.FormOrQuery("ft"), "+") {
		segs := strings.Split(ft, ":")
		if len(segs) > 1 {
			if filter, ok := filters[segs[0]]; ok {
				c.SetDB(filter.Handler(filter.Field, segs[1:], c))
			}
		}
	}
}

// GetFilters get all permitted field filters then overwrite with resource
func (c *Context) GetFilters() map[string]*Filter {
	if c.filters == nil {
		res := c.Resource
		filters := make(map[string]*Filter)
		for field := range c.GetPermitter().permittedFields.Fields {
			if !res.Struct.FieldMap[field].Xfilter {
				filters[field] = c.Resource.NewFilter(field)
			}
		}
		for name, filter := range res.filters {
			c.filters[name] = filter
		}
		c.filters = filters
	}
	return c.filters
}

// GetFilters get all permitted field filters then overwrite with resource
func (c *Context) GetFilterNames() (names []string) {
	res := c.Resource
	for field := range res.Struct.GetPermittedFields(c.Holds, c.PermType).Fields {
		if !res.Struct.FieldMap[field].Xfilter {
			names = append(names, field)
		}
	}
	for name, _ := range res.filters {
		names = append(names, name)
	}
	return
}

// NewFilter create Filter from struct field name
func (res *Resource) NewFilter(name string) *Filter {
	filter := &Filter{Name: name}
	if filter.Field == "" {
		filter.Field = res.Struct.FieldMap[filter.Name].Field.DBName
	}
	if filter.Handler == nil {
		filter.Handler = res.defaultFilterHandler()
	}
	return filter
}

// AddFilter add or overwrite the default filter handler
func (res *Resource) AddFilter(name, field string, handler FilterHandler) {
	if res.filters == nil {
		res.filters = make(map[string]*Filter)
	}
	res.filters[name] = &Filter{
		Name:    name,
		Field:   field,
		Handler: handler,
	}
}

func (res *Resource) defaultFilterHandler() FilterHandler {
	return func(column string, params []string, context *Context) *gorm.DB {
		scope := context.GetDB()
		switch len(params) {
		case 1:
			Value, err := strconv.ParseBool(params[0])
			if err != nil {
				return scope
			}
			return scope.Where(fmt.Sprintf("%v = ?", scope.NewScope(nil).Quote(column)), Value)

		case 2:
			switch params[0] {
			case "eq":
				return scope.Where(fmt.Sprintf("%v = ?", scope.NewScope(nil).Quote(column)), params[1])
			case "gt":
				return scope.Where(fmt.Sprintf("%v > ?", scope.NewScope(nil).Quote(column)), params[1])
			case "gteq":
				return scope.Where(fmt.Sprintf("%v >= ?", scope.NewScope(nil).Quote(column)), params[1])
			case "lt":
				return scope.Where(fmt.Sprintf("%v < ?", scope.NewScope(nil).Quote(column)), params[1])
			case "lteq":
				return scope.Where(fmt.Sprintf("%v <= ?", scope.NewScope(nil).Quote(column)), params[1])
			}
			return scope
		}
		return scope

	}
}
