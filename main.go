package main

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/misterhex/gogogocrawler/crawlers"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"math/rand"
	"time"
)

type Movie struct {
	Id        bson.ObjectId "_id"
	RawSource string
	Source    string
	Category  string
	Episode   string
	Origin    string
	ScrapTime time.Time
}

const MongodbConnString = "mongodb://goblintechie:test1234@ds039850.mongolab.com:39850/goblintechdb"
const Take = 50

func main() {

	go func() {
		var movieResult = crawlers.CrawlMovie(takeMostOutdated(getCategories(), Take))

		for {
			select {
			case movie := <-movieResult:
				saveMovieIfNotExistOrOutdated(movie)

			case <-time.After(time.Minute * 2):
				log.Println("no more movie detected ... try to re-run")
				movieResult = crawlers.CrawlMovie(takeMostOutdated(getCategories(), Take))
			}
		}
	}()

	<-make(chan int)
}

func saveMovieIfNotExistOrOutdated(movie crawlers.Movie) {
	session, err := mgo.Dial(MongodbConnString)
	defer session.Close()
	c := session.DB("goblintechdb").C("movies")

	queriedMovie := Movie{}
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
		if d.Minutes() > 10 {
			changeInfo, err := c.Upsert(bson.M{"_id": queriedMovie.Id}, movie)
			if err != nil {
				log.Println(err)
			} else {
				log.Printf("**** Upserted **** %v %v %v\n\n last updated since %v\n\n", queriedMovie.Id, changeInfo, movie, d)
			}
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

func takeMostOutdated(categories []string, take int) []string {

	session, err := mgo.Dial(MongodbConnString)
	defer session.Close()
	c := session.DB("goblintechdb").C("movies")

	var result []string

	err = c.Find(nil).Sort("scraptime").Distinct("category", &result)

	if err != nil {
		log.Println("unable to take most outdated")
		shuffle(categories)
		return categories[:take]
	} else {

		log.Println(result)
		origin := "http://www.gogoanime.com/category/"
		dbCategories := make([]string, 0)
		for _, shortCat := range result {
			var url = origin + shortCat
			dbCategories = append(dbCategories, url)
		}

		diff := difference(categories, dbCategories)

		log.Println(len(categories))
		log.Println(len(dbCategories))
		log.Println(len(diff))

		r := append(diff, categories...)

		return r[:take]
	}
}

func difference(slice1 []string, slice2 []string) []string {
	var diff []string

	// Loop two times, first to find slice1 strings not in slice2,
	// second loop to find slice2 strings not in slice1
	for i := 0; i < 2; i++ {
		for _, s1 := range slice1 {
			found := false
			for _, s2 := range slice2 {
				if s1 == s2 {
					found = true
					break
				}
			}
			// String not found. We add it to return slice
			if !found {
				diff = append(diff, s1)
			}
		}
		// Swap the slices, only if it was the first loop
		if i == 0 {
			slice1, slice2 = slice2, slice1
		}
	}

	return diff
}

func shuffle(slice []string) {
	rand.Seed(time.Now().UnixNano())
	n := len(slice)
	for i := n - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
}
