package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

var (
	db                 *sql.DB
	templatePath       string
	staticPath         string
	stylesheetfilename string
	javascriptfilename string
	memcacheClient     *memcache.Client
)

type ArticleMonth struct {
	Year  int
	Month int
	Count int
}

func loadLatestArticle() (*Article, error) {
	articleCacheKey, err := memcacheClient.Get("growse.com-latest")
	var article Article
	if err == nil {
		articleFromCache, err := memcacheClient.Get(string(articleCacheKey.Value))
		if err == nil {
			json.Unmarshal(articleFromCache.Value, &article)
		}
	} else {
		err := db.QueryRow(`Select id,datestamp,shorttitle,title,markdown from articles where published=true order by datestamp desc limit 1`).Scan(&article.Id, &article.Timestamp, &article.Slug, &article.Title, &article.Markdown)
		switch {
		case err == sql.ErrNoRows:
			return nil, fmt.Errorf("No article found")
		case err != nil:
			return nil, err
		default:
			err := article.cacheArticle()
			if err == nil {
				memcacheClient.Set(&memcache.Item{Key: "growse.com-latest", Value: []byte(article.getCacheKey())})
			}
		}
	}
	return &article, nil
}

func ArticleHandler(c *gin.Context) {
	year, err := strconv.Atoi("2" + c.Params.ByName("year"))
	if err != nil {
		c.String(404, err.Error())
		return
	}

	month, err := strconv.Atoi(c.Params.ByName("month"))
	if err != nil {
		c.String(404, err.Error())
		return
	}

	day, err := strconv.Atoi(c.Params.ByName("day"))
	if err != nil {
		c.String(404, err.Error())
		return
	}

	slug := c.Params.ByName("slug")

	article, err := GetArticle(year, month, day, slug)
	if err != nil {
		c.String(404, "404 Not Found")
		return
	}

	index, months, err := loadIndex()
	if err != nil {
		c.String(500, err.Error())
		return
	}
	obj := gin.H{"Index": index, "Title": article.Title, "Months": months, "Article": article, "Stylesheet": stylesheetfilename, "Javascript": javascriptfilename}

	c.HTML(200, "article.tmpl", obj)
}

func LatestArticleHandler(c *gin.Context) {
	article, err := loadLatestArticle()
	if err != nil {
		c.String(404, err.Error())
		return
	}
	index, months, err := loadIndex()
	if err != nil {
		c.String(500, err.Error())
		return
	}
	obj := gin.H{"Index": index, "Title": article.Title, "Months": months, "Article": article, "Stylesheet": stylesheetfilename, "Javascript": javascriptfilename}

	c.HTML(200, "article.tmpl", obj)
}

func loadIndex() (*[]Article, *[]ArticleMonth, error) {
	indexFromCache, err := memcacheClient.Get("growse.com-index")
	var articles []Article
	if err != nil {

		rows, err := db.Query("Select id, datestamp,shorttitle,title from articles where published=true order by datestamp desc;")
		if err != nil {
			return nil, nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var article Article
			rows.Scan(&article.Id, &article.Timestamp, &article.Slug, &article.Title)
			articles = append(articles, article)
		}
		indexToCache, _ := json.Marshal(articles)
		memcacheItem := memcache.Item{Key: "growse.com-index", Value: indexToCache}
		memcacheClient.Set(&memcacheItem)
	} else {
		json.Unmarshal(indexFromCache.Value, &articles)
	}

	monthRows, err := db.Query("select date_part('year',datestamp) as year, date_part('month',datestamp) as month, count(*) as count  from articles group by date_part('year',datestamp), date_part('month',datestamp) order by year asc, month asc")
	if err != nil {
		return nil, nil, err
	}
	defer monthRows.Close()
	var months []ArticleMonth
	for monthRows.Next() {
		var month ArticleMonth
		monthRows.Scan(&month.Year, &month.Month, &month.Count)
		months = append(months, month)

	}
	return &articles, &months, nil
}

func init() {
	flag.StringVar(&templatePath, "templatePath", "", "The path to the templates")
	flag.StringVar(&staticPath, "staticPath", "", "The path to the static files")

	flag.Parse()

	if templatePath == "" {
		log.Fatalf("No template directory supplied")
	}
	if staticPath == "" {
		log.Fatalf("No static directory supplied")
	}
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		log.Fatalf("No such file or directory: %s", templatePath)

	}
	if _, err := os.Stat(staticPath); os.IsNotExist(err) {
		log.Fatalf("No such file or directory: %s", staticPath)
	}

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
	router.LoadHTMLTemplates(path.Join(templatePath, "*.tmpl"))
	router.Static("/static/", staticPath)
	//Get latest updated stylesheet
	stylesheets, _ := ioutil.ReadDir(path.Join(staticPath, "css"))
	var lastTimeCss time.Time
	for _, file := range stylesheets {
		if !file.IsDir() && (lastTimeCss.IsZero() || file.ModTime().After(lastTimeCss)) && strings.HasSuffix(file.Name(), ".www.css") {
			lastTimeCss = file.ModTime()
			stylesheetfilename = file.Name()
		}
	}
	if stylesheetfilename == "" {
		log.Fatal("No stylesheet found in staticpath. Perhaps run Grunt first?")
	}

	javascripts, _ := ioutil.ReadDir(path.Join(staticPath, "js"))
	var lastTimeJs time.Time

	for _, file := range javascripts {
		if !file.IsDir() && (lastTimeJs.IsZero() || file.ModTime().After(lastTimeJs)) && strings.HasSuffix(file.Name(), ".www.js") {
			lastTimeJs = file.ModTime()
			javascriptfilename = file.Name()
		}
	}
	if javascriptfilename == "" {
		log.Fatal("No javascript found in staticpath. Perhaps run Grunt first?")
	}

	//Caching time
	memcacheClient = memcache.New("/tmp/memcache.sock")

	//Ugly hack to deal with the fact that httprouter can't cope with both /static/ and /:year existing
	//All years will begin with 2. So this sort of helps. Kinda.
	router.GET("/2:year/:month/:day/:slug", ArticleHandler)
	router.GET("/", LatestArticleHandler)
	router.Run(":8080")
	db.Close()
}
