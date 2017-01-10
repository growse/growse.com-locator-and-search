package main

import (
	"github.com/blevesearch/bleve"
	"gopkg.in/gin-gonic/gin.v1"
	"os/exec"
	"log"
	"os"
	"io/ioutil"
	"errors"
	"path"
)

var bleveIndex bleve.Index
var repoLocation = "/var/tmp/growse.com-jekyll-git"
var remoteGit = "https://github.com/growse/www.growse.com.git"
var blevePath = "growse.com.bleve"

func BleveInit() {
	updateGitRepo(remoteGit, repoLocation, "jekyll")
	openIndex(blevePath)
	addFilesToIndex(repoLocation + "/_posts", bleveIndex)
}

func openIndex(path string) {

	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Printf("%v doesn't exist, creating new index", path)
		mapping := bleve.NewIndexMapping()
		bleveIndex, _ = bleve.New(path, mapping)
	} else {
		log.Printf("%v already exists. opening", path)
		bleveIndex, _ = bleve.Open(path)
	}
}

func updateGitRepo(remoteLocation string, localLocation string, tag string) error {
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
	for _, file := range (fileinfos) {
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

func BleveSearchQuery(c *gin.Context) {

	var searchForm struct {
		SearchTerm string `form:"a" binding:"required"`
	}
	c.Bind(&searchForm)
	log.Printf("Searching for %v", searchForm.SearchTerm)
	docCount, _ := bleveIndex.DocCount()
	log.Printf("Number of docs: %v", docCount)
	query := bleve.NewMatchQuery(searchForm.SearchTerm)
	searchRequest := bleve.NewSearchRequest(query)
	searchResults, err := bleveIndex.Search(searchRequest)

	if err != nil {
		log.Printf("Error doing search: %v", err)
		c.String(500, "ERROR")
	} else {

		log.Println(searchResults)

		log.Printf("Search results: %v", searchResults.MaxScore)
		c.JSON(200, gin.H{
			"timeTaken":searchResults.Took,
			"totalHits":searchResults.Total,
			"sdf":searchResults.Hits,
		})
	}
}

func addFileToIndex(filename string, index bleve.Index) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return errors.New("File does not exist")
	}
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("Error reading file: %v", err)
		return err
	}
	log.Printf("%v contains %v bytes", filename, len(bytes))

	//err = index.Index(filename, string(bytes))
	data := struct {
		Name string
	}{
		Name: string(bytes),
	}

	// index some data
	index.Index(filename, data)
	if err != nil {
		return err
	}
	return nil
}
