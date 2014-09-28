package crawlers

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"math/rand"
	"strings"
	"time"
)

type Movie struct {
	Source   string
	Category string
	Episode  string
	Origin   string
}

const origin = "http://www.gogoanime.com"

func CrawlMovie() chan Movie {

	fmt.Println("starting to crawl movies ...")

	out := make(chan Movie)

	go func() {

		for _, category := range getCategories()[:1] {
			fmt.Printf("crawling %v\n\n", category)

			var episodes = getMovieEpisode(category, make([]string, 0))

			fmt.Printf("category %v -> episodes %v\n\n", category, episodes)

			for _, episode := range episodes {
				for _, movie := range getMovies(category, episode) {

					fmt.Printf("episode %v -> source %v\n\n", category, movie.Source)

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
				Category: category,
				Episode:  episode,
				Source:   src,
				Origin:   origin,
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

	// shuffle(slice)

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
