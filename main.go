package main

import (
	"./crawlers"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const MongodbConnString = "mongodb://goblintechie:test1234@ds039850.mongolab.com:39850/goblintechdb"

func main() {

	go func() {

		var movieResult = crawlers.CrawlMovie()

		for movie := range movieResult {

			saveMovieIfNotExist(movie)
		}
	}()

	<-make(chan int)
}

func saveMovieIfNotExist(movie crawlers.Movie) {

	session, err := mgo.Dial(MongodbConnString)
	defer session.Close()
	c := session.DB("goblintechdb").C("movies")

	queryMovieFromDb := crawlers.Movie{}
	err = c.Find(bson.M{
		"source": movie.Source,
	}).One(&queryMovieFromDb)

	if err != nil {
		c.Insert(movie)

		fmt.Printf("**** Saved **** %v\n\n", movie)
	}
}
