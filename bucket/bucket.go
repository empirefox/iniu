package bucket

import (
	bucketdb "github.com/empirefox/iniu/gorm"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
)

type Bucket struct {
	Id      int64     `form:"id" json:"id"`
	Name    string    `form:"name" json:"name" binding:"required"`
	Ak      string    `form:"ak" json:"ak" binding:"required"`
	Sk      string    `form:"sk" json:"sk" binding:"required"`
	Uptoken string    `form:"uptoken" json:"uptoken"`
	Expires time.Time `form:"expires" json:"expires"`
	Life    int64     `form:"life" json:"life"`
}

func (bucket *Bucket) Save() error {
	bucketdb := &bucketdb.Bucket{
		Id:      bucket.Id,
		Name:    bucket.Name,
		Ak:      bucket.Ak,
		Sk:      bucket.Sk,
		Uptoken: bucket.Uptoken,
		Expires: bucket.Expires,
		Life:    bucket.Life,
	}
	if bucketdb.Id == 0 {
		return bucketdb.Save()
	}
	return bucketdb.Updates()
}

func Buckets() {
	return func(c martini.Context, w http.ResponseWriter, r *http.Request) {
	}
}

func UpdateBucket() []martini.Handler {
	var bind martini.Handler = binding.Bind(Bucket{})
	var update martini.Handler = func(c martini.Context, bucket Bucket, w http.ResponseWriter, r *http.Request) {

	}
	return []martini.Handler{bind, update}
}

func Upload() {
	return func(c martini.Context, w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				glog.Infoln("上传图片错误:", err) // 这里的err其实就是panic传入的内容
				fail := &UpJsonRet{
					UpTime: r.FormValue("up_time"),
					Json:   failUpJson,
				}
				resUpJson(w, fail)
			}
		}()
		//读取KindEditor传过来的内容
		r.ParseMultipartForm(1 << 20)
		if !strings.EqualFold(r.Method, "POST") {
			return
		}
		if !strings.EqualFold(r.FormValue("dir"), "IMAGE") {
			return
		}

		imgName := r.FormValue("localUrl")
		bucketName := r.FormValue("bucket")
		if imgName == "" || bucketName == "" {
			glog.Infoln("imgName|bucketName为空")
			return
		}

		if strings.ContainsAny(imgName, "/\\:") {
			i := strings.LastIndexAny(imgName, "/\\:")
			runes := []rune(imgName)
			imgName = string(runes[i+1:])
		}
		imgName = time.Now().Format(IMG_PRE_FMT) + imgName

		imgFile, _, err := r.FormFile("imgFile")
		if err != nil {
			glog.Infoln("获取图片错误:", err)
			return
		}
		defer imgFile.Close()

		//取得bucket
		bucket := bucket(r)

		//上传内容到Qiniu
		var ret qio.PutRet
		uptoken := bucket.Uptoken
		extra := &qio.PutExtra{
		//Params:    params,
		//MimeType:  mieType,
		//Crc32:     crc32,
		//CheckCrc:  CheckCrc,
		}

		// ret       	变量用于存取返回的信息，详情见 qio.PutRet
		// uptoken   	为业务服务器端生成的上传口令
		// key:imgName	为文件存储的标识
		// r:imgFile   	为io.Reader类型，用于从其读取数据
		// extra     	为上传文件的额外信息,可为空， 详情见 qio.PutExtra, 可选
		err = qio.Put(nil, &ret, uptoken, imgName, imgFile, extra)

		if err != nil {
			//上传产生错误
			glog.Infoln("qio.Put failed:", err)
			bucket.LogErr()
			return
		}

		//上传成功，返回给KindEditor
		successJson := &UpJson{
			Error: 0,
			Url:   bucket.ImgUrl(imgName),
		}
		success := &UpJsonRet{
			UpTime: r.FormValue("up_time"),
			Json:   successJson,
		}
		//w.Header().Set("Content-type", "application/json")
		resUpJson(w, success)
	}
}
