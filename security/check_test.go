//GORM_DIALECT=mysql DB_URL="gorm:gorm@/gorm?charset=utf8&parseTime=True"
//GORM_DIALECT=postgres DB_URL="user=gorm dbname=gorm sslmode=disable" PORT=8080 goconvey
//GORM_DIALECT=sqlite3 DB_URL=/tmp/gorm.DB go test
package security

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bradrydzewski/go.auth"
	"github.com/go-martini/martini"
	"github.com/golang/glog"
	. "github.com/jinzhu/copier"
	"github.com/martini-contrib/render"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/empirefox/shirolet"
)

var (
	origin = originAuthConfig{}
)

type originAuthConfig struct {
	config auth.AuthConfig
}

func (o *originAuthConfig) init() {
	Copy(&o.config, auth.Config)
}

func (o originAuthConfig) recovery() {
	auth.Config = &o.config
}

func init() {
	origin.init()
}

func TestConfig(t *testing.T) {
	Convey("Config", t, func() {

		Convey("default value", func() {
			config := Config{}

			Convey("should config path", func() {
				config.path()
				So(AuthMap["/logout"], ShouldEqual, Logout)
				So(AuthMap["/login/openid"], ShouldBeNil)
				So(AuthMap["/login/google"], ShouldBeNil)
				So(AuthMap["/login/github"], ShouldBeNil)
			})
			Convey("should config go.auth", func() {
				config.auth()
				So(auth.Config.CookieSecret, ShouldNotBeEmpty)
				So(*auth.Config, ShouldResemble, auth.AuthConfig{
					CookieSecret:         auth.Config.CookieSecret,
					CookieName:           "_sess",
					CookieExp:            time.Hour * 24 * 14,
					CookieMaxAge:         0,
					CookieSecure:         false,
					CookieHttpOnly:       true,
					LoginRedirect:        "/auth/login",
					LoginSuccessRedirect: "/",
				})
			})
		})

		Convey("set value", func() {
			config := Config{
				CookieSecret:          []byte("CookieSecret"),
				CookieName:            "CookieName",
				CookieExp:             time.Duration(100),
				CookieMaxAge:          101,
				DisableCookieSecure:   true,
				DisableCookieHttpOnly: true,
				LoginRedirect:         "/test/login",
				LoginSuccessRedirect:  "/test/redirect",
				PathLogout:            "/test/logout",
				PathOpenId:            "/test/openid",
				Google:                []string{"/test/google", "ak", "sk", "redrect"},
				Github:                []string{"ak", "sk", "scope"},
			}

			Convey("should config path", func() {
				config.path()
				So(AuthMap["/logout"], ShouldBeNil)
				So(AuthMap["/login/openid"], ShouldBeNil)
				So(AuthMap["/login/google"], ShouldBeNil)

				So(AuthMap["/test/logout"], ShouldEqual, Logout)
				So(AuthMap["/login/github"], ShouldNotBeNil)
				So(AuthMap["/test/openid"], ShouldNotBeNil)
				So(AuthMap["/test/google"], ShouldNotBeNil)
			})
			Convey("should config go.auth", func() {
				config.auth()
				So(auth.Config.CookieSecret, ShouldNotBeEmpty)
				So(*auth.Config, ShouldResemble, auth.AuthConfig{
					CookieSecret:         auth.Config.CookieSecret,
					CookieName:           "CookieName",
					CookieExp:            time.Duration(100),
					CookieMaxAge:         101,
					CookieSecure:         false,
					CookieHttpOnly:       false,
					LoginRedirect:        "/test/login",
					LoginSuccessRedirect: "/test/redirect",
				})
			})
		})
		AuthMap = map[string]func(w http.ResponseWriter, r *http.Request){}
		origin.recovery()
	})
}

func TestPathHandle(t *testing.T) {
	w := httptest.NewRecorder()
	Convey("pathHandle", t, func() {
		handled := false
		AuthMap["/pathHandle"] = func(w http.ResponseWriter, r *http.Request) {
			handled = true
		}
		Convey("should pass thought", func() {
			r, _ := http.NewRequest("GET", "/pathHandle", nil)
			So(pathHandle(w, r), ShouldBeTrue)
			So(handled, ShouldBeTrue)
		})
		Convey("should not pass thought", func() {
			r, _ := http.NewRequest("GET", "/wrong/pathHandle", nil)
			glog.Infoln("pathHandle", AuthMap)
			So(pathHandle(w, r), ShouldBeFalse)
			So(handled, ShouldBeFalse)
		})
	})
	AuthMap = map[string]func(w http.ResponseWriter, r *http.Request){}
}

func TestSudo(t *testing.T) {
	var a *Account
	m := martini.Classic()
	m.Get("/sudo", sudo, func(account *Account) {
		a = account
	})
	w := httptest.NewRecorder()
	Convey("sudo", t, func() {
		Convey("should skip the sudo account", func() {
			r, _ := http.NewRequest("GET", "/sudo", nil)
			m.ServeHTTP(w, r)
			var holds shirolet.Holds
			So(a.Holds, ShouldResemble, holds)
		})
		Convey("should map default super account", func() {
			superdo = true
			r, _ := http.NewRequest("GET", "/sudo", nil)
			m.ServeHTTP(w, r)
			So(a.Permitted(shirolet.NewPermit("*")), ShouldBeTrue)
			So(a.Holds, ShouldResemble, shirolet.NewHolds("*"))
			superdo = false
		})
		Convey("should map perm from url", func() {
			superdo = true
			r, _ := http.NewRequest("GET", "/sudo?perm=a:b:c|d", nil)
			m.ServeHTTP(w, r)
			So(a.Permitted(shirolet.NewPermit("a:b:c")), ShouldBeTrue)
			So(a.Holds, ShouldResemble, shirolet.NewHolds("a:b:c|d"))
			superdo = false
		})
	})
}

func Passed(a *Account, c *Oauth) bool {
	return true
}

func Rejected(a *Account, c *Oauth) bool {
	return false
}

func TestCheck(t *testing.T) {
	triggered := false

	var result = func(prev []ValidateFunc) {
		triggered = true
	}

	Convey("WannaCheck", t, func() {
		m := martini.Classic()
		m.Use(render.Renderer())
		m.Use(Prepare(Config{}, false))
		r, _ := http.NewRequest("GET", "/check", nil)
		w := httptest.NewRecorder()
		Convey("should not Rejected when not triggered", func() {
			triggered = false
			m.Get("/check", WannaCheck(Rejected), result)
			m.ServeHTTP(w, r)
			So(triggered, ShouldBeTrue)
		})
		Convey("should check multi validates", func() {
			m.Get("/check", WannaCheck(Passed), WannaCheck(Rejected), CheckAll(), result)
			triggered = false
			m.ServeHTTP(w, r)
			So(triggered, ShouldBeFalse)
		})
	})

	Convey("Trigger Checks", t, func() {
		m := martini.Classic()
		m.Use(render.Renderer())
		m.Use(Prepare(Config{}, false))
		r, _ := http.NewRequest("GET", "/check", nil)
		w := httptest.NewRecorder()
		Convey("CheckAll", func() {
			Convey("should pass all", func() {
				triggered = false
				m.Get("/check", CheckAll(Passed, Passed, Passed), result)
				m.ServeHTTP(w, r)
				So(triggered, ShouldBeTrue)
			})
			Convey("should rejected all", func() {
				triggered = false
				m.Get("/check", CheckAll(Passed, Passed, Rejected, Passed), result)
				m.ServeHTTP(w, r)
				So(triggered, ShouldBeFalse)
			})
		})
		Convey("CheckThis", func() {
			Convey("should pass all", func() {
				triggered = false
				m.Get("/check", CheckThis(Passed, Passed, Passed), result)
				m.ServeHTTP(w, r)
				So(triggered, ShouldBeTrue)
			})
			Convey("should rejected all", func() {
				triggered = false
				m.Get("/check", CheckThis(Passed, Passed, Rejected, Passed), result)
				m.ServeHTTP(w, r)
				So(triggered, ShouldBeFalse)
			})
		})
		Convey("CheckAny", func() {
			Convey("should rejected", func() {
				triggered = false
				m.Get("/check", CheckAny(Rejected, Rejected, Rejected), result)
				m.ServeHTTP(w, r)
				So(triggered, ShouldBeFalse)
			})
			Convey("should pass", func() {
				triggered = false
				m.Get("/check", CheckAny(Passed, Passed, Rejected, Passed), result)
				m.ServeHTTP(w, r)
				So(triggered, ShouldBeTrue)
			})
		})
		Convey("AuthLogin", func() {
			Convey("should pass login", func() {
				triggered = false
				superdo = true
				m.Get("/check", AuthLogin, result)
				m.ServeHTTP(w, r)
				So(w.Header().Get("Location"), ShouldNotEqual, auth.Config.LoginRedirect)
				superdo = false
			})
			Convey("should redirect login", func() {
				triggered = false
				m.Get("/check", AuthLogin, result)
				m.ServeHTTP(w, r)
				So(w.Header().Get("Location"), ShouldEqual, auth.Config.LoginRedirect)
			})
		})
	})
}
