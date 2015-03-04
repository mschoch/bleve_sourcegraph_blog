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

	qString := `+description:text summary:"text indexing" summary:believe~2 -description:lucene duration:<30`
	q := bleve.NewQueryStringQuery(qString)
	req := bleve.NewSearchRequest(q)
	req.Highlight = bleve.NewHighlightWithStyle("ansi")
	req.Fields = []string{"summary", "speaker", "description", "duration"}
	res, err := index.Search(req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res)
}
