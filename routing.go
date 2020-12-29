package main

import (
	_ "time"

	"github.com/gin-gonic/gin"
)

func BuildRoutes(router *gin.Engine) {
	authorized := router.Group("/where/")
	authorized.Use(AuthRequired())
	{
		authorized.GET("", func(c *gin.Context) {
			c.Redirect(301, "/where/ui/")
		})
		//authorized.GET("", PingHandler)
		authorized.Static("ui", "/var/www/owntracks-frontend")
		otRecorderAPI := authorized.Group("data")
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

	router.POST("/search/", BleveSearchQuery)
	router.GET("/location/", LocationHandler)
	router.HEAD("/location/", LocationHeadHandler)

}

func PingHandler(c *gin.Context) {
	url := c.Request.URL
	url.Host = "tracker.growse.com"
	url.Path = "/"
	c.Redirect(302, url.String())
}
