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

func main() {

	go func() {
		movieResult := crawlers.CrawlMovie(shuffle(getCategories()))

		for {
			select {
			case movie := <-movieResult:
				defer func() {
					if e := recover(); e != nil {
						log.Println("panic when saving movie: ", e)
					}
				}()
				saveMovie(movie)

			case <-time.After(time.Minute * 6):
				log.Println("no more movie detected ... try to re-run")
				movieResult = crawlers.CrawlMovie(shuffle(getCategories()))
			}
		}
	}()

	<-make(chan int)
}

func saveMovie(movie crawlers.Movie) {
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

func shuffle(slice []string) []string {
	rand.Seed(time.Now().UnixNano())
	n := len(slice)
	for i := n - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
	return slice
}
