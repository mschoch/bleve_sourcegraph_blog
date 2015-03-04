package main

import (
	"fmt"
	"log"

	"github.com/blevesearch/bleve"
)

func main() {

	index, err := bleve.Open("gopherconin.bleve")
	if err != nil {
		log.Fatal(err)
	}

	PhraseSearch(index)
}

func PhraseSearch(index bleve.Index) {
	phrase := []string{"quality", "search", "results"}
	q := bleve.NewPhraseQuery(phrase, "description")
	req := bleve.NewSearchRequest(q)
	req.Highlight = bleve.NewHighlightWithStyle("ansi")
	req.Fields = []string{"summary", "speaker"}
	res, err := index.Search(req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res)
}
