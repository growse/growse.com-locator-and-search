package main

import (
	"github.com/blevesearch/bleve"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestAddingFilesToIndexAddsTheFilesToTheIndex(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "testprefix")
	assert.Nil(t, err)
	_, err = ioutil.TempFile(tempdir, "testprefix1")
	assert.Nil(t, err)
	_, err = ioutil.TempFile(tempdir, "testprefix2")
	assert.Nil(t, err)

	mapping := bleve.NewIndexMapping()
	index, err := bleve.New("testIndex", mapping)
	assert.Nil(t, err)

	err = addFilesToIndex(tempdir, index)
	assert.Nil(t, err)

	count, err := index.DocCount()
	assert.Nil(t, err)
	assert.Equal(t, uint64(2), count)

	os.RemoveAll("testIndex")
}

func TestAddSingleFileToIndexAddsTheFileToTheIndex(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "testprefix")
	assert.Nil(t, err)

	tempFile, err := ioutil.TempFile(tempDir, "testprefix1")
	assert.Nil(t, err)

	mapping := bleve.NewIndexMapping()
	index, err := bleve.New("testIndex", mapping)
	assert.Nil(t, err)

	err = addFileToIndex(tempFile.Name(), index)
	assert.Nil(t, err)
	count, err := index.DocCount()
	assert.Nil(t, err)
	assert.Equal(t, uint64(1), count)

	os.RemoveAll("testIndex")
}

func TestSearchingIndexForFileContentReturnsResult(t *testing.T) {
	content := "title:content\n---\ncontent"
	tempDir, err := ioutil.TempDir("", "testprefix")
	assert.Nil(t, err)

	tempFile, err := ioutil.TempFile(tempDir, "testprefix1")
	assert.Nil(t, err)

	ioutil.WriteFile(tempFile.Name(), []byte(content), os.ModePerm)

	mapping := bleve.NewIndexMapping()
	index, err := bleve.New("testIndex", mapping)
	assert.Nil(t, err)

	err = addFileToIndex(tempFile.Name(), index)
	assert.Nil(t, err)
	count, err := index.DocCount()
	assert.Nil(t, err)
	assert.Equal(t, uint64(1), count)

	searchForm := SearchForm{SearchTerm: "content"}
	searchResult, err := searchIndexForThings(index, searchForm)

	assert.Nil(t, err)
	assert.Equal(t, 0, searchResult.Status.Failed)
	assert.Equal(t, 1, searchResult.Status.Successful)
	assert.Equal(t, 0, len(searchResult.Status.Errors))
	assert.Equal(t, 1, len(searchResult.Hits))
	os.RemoveAll("testIndex")
}

func TestSearchingIndexForNonFileContentReturnsResult(t *testing.T) {
	content := "content"
	tempDir, err := ioutil.TempDir("", "testprefix")
	assert.Nil(t, err)

	tempFile, err := ioutil.TempFile(tempDir, "testprefix1")
	assert.Nil(t, err)

	ioutil.WriteFile(tempFile.Name(), []byte(content), os.ModePerm)

	mapping := bleve.NewIndexMapping()
	index, err := bleve.New("testIndex", mapping)
	assert.Nil(t, err)

	err = addFileToIndex(tempFile.Name(), index)
	assert.Nil(t, err)
	count, err := index.DocCount()
	assert.Nil(t, err)
	assert.Equal(t, uint64(1), count)

	query := bleve.NewMatchQuery("not")
	searchRequest := bleve.NewSearchRequest(query)
	searchResult, err := index.Search(searchRequest)
	assert.Nil(t, err)
	assert.Equal(t, uint64(0), searchResult.Total)
	os.RemoveAll("testIndex")
}
