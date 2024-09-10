package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

type Tournament struct {
	Name      string `json:"name"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

func main() {
	start := time.Now()
	tournaments := []Tournament{}
	games := []string{
		"https://liquipedia.net/leagueoflegends/Main_Page",
		"https://liquipedia.net/counterstrike/Main_Page",
		"https://liquipedia.net/valorant/Main_Page",
		"https://liquipedia.net/pubg/Main_Page",
	}
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
			// 	defer wg.Done()
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
		tournament.Name = strings.Replace(tournament.Name, "[e][h]", "", -1)

		h.ForEach("div", func(_ int, el *colly.HTMLElement) {
			switch el.ChildText("div.infobox-description") {
			case "Start Date:":
				tournament.StartDate = formatTime(el.ChildText("div.infobox-description + div"))
			case "End Date:":
				tournament.EndDate = formatTime(el.ChildText("div.infobox-description + div"))
			default:
			}
		})

		tournaments = append(tournaments, tournament)
	})

	// wg := sync.WaitGroup{}
	for _, game := range games {
		// wg.Add(1)
		// go func() {
		// defer wg.Done()
		collector.Visit(game)
		// }()
	}
	// wg.Wait()
	collector.Wait()

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", " ")

	enc.Encode(tournaments)

	duration := time.Since(start)
	fmt.Println(duration)
}

func formatTime(oldTime string) string {
	t, err := time.Parse("2006-01-02", oldTime)
	if err != nil {
		fmt.Println(err)
	}

	return t.Format("02/01/2006")
}
