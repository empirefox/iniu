package security

import (
	"errors"
	"flag"
	"net/http"
	"time"

	"github.com/bradrydzewski/go.auth"
	"github.com/dchest/uniuri"
	"github.com/go-martini/martini"
	"github.com/golang/glog"
	"github.com/martini-contrib/render"

	. "github.com/empirefox/iniu/base"
)

const (
	PathLogout = "/logout"
	PathOpenId = "/login/openid"
	PathGoogle = "/login/google"
	PathGithub = "/login/github"

	NO_CHECK = iota - 1
	WANNA_CHECK
	CHECK_ALL
	CHECK_THIS
	CHECK_ANY
	CHECK_LOGIN
)

var (
	superdo = false
	AuthMap = map[string]func(w http.ResponseWriter, r *http.Request){}
)

func init() {
	flag.BoolVar(&superdo, "superdo", false, "when true ,you can set a perm value in url. default is false")
	flag.Parse()
}

type Config struct {
	CookieSecret          []byte
	CookieName            string
	CookieExp             time.Duration
	CookieMaxAge          int
	DisableCookieSecure   bool
	DisableCookieHttpOnly bool
	LoginRedirect         string
	LoginSuccessRedirect  string
	PathLogout            string
	PathOpenId            string
	Google                []string // path, client, secret, scope
	Github                []string // path, client, secret, redirect
}

func (config *Config) path() {
	if config.PathLogout == "" {
		config.PathLogout = PathLogout
	}
	AuthMap[config.PathLogout] = Logout

	if config.PathOpenId != "" {
		AuthMap[config.PathOpenId] = auth.OpenId(auth.GoogleOpenIdEndpoint).ServeHTTP
	}

	AddHandler(auth.Google, config.Google, PathGoogle)
	AddHandler(auth.Github, config.Github, PathGithub)
}

func (src *Config) auth() {
	if src.CookieSecret == nil {
		auth.Config.CookieSecret = []byte(uniuri.New())
	} else {
		auth.Config.CookieSecret = src.CookieSecret
	}
	if src.CookieName != "" {
		auth.Config.CookieName = src.CookieName
	}
	if src.CookieExp != time.Duration(0) {
		auth.Config.CookieExp = src.CookieExp
	}
	auth.Config.CookieMaxAge = src.CookieMaxAge
	// MARTINI_ENV
	auth.Config.CookieSecure = !src.DisableCookieSecure && martini.Env != martini.Dev
	auth.Config.CookieHttpOnly = !src.DisableCookieHttpOnly
	if src.LoginRedirect != "" {
		auth.Config.LoginRedirect = src.LoginRedirect
	}
	if src.LoginSuccessRedirect != "" {
		auth.Config.LoginSuccessRedirect = src.LoginSuccessRedirect
	}
}

func AddHandler(handler func(client, secret, redirectOrScope string) *auth.AuthHandler, vars []string, rbPath ...string) {
	switch len(vars) {
	case 3:
		if len(rbPath) > 0 && rbPath[0] != "" {
			AuthMap[rbPath[0]] = handler(vars[0], vars[1], vars[2]).ServeHTTP
		} else {
			panic(errors.New("auth handler error"))
		}
	case 4:
		AuthMap[vars[0]] = handler(vars[1], vars[2], vars[3]).ServeHTTP
	}
}

var pathHandle = func(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == "GET" {
		if authHandler, ok := AuthMap[r.URL.Path]; ok {
			authHandler(w, r)
			return true
		}
	}
	return false
}

var sudo = func(c martini.Context, r *http.Request) {
	r.ParseForm()
	perm := r.FormValue("perm")
	if perm == "" {
		perm = "*"
	}
	glog.Infoln("sudo perm", perm)
	c.Map(&Account{
		Name:      "empirefox",
		Enabled:   true,
		HoldsPerm: perm,
	})
	c.Map(&Oauth{
		Oid:       "empirefox@sina.com",
		Provider:  "google",
		Name:      "empirefox@sina",
		Enabled:   true,
		Validated: true,
	})
}

var login = func(c martini.Context, redirect bool, r render.Render, req *http.Request) {
	user, err := auth.GetUserCookie(req)
	// does not logged in
	if err != nil || user.Id() == "" {
		if redirect {
			r.Redirect(auth.Config.LoginRedirect)
		} else {
			c.Map(&Account{})
			c.Map(&Oauth{})
		}
		return
	}
	account, current := FindAccount(user.Provider(), user.Id())
	c.Map(account)
	c.Map(current)
}

// will init Forms
//      check if superdo is been set
var Prepare = func(config Config, mustLogin bool, vs ...ValidateFunc) martini.Handler {
	InitForms()
	config.auth()
	config.path()
	if len(vs) == 0 {
		vs = []ValidateFunc{}
	}

	return func(c martini.Context, w http.ResponseWriter, req *http.Request, r render.Render) {
		if pathHandle(w, req) {
			return
		}

		// should map before sudo to avoid panic when get []ValidateFunc
		c.Map(vs)

		if superdo {
			sudo(c, req)
			return
		}

		login(c, mustLogin, r, req)
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	auth.DeleteUserCookie(w, r)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

var CheckWeb = func(method string) martini.Handler {
	return func(t Table, r render.Render, a *Account, model Model) {
		p := WebPerm(t, method)
		if !a.Permitted(p) {
			r.JSON(http.StatusUnauthorized, "")
		}
	}
}

type ValidateFunc func(*Account, *Oauth) bool

var WannaCheck = func(now ...ValidateFunc) martini.Handler {
	return func(c martini.Context, prev []ValidateFunc) {
		c.Map(append(prev, now...))
	}
}

var CheckAll = func(now ...ValidateFunc) martini.Handler {
	return func(prev []ValidateFunc, r render.Render, a *Account, c *Oauth) {
		for _, f := range append(prev, now...) {
			if !f(a, c) {
				r.JSON(http.StatusForbidden, "")
			}
		}
	}
}

var CheckThis = func(now ...ValidateFunc) martini.Handler {
	return func(r render.Render, a *Account, c *Oauth) {
		for _, f := range now {
			if !f(a, c) {
				r.JSON(http.StatusForbidden, "")
			}
		}
	}
}

var CheckAny = func(now ...ValidateFunc) martini.Handler {
	return func(prev []ValidateFunc, r render.Render, a *Account, c *Oauth) {
		for _, f := range append(prev, now...) {
			if f(a, c) {
				return
			}
			r.JSON(http.StatusForbidden, "")
		}
	}
}

var AuthLogin = func(r render.Render, a *Account) {
	if a.Id == 0 {
		if a.Name == "empirefox" {
			return
		}
		r.Redirect(auth.Config.LoginRedirect)
	}
}
