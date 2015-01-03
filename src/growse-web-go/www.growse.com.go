package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/growse/concurrent-expiring-map"
	"github.com/lib/pq"
	"github.com/mailgun/mailgun-go"
	"github.com/oxtoacart/bpool"
	"gopkg.in/fsnotify.v1"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path"
	"runtime/debug"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"
)

var (
	db                 *sql.DB
	stylesheetfilename string
	javascriptfilename string
	configuration      Configuration
	gun                mailgun.Mailgun
	templates          *template.Template
	bufPool            *bpool.BufferPool
	memoryCache        cmap.ConcurrentMap
)

type Configuration struct {
	DbUser             string
	DbName             string
	DbPassword         string
	DbHost             string
	TemplatePath       string
	StaticPath         string
	CpuProfile         string
	GeocodeApiURL      string
	MailgunKey         string
	Production         bool
	DefaultCacheExpiry time.Duration
}

type ArticleMonth struct {
	FirstOfTheYear bool
	Year           int
	Month          time.Month
	Count          int
}

var funcMap = template.FuncMap{
	"RenderFloat": RenderFloat,
}

func GetLatestArticle() (*Article, error) {
	var article Article
	articleBytes, ok := memoryCache.Get("growse.com-latest")
	if ok {
		article, err := FromBytes(articleBytes)
		if err != nil {
			return nil, err
		} else {
			return &article, nil
		}
	} else {
		err := db.QueryRow(`Select id,datestamp,shorttitle,title,markdown from articles where published=true order by datestamp desc limit 1`).Scan(&article.Id, &article.Timestamp, &article.Slug, &article.Title, &article.Markdown)
		switch {
		case err == sql.ErrNoRows:
			return nil, fmt.Errorf("No article found")
		case err != nil:
			return nil, err
		default:
			articleBytes, err := article.ToBytes()
			if err != nil {
				InternalError(err)
			} else {
				memoryCache.Set("growse.com-latest", articleBytes, time.Now().Add(configuration.DefaultCacheExpiry))
			}
			return &article, nil
		}
	}

}

func MonthHandler(c *gin.Context) {
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

	var resultSlug []byte

	resultSlug, ok := memoryCache.Get(fmt.Sprintf("growse.com-bymonth-%d-%d", year, month))

	if ok {
		c.Redirect(302, string(resultSlug))
	} else {
		var article Article
		err := db.QueryRow("select id,datestamp, shorttitle,title from articles where date_part('year',datestamp at time zone 'UTC')=$1 and date_part('month',datestamp at time zone 'UTC')=$2 order by datestamp desc limit 1", year, month).Scan(&article.Id, &article.Timestamp, &article.Slug, &article.Title)
		if err != nil {
			c.String(404, err.Error())
		}
		redirect := article.GetAbsoluteUrl()
		memoryCache.Set(fmt.Sprintf("growse.com-bymonth-%d-%d", year, month), []byte(redirect), time.Now().Add(configuration.DefaultCacheExpiry))
		c.Redirect(302, redirect)
	}
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

	//Check the page cache
	var cachedBytes []byte
	cacheKey := getCacheKey(year, month, day, slug)

	cachedBytes, ok := memoryCache.Get(cacheKey)

	if ok {
		c.Data(200, "text/html", cachedBytes)
		return
	}
	log.Printf("Cache MISS: %v", cacheKey)

	//Cache miss, load from DB
	article, err := GetArticle(year, month, day, slug)
	if err != nil {
		c.String(404, err.Error())
		return
	}

	//Get the indeces from DB
	index, months, err := LoadArticleIndex()
	if err != nil {
		InternalError(err)
		c.String(500, "Internal Error")
		return
	}

	lastlocation, err := GetLastLoction()
	if err != nil {
		InternalError(err)
		c.String(500, "Internal Error")
	}

	obj := gin.H{"Index": index, "Title": article.Title, "Months": months, "Article": article, "CurrentYear": time.Now().Year(), "Stylesheet": stylesheetfilename, "Javascript": javascriptfilename, "LastLocation": lastlocation}

	buf := bufPool.Get()
	buf.Reset()
	defer bufPool.Put(buf)
	err = templates.ExecuteTemplate(buf, "article.html", obj)
	pageBytes := buf.Bytes()
	//Cache the page

	if err == nil {
		memoryCache.Set(article.getCacheKey(), pageBytes, time.Now().Add(configuration.DefaultCacheExpiry))
		c.Data(200, "text/html", pageBytes)
	} else {
		InternalError(err)
		c.String(500, "Internal Error")
	}
}

func LatestArticleHandler(c *gin.Context) {
	//Get article latest
	article, err := GetLatestArticle()
	if err != nil {
		c.String(404, err.Error())
		return
	}

	var cacheBytes []byte
	cacheBytes, ok := memoryCache.Get(article.getCacheKey())

	if ok {
		c.Data(200, "text/html", cacheBytes)
		return
	}
	log.Printf("Cache MISS: %v", article.getCacheKey())

	lastlocation, err := GetLastLoction()
	if err != nil {
		InternalError(err)
		c.String(500, "Internal Error")
	}

	index, months, err := LoadArticleIndex()
	if err != nil {
		InternalError(err)
		c.String(500, "Internal Error")
		return
	}

	obj := gin.H{"Index": index, "Title": article.Title, "Months": months, "Article": article, "Stylesheet": stylesheetfilename, "Javascript": javascriptfilename, "LastLocation": lastlocation}

	buf := bufPool.Get()
	defer bufPool.Put(buf)
	err = templates.ExecuteTemplate(buf, "article.html", obj)
	pageBytes := buf.Bytes()

	if err == nil {
		memoryCache.Set(article.getCacheKey(), pageBytes, time.Now().Add(configuration.DefaultCacheExpiry))
		c.Data(200, "text/html", pageBytes)
	} else {
		InternalError(err)
		c.String(500, "Internal Error")
	}

}

func WhereHandler(c *gin.Context) {
	avgspeed, err := GetAverageSpeed()
	if err != nil {
		InternalError(err)
		c.String(500, "Internal Error")
		return
	}

	totaldistance, err := GetTotalDistance()
	if err != nil {
		InternalError(err)
		c.String(500, "Internal Error")
		return
	}

	lastlocation, err := GetLastLoction()
	if err != nil {
		InternalError(err)
		c.String(500, "Internal Error")
		return
	}
	obj := gin.H{"Title": "Where", "Stylesheet": stylesheetfilename, "Javascript": javascriptfilename, "Avgspeed": avgspeed, "Totaldistance": totaldistance, "LastLocation": lastlocation}
	buf := bufPool.Get()
	defer bufPool.Put(buf)

	err = templates.ExecuteTemplate(buf, "where.html", obj)
	pageBytes := buf.Bytes()
	if err == nil {
		c.Data(200, "text/html", pageBytes)
	} else {
		InternalError(err)
		c.String(500, "Internal Error")
	}
}

func WhereLineStringHandler(c *gin.Context) {
	linestring, err := GetLineStringAsJSON(c.Params.ByName("year"))
	if err != nil {
		InternalError(err)
		c.String(500, "Internal Error")
	}
	c.Data(200, "application/json", []byte(linestring))
}

func SearchPostHandler(c *gin.Context) {
	var searchForm struct {
		SearchTerm string `form:"a" binding:"required"`
	}
	c.Bind(&searchForm)
	c.Redirect(303, fmt.Sprintf("/search/%s/", searchForm.SearchTerm))
}

func SearchHandler(c *gin.Context) {
	searchterm := c.Params.ByName("searchterm")

	cacheKey := fmt.Sprintf("search-%v", searchterm)
	page, ok := memoryCache.Get(cacheKey)
	if ok {
		c.Data(200, "text/html", page)
		return
	}
	log.Printf("Cache MISS: %v", cacheKey)

	articles, err := SearchArticle(searchterm)
	if err != nil {
		InternalError(err)
		c.String(500, "Internal Error")
		return
	}

	lastlocation, err := GetLastLoction()
	if err != nil {
		InternalError(err)
		c.String(500, "Internal Error")
		return
	}

	obj := gin.H{"Searchterm": searchterm, "SearchResults": articles, "Title": fmt.Sprintf("%v :: search", searchterm), "Stylesheet": stylesheetfilename, "Javascript": javascriptfilename, "LastLocation": lastlocation}
	buf := bufPool.Get()
	defer bufPool.Put(buf)

	err = templates.ExecuteTemplate(buf, "search.html", obj)
	pageBytes := buf.Bytes()
	if err == nil {
		memoryCache.Set(cacheKey, pageBytes, time.Now().Add(5*time.Minute))
		c.Data(200, "text/html", pageBytes)
	} else {
		InternalError(err)
		c.String(500, "Internal Error")
	}
}

func RobotsHandler(c *gin.Context) {
	c.File(path.Join(configuration.TemplatePath, "robots.txt"))
}

func LoadTemplates() {
	templateGlob := path.Join(configuration.TemplatePath, "*.html")
	templates = template.Must(template.New("Yay Templates").Funcs(funcMap).ParseGlob(templateGlob))
}

func InternalError(err error) {
	log.Printf("%v", err)
	debug.PrintStack()
	if configuration.Production {
		m := mailgun.NewMessage("Sender <blogbot@growse.com>", "ERROR: www.growse.com", fmt.Sprintf("%v\n%v", err, string(debug.Stack())), "sysadmin@growse.com")
		response, id, _ := gun.Send(m)
		fmt.Printf("Response ID: %s\n", id)
		fmt.Printf("Message from server: %s\n", response)
	}
}

func main() {
	//Flags
	configFile := flag.String("configFile", "config.json", "File path to the JSON configuration")
	flag.Parse()

	//Config parsing
	file, err := os.Open(*configFile)
	if err != nil {
		log.Fatalf("Unable to open configuration file: %v", err)
	}

	decoder := json.NewDecoder(file)

	err = decoder.Decode(&configuration)
	if err != nil {
		log.Fatalf("Unable to parse configuration file: %v", err)
	}

	if configuration.TemplatePath == "" {
		log.Fatalf("No template directory supplied")
	}
	if configuration.StaticPath == "" {
		log.Fatalf("No static directory supplied")
	}
	if _, err := os.Stat(configuration.TemplatePath); os.IsNotExist(err) {
		log.Fatalf("No such file or directory: %s", configuration.TemplatePath)

	}
	if _, err := os.Stat(configuration.StaticPath); os.IsNotExist(err) {
		log.Fatalf("No such file or directory: %s", configuration.StaticPath)
	}

	gun = mailgun.NewMailgun("valid-mailgun-domain", "private-mailgun-key", "public-mailgun-key")

	//Initialize the template output buffer pool
	bufPool = bpool.NewBufferPool(16)

	//Get around auto removing of pq
	yay := pq.ListenerEventConnected
	log.Printf("Here's the number zero: %v", yay)

	//Database time

	connectionString := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s", configuration.DbHost, configuration.DbUser, configuration.DbName, configuration.DbPassword)
	db, err = sql.Open("postgres", connectionString)
	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}

	//Get the router
	router := gin.Default()
	gin.SetMode(gin.ReleaseMode)

	//Load the templates. Don't use gin for this, because we want to render to a buffer later
	LoadTemplates()
	//Static is over here
	router.Static("/static/", configuration.StaticPath)

	//Get latest updated stylesheet
	stylesheets, _ := ioutil.ReadDir(path.Join(configuration.StaticPath, "css"))
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
	//Get the latest javascript file
	javascripts, _ := ioutil.ReadDir(path.Join(configuration.StaticPath, "js"))
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
		watcher.Add(path.Join(configuration.StaticPath, "css"))
		watcher.Add(path.Join(configuration.StaticPath, "js"))
		watcher.Add(configuration.TemplatePath)
		go func() {
			for {
				select {
				case event := <-watcher.Events:
					if event.Op&fsnotify.Create == fsnotify.Create {
						if strings.HasSuffix(event.Name, ".www.css") {
							log.Printf("New CSS Detected: %s", path.Base(event.Name))
							stylesheetfilename = path.Base(event.Name)
							memoryCache.Flush()

						}
						if strings.HasSuffix(event.Name, ".www.js") {
							log.Printf("New JS Detected: %s", path.Base(event.Name))
							javascriptfilename = path.Base(event.Name)
							memoryCache.Flush()
						}
					} else if strings.HasSuffix(event.Name, ".html") {
						log.Printf("Template changed: %s", path.Base(event.Name))
						LoadTemplates()
						memoryCache.Flush()
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
	memoryCache = cmap.New()

	//Cpu Profiling time
	if configuration.CpuProfile != "" {
		f, err := os.Create(configuration.CpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	//Catch SIGTERM to stop the profiling
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
	router.GET("/where/", WhereHandler)
	router.GET("/where/linestring/:year/", WhereLineStringHandler)
	router.GET("/", LatestArticleHandler)
	router.GET("/robots.txt", RobotsHandler)
	router.POST("/search/", SearchPostHandler)
	router.GET("/search/:searchterm/", SearchHandler)
	router.Run(":8080")
}
