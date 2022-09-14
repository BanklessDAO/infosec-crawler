package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gocolly/colly"
)

type Article struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Link    string `json:"link"`
}

func main() {
	// Instantiate default collector
	c := colly.NewCollector(
		// Visit only domains: rekt.news, hackerspaces.org
		colly.AllowedDomains("rekt.news", "hackerspaces.org"),
	)

	articles := make([]Article, 0)

	// On every a element which has href attribute call callback
	c.OnHTML("article", func(e *colly.HTMLElement) {
		// create an article struct and fill it with the data we need
		post := Article{
			Title:   e.ChildText(".post-title"),
			Content: strings.ReplaceAll(e.ChildText("section > p:nth-child(1)"), "’", "'"),
			Link:    e.Request.AbsoluteURL(e.ChildAttr("article a[href]", "href")),
		}

		// append the article to the list of articles and filter out tagged articles
		if !strings.Contains(post.Link, "tag=") {
			articles = append(articles, post)
		}

		fmt.Printf("Article found: %q\n", post)
		// Visit link found on page
		// Only those links are visited which are in AllowedDomains
		c.Visit(post.Link)
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		// Abort if the URL is a tag, we only care about articles
		if r.URL.RawQuery != "" && strings.Contains(string(r.URL.RawQuery), "tag") {
			r.Abort()
		} else {
			fmt.Println("Visiting", r.URL.String())
		}
	})

	c.OnResponse(func(r *colly.Response) {
		fmt.Println("Got a response from", r.Request.URL)
	})

	// Set error handler
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Got this error:", err)
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished", r.Request.URL)
		js, err := json.MarshalIndent(articles, "", "    ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Writing data to file")
		if err := os.WriteFile("../articles.json", js, 0664); err == nil {
			fmt.Println("Data written to file successfully")
		}

	})

	// Start scraping on https://rekt.news
	c.Visit("https://rekt.news/")
}
