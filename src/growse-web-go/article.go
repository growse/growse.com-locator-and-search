package main

import (
	"encoding/json"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/russross/blackfriday"
	"html/template"
	"time"
)

type Article struct {
	Id        int
	Timestamp time.Time
	Slug      string
	Title     string
	Markdown  string
}

func (article *Article) getCacheKey() string {
	return Truncate(fmt.Sprintf("growse.com-article-%d-%02d-%02d-%s", article.Timestamp.Year(), article.Timestamp.Month(), article.Timestamp.Day(), article.Slug), 250)
}

func getCacheKey(year int, month int, day int, slug string) string {
	return Truncate(fmt.Sprintf("growse.com-article-%d-%02d-%02d-%s", year, month, day, slug), 250)
}

func (article *Article) cacheArticle() error {
	articleAsJson, err := json.Marshal(article)
	if err != nil {
		return err
	}
	memcacheItem := memcache.Item{Key: article.getCacheKey(), Value: articleAsJson}
	err = memcacheClient.Set(&memcacheItem)
	return err
}

func getArticleFromCache(cacheKey string) (*Article, error) {
	var article Article
	articleFromCache, err := memcacheClient.Get(cacheKey)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(articleFromCache.Value, &article)
	if err != nil {
		return nil, err
	}
	return &article, nil
}

func GetArticle(year int, month int, day int, slug string) (*Article, error) {
	cacheKey := getCacheKey(year, month, day, slug)
	articleFromCache, err := getArticleFromCache(cacheKey)
	if err == nil {
		return articleFromCache, nil
	}
	var article Article
	err = db.QueryRow(`Select id,datestamp,shorttitle,title,markdown from articles where published=true and shorttitle=$1 limit 1`, slug).Scan(&article.Id, &article.Timestamp, &article.Slug, &article.Title, &article.Markdown)
	if err == nil {
		article.cacheArticle()
		return &article, nil
	}
	return nil, err
}

func (article *Article) GetAbsoluteUrl() string {
	return fmt.Sprintf("/%d/%02d/%02d/%s", article.Timestamp.Year(), article.Timestamp.Month(), article.Timestamp.Day(), article.Slug)
}

func (article *Article) Rendered() template.HTML {
	return template.HTML(blackfriday.MarkdownCommon(([]byte)(article.Markdown)))
}

func Truncate(input string, length int) string {
	if len(input) < length {
		return input
	} else {
		return input[:length]
	}
}
