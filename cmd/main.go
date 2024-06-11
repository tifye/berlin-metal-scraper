package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/charmbracelet/log"
	"github.com/gocolly/colly"
)

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
	col.Visit("https://berlinmetal.eu/")

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

func parseRawData(rawEvents []rawEventData) []Event {
	events := make([]Event, 0)
	for _, rawEvent := range rawEvents {
		event, err := parseEventString(rawEvent.eventString)
		if err != nil {
			log.Info("err parsing raw event data", "raw", rawEvent.eventString)
			continue
		}

		event.Links = cleanEventLinks(rawEvent.links)
		events = append(events, event)
	}

	return events
}

func cleanEventLinks(rawLinks []rawEventDataLink) []EventLink {
	links := make([]EventLink, 0, len(rawLinks))
	for _, rawLink := range rawLinks {
		title := rawLink.text
		title = strings.ReplaceAll(title, "ⓘ", "Information")
		links = append(links, EventLink{Title: title, Url: rawLink.url})
	}
	return links
}

func parseEventString(eventString string) (Event, error) {
	eventString = strings.Trim(eventString, "@ ")
	dateStr, titleAt, found := strings.Cut(eventString, " ")
	if !found {
		return Event{}, fmt.Errorf("malformed raw event string %s; expected format", eventString)
	}

	title, at, _ := strings.Cut(titleAt, "@")

	date, err := parseRawEventDate(dateStr)
	if err != nil {
		return Event{}, err
	}

	return Event{
		Title: title,
		At:    at,
		Date:  date,
	}, nil
}

func parseRawEventDate(dateStr string) (time.Time, error) {
	monthStr, dayStr, found := strings.Cut(dateStr, "-")
	if !found {
		return time.Time{}, fmt.Errorf("malformed event date string %s; expected mm-dd", dateStr)
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil {
		return time.Time{}, err
	}

	day, err := strconv.Atoi(dayStr)
	if err != nil {
		return time.Time{}, err
	}

	loc, _ := time.LoadLocation("Europe/Berlin")
	currentTime := time.Now().In(loc)

	year := currentTime.Year()
	if month < int(currentTime.Month()) {
		year += 1
	}

	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc)
	return date, nil
}
