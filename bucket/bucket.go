package bucket

import (
	"github.com/empirefox/iniu/comm"
	bucketdb "github.com/empirefox/iniu/gorm"
	"github.com/go-martini/martini"
	"github.com/golang/glog"
	"github.com/martini-contrib/binding"
	"io"
	"net/http"
	"time"
)

//binding包需要显示写出form
type Bucket struct {
	Id          int64     `form:"Id" hidden:"true"`
	Name        string    `form:"Name" binding:"required"`
	Description string    `form:"Description"`
	Ak          string    `form:"Ak" binding:"required"`
	Sk          string    `form:"Sk" binding:"required"`
	Uptoken     string    `form:"Uptoken"`
	Life        int64     `form:"Life" input-type:"number"`
	Expires     time.Time `form:"Expires" input-type:"date"`
}

func (bucket *Bucket) Validate(errors *binding.Errors, req *http.Request) {
	glog.Infoln(bucket)
}

func (bucket *Bucket) Save() error {
	bucketdb := &bucketdb.Bucket{
		Id:          bucket.Id,
		Name:        bucket.Name,
		Description: bucket.Description,
		Ak:          bucket.Ak,
		Sk:          bucket.Sk,
		Uptoken:     bucket.Uptoken,
		Expires:     bucket.Expires,
		Life:        bucket.Life,
	}
	return bucketdb.Save()
}

//更新Bucket信息,需要先绑定Bucket
func UpdateBucketHandler(okPath string) martini.Handler {
	return func(bucket Bucket, w http.ResponseWriter, r *http.Request) {
		err := bucket.Save()
		if err != nil {
			io.WriteString(w, "保存错误，查看日志")
			glog.Error(err)
		} else {
			http.Redirect(w, r, okPath, http.StatusFound)
		}
	}
}

func UpdateBucketHandlers(okPath string) []martini.Handler {
	return []martini.Handler{binding.Bind(Bucket{}), UpdateBucketHandler(okPath)}
}

//查看Bucket的json信息,不需要绑定
func ViewBucket() martini.Handler {
	return func(params martini.Params) string {
		bucket, _ := bucketdb.FindByName(params["name"])
		return comm.ToJsonFunc(bucket)
	}
}

type RemoveReqData struct {
	Id int64 `json:"Id" binding:"required"`
}

func RemoveBucket() martini.Handler {
	return func(data RemoveReqData, w http.ResponseWriter, r *http.Request) string {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		err := bucketdb.Delete(data.Id)
		if err != nil {
			glog.Errorln("删除Id错误：", data.Id)
			return `{"error":1}`
		}
		return `{"error":0}`
	}
}
