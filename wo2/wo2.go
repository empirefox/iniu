package wo2

import (
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/asaskevich/govalidator"
	"github.com/buger/jsonparser"
	mpoauth2 "github.com/chanxuehong/wechat.v2/mp/oauth2"
	"github.com/chanxuehong/wechat.v2/oauth2"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/mcuadros/go-defaults"
)

var log = logrus.New()

type SecurityHandler interface {
	Login(userinfo *mpoauth2.UserInfo) (ret interface{}, err error)
	ParseToken(c *gin.Context) (tok *jwt.Token, user interface{}, err error)
}

type Config struct {
	WxAppId      string `valid:"required"`
	WxAppSecret  string `valid:"required"`
	WxOauthPath  string `default:"/oauth/wechat"`
	GinClaimsKey string `default:"claims"`
	GinUserKey   string `default:"user"`

	SecurityHandler SecurityHandler

	endpoint *mpoauth2.Endpoint
}

// Middleware proccess Login related logic.
// It does not block the user handler, just try to retrieve Token.Claims.
func (config *Config) Middleware() gin.HandlerFunc {
	if config == nil {
		panic("goauth config is nil")
	}
	config.loadDefault()

	return func(c *gin.Context) {
		if c.Request.URL.Path == config.WxOauthPath && c.Request.Method == "POST" {
			if err := config.authHandle(c); err != nil {
				c.AbortWithError(http.StatusUnauthorized, err)
			}
		} else {
			tok, user, err := config.SecurityHandler.ParseToken(c)
			if err == nil && tok.Valid {
				c.Set(config.GinClaimsKey, tok.Claims)
				c.Set(config.GinUserKey, user)
			}
		}
	}
}

func (config *Config) MustAuthed(c *gin.Context) {
	_, ok1 := c.Keys[config.GinClaimsKey]
	_, ok2 := c.Keys[config.GinUserKey]
	if !ok1 || !ok2 {
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}

func (config *Config) loadDefault() {
	if result, err := govalidator.ValidateStruct(config); !result {
		panic(err)
	}

	defaults.SetDefaults(config)
	config.endpoint = mpoauth2.NewEndpoint(config.WxAppId, config.WxAppSecret)
}

func (config *Config) authHandle(c *gin.Context) error {
	raw, _ := ioutil.ReadAll(c.Request.Body)
	log.Debugf("Code Body:%s\n", raw)
	code, err := jsonparser.GetUnsafeString(raw, "code")
	if err != nil {
		return err
	}

	client := &oauth2.Client{Endpoint: config.endpoint}
	tok, err := client.ExchangeToken(code)
	if err != nil {
		return err
	}

	userinfo, err := mpoauth2.GetUserInfo(tok.AccessToken, tok.OpenId, "", nil)
	if err != nil {
		return err
	}

	ret, err := config.SecurityHandler.Login(userinfo)
	if err != nil {
		return err
	}
	c.JSON(200, ret)
	c.Abort()
	return nil
}
