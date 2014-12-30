package main

import (
	"testing"
	"time"
)

func TestEncodeDecode(t *testing.T) {
	article := Article{1, time.Now(), "slug", "title", "markdown"}
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
		t.Error("Bad article Timestamp: %v", article2.Timestamp)
	}
	if article2.Slug != article.Slug {
		t.Error("Bad article Slug: %v", article2.Slug)
	}
	if article2.Title != article.Title {
		t.Error("Bad article Title: %v", article2.Title)
	}
	if article2.Markdown != article.Markdown {
		t.Error("Bad article Markdown: %v", article2.Markdown)
	}
}
