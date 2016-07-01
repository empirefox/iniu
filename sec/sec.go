package security

import (
	"fmt"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	mpoauth2 "github.com/chanxuehong/wechat.v2/mp/oauth2"
	"github.com/dchest/uniuri"
	"github.com/delaemon/sonyflake"
	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/empirefox/shirolet"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

var (
	log = logrus.New()
)

type Claims struct {
	jwt.StandardClaims
	Id     uint64 `json:"jti,omitempty"`
	UserId uint   `json:"uid,omitempty"`
}

type User struct {
	gorm.Model
	UserInfo *UserInfo

	OpenId  string
	UnionId string

	Holds    string
	Key      string
	Verified bool
}

type UserInfo struct {
	gorm.Model
	UserID uint

	UserInfoOut

	Email string
	Phone string
}

type UserInfoOut struct {
	Nickname     string
	Sex          int
	City         string
	Province     string
	Country      string
	HeadImageURL string
	Privilege    string
}

type Config struct {
	SignAlg   string        `default:"HS256"`
	TokenLife time.Duration `default:"30"`
	DB        *gorm.DB
	Sonyflake *sonyflake.Sonyflake
}

func (config *Config) Login(userinfo *mpoauth2.UserInfo) (interface{}, error) {
	attr := &User{
		OpenId:  userinfo.OpenId,
		UnionId: userinfo.UnionId,
		Key:     uniuri.NewLen(32),
	}
	if userinfo.Nickname != "" {
		attr.UserInfo = &UserInfo{
			UserInfoOut: UserInfoOut{
				Nickname:     userinfo.Nickname,
				Sex:          userinfo.Sex,
				City:         userinfo.City,
				Province:     userinfo.Province,
				Country:      userinfo.Country,
				HeadImageURL: userinfo.HeadImageURL, // TODO Save to our cdn
				Privilege:    strings.Join(userinfo.Privilege, "|"),
			},
		}
	}

	var user User
	db := config.DB.Where("open_id = ?", userinfo.OpenId).Attrs(attr).Assign(&User{Key: uniuri.NewLen(32)}).FirstOrCreate(&user)
	if db.Error != nil {
		return nil, db.Error
	}

	err := db.Related(user.UserInfo).Error

	var useroutput interface{}
	if err != nil {
		copied := *userinfo
		copied.OpenId = ""
		copied.UnionId = ""
		useroutput = &copied
	} else {
		useroutput = &user.UserInfo.UserInfoOut
	}

	return config.NewToken(&user, useroutput)
}

// NewToken generate {token,user} json
func (config *Config) NewToken(user *User, useroutput interface{}) (gin.H, error) {
	id, err := config.Sonyflake.NextID()
	if err != nil {
		log.Debugln("sonyflake timeout")
		return nil, err
	}
	claims := &Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(config.TokenLife * time.Minute).Unix(),
			IssuedAt:  time.Now().Unix(),
			Subject:   "Weixin",
		},
		Id:     id,
		UserId: user.ID,
	}

	token := jwt.NewWithClaims(jwt.GetSigningMethod(config.SignAlg), claims)
	tokenString, err := token.SignedString(user.Key)
	if err != nil {
		log.Infoln("Sign token err:", err)
		return nil, err
	}
	return gin.H{"token": tokenString, "user": useroutput}, nil
}

func (config *Config) RevokeToken(tok *jwt.Token) error {
	claims := tok.Claims.(*Claims)
	return config.DB.Model(&User{}).Where(claims.UserId).UpdateColumn("key", uniuri.NewLen(32)).Error
}

type userReq struct {
	holds shirolet.Holds
}

func (u *userReq) GetHolds() shirolet.Holds { return u.holds }

// ParseToken need query param typ=n/m
func (config *Config) ParseToken(c *gin.Context) (*jwt.Token, interface{}, error) {
	user := new(User)
	tok, err := request.ParseFromRequestWithClaims(c.Request, request.OAuth2Extractor, &Claims{}, func(tok *jwt.Token) (interface{}, error) {
		if tok.Method.Alg() != config.SignAlg {
			return nil, fmt.Errorf("Unexpected signing method: %v", tok.Header["alg"])
		}
		claims := tok.Claims.(*Claims)

		if err := config.DB.Where(claims.UserId).First(user).Error; err != nil {
			return nil, err
		}
		return []byte(user.Key), nil
	})

	if err != nil {
		return nil, nil, err
	}

	wu := new(userReq)
	if user.Holds != "" {
		wu.holds = shirolet.NewHoldsRaw(user.Holds)
	}

	return tok, wu, err
}
