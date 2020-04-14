package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	jsoniter "github.com/json-iterator/go"
)

var startDate time.Time
var endDate time.Time

type formats struct {
	JSON int
	CSV  int
}

// Formats is an Enum for data formats. Treat as read-only.
var Formats = &formats{
	JSON: 0,
	CSV:  1,
}

func main() {

	DataFormat := Formats.JSON // default data format
	fmt.Println(DataFormat)

	err := parseArgs(os.Args)
	if err != "" {
		fmt.Println(err)
		return
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

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Scraping", r.URL)
	})

	c.OnHTML(".chart-list__element", func(e *colly.HTMLElement) {

		var line ChartLine

		line.Rank = strings.TrimSpace(e.DOM.Find(".chart-element__rank__number").Text())
		line.Song = strings.TrimSpace(e.DOM.Find(".chart-element__information__song").Text())
		line.Artist = strings.TrimSpace(e.DOM.Find(".chart-element__information__artist").Text())
		line.LastWeek = strings.TrimSpace(e.DOM.Find(".chart-element__meta.text--last").Text())
		line.Trend = strings.TrimSpace(e.DOM.Find(".chart-element__trend").Text())
		// correct a typo in Billboard's data:
		if line.Trend == "Failing" {
			line.Trend = "Falling"
		}
		line.Movement = strings.TrimSpace(e.DOM.Find(".chart-element__information__delta__text.text--default").Text())
		line.Peak = strings.TrimSpace(e.DOM.Find(".chart-element__meta.text--peak").Text())
		line.Weeks = strings.TrimSpace(e.DOM.Find(".chart-element__meta.text--week").Text())
		thisWeek = append(thisWeek, line)
	})

	for week := startDate; week.Before(endDate); week = week.AddDate(0, 0, 7) {

		createDataDirectories(week)

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

		var json = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonOut, _ := json.MarshalIndent(thisWeek, "", " ")
		if _, err := f.Write(jsonOut); err != nil {
			log.Println(err)
		}
		f.Close()

		// throttle our requests
		log.Println("self-throttle: pausing for two seconds")
		time.Sleep(2 * time.Second)
	}

	log.Println("-- Done.")
}

func parseArgs(args []string) string {
	if len(args) != 3 && len(args) != 2 {
		err := "Usage: Billboard_Hot.100 startDate [endDate]\n" +
			"Dates must be in the form: YYYY-MM-DD\n" +
			"Up to twenty weeks at a time may be downloaded per execution.\n" +
			"endDate defaults to twenty weeks past the startDate.\n\n"
		return err
	}

	// Hot 100 Charts released on Saturdays for the coming week: the week of Monday's date
	// first week ever: released Aug 2, 1958, for the week of August 4th, 1958
	now := time.Now()
	firstWeek, _ := time.Parse("2006-01-02", "1958-08-04")
	startDate, _ = time.Parse("2006-01-02", args[1]) // command line argument one
	if startDate.Before(firstWeek) {
		startDate = firstWeek
	}

	if len(args) == 3 { // endDate was specified
		endDate, _ = time.Parse("2006-01-02", args[2]) // command line argument two
	} else {
		endDate = startDate.AddDate(0, 0, 140) // set default endDate to twenty weeks
	}
	if endDate.After(now) {
		endDate = now
	}
	if endDate.Before(startDate) {
		endDate = startDate
	}

	// backup in time to the nearest Monday...
	for startDate.Weekday(); startDate.Weekday() != 1; startDate = startDate.AddDate(0, 0, -1) {
		//log.Println(startDate.String())
	}

	//fmt.Println("Start Date:", startDate, "-- End date:", endDate)
	return ""
}

func createDataDirectories(week time.Time) {
	// create data directory if necessary
	if err := os.Mkdir("data", 0755); os.IsExist(err) {
		//log.Println("data directory already exists")
	}
	// create data/year directory if necessary
	if err := os.Mkdir("data/"+week.Format("2006"), 0755); os.IsExist(err) {
		//log.Println("data/" + week.Format("2006") + " directory already exists")
	}
}
