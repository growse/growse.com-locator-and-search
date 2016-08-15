package main

import (
	"bytes"
	"compress/gzip"
	"database/sql"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/feeds"
	"github.com/russross/blackfriday"
	"html/template"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Article struct {
	Id        int
	Timestamp time.Time
	Slug      string
	Title     string
	Markdown  string
}

func (article Article) ToBytes() ([]byte, error) {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(article)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func FromBytes(data []byte) (Article, error) {
	var article Article
	var buffer bytes.Buffer
	buffer.Write(data)
	dec := gob.NewDecoder(&buffer)
	err := dec.Decode(&article)
	return article, err
}

func (article *Article) getCacheKey() string {
	return Truncate(fmt.Sprintf("article-%d-%02d-%02d-%s", article.Timestamp.UTC().Year(), article.Timestamp.UTC().Month(), article.Timestamp.UTC().Day(), article.Slug), 250)
}

func getCacheKey(year int, month int, day int, slug string) string {
	return Truncate(fmt.Sprintf("article-%d-%02d-%02d-%s", year, month, day, slug), 250)
}

func SearchArticle(searchterm string) (*[]Article, error) {
	rows, err := db.Query("select id,datestamp,shorttitle,title,searchtext from articles where idxfti @@ plainto_tsquery('english',$1) and published=true order by ts_rank(idxfti,plainto_tsquery('english',$1)) desc", searchterm)

	if err != nil {
		return nil, err
	}
	var searchresults []Article
	for rows.Next() {
		var article Article
		rows.Scan(&article.Id, &article.Timestamp, &article.Slug, &article.Title, &article.Markdown)
		searchresults = append(searchresults, article)
	}
	return &searchresults, nil
}

func GetArticle(year int, month int, day int, slug string) (*Article, error) {
	var article Article
	err := db.QueryRow(`Select id,datestamp,shorttitle,title,markdown from articles where published=true and shorttitle=$1 and date_part('year',datestamp at time zone 'UTC')=$2 and date_part('month',datestamp at time zone 'UTC')=$3 and date_part('day',datestamp at time zone 'UTC')=$4 limit 1`, slug, year, month, day).Scan(&article.Id, &article.Timestamp, &article.Slug, &article.Title, &article.Markdown)
	if err == nil {
		return &article, nil
	}
	return nil, err
}

func LoadArticleIndex() (*[]Article, *[]ArticleMonth, error) {
	var articles []Article
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

func (article *Article) GetAbsoluteUrl() string {
	return fmt.Sprintf("/%d/%02d/%02d/%s/", article.Timestamp.UTC().Year(), article.Timestamp.UTC().Month(), article.Timestamp.UTC().Day(), article.Slug)
}

func removePunctuation(text string) string {
	//pattern := regexp.MustCompile("([^\\s\\w]|_)+")
	pattern := regexp.MustCompile("[^\\P{P}-]+")
	return pattern.ReplaceAllString(text, "")
}

func SmartTruncateWithHighlight(input string, searchterm string, surroundingWords int, suffix string, highlightFormat string) string {
	words := strings.Split(input, " ")
	searchterm = removePunctuation(searchterm)

	index := indexOfCaseInsensitive(words, searchterm)
	result := ""
	if index >= 0 {
		words[index] = fmt.Sprintf(highlightFormat, words[index])
		startIndex := index - surroundingWords
		endIndex := index + surroundingWords
		if startIndex < 0 {
			startIndex = 0
			endIndex = 2 * surroundingWords
		}
		if endIndex >= len(words) {
			endIndex = len(words) - 1
			startIndex = endIndex - (2 * surroundingWords)
			if startIndex < 0 {
				startIndex = 0
			}
		}
		result = strings.Join(words[startIndex:endIndex+1], " ")
		if startIndex > 0 {
			result = suffix + result
		}
		if endIndex < len(words)-1 {
			result += suffix
		}
	} else {
		if 2*surroundingWords > len(words)-1 {
			result = strings.Join(words, " ")
		} else {
			result = strings.Join(words[0:2*surroundingWords], " ") + suffix
		}
	}
	return result
}

func indexOfCaseInsensitive(slice []string, thing string) int {
	for index, item := range slice {
		if strings.ToLower(removePunctuation(item)) == strings.ToLower(thing) {
			return int(index)
		}
	}
	return -1
}

func (article *Article) RenderAsSearchResult(searchterm string) template.HTML {
	return template.HTML(SmartTruncateWithHighlight(article.Markdown, searchterm, 15, "...", "<b>%s</b>"))
}

func (article *Article) Rendered() template.HTML {
	return template.HTML(blackfriday.MarkdownCommon(([]byte)(article.Markdown)))
}

func (article *Article) GetUnsafeAbsoluteUrl() template.URL {
	return template.URL(article.GetAbsoluteUrl())
}

func Truncate(input string, length int) string {
	if len(input) < length {
		return input
	} else {
		return string(input[:length])
	}
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
			return nil, errors.New("No article found")
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

/* HTTP Handlers */

func MarkdownPreviewHandler(c *gin.Context) {
	var input struct {
		Markdown string `form:"markdown" binding:"required"`
	}
	c.Bind(&input)
	rendered := blackfriday.MarkdownCommon(([]byte)(input.Markdown))
	c.Data(200, "text/html", rendered)
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
	log.Printf("Article cache MISS: %v for request: %v", cacheKey, c.Request)

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
		lastlocation = nil
	}

	totaldistance, err := GetTotalDistance()

	if err != nil {
		InternalError(err)
		c.String(500, "Internal Error")
		return
	}

	obj := gin.H{
		"Index":         index,
		"Title":         article.Title,
		"Months":        months,
		"Article":       article,
		"CurrentYear":   time.Now().Year(),
		"Stylesheet":    stylesheetfilename,
		"Javascript":    javascriptfilename,
		"LastLocation":  lastlocation,
		"TotalDistance": totaldistance,
		"Production":    configuration.Production,
		"DisqusUrl":     template.JS(fmt.Sprintf("var disqus_url = 'https://www.growse.com%s';", article.GetAbsoluteUrl()))}

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
	log.Printf("Latest article cache MISS: %v for request: %v", article.getCacheKey(), c.Request)

	lastlocation, err := GetLastLoction()
	if err != nil {
		InternalError(err)
		c.String(500, "Internal Error")
	}

	totaldistance, err := GetTotalDistance()

	if err != nil {
		InternalError(err)
		c.String(500, "Internal Error")
		return
	}

	index, months, err := LoadArticleIndex()
	if err != nil {
		InternalError(err)
		c.String(500, "Internal Error")
		return
	}

	obj := gin.H{
		"Index":         index,
		"Title":         article.Title,
		"Months":        months,
		"Article":       article,
		"Stylesheet":    stylesheetfilename,
		"Javascript":    javascriptfilename,
		"LastLocation":  lastlocation,
		"TotalDistance": totaldistance,
		"Production":    configuration.Production,
		"DisqusUrl":     template.JS(fmt.Sprintf("var disqus_url = 'https://www.growse.com%s';", article.GetAbsoluteUrl()))}

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
	log.Printf("Search cache MISS: %v for request: %v", cacheKey, c.Request)

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

func RSSHandler(c *gin.Context) {
	feed := &feeds.Feed{
		Title:       "growse.com",
		Link:        &feeds.Link{Href: "https://www.growse.com"},
		Description: "ARGLEFARGLE",
		Created:     time.Now(),
	}
	feed.Items = []*feeds.Item{}

	rows, err := db.Query("select title,shorttitle,datestamp from articles where published=true order by datestamp desc limit 20")
	if err != nil {
		InternalError(err)
		c.String(500, "Internal Error")
		return
	}

	for rows.Next() {
		article := Article{}
		rows.Scan(&article.Title, &article.Slug, &article.Timestamp)
		item := feeds.Item{
			Title:   article.Title,
			Link:    &feeds.Link{Href: fmt.Sprintf("https://www.growse.com%s", article.GetAbsoluteUrl())},
			Created: article.Timestamp,
		}
		feed.Items = append(feed.Items, &item)
	}

	rss, err := feed.ToRss()
	if err != nil {
		InternalError(err)
		c.String(500, "Internal Error")
	}
	c.Data(200, "application/xml", []byte(rss))

}

const (
	sitemapHeader = "<?xml version=\"1.0\" encoding=\"UTF-8\"?><urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">"
	sitemapFooter = "</urlset>"
	sitemapUrl    = "<url><loc><![CDATA[https://www.growse.com%s]]></loc><lastmod>%s</lastmod><changefreq>weekly</changefreq><priority>0.5</priority></url>"
)

func UncompressedSiteMapHandler(c *gin.Context) {
	SiteMapHandler(c, false)
}

func CompressedSiteMapHandler(c *gin.Context) {
	SiteMapHandler(c, true)
}

func SiteMapHandler(c *gin.Context, compressed bool) {
	cacheKey := fmt.Sprintf("sitemap-%v", compressed)
	cachedBytes, ok := memoryCache.Get(cacheKey)
	mimeType := "text/xml"
	if compressed {
		mimeType = "application/x-gzip"
	}
	if ok {
		c.Data(200, mimeType, cachedBytes)
		return
	}
	log.Printf("Sitemap handler cache MISS: %v for request: %v", cacheKey, c.Request)
	rows, err := db.Query("Select id, datestamp,shorttitle,title from articles where published=true order by datestamp desc;")
	if err != nil {
		InternalError(err)
		c.String(500, "Internal Error")
		return
	}
	defer rows.Close()

	buf := bufPool.Get()
	defer bufPool.Put(buf)
	buf.WriteString(sitemapHeader)
	for rows.Next() {
		var article Article
		rows.Scan(&article.Id, &article.Timestamp, &article.Slug, &article.Title)
		buf.WriteString(fmt.Sprintf(sitemapUrl, article.GetAbsoluteUrl(), article.Timestamp.Format("2006-01-02T15:04:05Z")))
	}

	buf.WriteString(sitemapFooter)
	var pageBytes []byte
	if compressed {
		compressedBuf := bufPool.Get()
		zip := gzip.NewWriter(compressedBuf)

		_, err = zip.Write(buf.Bytes())
		zip.Close()
		if err != nil {
			InternalError(err)
			c.String(500, "Internal Error")
			return
		}
		pageBytes = compressedBuf.Bytes()
	} else {
		pageBytes = buf.Bytes()
	}
	memoryCache.Set(cacheKey, pageBytes, time.Now().Add(configuration.DefaultCacheExpiry))
	c.Data(200, mimeType, pageBytes)

}
