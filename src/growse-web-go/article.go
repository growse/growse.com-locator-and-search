package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/russross/blackfriday"
	"html/template"
	"regexp"
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
	pattern := regexp.MustCompile("([^\\s\\w]|_)+")
	return pattern.ReplaceAllString(text, "")
}

func smartTruncate(input string, searchterm string, surroundingWords int, suffix string) string {
	words := strings.Split(input, " ")
	searchterm = removePunctuation(searchterm)
	var trimmedWords []string
	/*for _, word := range words {
		trimmedWords = append(trimmedWords, removePunctuation(word))
	}*/
	trimmedWords = words
	index := indexOfCaseInsensitive(trimmedWords, searchterm)
	result := ""
	if index >= 0 {
		startIndex := index - surroundingWords
		endIndex := index + surroundingWords
		if startIndex < 0 {
			startIndex = 0
			endIndex = 2 * surroundingWords
		}
		if endIndex > len(trimmedWords) {
			endIndex = len(trimmedWords) - 1
			startIndex = endIndex - (2 * surroundingWords)
			if startIndex < 0 {
				startIndex = 0
			}
		}
		result = strings.Join(trimmedWords[startIndex:endIndex+1], " ")
		if startIndex > 0 {
			result = suffix + result
		}
		if endIndex < len(trimmedWords)-1 {
			result += suffix
		}
	} else {
		if 2*surroundingWords > len(trimmedWords)-1 {
			result = strings.Join(trimmedWords, " ")
		} else {
			result = strings.Join(trimmedWords[0:2*surroundingWords], " ") + suffix
		}
	}
	return result
}

func indexOfCaseInsensitive(slice []string, thing string) int {
	for index, item := range slice {
		if strings.ToLower(item) == strings.ToLower(thing) {
			return index
		}
	}
	return -1
}

func (article *Article) RenderAsSearchResult(searchterm string) template.HTML {
	return template.HTML(smartTruncate(article.Markdown, searchterm, 15, "..."))
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
