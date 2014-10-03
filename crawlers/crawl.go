package crawlers

import (
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type Movie struct {
	RawSource string
	Source    string
	Category  string
	Episode   string
	Origin    string
	ScrapTime time.Time
}

const origin = "http://www.gogoanime.com"

func CrawlMovie() chan Movie {

	log.Println("starting to crawl movies ...")

	out := make(chan Movie)

	go func() {

		for _, category := range getCategories() {
			log.Printf("crawling %v\n\n", category)

			var episodes = getMovieEpisode(category, make([]string, 0))

			log.Printf("category %v -> episodes %v\n\n", category, episodes)

			for _, episode := range episodes {
				for _, movie := range getMovies(category, episode) {

					log.Printf("episode %v -> source %v\n\n", category, movie.Source)

					out <- movie
				}
			}
		}
	}()

	return out
}

func getMovieEpisode(category string, episodes []string) []string {

	var doc *goquery.Document
	var e error

	if doc, e = goquery.NewDocument(category); e != nil {
		return episodes
	}

	doc.Find("div.postlist table > tbody > tr > td > a").Each(func(i int, s *goquery.Selection) {

		if href, success := s.Attr("href"); success {
			episodes = append(episodes, href)
		}
	})

	if doc.Find(".wp-pagenavi").Length() != 0 {
		var isLast = true
		doc.Find(".wp-pagenavi").Children().Each(func(i int, s *goquery.Selection) {
			html, e := s.Html()
			if e != nil && html == "Next" {
				isLast = false
			}
		})

		if !isLast {

			var current = doc.Find(".wp-pagenavi span.current")

			var closestNode = current.Next()

			if nextPage, success := closestNode.Attr("href"); success {
				return getMovieEpisode(nextPage, episodes)
			}
		}
	}

	return episodes
}

func getMovies(category string, episode string) []Movie {
	var doc *goquery.Document
	var e error

	var movies = make([]Movie, 0)

	if doc, e = goquery.NewDocument(episode); e != nil {
		return movies
	}

	doc.Find("iframe").Each(func(i int, s *goquery.Selection) {

		if src, success := s.Attr("src"); success && (strings.HasSuffix(src, "mp4") || strings.HasSuffix(src, "flv")) {
			movie := Movie{
				Category:  category,
				Episode:   episode,
				Source:    src,
				Origin:    origin,
				ScrapTime: time.Now(),
				RawSource: getRawSource(src),
			}
			movies = append(movies, movie)
		}
	})

	return movies
}

func getCategories() []string {

	var doc *goquery.Document
	var e error
	var slice = make([]string, 0)

	if doc, e = goquery.NewDocument("http://www.gogoanime.com/watch-anime-list"); e != nil {
		return slice
	}

	doc.Find(".cat-item a").Each(func(i int, s *goquery.Selection) {

		if href, success := s.Attr("href"); success {
			slice = append(slice, href)
		}
	})
	slice = append(slice, "http://www.gogoanime.com/category/miscellaneous")

	shuffle(slice)

	return slice
}

func shuffle(slice []string) {
	rand.Seed(time.Now().UnixNano())
	n := len(slice)
	for i := n - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
}

func getRawSource(source string) string {

	res, err := http.Get(source)
	if err != nil {
		log.Fatal(err)
	}

	htmlBytes, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	s := string(htmlBytes)

	return ParseRawSource(s)
}

func ParseRawSource(html string) string {
	re, err := regexp.Compile("url:\\s\\'http\\:.*\\'\\,")
	if err != nil {
		log.Fatal(err)
	}
	slice := re.FindAllString(html, -1)
	s2 := slice[len(slice)-1]
	r := strings.Replace(s2, "url: '", "", -1)
	r = strings.Replace(r, "',", "", -1)
	r, _ = url.QueryUnescape(r)

	return r
}
