package main

import (
	"log"

	"github.com/blevesearch/bleve"
)

type Person struct {
	Name string
}

func main() {
	mapping := bleve.NewIndexMapping()
	index, err := bleve.New("people.bleve", mapping)
	if err != nil {
		log.Fatal(err)
	}

	person := Person{"Marty Schoch"}
	err = index.Index("m1", person)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Indexed Document")
}
