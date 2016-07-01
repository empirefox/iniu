package webcore

import (
	"net/http"
	"strings"

	"github.com/empirefox/iniu/base"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

var (
	GinUserKey = "user"
)

type Resource struct {
	DB *gorm.DB

	Struct *base.StructMeta

	// used by Context
	scopeMap map[string]*SearchScope

	// output to client
	scopes []*SearchScopeGroup

	// used to overwrite default filters
	filters map[string]*Filter

	SearchHandler func(keyword string, context *Context) *gorm.DB
}

func NewResource(m base.Model, db *gorm.DB) *Resource {
	sm, ok := base.StructMetaMap[base.Formname(m)]
	if !ok {
		return nil
	}
	res := &Resource{
		Struct: sm,
		DB:     db,
	}
	res.SearchAttrs()
	return res
}

// Get /:table/searchmetas?typ=n/m
func (res *Resource) SearchMetas(gc *gin.Context) {
	c := NewContext(gc, res)
	h := gin.H{
		"Filters":     c.GetFilterNames(),
		"ScopeGroups": res.scopes,
		"GroupBase":   res.Struct.GroupBase,
	}
	if c.GetPermitter().IsPermitted("Pos") {
		h["PosInfo"] = &res.Struct.PosInfo
	}
	c.JSON(200, h)
}

// Get /:table/metas?typ=n/m
func (res *Resource) FormMetas(gc *gin.Context) {
	c := NewContext(gc, res)
	out := res.Struct.GetPermittedMetas(c.Holds, c.PermType)
	if out == nil {
		c.AbortWithStatus(404)
	} else {
		c.JSON(200, out)
	}
}

// Get /:table/default
func (res *Resource) Default(gc *gin.Context) {
	c := NewContext(gc, res)
	c.PermType = base.CREATE
	m := res.Struct.Default()
	c.PermittedJSON(m)
}

// Post /:table/+1?typ=n/m
// TODO add related create
func (res *Resource) Create(gc *gin.Context) {
	c := NewContext(gc, res)
	m := res.Struct.New()
	_, err := c.PermittedBind(m)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	db := c.GetDB().Create(m)
	if err := db.Error; err != nil {
		c.AbortCheck404Error(err)
	} else {
		c.PermittedJSON(m)
	}
}

// Post /:table/1/:id?typ=n/m
// TODO add related update and batch command
func (res *Resource) Update(gc *gin.Context) {
	c := NewContext(gc, res)
	m := res.Struct.New()
	updates, err := c.PermittedBind(m)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	db := c.GetDB().Model(res.Struct.New()).Where(gc.Param("id")).Updates(updates.Fields)
	if err := db.Error; err != nil {
		c.AbortCheck404Error(err)
	} else {
		c.JSON(200, gin.H{"RowsAffected": db.RowsAffected})
	}
}

// Get /:table/1/:id?typ=n/m
func (res *Resource) OneById(gc *gin.Context) {
	c := NewContext(gc, res)
	m := res.Struct.New()
	if err := c.GetDB().Where(gc.Param("id")).First(m).Error; err != nil {
		c.AbortCheck404Error(err)
	} else {
		c.PermittedJSON(m)
	}
}

// Get /:table/ls?typ=n/m
//		&q(query)=bmw
//		&st(start)=100&sz(size)=20
//		&sp(scope)=2016style+white
//		&ft(filter)=Price:gteq:10+Price:lteq:20+Discount:true
//		&ob(order)=Price:desc
func (res *Resource) ListMany(gc *gin.Context) {
	c := NewContext(gc, res)
	result, err := c.FindMany()
	if err != nil {
		c.AbortCheck404Error(err)
	} else {
		c.PermittedJSON(result)
	}
}

// Get /:table/pkshows?typ=n/m
//		&q(query)=bmw
//		&st(start)=100&sz(size)=20
//		&sp(scope)=2016style+white
//		&ft(filter)=Price:gteq:10+Price:lteq:20+Discount:true
//		&ob(order)=Price:desc
func (res *Resource) PkShows(gc *gin.Context) {
	c := NewContext(gc, res)
	if !c.GetPermitter().IsPermitted(res.Struct.PkShowFields[1]) {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	c.HandleSearch()
	data := res.Struct.GetPkShow(c.GetDB())
	c.JSON(200, gin.H{
		"Fields": res.Struct.PkShowFields,
		"Data":   data,
	})
}

// Delete /:table/del?id=1+2+5
func (res *Resource) Delete(gc *gin.Context) {
	c := NewContext(gc, res)
	if !res.Struct.CanDelete(c.Holds) {
		c.AbortWithStatus(403)
		return
	}

	idQuery := c.FormOrQuery("id")
	if idQuery == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	db := res.DB.Where(strings.Split(idQuery, "+")).Delete(res.Struct.New())
	if err := db.Error; err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	} else {
		c.JSON(200, gin.H{"RowsAffected": db.RowsAffected})
	}
}
