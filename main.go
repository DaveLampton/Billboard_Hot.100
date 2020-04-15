package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
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

// ChartLine represents a single song on the chart of a particular week.
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

func main() {

	err := parseArgs(os.Args)
	if err != "" {
		fmt.Println(err)
		return
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
		if _, err := os.Stat("data/" + week.Format("2006") + "/" + week.Format("2006-01-02") + ".json"); os.IsNotExist(err) {
			fmt.Printf("File does not exist. Creating...\n")

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
			log.Println("self-throttle: pausing for three seconds")
			time.Sleep(3 * time.Second)
		} /*  else {
			fmt.Println("Data file already exists.")
		} */
	}

	log.Println("-- Done.")
}

func parseArgs(args []string) string {

	var argCount = len(args) - 1
	dataFormat := Formats.JSON // default format
	verify := flag.Bool("verify", false, "Verify that all JSON and CSV files"+
		" in the data directory each contain 100 songs. Invalid"+
		" data files will be deleted.")
	CSV := flag.Bool("csv", false, "use the CSV data format")
	flag.Parse()

	if len(args) < 2 || len(args) > 5 {
		err := "Usage: Billboard_Hot.100 startDate [endDate]\n" +
			"Dates must be in the form: YYYY-MM-DD\n" +
			"Up to twenty weeks at a time may be downloaded per execution.\n" +
			"endDate defaults to twenty weeks past the startDate.\n\n"
		return err
	}

	if *CSV {
		argCount--
		dataFormat = Formats.CSV
	}

	if *verify {
		argCount--
		fmt.Println("Verify flag specified. startDate and endDate are ignored.")
		fmt.Println("Checking all data... deleting invalid files.")
		return verifyData()
	}

	if dataFormat == 1 {
		fmt.Println("Using CSV data format.")
	} else {
		fmt.Println("Using JSON data format.")
	}
	// Hot 100 Charts released on Saturdays for the coming week: the week of Monday's date
	// first week ever: released Aug 2, 1958, for the week of August 4th, 1958
	now := time.Now()
	firstWeek, _ := time.Parse("2006-01-02", "1958-08-04")
	startDate, _ = time.Parse("2006-01-02", args[len(args)-argCount]) //
	if startDate.Before(firstWeek) {
		startDate = firstWeek
	}
	fmt.Println(startDate.Format("2006-01-02"))

	if argCount > 1 { // endDate was specified
		endDate, _ = time.Parse("2006-01-02", args[len(args)-1]) //
	} else {
		endDate = startDate.AddDate(0, 0, 140) // set default endDate to twenty weeks
	}
	//fmt.Println(endDate.Format("2006-01-02"))

	if endDate.After(now) {
		endDate = now
	}
	if endDate.Before(startDate) {
		endDate = startDate
	}
	fmt.Println(endDate.Format("2006-01-02"))

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

func verifyData() string {
	err := filepath.Walk("data",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			switch {
			case err != nil:
				log.Println("filepath.Walk: ", err)
			case info.IsDir():
				//fmt.Println(path, "Directory")
			default:
				//fmt.Println(path, info.Size())
				f, err := os.Open(path)
				if err != nil {
					log.Println("OpenFile:", err)
				}
				var contents = make([]byte, info.Size())
				_, err = f.Read(contents)
				if err != nil {
					log.Println("Reading:", err)
				}
				f.Close()
				//fmt.Print("File contents:\n", string(contents))

				theWeek := make([]ChartLine, 100)
				err = jsoniter.Unmarshal(contents, &theWeek)
				if err != nil {
					log.Println("Unmarshalling:", err)
				}
				// for some rerason several charts in late 1976 to early 1977
				// have only 99 songs in the Hot100, so I'll accept 99 songs
				if len(theWeek) != 100 && len(theWeek) != 99 {
					fmt.Println(path, ": len(theWeek):", len(theWeek))
					// delete file
					var err = os.Remove(path)
					if err != nil {
						fmt.Println("Problem deleteing file", path)
					}
					fmt.Println("File Deleted")
				}
			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}
	return "Any invalid/incomplete files have been removed. Remaining data checks out OK."
}
