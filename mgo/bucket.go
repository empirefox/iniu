//deprecated
package mgo

import (
	"errors"
	"flag"
	"fmt"
	. "github.com/empirefox/iniu/comm"
	. "github.com/empirefox/iniu/conf"
	"github.com/golang/glog"
	"github.com/qiniu/api/auth/digest"
	"github.com/qiniu/api/rs"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"os"
	"time"
)

var (
	DbUrl  string
	DbName string
)

func init() {
	flag.Set("stderrthreshold", "INFO")
	DbUrl = os.Getenv("DB_URL")
	DbName = os.Getenv("DB_NAME")
	if DbUrl == "" || DbName == "" {
		panic("数据库环境变量没有正确设置")
	}
	glog.Infoln(DbUrl)
	glog.Infoln(DbName)
}

//Bucket 指七牛bucket的相关信息
type Bucket struct {
	Id       bson.ObjectId `bson:"_id"` //For mgo
	Name     string
	Ak       string
	Sk       string
	Uptoken  string
	Expires  time.Time
	Life     int64
	HasError bool
	Errors   int
}

//内存中new一个uptoken,有效期为从现在开始的第X天
func (this *Bucket) NewUptoken() error {
	if this.Name == "" || this.Ak == "" || this.Sk == "" {
		return errors.New("Bucket的Name/Ak/Sk为空，无法生成Uptoken")
	}
	if this.Life == 0 {
		this.Life = 380
	}
	this.Expires = time.Now().Add(time.Duration(this.Life) * DAY)
	putPolicy := rs.PutPolicy{
		Scope:   this.Name,
		Expires: uint32(this.Expires.Unix()),
		//CallbackUrl: callbackUrl,
		//CallbackBody:callbackBody,
		//ReturnUrl:   returnUrl,
		//ReturnBody:  returnBody,
		//AsyncOps:    asyncOps,
		//EndUser:     endUser,
	}
	this.Uptoken = putPolicy.Token(&digest.Mac{this.Ak, []byte(this.Sk)})
	this.HasError = false
	return nil
}

//恢复uptoken
func recUptoken(old string) func(this *Bucket) {
	return func(this *Bucket) {
		if err := recover(); err != nil {
			this.Uptoken = old
		}
	}
}

//更新uptoken,去除Err标志
func (this *Bucket) ReUptoken() {
	defer recUptoken(this.Uptoken)(this)

	err := this.NewUptoken()
	if err != nil {
		panic(err)
	}

	err = this.Save()
	if err != nil {
		panic(err)
	}

	this.NoErr()
}

//生成img的url
func (this *Bucket) ImgUrl(key string) string {
	return this.ImgBaseUrl() + key
}

func (this *Bucket) ImgBaseUrl() string {
	return fmt.Sprintf("http://%v.qiniudn.com/", this.Name)
}

func (this *Bucket) LogErrS(session *mgo.Session) {
	c := newC(session, C_BUCKET)
	err := c.UpdateId(this.Id,
		bson.M{"$inc": bson.M{
			"errors": 1,
		},
			"$set": bson.M{
				"has_error": true,
			}})
	ErrLog(err)
}

func (this *Bucket) NoErrS(session *mgo.Session) {
	c := newC(session, C_BUCKET)
	err := c.UpdateId(this.Id,
		bson.M{"$set": bson.M{
			"has_error": false,
		}})
	ErrLog(err)
}

//保存
func (this *Bucket) SaveS(session *mgo.Session) error {
	if this.Uptoken == "" {
		this.NewUptoken()
	}
	c := newC(session, C_BUCKET)
	var err error
	if len(this.Id) == 0 {
		this.Id = bson.NewObjectId()
		err = c.Insert(this)
	} else {
		err = c.UpdateId(this.Id, this)
	}
	return ErrLog(err)
}

//查找
func (this *Bucket) FindS(session *mgo.Session) error {
	c := newC(session, C_BUCKET)
	err := c.Find(bson.M{"name": this.Name}).One(this)
	if err != nil {
		glog.Infoln("找不到Bucket")
		return err
	}
	return nil
}

//删除
func (this *Bucket) DelS(session *mgo.Session) error {
	c := newC(session, C_BUCKET)
	err := c.RemoveId(this.Id)
	if err != nil {
		glog.Infoln("无法删除Bucket")
		return err
	}
	return nil
}

func (this *Bucket) Save() error {
	session := NewS()
	defer session.Close()
	return this.SaveS(session)
}

func (this *Bucket) Find() error {
	session := NewS()
	defer session.Close()
	return this.FindS(session)
}

func (this *Bucket) Del() error {
	session := NewS()
	defer session.Close()
	return this.DelS(session)
}

func (this *Bucket) LogErr() {
	session := NewS()
	defer session.Close()
	this.LogErrS(session)
}

func (this *Bucket) NoErr() {
	session := NewS()
	defer session.Close()
	this.NoErrS(session)
}

func Buckets(bs *[]Bucket) error {
	session := NewS()
	defer session.Close()
	c := newC(session, C_BUCKET)
	return c.Find(nil).All(bs)
}

func newC(session *mgo.Session, cName string) *mgo.Collection {
	return session.DB(DbName).C(cName)
}

//S 新建一个链接Session，处理完毕后应Close,Handler应该有一个recover
func NewS() *mgo.Session {
	glog.Infoln("连接到mongodb:", DbUrl)
	session, err := mgo.Dial(DbUrl)
	if err != nil {
		glog.Errorln(err)
		panic("连接mongodb错误")
	}
	session.SetMode(mgo.Monotonic, true)
	return session
}
