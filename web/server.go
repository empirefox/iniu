package web

import (
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/empirefox/gotool/paas"
	"github.com/empirefox/iniu/base"
	"github.com/empirefox/iniu/webcore"
	"github.com/empirefox/iniu/wo2"
	"github.com/gin-gonic/gin"
	"github.com/itsjamie/gin-cors"
	"github.com/jinzhu/gorm"
)

var log = logrus.New()

type serverInjector interface {
	Serve(*gin.Engine, *gorm.DB)
}

type Server struct {
	*gin.Engine
	DB         *gorm.DB
	AuthConfig *wo2.Config
	Origins    string

	authMiddleWare gin.HandlerFunc
}

func NewServer(authConfig *wo2.Config, db *gorm.DB, origins string) *Server {
	s := &Server{
		DB:         db,
		AuthConfig: authConfig,
		Origins:    origins,
	}

	s.authMiddleWare = authConfig.Middleware()

	s.Use(s.Cors("GET, PUT, POST, DELETE"))

	webcore.RegisterCallbacks(db)
	return s
}

func (s *Server) StartRun() {
	s.POST(s.AuthConfig.WxOauthPath, s.authMiddleWare, s.Ok)
	optPaths := make(map[string]bool)
	rs := s.Routes()
	for _, r := range rs {
		if r.Method == "OPTIONS" {
			optPaths[r.Path] = true
		}
	}
	for _, r := range rs {
		if !optPaths[r.Path] {
			s.OPTIONS(r.Path, s.Ok)
		}
	}
	s.Run(paas.BindAddr)
}

func (s *Server) AddResource(m base.Model) {
	res := webcore.NewResource(m, s.DB)
	if res == nil {
		log.WithField("Model", base.Formname(m)).Fatalln("NewResource failed")
	}

	table := s.Group("/"+res.Struct.Name, s.authMiddleWare)
	table.GET("/searchmetas", res.SearchMetas)
	table.GET("/metas", res.FormMetas)
	table.GET("/default", res.Default)
	table.POST(`/\+1`, res.Create)
	table.POST("/1", res.Update)
	table.GET("/1/:id", res.OneById)
	table.GET("/ls", res.ListMany)
	table.GET("/pkshows", res.PkShows)
	table.DELETE("/del", res.Delete)

	pos := table.Group("/pos")
	pos.POST("/get", res.GetPos)
	pos.POST("/re", res.RePos)
	pos.POST("/bound", res.PosBound)
	pos.POST("/x", res.XchIpGtOrLt)
	pos.POST("/next", res.PosGtOrLtAnd)
	pos.POST("/save", res.SavePosAll)

	table.OPTIONS("/searchmetas", s.Ok)

	if injector, ok := m.(serverInjector); ok {
		injector.Serve(s.Engine, s.DB)
	}
}

func (s *Server) Cors(method string) gin.HandlerFunc {
	return cors.Middleware(cors.Config{
		Origins:         s.Origins,
		Methods:         method,
		RequestHeaders:  "Origin, Authorization, Content-Type",
		ExposedHeaders:  "",
		MaxAge:          480 * time.Hour,
		Credentials:     false,
		ValidateHeaders: false,
	})
}

func (s *Server) Ok(c *gin.Context)       { c.AbortWithStatus(http.StatusOK) }
func (s *Server) NotFound(c *gin.Context) { c.AbortWithStatus(http.StatusNotFound) }
