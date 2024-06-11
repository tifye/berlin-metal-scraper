package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/log"
)

func parseRawData(rawEvents []rawEventData) []Event {
	loc, _ := time.LoadLocation("Europe/Berlin")
	events := make([]Event, 0)
	for _, rawEvent := range rawEvents {
		event, err := parseEventString(rawEvent.eventString, loc)
		if err != nil {
			log.Info("err parsing raw event data", "raw", rawEvent.eventString)
			continue
		}

		event.Links = cleanEventLinks(rawEvent.links)
		events = append(events, event)
	}

	return events
}

func parseEventString(eventString string, loc *time.Location) (Event, error) {
	dateStr, title, at, err := cutEventString(eventString)
	if err != nil {
		return Event{}, nil
	}

	date, err := parseRawEventDate(dateStr, loc)
	if err != nil {
		return Event{}, err
	}

	return Event{
		Title: title,
		At:    at,
		Date:  date,
	}, nil
}

func parseRawEventDate(dateStr string, loc *time.Location) (time.Time, error) {
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

	currentTime := time.Now().In(loc)
	year := currentTime.Year()
	if month < int(currentTime.Month()) {
		year += 1
	}

	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc)
	return date, nil
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

func cutEventString(eventString string) (date, title, at string, err error) {
	eventString = strings.Trim(eventString, "@ ")
	date, titleAt, found := strings.Cut(eventString, " ")
	if !found {
		return "", "", "", fmt.Errorf("malformed raw event string %s; expected format", eventString)
	}

	title, at, _ = strings.Cut(titleAt, "@")
	return
}
