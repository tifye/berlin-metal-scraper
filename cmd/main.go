package main

import (
	"fmt"

	"github.com/PuerkitoBio/goquery"
	"github.com/charmbracelet/log"
	"github.com/gocolly/colly"
)

type eventDataLink struct {
	text string
	url  string
}

type eventData struct {
	eventString string
	genre       string
	links       []eventDataLink
}

func main() {
	col := colly.NewCollector()

	var eventsData []eventData
	col.OnHTML("div#scrollbar", func(el *colly.HTMLElement) {
		eventsData = scrapeConcerts(el)
	})
	col.Visit("https://berlinmetal.eu/")

	for _, eventData := range eventsData {
		log.Info("", "text", eventData.eventString, "genre", eventData.genre, "links", eventData.links)
	}
}

func scrapeConcerts(el *colly.HTMLElement) []eventData {
	eventsData := make([]eventData, 0)
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
		linksData := make([]eventDataLink, 0, len(linkElements.Nodes))
		for _, linkNode := range linkElements.Nodes {
			linkData := eventDataLink{
				text: linkNode.FirstChild.Data,
				url:  linkNode.Attr[0].Val,
			}
			linksData = append(linksData, linkData)
		}

		eventsData = append(eventsData, eventData{
			eventString: eventString,
			genre:       genre,
			links:       linksData,
		})
	})

	return eventsData
}
