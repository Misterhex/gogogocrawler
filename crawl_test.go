package main

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestShouldBeAbleToExtractPlaylist(t *testing.T) {
	file, err := os.Open("crawl_test_videofun.html")
	if err != nil {
		log.Fatal(err)
	}

	buf, err := ioutil.ReadAll(file)

	s := string(buf)
	r, e := ParseRawSource(s)

	if r == "" || e != nil {
		t.Error("")
	}
}

func TestShouldReturnHackLegendWhenFiltered(t *testing.T) {

	categories := make([]string, 0)
	categories = append(categories, "http://www.gogoanime.com/category/a-channel")

	result := FilterCategories(categories, "a")

	if len(result) == 0 {
		t.Error("should not be empty result")
	}

	if len(result) > 0 && result[0] != "http://www.gogoanime.com/category/a-channel" {
		t.Error("should be return a channel category")
	}
}

func TestShouldReturnCategoriesThatStartWithNumberOnlyWhenPassedHex(t *testing.T) {

	categories := make([]string, 0)
	categories = append(categories, "http://www.gogoanime.com/category/a-channel")
	categories = append(categories, "http://www.gogoanime.com/category/07-ghost")
	result := FilterCategories(categories, "#")

	if len(result) == 0 {
		t.Error("should not be empty result")
	}

	if len(result) > 0 && result[0] != "http://www.gogoanime.com/category/07-ghost" {
		t.Error("should be return 07 ghost category")
	}
}
