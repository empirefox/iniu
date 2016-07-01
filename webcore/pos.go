package webcore

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/empirefox/iniu/base"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

const (
	DEL_AT        = "deleted_at"
	DEL_AT_CLAUSE = "deleted_at IS NULL"
)

var (
	DEL_AT_MAP = map[string]interface{}{DEL_AT: nil}
)

type IdPos struct {
	ID  uint
	Pos int64
}

// Post /:table/pos/get
func (res *Resource) GetPos(gc *gin.Context) {
	c := NewContext(gc, res)
	if code := res.checkPos(c); code != 0 {
		c.AbortWithStatus(code)
		return
	}

	var ids []uint
	if err := c.BindJSON(&ids); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var ips []IdPos
	db := c.GetDB().Table(res.Struct.TableName).Where(ids)
	if res.Struct.SoftDel {
		db = db.Where(DEL_AT_MAP)
	}
	if err := db.Find(&ips).Error; err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(200, ips)
}

// Post /:table/pos/re
func (res *Resource) RePos(gc *gin.Context) {
	c := NewContext(gc, res)
	c.PermType = base.UPDATE
	if code := res.checkPos(c); code != 0 {
		c.AbortWithStatus(code)
		return
	}

	var where []string
	params := []interface{}{4}
	if res.Struct.GroupBase != "" {
		if value := c.FormOrQuery(res.Struct.GroupBase); value != "" {
			where = append(where, fmt.Sprintf("%v = ?", res.Struct.GroupBase))
			params = append(params, value)
		} else {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}
	if res.Struct.SoftDel {
		where = append(where, DEL_AT_CLAUSE)
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = " WHERE " + strings.Join(where, " AND ")
	}

	db := c.GetDB().Exec(fmt.Sprintf(`
UPDATE %v
SET pos = vp * ?
FROM
(
    SELECT row_number() OVER (ORDER BY pos) AS vp, id AS vid
    FROM %v%v
) AS v
WHERE id = vid`,
		res.Struct.TableName, res.Struct.TableName, whereClause), params...)

	if err := db.Error; err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(200, "")
}

type PosBoundData struct {
	Id      uint
	Biggest bool
	Pos     int64
}

// Post /:table/pos/bound
func (res *Resource) PosBound(gc *gin.Context) {
	c := NewContext(gc, res)
	c.PermType = base.UPDATE
	if code := res.checkPos(c); code != 0 {
		c.AbortWithStatus(code)
		return
	}

	var data PosBoundData
	if err := c.BindJSON(&data); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	where := []string{"id = ?"}
	params := []interface{}{4, data.Id}
	if res.Struct.GroupBase != "" {
		if value := c.FormOrQuery(res.Struct.GroupBase); value != "" {
			where = append(where, fmt.Sprintf("%v = ?", res.Struct.GroupBase))
			params = append(params, value)
		} else {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}
	if res.Struct.SoftDel {
		where = append(where, DEL_AT_CLAUSE)
	}

	db := c.GetDB()
	if data.Biggest {
		db = db.Raw(fmt.Sprintf(
			`UPDATE %v SET pos = ((SELECT pos FROM %v ORDER BY pos DESC LIMIT 1) + ?) WHERE %v RETURNING pos`,
			res.Struct.TableName, res.Struct.TableName, strings.Join(where, " AND ")), params...,
		)
	} else {
		db = db.Raw(fmt.Sprintf(
			`UPDATE %v SET pos = ((SELECT pos FROM %v ORDER BY pos LIMIT 1) - ?) WHERE %v RETURNING pos`,
			res.Struct.TableName, res.Struct.TableName, strings.Join(where, " AND ")), params...,
		)
	}

	if err := db.Scan(&data).Error; err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(200, &data)
}

// Post /:table/pos/x
func (res *Resource) XchIpGtOrLt(gc *gin.Context) {
	c := NewContext(gc, res)
	c.PermType = base.UPDATE
	if code := res.checkPos(c); code != 0 {
		c.AbortWithStatus(code)
		return
	}

	ips, err := res.ipGtOrLtAnd(c, 2)
	if err != nil {
		return
	}

	switch len(ips) {
	case 2:
		ips[0].Pos, ips[1].Pos = ips[1].Pos, ips[0].Pos
		if err = res.saveIPs(ips); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	case 0:
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(200, ips)
}

// Post /:table/pos/next
func (res *Resource) PosGtOrLtAnd(gc *gin.Context) {
	c := NewContext(gc, res)
	c.PermType = base.CREATE
	if code := res.checkPos(c); code != 0 {
		c.AbortWithStatus(code)
		return
	}

	ips, err := res.ipGtOrLtAnd(c, 0)
	if err != nil {
		return
	}
	c.JSON(200, ips)
}

// Post /:table/pos/save
func (res *Resource) SavePosAll(gc *gin.Context) {
	c := NewContext(gc, res)
	c.PermType = base.UPDATE
	if code := res.checkPos(c); code != 0 {
		c.AbortWithStatus(code)
		return
	}

	var ips []IdPos
	if err := c.BindJSON(&ips); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	if err := res.saveIPs(ips); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(200, "")
}

func (res *Resource) saveIPs(ips []IdPos) error {
	tx := res.DB.Begin()
	for i := range ips {
		tx = tx.Table(res.Struct.TableName).Where("id = ?", ips[i].ID).UpdateColumn("pos", ips[i].Pos)
		if err := tx.Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func (res *Resource) checkPos(c *Context) int {
	if !res.Struct.HasPos {
		return http.StatusBadRequest
	}
	if !c.GetPermitter().IsPermitted("Pos") {
		return http.StatusForbidden
	}
	return 0
}

type IpGtOrLtAndData struct {
	IsGreaterThan bool `json:"Gt"`
	Id            uint
	Size          int
}

func (res *Resource) ipGtOrLtAnd_(db *gorm.DB, data *IpGtOrLtAndData) ([]IdPos, error) {
	dir, o := "<=", "pos DESC"
	if data.IsGreaterThan {
		dir, o = ">=", "pos"
	}
	var ips []IdPos
	err := db.Table(res.Struct.TableName).Order(o).
		Where(fmt.Sprintf("pos %v (SELECT pos FROM %v t2 WHERE t2.id = ? LIMIT 1)", dir, res.Struct.TableName), data.Id).
		Limit(data.Size).Find(&ips).Error
	return ips, err
}

func (res *Resource) ipGtOrLtAnd(c *Context, size int) (ips []IdPos, err error) {
	var data IpGtOrLtAndData
	err = c.BindJSON(&data)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if size > 0 {
		data.Size = size
	}
	db := c.GetDB()
	if res.Struct.GroupBase != "" {
		if value := c.FormOrQuery(res.Struct.GroupBase); value != "" {
			db = db.Where(res.Struct.GroupBase, value)
		} else {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}
	if res.Struct.SoftDel {
		db = db.Where(DEL_AT_MAP)
	}
	ips, err = res.ipGtOrLtAnd_(db, &data)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	return
}
