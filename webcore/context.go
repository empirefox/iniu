package webcore

import (
	"net/http"
	"net/url"

	"github.com/empirefox/ffgen/ffgen"
	"github.com/empirefox/iniu/base"
	"github.com/empirefox/shirolet"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type Pagination struct {
	Total int
	Start int
	Size  int
}

type Context struct {
	*gin.Context
	DB       *gorm.DB
	Resource *Resource
	Holds    shirolet.Holds

	filters   map[string]*Filter
	permitter *permitter
	PermType  base.PermType

	Pagination Pagination
	query      url.Values
}

func NewContext(gc *gin.Context, res *Resource) *Context {
	c := &Context{
		Context:  gc,
		Resource: res,
	}
	c.PreInit()
	return c
}

// Clone clone current context
func (context *Context) Clone() *Context {
	var clone = *context
	return &clone
}

// GetDB get db from current context
func (context *Context) GetDB() *gorm.DB {
	if context.DB != nil {
		return context.DB
	}
	return context.Resource.DB
}

// SetDB set db into current context
func (context *Context) SetDB(db *gorm.DB) {
	context.DB = db
}

type holder interface {
	GetHolds() shirolet.Holds
}

func (c *Context) PreInit() {
	if c.PermType == 0 {
		switch c.Request.URL.Query().Get("typ") {
		case "n": // new
			c.PermType = base.CREATE
		case "m": // modify
			c.PermType = base.UPDATE
		default: // view
			c.PermType = base.READ
		}
	}

	hd, ok := c.Keys[GinUserKey].(holder)
	if ok {
		c.Holds = hd.GetHolds()
	}
}

func (c *Context) FormOrQuery(name string) string {
	if value := c.Request.Form.Get(name); value != "" {
		return value
	}
	if c.query == nil {
		c.query = c.Request.URL.Query()
	}
	return c.query.Get(name)
}

func (c *Context) PermittedBind(m interface{}) (*ffgen.Unmarshaled, error) {
	if unmarshaler, ok := m.(ffgen.Unmarshaler); ok {
		w := ffgen.NewUnmarshalerWrapper(unmarshaler, c.GetPermitter())
		return w.Unmarshaled, c.BindJSON(w)
	}
	return nil, c.BindJSON(m)
}

func (c *Context) PermittedJSON(m interface{}) {
	if marshaler, ok := m.(ffgen.Marshaler); ok {
		m = ffgen.NewMarshalerWrapper(marshaler, c.GetPermitter())
	}
	c.JSON(200, m)
}

func (c *Context) AbortCheck404Error(err error) {
	if err == gorm.ErrRecordNotFound {
		c.AbortWithStatus(http.StatusNotFound)
	} else {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
}
