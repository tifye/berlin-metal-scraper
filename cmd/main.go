package main

import (
	"fmt"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/charmbracelet/log"
	"github.com/gocolly/colly"
)

type Event struct {
	Title string
	At    string
	Date  time.Time
	Links []EventLink
}

type EventLink struct {
	Title string
	Url   string
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

	events := parseRawData(eventsData)
	for _, event := range events {
		fmt.Println(event)
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
