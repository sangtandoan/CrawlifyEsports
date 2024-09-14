package main

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

type Tournament struct {
	Name      string `json:"name"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

const (
	lol  = "League of Legends"
	cs2  = "Counter Strike 2"
	val  = "Valorant"
	pubg = "PUBG"
)

var (
	games = map[string]string{
		lol:  "https://liquipedia.net/leagueoflegends/Main_Page",
		cs2:  "https://liquipedia.net/counterstrike/Main_Page",
		val:  "https://liquipedia.net/valorant/Main_Page",
		pubg: "https://liquipedia.net/pubg/Main_Page",
	}

	// channels = map[string]chan Tournament{
	// 	lol:  make(chan Tournament),
	// 	cs2:  make(chan Tournament),
	// 	val:  make(chan Tournament),
	// 	pubg: make(chan Tournament),
	// }

	tournaments = make(map[string][]Tournament) // map[string][]tournament{}, init empty map
)

func main() {
	start := time.Now()
	// closeChan := make(chan bool)
	// done := make(chan bool)

	wg := &sync.WaitGroup{}
	for key, value := range games {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// scrapingForGame(value, channels[key], done)
			scrapingForGame(value, tournaments, key)
		}()
	}

	// go func() {
	// loop:
	// 	for {
	// 		select {
	// 		case t := <-channels[lol]:
	// 			tournaments[lol] = append(tournaments[lol], t)
	// 		case t := <-channels[cs2]:
	// 			tournaments[cs2] = append(tournaments[cs2], t)
	// 		case t := <-channels[val]:
	// 			tournaments[val] = append(tournaments[val], t)
	// 		case t := <-channels[pubg]:
	// 			tournaments[pubg] = append(tournaments[pubg], t)
	// 		case <-closeChan:
	// 			break loop
	// 		}
	// 	}
	// }()
	//
	// go func() {
	// 	for i := 0; i < len(games); i++ {
	// 		<-done
	// 	}
	//
	// 	close(closeChan)
	// }()

	wg.Wait()
	// Convert struct to json and print to terminal
	// enc := json.NewEncoder(os.Stdout)
	// enc.SetIndent("", " ")
	//
	// enc.Encode(tournaments)
	//

	// Sort tournaments based on StartDate
	for _, tournament := range tournaments {
		slices.SortFunc(tournament, func(a, b Tournament) int {
			startDateA, _ := time.Parse("02-01-2006", a.StartDate)
			startDateB, _ := time.Parse("02-01-2006", b.StartDate)
			if startDateA.Before(startDateB) {
				return -1
			} else {
				return 1
			}
		})
	}

	for key, tournament := range tournaments {
		renderToTable(key, tournament)
		fmt.Println()
	}

	duration := time.Since(start)
	fmt.Println(duration)
}

// func scrapingForGame(link string, ch chan<- Tournament, done chan<- bool) {
// 	collector := colly.NewCollector(colly.Async(true))
//
// 	collector.OnError(func(r *colly.Response, err error) {
// 		fmt.Println(err)
// 	})
//
// 	collector.OnHTML("ul#tournaments-menu-upcoming", func(h *colly.HTMLElement) {
// 		// wg := sync.WaitGroup{}
// 		h.ForEach("a.dropdown-item", func(_ int, el *colly.HTMLElement) {
// 			// wg.Add(1)
//
// 			// go func() {
// 			// 	defer wg.Done()
// 			// Turn relative path in href into absolute path
// 			link := el.Request.AbsoluteURL(el.Attr("href"))
// 			el.Request.Visit(link)
// 			// }()
// 		})
// 		// wg.Wait()
// 	})
//
// 	collector.OnHTML("ul#tournaments-menu-ongoing", func(h *colly.HTMLElement) {
// 		// wg := sync.WaitGroup{}
//
// 		h.ForEach("a.dropdown-item", func(_ int, el *colly.HTMLElement) {
// 			// wg.Add(1)
//
// 			// go func() {
// 			// defer wg.Done()
// 			// Turn relative path in href into absolute path
// 			link := el.Request.AbsoluteURL(el.Attr("href"))
//
// 			el.Request.Visit(link)
// 			// }()
// 		})
// 		// wg.Wait()
// 	})
//
// 	collector.OnHTML("div.fo-nttax-infobox", func(h *colly.HTMLElement) {
// 		tournament := Tournament{}
// 		tournament.Name = h.ChildText("div:nth-child(1) > div.infobox-header")
// 		if tournament.Name == "Upcoming Matches" {
// 			return
// 		}
//
// 		tournament.Name = strings.ReplaceAll(tournament.Name, "[e][h]", "")
//
// 		h.ForEach("div", func(_ int, el *colly.HTMLElement) {
// 			selectorForDate := "div.infobox-description + div"
// 			switch el.ChildText("div.infobox-description") {
// 			case "Start Date:":
// 				tournament.StartDate = formatTime(el.ChildText(selectorForDate))
// 			case "End Date:":
// 				tournament.EndDate = formatTime(el.ChildText(selectorForDate))
// 			case "Date:":
// 				tournament.StartDate = formatTime(el.ChildText(selectorForDate))
// 				tournament.EndDate = formatTime(el.ChildText(selectorForDate))
// 			}
// 		})
//
// 		// tournaments[key] = append(tournaments[key], tournament)
// 		ch <- tournament
// 	})
//
// 	collector.Visit(link)
// 	// wg.Wait()
// 	collector.Wait()
// 	done <- true
// }

func scrapingForGame(link string, tournaments map[string][]Tournament, key string) {
	collector := colly.NewCollector(colly.Async(true))

	collector.OnError(func(r *colly.Response, err error) {
		fmt.Println(err)
	})

	collector.OnHTML("ul#tournaments-menu-upcoming", func(h *colly.HTMLElement) {
		// wg := sync.WaitGroup{}
		h.ForEach("a.dropdown-item", func(_ int, el *colly.HTMLElement) {
			// wg.Add(1)

			// go func() {
			// 	defer wg.Done()
			// Turn relative path in href into absolute path
			link := el.Request.AbsoluteURL(el.Attr("href"))
			el.Request.Visit(link)
			// }()
		})
		// wg.Wait()
	})

	collector.OnHTML("ul#tournaments-menu-ongoing", func(h *colly.HTMLElement) {
		// wg := sync.WaitGroup{}

		h.ForEach("a.dropdown-item", func(_ int, el *colly.HTMLElement) {
			// wg.Add(1)

			// go func() {
			// defer wg.Done()
			// Turn relative path in href into absolute path
			link := el.Request.AbsoluteURL(el.Attr("href"))

			el.Request.Visit(link)
			// }()
		})
		// wg.Wait()
	})

	collector.OnHTML("div.fo-nttax-infobox", func(h *colly.HTMLElement) {
		tournament := Tournament{}
		tournament.Name = h.ChildText("div:nth-child(1) > div.infobox-header")
		if tournament.Name == "Upcoming Matches" {
			return
		}

		tournament.Name = strings.ReplaceAll(tournament.Name, "[e][h]", "")

		h.ForEach("div", func(_ int, el *colly.HTMLElement) {
			selectorForDate := "div.infobox-description + div"
			switch el.ChildText("div.infobox-description") {
			case "Start Date:":
				tournament.StartDate = formatTime(el.ChildText(selectorForDate))
			case "End Date:":
				tournament.EndDate = formatTime(el.ChildText(selectorForDate))
			case "Date:":
				tournament.StartDate = formatTime(el.ChildText(selectorForDate))
				tournament.EndDate = formatTime(el.ChildText(selectorForDate))
			}
		})

		tournaments[key] = append(tournaments[key], tournament)
	})

	collector.Visit(link)
	// wg.Wait()
	collector.Wait()
}

// formatTime change default format from 2006-01-02 to 02-01-2006
func formatTime(oldTime string) string {
	if strings.Contains(oldTime, "-??") {
		oldTime = strings.ReplaceAll(oldTime, "-??", "")
		t, err := time.Parse("2006-01", oldTime)
		if err != nil {
			fmt.Println(err)
		}

		return "??-" + t.Format("01-2006")
	}

	t, err := time.Parse("2006-01-02", oldTime)
	if err != nil {
		fmt.Println(err)
	}

	return t.Format("02-01-2006")
}

func renderToTable(gameName string, tournaments []Tournament) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetAutoIndex(true)
	t.SetTitle(gameName)
	// Use t.Style().something.something to style a specific thing
	// Use t.SetStyle() to style for a whole table
	t.Style().Title.Align = text.AlignCenter

	t.AppendHeader(table.Row{"Tournament", "Start Date", "End Date"})

	for _, tournament := range tournaments {
		t.AppendRow(table.Row{tournament.Name, tournament.StartDate, tournament.EndDate})
	}

	t.SetColumnConfigs([]table.ColumnConfig{
		{
			// Specify name of column to apply config to
			Name:        "Tournament",
			AlignHeader: text.AlignCenter,
		},
	})

	t.Render()
}
