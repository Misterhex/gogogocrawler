package crawlers

import (
	"fmt"
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
	r := GetRawSource(s)
	fmt.Println(r)

}
