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
	if len(os.Args) != 3 {
		fmt.Println("Usage:", os.Args[0], "StartDate_(YYYY-MM-DD)", "EndDate_(YYYY-MM-DD)")
		return
	}

	//var Hot100 map[string]chart

	// Hot 100 Charts released on Saturdays for the coming week: the week of Monday's date.
	// First week ever: released Aug 2, 1958, for the week of August 4th.
	startDate, _ := time.Parse("2006-01-02", os.Args[1])
	endDate, _ := time.Parse("2006-01-02", os.Args[2])

	// backup to the nearest Saturday...
	for startDate.Weekday(); startDate.Weekday() != 6; startDate = startDate.AddDate(0, 0, -1) {
		//log.Println(startDate.String())
	}

	type ChartLine struct {
		Rank     string
		Trend    string
		Song     string
		Artist   string
		LastWeek string
		Movement string
		Peak     string
		Weeks    string
	}

	var thisWeek []ChartLine

	c := colly.NewCollector()

	c.OnHTML(".chart-list__element", func(e *colly.HTMLElement) {

		var line ChartLine

		line.Rank = strings.TrimSpace(e.DOM.Find(".chart-element__rank__number").Text())
		line.Trend = strings.TrimSpace(e.DOM.Find(".chart-element__trend").Text())
		line.Song = strings.TrimSpace(e.DOM.Find(".chart-element__information__song").Text())
		line.Artist = strings.TrimSpace(e.DOM.Find(".chart-element__information__artist").Text())
		line.LastWeek = strings.TrimSpace(e.DOM.Find(".chart-element__meta.text--last").Text())
		line.Movement = strings.TrimSpace(e.DOM.Find(".chart-element__information__delta__text.text--default").Text())
		line.Peak = strings.TrimSpace(e.DOM.Find(".chart-element__meta.text--peak").Text())
		line.Weeks = strings.TrimSpace(e.DOM.Find(".chart-element__meta.text--week").Text())
		//NewLine, _ := json.Marshal(line)
		//fmt.Println(string(NewLine))
		thisWeek = append(thisWeek, line)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	for week := startDate; week.Before(endDate); week = week.AddDate(0, 0, 7) {

		thisWeek = make([]ChartLine, 0)

		// open file for this week
		log.Println("week", week.Format("2006.01.02"))
		f, err := os.OpenFile("data/"+week.Format("2006")+"/"+week.Format("2006-01-02")+".json",
			os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err)
		}
		// instead of defer.close(), we'll explicitly close the file below.

		c.Visit("https://www.billboard.com/charts/hot-100/" + week.Format("2006-01-02"))
		if err != nil {
			log.Println(string("Visit error: ") + err.Error())
		}

		json, _ := json.MarshalIndent(thisWeek, "", " ")
		if _, err := f.Write(json); err != nil {
			log.Println(err)
		}
		f.Close()
	}

}
