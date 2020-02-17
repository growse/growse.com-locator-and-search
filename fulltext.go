package main

import (
	"bytes"
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/lang/en"
	"github.com/blevesearch/bleve/mapping"
	htmlHighlight "github.com/blevesearch/bleve/search/highlight/format/html"
	"github.com/gin-gonic/gin"
	"github.com/grokify/html-strip-tags-go"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

type BlogPost struct {
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Published time.Time `json:"published"`
}

func (*BlogPost) Type() string {
	return "blogpost"
}

var bleveIndex bleve.Index

func BleveInit(webroot string, pathPattern *regexp.Regexp) {
	if _, err := os.Stat(webroot); err == nil {
		index, err := createInMemoryHTMLBlogPostBleveIndex()
		if err != nil {
			log.Fatalf("Error opening index: %v", err)
		}
		addHtmlFilesToIndex(webroot, pathPattern, index)
		bleveIndex = index
	} else {
		log.Print("No webroot provided for search")
	}
}

func buildIndexMapping() (*mapping.IndexMappingImpl, error) {

	indexMapping := bleve.NewIndexMapping()
	indexMapping.DefaultAnalyzer = en.AnalyzerName
	indexMapping.DefaultType = "blogPost"

	englishTextFieldMapping := bleve.NewTextFieldMapping()
	englishTextFieldMapping.Analyzer = en.AnalyzerName

	blogPostMapping := bleve.NewDocumentMapping()
	blogPostMapping.AddFieldMappingsAt("title", englishTextFieldMapping)
	blogPostMapping.AddFieldMappingsAt("content", englishTextFieldMapping)

	indexMapping.DefaultMapping = blogPostMapping

	return indexMapping, nil
}

func createInMemoryHTMLBlogPostBleveIndex() (bleve.Index, error) {
	indexMapping, err := buildIndexMapping()
	if err != nil {
		return nil, err
	}
	bleveIndex, err := bleve.NewMemOnly(indexMapping)
	return bleveIndex, err
}

func addHtmlFilesToIndex(sourceLocation string, regexPattern *regexp.Regexp, index bleve.Index) {
	err := filepath.Walk(sourceLocation, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error: %v", err)
		}
		if !info.IsDir() && filepath.Ext(info.Name()) == ".html" && regexPattern.MatchString(path) {
			err = addFileToIndex(sourceLocation, path, index)
			if err != nil {
				log.Printf("Error adding %v to index. Skipping: %v", path, err)
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("Error walking the webroot: %v", err)
	} else {
		count, _ := index.DocCount()
		log.Printf("Indexing complete. %v items added", count)
	}
}

func addFileToIndex(webroot string, filePath string, index bleve.Index) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return errors.New("file does not exist")
	}
	contentBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading file: %v", err)
		return err
	}

	data, err := extractBlogPostFromHTML(contentBytes)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unable to create indexable content from %v", filePath))
	}
	// index some data
	err = index.Index(getUrlFromFilePath(webroot, filePath), data)
	if err != nil {
		return err
	}
	return nil
}

func getUrlFromFilePath(webroot string, filePath string) string {
	return filePath[len(webroot):]
}

func extractBlogPostFromHTML(content []byte) (*BlogPost, error) {
	blogPost := BlogPost{}
	document, err := htmlquery.Parse(bytes.NewReader(content))
	if err != nil {
		log.Fatalf("Error parsing HTML: %v", err)
	}
	article := htmlquery.FindOne(document, "//article[@itemprop='blogPost']")
	if article == nil {
		return nil, errors.New("Can't find an article in that file")
	}
	title := htmlquery.FindOne(article, "/header/h1")
	blogPost.Title = htmlquery.InnerText(title)
	published := htmlquery.FindOne(article, "/header/time[@class='plain']")
	parsedDate, err := time.Parse("2006-01-02", htmlquery.SelectAttr(published, "datetime"))

	if err != nil {
		log.Printf("Error parsing date %v", err)
	} else {
		blogPost.Published = parsedDate
	}

	blogContent := htmlquery.FindOne(article, "/section[@itemprop='articleBody']")
	if blogContent == nil {
		return nil, errors.New("Could not extract content from HTML")
	}
	blogPost.Content = strip.StripTags(htmlquery.InnerText(blogContent))
	return &blogPost, nil
}

type SearchForm struct {
	SearchTerm string `form:"term" binding:"required"`
	Page       int    `form:"page" binding:"-"`
}

func searchIndexForThings(index bleve.Index, searchForm SearchForm) (*bleve.SearchResult, error) {
	query := bleve.NewMatchQuery(searchForm.SearchTerm)

	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Highlight = bleve.NewHighlightWithStyle(htmlHighlight.Name)
	searchRequest.Size = 10
	page := searchForm.Page

	if page < 1 {
		page = 1
	}
	searchRequest.From = (page - 1) * searchRequest.Size
	searchRequest.Fields = []string{"title", "published"}
	return index.Search(searchRequest)
}

func BleveSearchQuery(c *gin.Context) {

	if bleveIndex == nil {
		c.String(503, "Search index not defined")
		return
	}
	searchForm := SearchForm{}
	err := c.Bind(&searchForm)
	if err != nil {
		c.String(400, fmt.Sprintf("Unable to bind search params: %v", err))
	}
	log.Printf("Searching for %v", searchForm)

	searchResults, err := searchIndexForThings(bleveIndex, searchForm)

	if err != nil {
		log.Printf("Error doing search: %v", err)
		c.String(500, "ERROR")
	} else {
		c.JSON(200, gin.H{
			"timeTaken": searchResults.Took,
			"totalHits": searchResults.Total,
			"hits":      searchResults.Hits,
		})
	}
}
