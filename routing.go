package main

import (
	"github.com/gin-gonic/gin"
	"github.com/growse/gin-oauth2/google"
)

func BuildRoutes(router *gin.Engine, staticDir string) {

	google.SetupWithCreds(configuration.OAuth2CallbackUrl, configuration.ClientID, configuration.ClientSecret, []string{"openid", "email"}, []byte("secret"))
	authorized := router.Group("/auth/")
	authorized.Use(google.Auth())
	{
		//authorized.Static("static/", staticDir)
		authorized.GET("where/", WhereHandler)
		authorized.GET("where/osm/:year/:filtered/", OSMWhereHandler)
		authorized.GET("where/linestring/:year/", WhereLineStringHandlerNonFiltered)
		authorized.GET("where/linestring/:year/:filtered/", WhereLineStringHandler)
	}
	router.GET("/oauth2callback", google.LoginHandler)

	router.GET("/where/", func(c *gin.Context) {
		c.Redirect(301, "/auth/where/")
	})

	router.POST("/search/", SearchPostHandler)
	router.POST("/blevesearch", BleveSearchQuery)
	router.POST("/search/index", BleveIndexDocs)
	router.POST("/locator/", LocatorHandler)
	router.GET("/location/", LocationHandler)
	router.HEAD("/location/", LocationHeadHandler)
	router.GET("/search/:searchterm/", SearchHandler)
}
