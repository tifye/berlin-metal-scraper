package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/charmbracelet/log"
	"github.com/gocolly/colly"
)

type Event struct {
	Title  string      `json:"title"`
	At     string      `json:"at,omitempty"`
	Date   time.Time   `json:"date"`
	Genres []string    `json:"genres"`
	Links  []EventLink `json:"links,omitempty"`
}

type EventLink struct {
	Title string `json:"title"`
	Url   string `json:"url"`
}

type rawEventDataLink struct {
	text string
	url  string
}

type rawEventData struct {
	eventString string
	genre       string
	links       []rawEventDataLink
}

func main() {
	col := colly.NewCollector()

	var eventsData []rawEventData
	col.OnHTML("div#scrollbar", func(el *colly.HTMLElement) {
		eventsData = scrapeConcerts(el)
	})
	err := col.Visit("https://berlinmetal.eu/")
	if err != nil {
		log.Fatal(err)
	}

	outputJson := flag.Bool("json", false, "output JSON")
	flag.Parse()

	events := parseRawData(eventsData)
	if *outputJson {
		output, err := json.Marshal(events)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(output))
	} else {
		prettyPrintEvents(events)
	}
}

func prettyPrintEvents(events []Event) {
	for _, event := range events {
		fmt.Printf("\n%s\n%s\n%s\n", event.Date.Format(time.DateOnly), event.Title, strings.Join(event.Genres, ", "))
		if event.At != "" {
			fmt.Printf("%s\n", event.At)
		}
		for _, link := range event.Links {
			fmt.Printf("%s: %s\n", link.Title, link.Url)
		}
	}
}

func scrapeConcerts(el *colly.HTMLElement) []rawEventData {
	eventsData := make([]rawEventData, 0)
	elements := el.DOM.Find("p.konzerte")
	elements.Each(func(_ int, sel *goquery.Selection) {
		if len(sel.Nodes) < 1 {
			fmt.Println("found p.konzerte element with no child nodes")
			return
		}

		eventString := sel.Nodes[0].FirstChild.Data

		genre := ""
		sibling := sel.Next()
		if len(sibling.Nodes) > 0 &&
			len(sibling.Nodes[0].Attr) > 0 &&
			sibling.Nodes[0].Attr[0].Val == "genre" {
			genre = sibling.Nodes[0].FirstChild.Data
		}

		linkElements := sel.Find("a.konzertliste[href]")
		linksData := make([]rawEventDataLink, 0, len(linkElements.Nodes))
		for _, linkNode := range linkElements.Nodes {
			linkData := rawEventDataLink{
				text: linkNode.FirstChild.Data,
				url:  linkNode.Attr[0].Val,
			}
			linksData = append(linksData, linkData)
		}

		eventsData = append(eventsData, rawEventData{
			eventString: eventString,
			genre:       genre,
			links:       linksData,
		})
	})

	return eventsData
}
