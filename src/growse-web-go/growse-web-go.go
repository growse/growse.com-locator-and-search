package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/julienschmidt/httprouter"
	"github.com/lib/pq"
	"github.com/oxtoacart/bpool"
	"github.com/russross/blackfriday"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

var (
	templates          map[string]*template.Template
	bufpool            *bpool.BufferPool
	db                 *sql.DB
	stylesheetname     string
	javascriptfilename string
)

type Article struct {
	Id        int
	Timestamp time.Time
	Slug      string
	Title     string
	Markdown  string
	Rendered  string
}

func loadLatestArticle() (*Article, error) {
	var a Article
	err := db.QueryRow(`Select id,datestamp,shorttitle,title,markdown from articles where published=true order by datestamp desc limit 1`).Scan(&a.Id, &a.Timestamp, &a.Slug, &a.Title, &a.Markdown)
	switch {
	case err == sql.ErrNoRows:
		return nil, fmt.Errorf("No article found")
	case err != nil:
		return nil, err
	default:
		a.Rendered = (string)(blackfriday.MarkdownCommon(([]byte)(a.Markdown)))
		return &a, nil
	}
}

func loadArticle(slug string) (*Article, error) {
	var a Article
	err := db.QueryRow(`Select id,datestamp,shorttitle,title,markdown from articles where shorttitle=$1 and published=true`, slug).Scan(&a.Id, &a.Timestamp, &a.Slug, &a.Title, &a.Markdown)
	switch {
	case err == sql.ErrNoRows:
		return nil, fmt.Errorf("No article found")
	case err != nil:
		return nil, err
	default:
		a.Rendered = (string)(blackfriday.MarkdownCommon(([]byte)(a.Markdown)))
		return &a, nil
	}
}

func loadIndex() (*[]Article, error) {
	rows, err := db.Query("Select id, datestamp,shorttitle,title from articles where published=true order by datestamp desc;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var articles []Article
	for rows.Next() {
		var article Article
		rows.Scan(&article.Id, &article.Timestamp, &article.Slug, &article.Title)
		articles = append(articles, article)
	}
	return &articles, nil
}

func LatestArticleHandler(c *gin.Context) {
	article, err := loadLatestArticle()
	if err != nil {
		c.String(400, err.Error())
		return
	}
	index, err := loadIndex()
	if err != nil {
		c.String(500, err.Error())
		return
	}
	obj := gin.H{"Index": index, "Title": "yaaaaay", "Article": article, "Stylesheet": stylesheetname, "Javascript": javascriptfilename}

	c.HTML(200, "base.tmpl", obj)
}

func ArticleHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

}

func (article *Article) GetAbsoluteUrl() string {
	return ""
}

func RenderTemplate(w http.ResponseWriter, name string, data map[string]interface{}) error {
	tmpl, ok := templates[name]
	if !ok {
		return fmt.Errorf("The template %s was not found", name)
	}
	buf := bufpool.Get()
	err := tmpl.ExecuteTemplate(buf, "base", data)
	if err != nil {
		bufpool.Put(buf)
		return err
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
	bufpool.Put(buf)
	return nil
}

func init() {
	bufpool = bpool.NewBufferPool(64)
	templates = make(map[string]*template.Template)
}

func main() {
	yay := pq.ListenerEventConnected
	log.Print(yay)
	var err error
	db, err = sql.Open("postgres", "user=andrew dbname=www_growse_com sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()
	gin.SetMode(gin.ReleaseMode)
	staticPath := "/Users/andrew/Projects/growse-web-go/static/"
	router.LoadHTMLTemplates("/Users/andrew/Projects/growse-web-go/templates/*.tmpl")
	router.Static("/static/", staticPath)
	//Get latest updated stylesheet
	stylesheets, _ := ioutil.ReadDir(staticPath + "css/")
	var lastTime time.Time
	for _, file := range stylesheets {
		if !file.IsDir() && (lastTime == nil || file.ModTime() > lastTime) && strings.HasSuffix(file.Name(), ".www.css") {
			lastTime = file.ModTime()
			stylesheetname = file.Name()
		}
	}

	javascripts, _ := filepath.Glob(staticPath + "js/*.www.js")
	javascriptfilename = "/static/js/" + filepath.Base(javascripts[0])
	router.GET("/", LatestArticleHandler)
	//router.GET("/:year/:month/:day/:slug", ArticleHandler)
	router.Run(":8080")
	db.Close()
}
