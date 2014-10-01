package main

import (
	"./crawlers"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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

		var movieResult = crawlers.CrawlMovie()

		for movie := range movieResult {

			saveMovieIfNotExistOrOutdated(movie)
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
			fmt.Println(err)
		} else {
			fmt.Printf("**** Saved **** %v\n\n", movie)
		}
	} else {
		d := time.Since(queriedMovie.ScrapTime)
		if d.Minutes() > 10 {
			changeInfo, err := c.Upsert(bson.M{"_id": queriedMovie.Id}, movie)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Printf("**** Upserted **** %v %v %v\n\n", queriedMovie.Id, changeInfo, movie)
			}
		}
	}
}
