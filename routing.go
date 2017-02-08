package main

import "github.com/gin-gonic/gin"

func BuildRoutes(router *gin.Engine) {
	//Ugly hack to deal with the fact that httprouter can't cope with both /static/ and /:year existing
	//All years will begin with 2. So this sort of helps. Kinda.
	authorized := router.Group("/auth/")
	authorized.Use(AuthRequired())
	{
		authorized.GET("articles/", AdminArticleHandler)
		authorized.POST("articles/", AdminNewArticleHandler)
		authorized.PUT("articles/:id/", AdminUpdateArticleHandler)
		authorized.GET("articles/:id/", AdminGetArticleHandler)
		authorized.DELETE("articles/:id/", AdminDeleteArticleHandler)
		authorized.POST("preview/", MarkdownPreviewHandler)
		authorized.GET("where/", WhereHandler)
		authorized.GET("where/osm/:year/:filtered/", OSMWhereHandler)
		authorized.GET("where/linestring/:year/", WhereLineStringHandlerNonFiltered)
		authorized.GET("where/linestring/:year/:filtered/", WhereLineStringHandler)
	}
	router.GET("/oauth2callback", OauthCallback)

	router.GET("/2:year/:month/", MonthHandler)
	router.GET("/2:year/:month/:day/:slug/", ArticleHandler)
	router.GET("/rss/", RSSHandler)
	router.GET("/", LatestArticleHandler)
	router.GET("/robots.txt", RobotsHandler)

	//Redirects
	router.GET("/news/rss/", func(c *gin.Context) {
		c.Redirect(301, "/rss/")
	})
	router.GET("/where/", func(c *gin.Context) {
		c.Redirect(301, "/auth/where/")
	})

	//Sitemap
	router.GET("/sitemap.xml", UncompressedSiteMapHandler)
	router.GET("/sitemap.xml.gz", CompressedSiteMapHandler)

	router.POST("/search/", SearchPostHandler)
	router.POST("/blevesearch", BleveSearchQuery)
	router.POST("/search/index", BleveIndexDocs)
	router.POST("/locator/", LocatorHandler)
	router.GET("/search/:searchterm/", SearchHandler)

}
