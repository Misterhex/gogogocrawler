package main

import (
	"encoding/json"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"runtime/debug"
	"strings"
	"time"
)

const origin = "http://www.gogoanime.com"
const CategoryUrlPrefix = "http://www.gogoanime.com/category/"

type Config struct {
	SyncServerAddr    string
	MongoDbConnString string
}

type Movie struct {
	RawSource string
	Source    string
	Category  string
	Episode   string
	Origin    string
	ScrapTime time.Time
}

type MovieSchema struct {
	Id        bson.ObjectId "_id"
	RawSource string
	Source    string
	Category  string
	Episode   string
	Origin    string
	ScrapTime time.Time
}

var config = Config{}

func main() {

	parseConfig, err := ParseConfig()
	if err != nil {
		log.Fatal(err)
	}

	config = parseConfig
	log.Println(config)

	go func() {
		movieResult := crawlMovie(shuffle(filterBySyncServer(getCategories())))

		for {
			select {
			case movie := <-movieResult:
				saveMovie(movie)

			case <-time.After(time.Second * 45):
				log.Println("no more movie detected ... try to re-run")
				movieResult = crawlMovie(shuffle(filterBySyncServer(getCategories())))
			}
		}
	}()

	<-make(chan int)
}

func ParseConfig() (Config, error) {

	var config = Config{}

	configFile, err := os.Open("config.json")
	if err != nil {
		return config, err
	}

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&config); err != nil {
		return config, err
	}

	return config, nil
}

func filterBySyncServer(categories []string) []string {
	result := make([]string, 0)
	res, err := http.Get(config.SyncServerAddr)
	if err != nil {
		log.Println("error when trying to get alphabets from sync server")
		return result
	}

	defer res.Body.Close()

	bytes, err := ioutil.ReadAll(res.Body)

	if err != nil {
		log.Println("error when trying to get alphabets from sync server")
		return result
	}

	var receivedAlphabet = string(bytes)
	log.Printf("received %v from sync server\n", receivedAlphabet)
	for _, category := range categories {
		categoryTrimmed := strings.Replace(category, CategoryUrlPrefix, "", -1)
		if len(category) > 0 && string(strings.ToUpper(categoryTrimmed)[0]) == strings.ToUpper(receivedAlphabet) {
			result = append(result, category)
		}
	}

	log.Printf("filterBySyncServer returning %v\n", result)

	return result
}

func FilterCategories(categories []string, startWithAlphabet string) []string {

	result := make([]string, 0)

	startWithAlphabet = strings.ToUpper(startWithAlphabet)

	if startWithAlphabet == "#" {
		r, _ := regexp.Compile("[0-9]")

		for _, v := range categories {

			trimmed := strings.Replace(v, CategoryUrlPrefix, "", -1)

			if len(trimmed) > 0 && r.MatchString(string(trimmed[0])) {
				result = append(result, v)
			}
		}

	} else {

		for _, v := range categories {

			trimmed := strings.Replace(v, CategoryUrlPrefix, "", -1)
			trimmed = strings.ToUpper(trimmed)

			if len(trimmed) > 0 && string(trimmed[0]) == startWithAlphabet {
				result = append(result, v)
			}
		}
	}

	return result
}

func saveMovie(movie Movie) {

	session, err := mgo.Dial(config.MongoDbConnString)
	if err != nil {
		log.Println("could not open session to mongodb")
		return
	}

	defer session.Close()
	c := session.DB("goblintechdb").C("movies")

	queriedMovie := MovieSchema{}
	err = c.Find(bson.M{
		"source": movie.Source,
	}).One(&queriedMovie)

	if err != nil {
		err = c.Insert(movie)
		if err != nil {
			log.Println(err)
		} else {
			log.Printf("**** Saved **** %v\n\n", movie)
		}
	} else {
		d := time.Since(queriedMovie.ScrapTime)
		changeInfo, err := c.Upsert(bson.M{"_id": queriedMovie.Id}, movie)
		if err != nil {
			log.Println("error when trying to upsert", err)
		} else {
			log.Printf("**** Upserted **** %v %v %v\n\n last updated since %v\n\n", queriedMovie.Id, changeInfo, movie, d)
		}
	}
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

	return slice
}

func shuffle(slice []string) []string {
	rand.Seed(time.Now().UnixNano())
	n := len(slice)
	for i := n - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
	return slice
}

func crawlMovie(categories []string) chan Movie {

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
		log.Println("getRawSource, error when trying http get", err)
		return "", err
	}

	htmlBytes, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Println("getRawSource, error when trying to read bytes", err)
		return "", nil
	}

	s := string(htmlBytes)

	return ParseRawSource(s)
}

func ParseRawSource(html string) (string, error) {
	re, err := regexp.Compile("url:\\s\\'http\\:.*\\'\\,")
	if err != nil {
		panic("cannot compile regex")
	}

	slice := re.FindAllString(html, -1)
	if len(slice) == 0 {
		return "", errors.New("unable to parse raw source")
	} else {
		s2 := slice[len(slice)-1]
		r := strings.Replace(s2, "url: '", "", -1)
		r = strings.Replace(r, "',", "", -1)
		r, err = url.QueryUnescape(r)

		if err != nil {
			return "", err
		}

		return r, nil
	}
}

func IsVideoContentType(source string) (isVideo bool) {
	defer func() {
		if e := recover(); e != nil {
			log.Println("panic when checking IsVideoContentType: ", e)
			debug.PrintStack()
			isVideo = false
		}
	}()

	isVideo = false

	res, err := http.Get(source)
	if err != nil {
		log.Println("error when trying to determine video type")
		return isVideo
	}

	defer res.Body.Close()

	var contentType = res.Header.Get("content-type")

	log.Printf("source  %v -> %v\n", source, contentType)

	if contentType == "video/mp4" || contentType == "video/x-flv" {
		isVideo = true
	}

	return isVideo
}
