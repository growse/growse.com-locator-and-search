package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
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

func GetArticle(year int, month int, day int, slug string) (*Article, error) {
	var article Article
	err := db.QueryRow(`Select id,datestamp,shorttitle,title,markdown from articles where published=true and shorttitle=$1 and date_part('year',datestamp at time zone 'UTC')=$2 and date_part('month',datestamp at time zone 'UTC')=$3 and date_part('day',datestamp at time zone 'UTC')=$4 limit 1`, slug, year, month, day).Scan(&article.Id, &article.Timestamp, &article.Slug, &article.Title, &article.Markdown)
	if err == nil {
		return &article, nil
	}
	return nil, err
}

func (article *Article) GetAbsoluteUrl() string {
	return fmt.Sprintf("/%d/%02d/%02d/%s/", article.Timestamp.UTC().Year(), article.Timestamp.UTC().Month(), article.Timestamp.UTC().Day(), article.Slug)
}

func (article *Article) Rendered() template.HTML {
	return template.HTML(blackfriday.MarkdownCommon(([]byte)(article.Markdown)))
}

func Truncate(input string, length int) string {
	if len(input) < length {
		return input
	} else {
		return string(input[:length])
	}
}
