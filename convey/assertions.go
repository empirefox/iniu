package cv

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	zeroTimeRemove = regexp.MustCompile(`,"\S+":"0001-01-01 00:00:00 \+0000 UTC"`)
)

type Request func(m martini.Router) (*http.Request, error)

type structure map[string]interface{}
type slice []structure

func newStruct() structure {
	return structure{}
}

func newSlice() slice {
	return slice{}
}

// expected: data | [data, code]
func ShouldResponseOk(actual interface{}, expected ...interface{}) string {
	request, ok := actual.(func(m martini.Router) (*http.Request, error))
	if !ok {
		return "actual is not a struct of request"
	}

	m := martini.Classic()
	m.Use(render.Renderer())

	// routing

	res := httptest.NewRecorder()
	req, _ := request(m)
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}

	m.ServeHTTP(res, req)

	ecode := 200
	if len(expected) > 1 {
		c, ok := expected[1].(int)
		if !ok {
			return "can not know the status code"
		}
		ecode = c
	}

	if res.Code != ecode {
		return fmt.Sprintf("expected status code:%d, but got:%d\n", ecode, res.Code)
	}

	if ecode != 200 {
		return ""
	}

	if expStr, ok := expected[0].(string); ok {
		if expStr == "" {
			body := res.Body.String()
			switch body {
			case ``, `""`, `''`, `null`, `<nil>`, `nil`:
				return ""
			}
			return "should response empty string,but got this:\n" + body
		}
	}

	var left, right interface{}
	var leftPtr, rightPtr interface{}
	t := reflect.TypeOf(expected[0])
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() == reflect.Slice {
		l := slice{}
		r := slice{}
		left, right = l, r
		leftPtr, rightPtr = &l, &r
	} else {
		l := structure{}
		r := structure{}
		left, right = l, r
		leftPtr, rightPtr = &l, &r
	}

	bright, err := json.Marshal(expected[0])
	if err != nil {
		return "can not marshal expected!"
	}
	err = json.Unmarshal(bright, rightPtr)
	if err != nil {
		return "can not Unmarshal back expected!"
	}
	err = json.Unmarshal(res.Body.Bytes(), leftPtr)
	if err != nil {
		return "can not Unmarshal response body!"
	}

	return ShouldResemble(left, right)
}

func Res2Struct(res *httptest.ResponseRecorder, v interface{}) error {
	s := res.Body.String()
	r := zeroTimeRemove.ReplaceAllString(s, "")
	return json.Unmarshal([]byte(r), v)
	//return json.Unmarshal(res.Body.Bytes(), v)
}
