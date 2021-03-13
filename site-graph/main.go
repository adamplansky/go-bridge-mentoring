package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func ExampleScrape() {
	// Request the HTML page.
	res, err := http.Get("https://github.com/adamplansky/go-bridge-mentoring")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Find the review items
	doc.Find("h1").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		as := s.Find("a")
		for _, a := range as.Nodes {
			fmt.Printf("<a>%s</a>\n", a.FirstChild.Data)
		}

		//title := s.Find("i").Text()
		//fmt.Printf("Review %d: %s - %s\n", i, band, title)
		//fmt.Printf("%#v\n", s)
	})
}

func main() {
	ExampleScrape()
}
