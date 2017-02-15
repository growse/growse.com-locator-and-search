package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/blevesearch/bleve"
	"github.com/gin-gonic/gin"
	"github.com/mschoch/blackfriday-text"
	"github.com/russross/blackfriday"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	_ "github.com/blevesearch/bleve/search/highlight/highlighters/simple"
)

var bleveIndex bleve.Index

func BleveInit(remoteGit string, repoLocation string) {
	if (remoteGit != "" && repoLocation != "") {
		updateGitRepo(remoteGit, repoLocation, "jekyll")
		openIndex()
		addFilesToIndex(repoLocation + "/_posts", bleveIndex)
	} else {
		log.Print("No SearchIndex parameters supplied, skipping")
	}
}

func openIndex() {
	indexMapping := bleve.NewIndexMapping()
	bleveIndex, _ = bleve.NewMemOnly(indexMapping)
}

func updateGitRepo(remoteLocation string, localLocation string, tag string) error {
	log.Printf("cloning %s to %s", remoteLocation, localLocation)
	if _, err := os.Stat(localLocation); os.IsNotExist(err) {
		cmd := exec.Command("git", "clone", remoteLocation, localLocation)
		err := cmd.Run()
		if err != nil {
			log.Printf("Error cloning repo: %v", err)
			return err
		}
	} else {
		cmd := exec.Command("git", "-C", localLocation, "pull")
		err := cmd.Run()
		if err != nil {
			log.Printf("Error pulling repo: %v", err)
			return err
		}
	}

	cmd := exec.Command("git", "-C", localLocation, "checkout", tag)
	err := cmd.Run()
	if err != nil {
		log.Printf("Error cloning repo: %v", err)
		return err
	}
	return nil
}

func addFilesToIndex(sourceLocation string, index bleve.Index) error {
	fileinfos, err := ioutil.ReadDir(sourceLocation)
	if err != nil {
		return err
	}
	for _, file := range fileinfos {
		if !file.IsDir() {
			fullFileName := path.Join(sourceLocation, file.Name())
			log.Printf("Indexing file %v", fullFileName)
			err := addFileToIndex(fullFileName, index)
			if err != nil {
				log.Printf("Error indexing file: %v", err)
			}
		}
	}
	return nil
}

func BleveIndexDocs(c *gin.Context) {
	c.String(204, "Accepted")
}

var escapeChars = "\\+-=&|><!(){}[]^\"~*?:/ "

func escape(term string) string {
	escapedTerm := term
	for _, char := range escapeChars {
		escapedTerm = strings.Replace(escapedTerm, string(char), `\` + string(char), -1)
	}
	return escapedTerm
}

func BleveSearchQuery(c *gin.Context) {
	if (bleveIndex == nil) {
		c.String(503, "Search not defined")
		return
	}
	var searchForm struct {
		SearchTerm string `form:"a" binding:"required"`
	}
	c.Bind(&searchForm)
	log.Printf("Searching for %v", searchForm.SearchTerm)

	queryString := ""
	for _, term := range strings.Split(searchForm.SearchTerm, " ") {
		escapedTerm := escape(term)
		log.Printf("Escaped search term: %s", escapedTerm)
		queryString += fmt.Sprintf("Body:%s Title:%s^5 ", escapedTerm, escapedTerm)
	}
	log.Print(queryString)

	query := bleve.NewQueryStringQuery(queryString)
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Fields = []string{"Title"}

	searchRequest.Highlight = bleve.NewHighlight()
	searchResults, err := bleveIndex.Search(searchRequest)

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

func addFileToIndex(filePath string, index bleve.Index) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return errors.New("File does not exist")
	}
	contentBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading file: %v", err)
		return err
	}
	log.Printf("%v contains %v bytes", filePath, len(contentBytes))
	numberOfDelimiters := 0
	title := ""
	var body bytes.Buffer
	for _, line := range strings.Split(string(contentBytes), "\n") {
		if line == "---" {
			numberOfDelimiters += 1
		} else {
			if numberOfDelimiters < 2 {
				if strings.HasPrefix(line, "title:") {
					title = strings.Trim(strings.SplitN(line, ":", 2)[1], " \"")
				}
			} else {
				body.WriteString(line)
			}
		}
	}
	renderer := blackfridaytext.TextRenderer()
	textBytes := blackfriday.Markdown(body.Bytes(), renderer, 0)

	data := struct {
		Body  string
		Title string
	}{
		Title: title,
		Body:  string(textBytes),
	}

	// index some data
	_, filename := filepath.Split(filePath)
	err = index.Index(filename, data)
	if err != nil {
		return err
	}
	return nil
}
