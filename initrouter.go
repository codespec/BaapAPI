package main

import (
	"github.com/BaapAPI/handlers"
	"github.com/BaapAPI/middleware"
	"github.com/gin-gonic/gin"
)

func initrouter(router *gin.Engine) {

	router.Static("/css", "./static/css")
	router.Static("/img", "./static/img")
	router.LoadHTMLGlob("templates/*")

	router.GET("/", handlers.IndexHandler)
	router.GET("/login", handlers.LoginHandler)
	router.GET("/google/auth", handlers.GoogleAuthHandler)
	router.GET("/facebook/auth", handlers.FaceBookAuthHandler)

	authorized := router.Group("/battle")
	authorized.Use(middleware.AuthorizeRequest())
	{
		authorized.GET("/field", handlers.FieldHandler)
	}

}
