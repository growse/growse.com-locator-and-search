package main

import (
	"testing"
	"time"
)

func TestEncodeDecode(t *testing.T) {
	now := time.Now()
	article := Article{1, now, "slug", "title", "markdown"}
	bytes, err := article.ToBytes()
	if err != nil {
		t.Errorf("Failed to encode: %v", err)
	}
	article2, err := FromBytes(bytes)
	if err != nil {
		t.Errorf("Failed to decode: %v", err)
	}
	if article2.Id != article.Id {
		t.Errorf("Bad article Id: %v", article2.Id)
	}
	if article2.Timestamp != article.Timestamp {
		t.Errorf("Bad article Timestamp: %v", article2.Timestamp)
	}
	if article2.Slug != article.Slug {
		t.Errorf("Bad article Slug: %v", article2.Slug)
	}
	if article2.Title != article.Title {
		t.Errorf("Bad article Title: %v", article2.Title)
	}
	if article2.Markdown != article.Markdown {
		t.Errorf("Bad article Markdown: %v", article2.Markdown)
	}
}

func TestIndexOfContains(t *testing.T) {
	thingToTest := []string{"hello", "there", "awesomesauce", "wibble", "ø∂˚¨˙ƒ´˚¨˙"}
	actual := indexOfCaseInsensitive(thingToTest, "wibble")
	if actual != 3 {
		t.Errorf("Failed to get index of item. Expected: 3 Actual: %v", actual)
	}
}

func TestIndexOfCaseInsensitivity(t *testing.T) {

	thingToTest := []string{"hello", "there", "awesomesauce", "wibble", "ø∂˚¨˙ƒ´˚¨˙"}
	actual := indexOfCaseInsensitive(thingToTest, "WIBBLE")
	if actual != 3 {
		t.Errorf("Failed to get index of item. Expected: 3 Actual: %v", actual)
	}
}
func TestIndexOfUnicode(t *testing.T) {
	thingToTest := []string{"hello", "there", "awesomesauce", "wibble", "ø∂˚¨˙ƒ´˚¨˙"}
	actual := indexOfCaseInsensitive(thingToTest, "ø∂˚¨˙ƒ´˚¨˙")
	if actual != 4 {
		t.Errorf("Failed to get index of unicode item. Expected: 3 Actual: %v", actual)
	}
}

func TestIndexOfNotFound(t *testing.T) {
	thingToTest := []string{"hello", "there", "awesomesauce", "wibble", "ø∂˚¨˙ƒ´˚¨˙"}
	actual := indexOfCaseInsensitive(thingToTest, "NotFound")
	if actual != -1 {
		t.Errorf("Failed to detet item not found. Expected: -1 Actual: %v", actual)
	}
}

func TestSmartTruncate(t *testing.T) {
	inputString := "Lorem ipsum dolor Sit amet, consectetur adipiscing elit. In in erat pretium nisi ornare tempor. Phasellus molestie lectus tellus, a facilisis enim commodo at. Ut vel dui eu libero lacinia congue pretium et ex. Etiam commodo accumsan scelerisque. Suspendisse augue lorem, sodales id ex vel, scelerisque porta neque. Sed posuere sed ligula a accumsan. Nam pellentesque sodales nisl eu placerat. Quisque at odio nunc."
	actual := smartTruncate(inputString, "sit", 2, "...")
	expected := "...ipsum dolor Sit amet, consectetur..."
	if actual != expected {
		t.Errorf("Incorrect truncating. Expected \"%s\". Actual: \"%s\"", expected, actual)
	}
}

func TestSmartTruncateNotFound(t *testing.T) {
	inputString := "Lorem ipsum dolor Sit amet, consectetur adipiscing elit. In in erat pretium nisi ornare tempor. Phasellus molestie lectus tellus, a facilisis enim commodo at. Ut vel dui eu libero lacinia congue pretium et ex. Etiam commodo accumsan scelerisque. Suspendisse augue lorem, sodales id ex vel, scelerisque porta neque. Sed posuere sed ligula a accumsan. Nam pellentesque sodales nisl eu placerat. Quisque at odio nunc."
	expected := "Lorem ipsum dolor Sit amet, consectetur adipiscing elit. In in..."
	actual := smartTruncate(inputString, "wibble", 5, "...")
	if actual != expected {
		t.Errorf("Incorrect truncating. Expected \"%s\". Actual: \"%s\"", expected, actual)
	}
}
