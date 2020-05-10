package main

import (
	_ "time"

	"github.com/gin-gonic/gin"
)

func BuildRoutes(router *gin.Engine) {
	authorized := router.Group("/auth/")
	authorized.Use(AuthRequired())
	{
		authorized.GET("", PingHandler)
		api := authorized.Group("location/api/0")
		{
			api.GET("list", OTListUserHandler)
		}
		otRecorderAPI := authorized.Group("location")
		{
			restAPI := otRecorderAPI.Group("api/0")
			{
				restAPI.GET("list", OTListUserHandler)
				restAPI.GET("last", OTLastPosHandler)
				restAPI.GET("locations", OTLocationsHandler)
				restAPI.GET("version", OTVersionHandler)
			}
			wsAPI := otRecorderAPI.Group("ws")
			{
				wsAPI.GET("last", func(c *gin.Context) {
					wshandler(c.Writer, c.Request)
				})
			}
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
