package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

func main() {
	if len(os.Args) != 3 && len(os.Args) != 2 {
		fmt.Print("Usage:", os.Args[0], "startDate", "[endDate]\n"+
			"Dates must be in the form: YYYY-MM-DD\n"+
			"Up to twenty weeks at a time may be downloaded per execution.\n"+
			"endDate defaults to twenty weeks past the startDate.\n\n")
		return
	}

	// Hot 100 Charts released on Saturdays for the coming week: the week of Monday's date.
	// First week ever: released Aug 2, 1958, for the week of August 4th.
	now := time.Now()
	firstWeek, _ := time.Parse("2006-01-02", "1958-08-04")
	startDate, _ := time.Parse("2006-01-02", os.Args[1]) // command line argument one
	if startDate.Before(firstWeek) {
		startDate = firstWeek
	}
	var endDate time.Time
	if len(os.Args) == 3 { // endDate was specified
		endDate, _ = time.Parse("2006-01-02", os.Args[2]) // command line argument two
	} else {
		endDate = startDate.AddDate(0, 0, 140) // add twenty weeks
	}
	if endDate.After(now) {
		endDate = now
	}
	if endDate.Before(startDate) {
		fmt.Println("endDate must come after startDate")
		return
	}
	//fmt.Println("Start Date:", startDate, " End date:", endDate)

	// backup in time to the nearest Monday...
	for startDate.Weekday(); startDate.Weekday() != 1; startDate = startDate.AddDate(0, 0, -1) {
		//log.Println(startDate.String())
	}

	type ChartLine struct {
		Rank     string
		Song     string
		Artist   string
		LastWeek string
		Trend    string
		Movement string
		Peak     string
		Weeks    string
	}

	var thisWeek []ChartLine

	c := colly.NewCollector()

	c.OnHTML(".chart-list__element", func(e *colly.HTMLElement) {

		var line ChartLine

		line.Rank = strings.TrimSpace(e.DOM.Find(".chart-element__rank__number").Text())
		line.Song = strings.TrimSpace(e.DOM.Find(".chart-element__information__song").Text())
		line.Artist = strings.TrimSpace(e.DOM.Find(".chart-element__information__artist").Text())
		line.LastWeek = strings.TrimSpace(e.DOM.Find(".chart-element__meta.text--last").Text())
		line.Trend = strings.TrimSpace(e.DOM.Find(".chart-element__trend").Text())
		// Correct a typo in Billboard's data:
		if line.Trend == "Failing" {
			line.Trend = "Falling"
		}
		line.Movement = strings.TrimSpace(e.DOM.Find(".chart-element__information__delta__text.text--default").Text())
		line.Peak = strings.TrimSpace(e.DOM.Find(".chart-element__meta.text--peak").Text())
		line.Weeks = strings.TrimSpace(e.DOM.Find(".chart-element__meta.text--week").Text())
		thisWeek = append(thisWeek, line)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Scraping", r.URL)
	})

	// count of weeks retrieved in this execution run
	weeks := 0
	// Billboard.com seems to have a limit at around 20 requests per minute
	maxWeeks := 20
	for week := startDate; week.Before(endDate) && weeks < maxWeeks; week = week.AddDate(0, 0, 7) {

		// Create data directory if necessary
		if err := os.Mkdir("data", 0755); os.IsExist(err) {
			//log.Println("data directory already exists")
		}
		// Create year directory if necessary
		if err := os.Mkdir("data/"+week.Format("2006"), 0755); os.IsExist(err) {
			//log.Println("data/" + week.Format("2006") + " directory already exists")
		}

		thisWeek = make([]ChartLine, 0)

		// open file for this week
		log.Println("-- Week:", week.Format("2006-01-02"))
		f, err := os.OpenFile("data/"+week.Format("2006")+"/"+week.Format("2006-01-02")+".json",
			os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err)
		}
		// instead of defer.close(), the file is explicitly closed below.

		c.Visit("https://www.billboard.com/charts/hot-100/" + week.Format("2006-01-02"))
		if err != nil {
			log.Println(string("Visit error: ") + err.Error())
		}

		json, _ := json.MarshalIndent(thisWeek, "", " ")
		if _, err := f.Write(json); err != nil {
			log.Println(err)
		}
		f.Close()

		// increment weeks processed
		weeks++
	}

	log.Println("-- Done.")
}
