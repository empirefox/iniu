package webcore

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/empirefox/iniu/conf"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/now"
)

//		/:table/ls?typ=n/m
//		&q(query)=bmw
//		&st(start)=100&sz(size)=20
//		&sp(scope)=2016style+white
//		&ft(filter)=Price:gteq:10+Price:lteq:20+Discount:true
//		&ob(order)=Price:desc
func (c *Context) FindMany() (interface{}, error) {
	c.HandleSearch()
	result := c.Resource.Struct.Slice()
	err := c.GetDB().Set("gorm:order_by_primary_key", "DESC").Find(result).Error
	return result, err
}

//		/:table/ls?typ=n/m
//		&q(query)=bmw
//		&st(start)=100&sz(size)=20
//		&sp(scope)=2016style+white
//		&ft(filter)=Price:gteq:10+Price:lteq:20+Discount:true
//		&ob(order)=Price:desc
func (c *Context) HandleSearch() {
	c.HandleQueryScopes()
	c.HandleQueryFilters()
	c.HandleQueryOrder()

	keyword := c.FormOrQuery("q")
	if keyword != "" && c.Resource.SearchHandler != nil {
		c.SetDB(c.Resource.SearchHandler(keyword, c))
	}

	c.HandleQueryPage()
}

// HandleQueryPage handle page from query string:
// &st(start)=100&sz(size)=20
func (c *Context) HandleQueryPage() {
	db := c.GetDB()

	if db.Count(&c.Pagination.Total).Error != nil {
		return
	}

	c.Pagination.Start, _ = strconv.Atoi(c.FormOrQuery("st"))
	c.Pagination.Size, _ = strconv.Atoi(c.FormOrQuery("sz"))
	if c.Pagination.Size == 0 {
		c.Pagination.Size = conf.Defaults.PageSize
	}

	c.SetDB(db.Limit(c.Pagination.Size).Offset(c.Pagination.Start))
}

// HandleQueryOrder handle order from query string:
// &ob(order)=Price:desc
func (c *Context) HandleQueryOrder() {
	orders := strings.Split(c.FormOrQuery("ob"), ":")
	l := len(orders)
	if l < 1 || l > 2 {
		return
	}

	if c.GetPermitter().IsPermitted(orders[0]) {
		dst := ""
		if l == 2 && orders[1] == "desc" {
			dst = " DESC"
		}
		c.SetDB(c.GetDB().Order(c.Resource.Struct.FieldMap[orders[0]].Field.DBName+dst, true))
	}
}

// SearchAttrs set search attributes, when search resources, will use those columns to search
//     // Search products with its name, code, category's name, brand's name
//	   product.SearchAttrs("Name", "Code", "Category.Name", "Brand.Name")
func (res *Resource) SearchAttrs(columns ...string) {
	if len(columns) == 0 {
		columns = res.Struct.SearchAttrs
	}
	if len(columns) == 0 {
		return
	}
	res.SearchHandler = func(keyword string, context *Context) *gorm.DB {
		db := context.GetDB()
		var joinConditionsMap = map[string][]string{}
		var conditions []string
		var keywords []interface{}
		scope := db.NewScope(res.Struct.New())

		for _, column := range columns {
			currentScope, nextScope := scope, scope

			if strings.Contains(column, ".") {
				for _, field := range strings.Split(column, ".") {
					column = field
					currentScope = nextScope
					if field, ok := scope.FieldByName(field); ok {
						if relationship := field.Relationship; relationship != nil {
							nextScope = currentScope.New(reflect.New(field.Field.Type()).Interface())
							key := fmt.Sprintf("LEFT JOIN %v ON", nextScope.TableName())

							for index := range relationship.ForeignDBNames {
								if relationship.Kind == "has_one" || relationship.Kind == "has_many" {
									joinConditionsMap[key] = append(joinConditionsMap[key],
										fmt.Sprintf("%v.%v = %v.%v",
											nextScope.QuotedTableName(), scope.Quote(relationship.ForeignDBNames[index]),
											currentScope.QuotedTableName(), scope.Quote(relationship.AssociationForeignDBNames[index]),
										))
								} else if relationship.Kind == "belongs_to" {
									joinConditionsMap[key] = append(joinConditionsMap[key],
										fmt.Sprintf("%v.%v = %v.%v",
											currentScope.QuotedTableName(), scope.Quote(relationship.ForeignDBNames[index]),
											nextScope.QuotedTableName(), scope.Quote(relationship.AssociationForeignDBNames[index]),
										))
								}
							}
						}
					}
				}
			}

			var tableName = currentScope.Quote(currentScope.TableName())
			if field, ok := currentScope.FieldByName(column); ok && field.IsNormal {
				switch field.Field.Kind() {
				case reflect.String:
					conditions = append(conditions, fmt.Sprintf("upper(%v.%v) like upper(?)", tableName, scope.Quote(field.DBName)))
					keywords = append(keywords, "%"+keyword+"%")
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					if _, err := strconv.Atoi(keyword); err == nil {
						conditions = append(conditions, fmt.Sprintf("%v.%v = ?", tableName, scope.Quote(field.DBName)))
						keywords = append(keywords, keyword)
					}
				case reflect.Float32, reflect.Float64:
					if _, err := strconv.ParseFloat(keyword, 64); err == nil {
						conditions = append(conditions, fmt.Sprintf("%v.%v = ?", tableName, scope.Quote(field.DBName)))
						keywords = append(keywords, keyword)
					}
				case reflect.Bool:
					if value, err := strconv.ParseBool(keyword); err == nil {
						conditions = append(conditions, fmt.Sprintf("%v.%v = ?", tableName, scope.Quote(field.DBName)))
						keywords = append(keywords, value)
					}
				case reflect.Struct:
					// time ?
					if _, ok := field.Field.Interface().(time.Time); ok {
						if parsedTime, err := now.Parse(keyword); err == nil {
							conditions = append(conditions, fmt.Sprintf("%v.%v = ?", tableName, scope.Quote(field.DBName)))
							keywords = append(keywords, parsedTime)
						}
					}
				case reflect.Ptr:
					// time ?
					if _, ok := field.Field.Interface().(*time.Time); ok {
						if parsedTime, err := now.Parse(keyword); err == nil {
							conditions = append(conditions, fmt.Sprintf("%v.%v = ?", tableName, scope.Quote(field.DBName)))
							keywords = append(keywords, parsedTime)
						}
					}
				default:
					conditions = append(conditions, fmt.Sprintf("%v.%v = ?", tableName, scope.Quote(field.DBName)))
					keywords = append(keywords, keyword)
				}
			}
		}

		// join conditions
		if len(joinConditionsMap) > 0 {
			var joinConditions []string
			for key, values := range joinConditionsMap {
				joinConditions = append(joinConditions, fmt.Sprintf("%v %v", key, strings.Join(values, " AND ")))
			}
			db = db.Joins(strings.Join(joinConditions, " "))
		}

		// search conditions
		if len(conditions) > 0 {
			return db.Where(strings.Join(conditions, " OR "), keywords...)
		}
		return db
	}
}
