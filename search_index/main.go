package main

import (
	"log"

	"github.com/blevesearch/bleve"
)

type Person struct {
	Name string
}

func main() {
	index, err := bleve.Open("people.bleve")
	if err != nil {
		log.Fatal(err)
	}

	query := bleve.NewTermQuery("marty")
	request := bleve.NewSearchRequest(query)
	result, err := index.Search(request)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(result)
}
