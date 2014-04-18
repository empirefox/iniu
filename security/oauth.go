//group := Usernames{"123", "asd"}
//Use(oauth.Oauth(group))
//m.Get("/", Wanna(Usernames{"123", "asd"}))
package oauth

import (
	"github.com/bradrydzewski/go.auth"
	"github.com/dchest/uniuri"
	"github.com/deckarep/golang-set"
	"github.com/go-martini/martini"
	"net/http"
	"net/url"
)

func init() {

}

var (
	PathLogout = "/logout"
	OpenId     = auth.OpenId(auth.GoogleOpenIdEndpoint)
)

type Usernames []interface{}

type Expected interface{}

var Oauth = func(okPath string, expected ...Usernames) martini.Handler {
	auth.Config.CookieSecret = []byte(uniuri.New())
	auth.Config.LoginSuccessRedirect = okPath
	auth.Config.CookieSecure = martini.Env == martini.Prod

	return func(c martini.Context, w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			switch r.URL.Path {
			case auth.Config.LoginRedirect:
				OpenId.ServeHTTP(w, r)
			case PathLogout:
				Logout(w, r)
			}
		}
		names := initUsernames(expected)
		c.Map(names)
	}
}

func initUsernames(exps []Usernames) Usernames {
	var names Usernames
	for _, v := range exps {
		names = append(names, v...)
	}
	//if names == nil {
	//	names = make(Usernames)
	//}
	return names
}

func Logout(w http.ResponseWriter, r *http.Request) {
	auth.DeleteUserCookie(w, r)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

var Wanna = func(expected ...Usernames) martini.Handler {
	return func(c martini.Context, names Usernames, w http.ResponseWriter, r *http.Request) {
		user, err := auth.GetUserCookie(r)

		//if no active user session then authorize user
		if err != nil || user.Id() == "" {
			http.Redirect(w, r, auth.Config.LoginRedirect, http.StatusFound)
			return
		}

		//else, add the user to the URL and continue
		r.URL.User = url.User(user.Id())
		username := r.URL.User.Username()
		names = append(names, initUsernames(expected)...)
		set := mapset.NewSetFromSlice(names)
		if username == "" {
			http.Redirect(w, r, auth.Config.LoginRedirect, http.StatusFound)
		}
		if !set.Contains(username) {
			http.Redirect(w, r, PathLogout, http.StatusFound)
		}
	}
}
