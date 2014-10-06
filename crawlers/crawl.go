package crawlers

import (
	"errors"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
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

func CrawlMovie(categories []string) chan Movie {

	log.Println("starting to crawl movies ...")

	out := make(chan Movie)

	go func() {

		for _, category := range categories {
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

			if e == nil && strings.TrimSpace(html) == "Next" {
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
			if rawSource, e := getRawSource(src); e == nil {

				if IsVideoContentType(rawSource) {
					movie := Movie{
						Origin:    origin,
						Category:  strings.Replace(category, origin+"/category/", "", -1),
						Episode:   strings.Replace(episode, origin+"/", "", -1),
						Source:    src,
						RawSource: rawSource,
						ScrapTime: time.Now(),
					}
					movies = append(movies, movie)
				}
			}
		}
	})

	return movies
}

func getRawSource(source string) (string, error) {

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

func ParseRawSource(html string) (string, error) {
	re, err := regexp.Compile("url:\\s\\'http\\:.*\\'\\,")
	if err != nil {
		log.Fatal(err)
	}
	slice := re.FindAllString(html, -1)
	if len(slice) == 0 {
		return "", errors.New("unable to parse raw source")
	} else {
		s2 := slice[len(slice)-1]
		r := strings.Replace(s2, "url: '", "", -1)
		r = strings.Replace(r, "',", "", -1)
		r, _ = url.QueryUnescape(r)

		return r, nil
	}
}

func IsVideoContentType(source string) (isVideo bool) {

	defer func() {
		if e := recover(); e != nil {
			log.Println("panic in IsVideoContentType: ", e)
			isVideo = false
		}
	}()

	isVideo = false

	res, err := http.Get(source)
	if err != nil {

		log.Println("error when trying to determine video type")

		return isVideo
	} else {

		var contentType = res.Header.Get("content-type")

		log.Printf("source  %v -> %v\n", source, contentType)

		if contentType == "video/mp4" || contentType == "video/x-flv" {
			isVideo = true
		}

		return isVideo
	}
}
