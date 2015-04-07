package main

import (
    "github.com/gin-gonic/gin"
    "log"

    "github.com/russross/blackfriday"
    "regexp"
    "strings"
)

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
        rows.Scan(&article.Id, &article.Title, &article.Slug, &article.Timestamp)
        articles = append(articles, article)
    }
    buf := bufPool.Get()
    buf.Reset()
    defer bufPool.Put(buf)

    obj := gin.H{
        "Stylesheet": stylesheetfilename,
        "Javascript": javascriptfilename,
        "Articles":   articles,
        "Title":      "Wheeeeeeeeeeeeeeee",
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

func AdminNewArticleHandler(c *gin.Context) {
    type NewArticleForm struct {
        Title    string `form:"title" binding:"required"`
        Markdown string `form:"markdown" binding:"required"`
    }
    var form NewArticleForm
    c.Bind(&form)

    log.Printf("Title: %s", form.Title)
    slug := slugify(form.Title)
    log.Printf("Slug: %s", slug)
    searchtext := StripTags(string(blackfriday.MarkdownCommon(([]byte)(form.Markdown))))
    log.Printf("markdown: %s", form.Markdown)
    log.Printf("searchtext: %s", searchtext)

    _, err := db.Exec("insert into articles (title,shorttitle,markdown,searchtext, published) values ($1,$2,$3,$4,$5)", form.Title, slug, form.Markdown, searchtext, true)
    if err != nil {
        log.Printf("Error writing to db: %v", err)
    }
    c.Redirect(302, "/auth/articles/")
}
func AdminUpdateArticleHandler(c *gin.Context) {
    c.Redirect(302, "/auth/articles/")
}
func AdminDeleteArticleHandler(c *gin.Context) {
    c.Redirect(302, "/auth/articles/")
}

func slugify(title string) (slug string) {
    regex := regexp.MustCompile("[^a-zA-Z0-9]+")
    //re.sub("[^a-zA-Z0-9]+", "-", self.shorttitle.lower()).lstrip('-').rstrip('-')
    return strings.TrimPrefix(strings.TrimSuffix(string(regex.ReplaceAll([]byte(strings.ToLower(title)), []byte("-"))), "-"), "-")
}
