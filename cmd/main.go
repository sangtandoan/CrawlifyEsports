package main

import (
	"fmt"
	"os"
	"strings"
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
		if tournament.Name == "Upcoming Matches" {
			return
		}

		tournament.Name = strings.Replace(tournament.Name, "[e][h]", "", -1)

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

	// Convert struct to json and print to terminal
	// enc := json.NewEncoder(os.Stdout)
	// enc.SetIndent("", " ")
	//
	// enc.Encode(tournaments)

	renderToTable(tournaments)

	duration := time.Since(start)
	fmt.Println(duration)
}

// formatTime change default format from 2006-01-02 to 02-01-2006
func formatTime(oldTime string) string {
	if strings.Contains(oldTime, "-??") {
		oldTime = strings.Replace(oldTime, "-??", "", -1)
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

func renderToTable(tournaments []Tournament) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetAutoIndex(true)
	t.SetTitle("Tournaments")
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
