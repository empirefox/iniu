package security

import (
	"flag"
	"github.com/bradrydzewski/go.auth"
	"github.com/dchest/uniuri"
	"github.com/empirefox/iniu/comm"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"net/http"
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
	Skip    bool
	AuthMap = map[string]func(w http.ResponseWriter, r *http.Request){
		PathOpenId: auth.OpenId(auth.GoogleOpenIdEndpoint).ServeHTTP,
		PathLogout: Logout,
	}
)

func init() {
	flag.BoolVar(&Skip, "oauth-skip", false, "是否跳过认证，默认false")
	flag.Parse()
}

func AddGoogle(ak, sk, redirect string) {
	AuthMap[PathGoogle] = auth.Google(ak, sk, redirect).ServeHTTP
}

func AddGithub(ak, sk, scope string) {
	AuthMap[PathGithub] = auth.Github(ak, sk, scope).ServeHTTP
}

//准备登录逻辑，在内存中加载全部Form
var Prepare = func(okPath string, v ...ValidateFunc) martini.Handler {
	InitForms()

	if len(v) == 0 {
		v = []ValidateFunc{}
	}
	auth.Config.CookieSecret = []byte(uniuri.New())
	auth.Config.LoginSuccessRedirect = okPath
	auth.Config.CookieSecure = martini.Env == martini.Prod

	return func(c martini.Context, w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			if authHandler := AuthMap[r.URL.Path]; authHandler != nil {
				authHandler(w, r)
			}
		}
		c.Map(v)
		c.Map(&Account{})
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	auth.DeleteUserCookie(w, r)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type ValidateFunc func(*Account, *Oauth) bool

func addVFunc(a, b []ValidateFunc) []ValidateFunc {
	lb := len(b)
	if lb == 0 {
		return a
	}
	la := len(a)
	if la == 0 {
		return b
	}
	l := la + lb
	v := make([]ValidateFunc, l, l)
	v = append(v, a...)
	v = append(v, b...)
	return v
}

var WannaCheck = func(now ...ValidateFunc) martini.Handler {
	return func(c martini.Context, prev []ValidateFunc) {
		v := addVFunc(now, prev)
		c.MapTo(v, (*[]ValidateFunc)(nil))
	}
}

var CheckAll = func(now ...ValidateFunc) martini.Handler {
	return func(prev []ValidateFunc, r render.Render, req *http.Request) {
		v := addVFunc(now, prev)
		innerCheck(v, CHECK_ALL, r, req)
	}
}

var CheckThis = func(now ...ValidateFunc) martini.Handler {
	return func(r render.Render, req *http.Request) {
		innerCheck(now, CHECK_THIS, r, req)
	}
}

var CheckAny = func(now ...ValidateFunc) martini.Handler {
	return func(prev []ValidateFunc, r render.Render, req *http.Request) {
		v := addVFunc(now, prev)
		innerCheck(v, CHECK_ANY, r, req)
	}
}

var NoCheck = func() martini.Handler {
	return func() {
	}
}

func innerCheck(v []ValidateFunc, vType int, r render.Render, req *http.Request) {
	//跳过验证
	if Skip || len(v) == 0 {
		return
	}

	user, err := auth.GetUserCookie(req)
	//没有登录
	if err != nil || user.Id() == "" {
		r.Redirect(auth.Config.LoginRedirect, http.StatusFound)
		return
	}

	//验证过程
	a, c := FindAccount(user.Provider(), user.Id())
	for _, f := range v {
		switch vType {
		case CHECK_ALL, CHECK_THIS:
			if !f(a, c) {
				comm.JsonNoPerm(r)
				return
			}
			return
		case CHECK_ANY:
			if f(a, c) {
				return
			}
			comm.JsonNoPerm(r)
			return
		default:
			comm.Json(r, 1, "未知验证类型")
			return
		}
	}
}

//获取账户并map,如果没有登录则转到登录页面
var CheckLogin = func() martini.Handler {
	return func(c martini.Context, r render.Render, req *http.Request) {
		//跳过验证
		if Skip {
			return
		}

		user, err := auth.GetUserCookie(req)
		//没有登录
		if err != nil || user.Id() == "" {
			r.Redirect(auth.Config.LoginRedirect, http.StatusFound)
		}
		account, current := FindAccount(user.Provider(), user.Id())
		c.Map(account)
		c.Map(current)
	}
}

//获取账户并map,如果没有登录则map空账户
var CheckLoginQuiet = func() martini.Handler {
	return func(c martini.Context, r render.Render, req *http.Request) {
		//跳过验证
		if Skip {
			return
		}

		user, err := auth.GetUserCookie(req)
		//没有登录
		if err != nil || user.Id() == "" {
			c.Map(&Account{})
			c.Map(&Oauth{})
			return
		}
		account, current := FindAccount(user.Provider(), user.Id())
		c.Map(account)
		c.Map(current)
	}
}
