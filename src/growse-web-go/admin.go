package main

import "github.com/gin-gonic/gin"

func AdminArticleHandler(c *gin.Context) {
	rows, err := db.Query("select id,title,shorttitle,datestamp from articles order by datestamp desc")

	if err != nil {
		InternalError(err)
		c.String(500, "Internal Error")
		return
	}
	articles := []Article{}
	for rows.Next() {
		article := Article{}
		rows.Scan(&article.Id,&article.Title, &article.Slug, &article.Timestamp)
		articles = append(articles, article)
	}
	buf := bufPool.Get()
	buf.Reset()
	defer bufPool.Put(buf)

	obj := gin.H{
		"Stylesheet": stylesheetfilename,
		"Javascript": javascriptfilename,
		"Articles":    articles,
	}
	err = templates.ExecuteTemplate(buf, "admin_articlelist.html", obj)
	pageBytes := buf.Bytes()
	if err == nil {
		c.Data(200, "text/html", pageBytes)
	} else {
		InternalError(err)
		c.String(500, "Internal Error")
	}
}

func AdminNewArticleHandler(c *gin.Context){}
func AdminUpdateArticleHandler(c *gin.Context){}
func AdminDeleteArticleHandler(c *gin.Context){}
