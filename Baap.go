package main

import (
	"io"
	"os"
	"path/filepath"

	"github.com/BaapAPI/baaplogger"
	"github.com/BaapAPI/handlers"
	"github.com/BaapAPI/middleware"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

var (
	log    *baaplogger.Baaplogger
	ginlog *baaplogger.Logger
)

func init() {
	initlogger()
}

func main() {
	log.Informational("service start")
	gin.DefaultWriter = io.MultiWriter(ginlog, os.Stdout)
	router := gin.Default()

	store := sessions.NewCookieStore([]byte(handlers.RandToken(64)))
	store.Options(sessions.Options{
		Path:   "/",
		MaxAge: 86400 * 7,
	})
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(sessions.Sessions("goquestsession", store))
	router.Static("/css", "./static/css")
	router.Static("/img", "./static/img")
	router.LoadHTMLGlob("templates/*")

	router.GET("/", handlers.IndexHandler)
	router.GET("/login", handlers.LoginHandler)
	router.GET("/auth", handlers.AuthHandler)

	authorized := router.Group("/battle")
	authorized.Use(middleware.AuthorizeRequest())
	{
		authorized.GET("/field", handlers.FieldHandler)
	}

	router.Run("127.0.0.1:9090")
	log.Informational("service end")
}

func initlogger() {
	dir, err := filepath.Abs("log")
	if err != nil {
		panic(err)
	}

	// this for baap API log
	log = &baaplogger.Baaplogger{
		Level: baaplogger.LevelDebug,
		Log: &baaplogger.Logger{
			Filename:   filepath.Join(dir, "baap.log"),
			MaxSize:    500, // megabytes
			MaxBackups: 6,
			MaxAge:     28, // days
		},
	}

	// this for log inforamtion in gin
	ginlog = &baaplogger.Logger{
		Filename:   filepath.Join(dir, "gin.log"),
		MaxSize:    500, // megabytes
		MaxBackups: 6,
		MaxAge:     28, // days
	}

}
