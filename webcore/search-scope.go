package webcore

import (
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
)

type SearchScopeGroup struct {
	Name   string
	Scopes []*SearchScope
}

type SearchScope struct {
	Name   string
	Handle func(*Context) *gorm.DB
}

// HandleQueryScopes handle scopes from query string:
// &sp(scope)=2016style+white
func (c *Context) HandleQueryScopes() {
	if c.Resource.scopeMap != nil {
		for _, sp := range strings.Split(c.FormOrQuery("sp"), "+") {
			if scope, ok := c.Resource.scopeMap[sp]; ok {
				if scope.Handle != nil {
					c.SetDB(scope.Handle(c))
				}
			}
		}
	}
}

func (res *Resource) AddSearchScope(name, group string, handle func(*Context) *gorm.DB) {
	if res.scopeMap == nil {
		res.scopeMap = make(map[string]*SearchScope)
	}
	if _, ok := res.scopeMap[name]; ok {
		log.WithFields(logrus.Fields{
			"struct":      res.Struct.Name,
			"serachscope": name,
		}).Fatalln("duplicated scope")
	}
	scope := &SearchScope{
		Name:   name,
		Handle: handle,
	}

	res.scopeMap[name] = scope

	grp := findSearchScope(res.scopes, group)
	if grp == nil {
		grp = &SearchScopeGroup{Name: group}
		res.scopes = append(res.scopes, grp)
	}
	grp.Scopes = append(grp.Scopes, scope)
}

func findSearchScope(groups []*SearchScopeGroup, name string) *SearchScopeGroup {
	for _, group := range groups {
		if group.Name == name {
			return group
		}
	}
	return nil
}
