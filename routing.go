package main

import (
	"github.com/gin-gonic/gin"
)

func BuildRoutes(router *gin.Engine) {

	authorized := router.Group("/auth/")
	authorized.Use(AuthRequired())
	{
		authorized.GET("", PingHandler)
		authorized.GET("where/linestring/:year/", WhereLineStringHandler)
		api := authorized.Group("location/api/0")
		{
			api.GET("list", OTListUserHandler)
		}
	}
	router.GET("/oauth2callback", OauthCallback)

	router.GET("/where/", func(c *gin.Context) {
		c.Redirect(301, "/auth/where/")
	})

	router.POST("/search/", BleveSearchQuery)
	router.GET("/location/", LocationHandler)
	router.HEAD("/location/", LocationHeadHandler)
}

func PingHandler(c *gin.Context) {
	c.Status(201)
}
