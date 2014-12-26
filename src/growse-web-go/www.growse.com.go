package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"gopkg.in/fsnotify.v1"
	"gopkgs.com/memcache.v1"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"
)

var (
	db                 *sql.DB
	templatePath       string
	staticPath         string
	cpuProfile         string
	stylesheetfilename string
	javascriptfilename string
	memcacheUrl        string
	dbUser             string
	dbName             string
	dbPassword         string
	dbHost             string
	memcacheClient     *memcache.Client
)

type ArticleMonth struct {
	FirstOfTheYear bool
	Year           int
	Month          int
	Count          int
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

func MonthHandler(c *gin.Context) {
	log.Print("MonthHandler")
	year, err := strconv.Atoi("2" + c.Params.ByName("year"))
	if err != nil {
		c.String(404, "404 Not Found")
		return
	}

	month, err := strconv.Atoi(c.Params.ByName("month"))
	if err != nil {
		c.String(404, "404 Not Found")
		return
	}

	resultSlug, err := memcacheClient.Get(fmt.Sprintf("growse.com-bymonth-%d-%d", year, month))
	if err == nil {
		c.Redirect(302, string(resultSlug.Value))
	} else {
		var article Article
		err := db.QueryRow("select id,datestamp, shorttitle,title from articles where date_part('year',datestamp at time zone 'UTC')=$1 and date_part('month',datestamp at time zone 'UTC')=$2 order by datestamp desc limit 1", year, month).Scan(&article.Id, &article.Timestamp, &article.Slug, &article.Title)
		if err != nil {
			c.String(404, err.Error())
		}
		redirect := article.GetAbsoluteUrl()
		memcacheClient.Set(&memcache.Item{Key: fmt.Sprintf("growse.com-bymonth-%d-%d", year, month), Value: []byte(redirect)})
		c.Redirect(302, redirect)
	}
}

func ArticleHandler(c *gin.Context) {
	log.Print("ArticleHandler")
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
	obj := gin.H{"Index": index, "Title": article.Title, "Months": months, "Article": article, "CurrentYear": time.Now().Year(), "Stylesheet": stylesheetfilename, "Javascript": javascriptfilename}

	c.HTML(200, "article.html", obj)
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

	c.HTML(200, "article.html", obj)
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

	monthRows, err := db.Query("with t as (select date_part('year',datestamp at time zone 'UTC') as year, date_part('month',datestamp at time zone 'UTC') as month, count(*) as c from articles group by date_part('year',datestamp at time zone 'UTC'),date_part('month',datestamp at time zone 'UTC') order by year desc, month desc) select case when lag(year,1) over () = year then false else true end as first, year,month,c from t;")
	defer monthRows.Close()
	if err != nil {
		return nil, nil, err
	}
	var months []ArticleMonth
	for monthRows.Next() {
		var month ArticleMonth
		monthRows.Scan(&month.FirstOfTheYear, &month.Year, &month.Month, &month.Count)

		months = append(months, month)

	}
	return &articles, &months, nil
}

func main() {
	yay := pq.ListenerEventConnected
	log.Print(yay)
	var err error
	db, err = sql.Open("postgres", fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s", dbHost, dbUser, dbName, dbPassword))
	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()
	gin.SetMode(gin.ReleaseMode)
	router.LoadHTMLGlob(path.Join(templatePath, "*.html"))
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

	//Detect changes to css / js and update those paths.
	watcher, err := fsnotify.NewWatcher()
	if err == nil {
		defer watcher.Close()
		watcher.Add(path.Join(staticPath, "css"))
		watcher.Add(path.Join(staticPath, "js"))
		go func() {
			for {
				select {
				case event := <-watcher.Events:
					if event.Op&fsnotify.Create == fsnotify.Create {
						if strings.HasSuffix(event.Name, ".www.css") {
							log.Printf("New CSS Detected: %s", path.Base(event.Name))
							stylesheetfilename = path.Base(event.Name)
						}
						if strings.HasSuffix(event.Name, ".www.js") {
							log.Printf("New JS Detected: %s", path.Base(event.Name))
							javascriptfilename = path.Base(event.Name)
						}
					}
				case err := <-watcher.Errors:
					log.Println("error:", err)
				}
			}
		}()
	} else {
		log.Printf("Initify watcher failed: %s Continuing.", err)
	}

	//Caching time
	memcacheClient = memcache.New(memcacheUrl)
	memcacheClient.SetMaxIdleConnsPerAddr(10)

	if cpuProfile != "" {
		f, err := os.Create(cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			log.Printf("captured %v, stopping profiler and exiting..", sig)
			pprof.StopCPUProfile()
			os.Exit(1)
		}
	}()

	//Ugly hack to deal with the fact that httprouter can't cope with both /static/ and /:year existing
	//All years will begin with 2. So this sort of helps. Kinda.
	router.GET("/2:year/:month/", MonthHandler)
	router.GET("/2:year/:month/:day/:slug/", ArticleHandler)
	router.GET("/", LatestArticleHandler)
	router.Run(":8080")
}

func init() {
	flag.StringVar(&templatePath, "templatePath", "", "The path to the templates")
	flag.StringVar(&staticPath, "staticPath", "", "The path to the static files")
	flag.StringVar(&cpuProfile, "cpuprofile", "", "write cpu profile to file")
	flag.StringVar(&memcacheUrl, "memcacheUrl", "tcp://localhost:11211", "Url to connect to memcache")
	flag.StringVar(&dbUser, "dbUser", "www_growse_com", "Postgres username")
	flag.StringVar(&dbPassword, "dbPassword", "", "Postgres database password")
	flag.StringVar(&dbName, "dbName", "www_growse_com", "Postgres database name")
	flag.StringVar(&dbName, "dbHost", "/var/run/postgres", "Postgres database host")

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