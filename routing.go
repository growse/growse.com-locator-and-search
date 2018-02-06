package main

import (
	"github.com/gin-gonic/gin"
)

func BuildRoutes(router *gin.Engine) {

	authorized := router.Group("/auth/")
	authorized.Use(AuthRequired())
	{
		authorized.GET("ping", PingHandler)
		authorized.GET("where/", WhereHandler)
		authorized.GET("where/osm/:year/:filtered/", OSMWhereHandler)
		authorized.GET("where/linestring/:year/", WhereLineStringHandlerNonFiltered)
		authorized.GET("where/linestring/:year/:filtered/", WhereLineStringHandler)
	}
	router.GET("/oauth2callback", OauthCallback)

	router.GET("/where/", func(c *gin.Context) {
		c.Redirect(301, "/auth/where/")
	})

	router.POST("/search/", BleveSearchQuery)
	router.POST("/search/index", BleveIndexDocs)
	router.POST("/locator/", LocatorHandler)
	router.GET("/location/", LocationHandler)
	router.HEAD("/location/", LocationHeadHandler)
}

func PingHandler(c *gin.Context) {
	c.Status(201)
}
