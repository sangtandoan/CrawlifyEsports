package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/gocolly/colly"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

type Tournament struct {
	GameName  string `json:"-" csv:"game"`
	Name      string `json:"name" csv:"name"`
	StartDate string `json:"start_date" csv:"start_date"`
	EndDate   string `json:"end_date" csv:"end_date"`
	Tier      string `json:"tier" csv:"tier"`
}

type Match struct {
	GameName       string    `json:"-" csv:"game"`
	TournamentName string    `json:"tournament_name" csv:"tournament_name"`
	TeamLeft       string    `json:"team_left" csv:"team_left"`
	TeamRight      string    `json:"team_right" csv:"team_right"`
	StartTime      time.Time `json:"start_time" csv:"start_time"`
	Links          []string  `json:"links" csv:"links"`
}

type CommandLineArgs struct {
	tier        string
	chosenGames []string
	ongoing     bool
	upcoming    bool
}

var (
	games = map[string]string{
		"lol":  "https://liquipedia.net/leagueoflegends/Main_Page",
		"cs2":  "https://liquipedia.net/counterstrike/Main_Page",
		"val":  "https://liquipedia.net/valorant/Main_Page",
		"pubg": "https://liquipedia.net/pubg/Main_Page",
	}

	names = map[string]string{
		"lol":  "League of Legends",
		"cs2":  "Counter Strike 2",
		"val":  "Valorant",
		"pubg": "PUBG",
	}

	// channels = map[string]chan Tournament{
	// 	lol:  make(chan Tournament),
	// 	cs2:  make(chan Tournament),
	// 	val:  make(chan Tournament),
	// 	pubg: make(chan Tournament),
	// }

	tournaments = make(map[string][]Tournament) // map[string][]tournament{}, init empty map
	matches     = make(map[string][]Match)
)

// gameSlice implement flag.Value interface
type gamesSlice []string

// flag.Value interface method
func (gs *gamesSlice) String() string {
	return strings.Join(*gs, ",")
}

// flag.Value interface method
func (gs *gamesSlice) Set(value string) error {
	// *gs = append(*gs, value) uses this if want to have multiple flag to set this value
	*gs = strings.Split(value, ",") // uses this if just want to have one flag to set this value using csv
	return nil
}

func main() {
	start := time.Now()
	// closeChan := make(chan bool)
	// done := make(chan bool)

	args := parseCommandLineArgs()

	wg := &sync.WaitGroup{}
	for _, game := range args.chosenGames {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// scrapingForGame(value, channels[key], done)
			scrapingForGame(games[game], tournaments, game, args, matches)
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
	sortByStartDate()

	for key := range tournaments {
		renderToTable(key, tournaments[key], matches[key])
		fmt.Println()
	}

	err := os.Mkdir("data", 0755)
	if err != nil {
		panic(err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		exportJSON()
	}()

	go func() {
		defer wg.Done()
		exportCSV()
	}()
	wg.Wait()

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

func scrapingForGame(link string, tournaments map[string][]Tournament, key string, args CommandLineArgs, matches map[string][]Match) {
	collector := colly.NewCollector(colly.Async(true), colly.CacheDir(""))
	// Need to create a new Collector if clone not working
	matchesCollector := colly.NewCollector(colly.Async(true), colly.CacheDir(""))

	collector.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Cache-Control", "no-cache")
		r.Headers.Set("Pragma", "no-cache")
		r.Headers.Set("Expires", "0")

		// Add a timestamp to the URL to bypass cache
		r.URL.RawQuery += "&t=" + fmt.Sprint(time.Now().Unix())
	})

	collector.OnError(func(r *colly.Response, err error) {
		fmt.Println(err)
	})

	matchesCollector.OnHTML("div.main-container", func(h *colly.HTMLElement) {
		tournamentName := h.ChildText("div.tabs-static + div.fo-nttax-infobox-wrapper > div.fo-nttax-infobox > div:nth-child(1) > div.infobox-header")
		tournamentName = strings.ReplaceAll(tournamentName, "[e][h]", "")

		h.ForEach("table.infobox_matches_content", func(_ int, el *colly.HTMLElement) {
			startTime := el.ChildText("span.match-countdown")
			startTime = strings.TrimSpace(strings.ReplaceAll(startTime, "UTC", ""))
			t, _ := time.Parse("January 02, 2006 - 15:04", startTime)
			// Load loocal time
			loc, _ := time.LoadLocation("Local")
			// Change time to local time
			t = t.In(loc)
			match := Match{TournamentName: tournamentName, StartTime: t, GameName: key}
			match.TeamLeft = el.ChildText("td.team-left span.team-template-text")
			match.TeamRight = el.ChildText("td.team-right span.team-template-text")
			match.Links = []string{func() string {
				if attr, ok := el.DOM.Find("span.match-countdown > a:first-of-type").Attr("href"); ok {
					return el.Request.AbsoluteURL(strings.TrimSpace(attr))
				}
				return ""
			}()}

			if match.TeamLeft != "TBD" && match.TeamRight != "TBD" {
				matches[key] = append(matches[key], match)
			}
		})
	})

	collector.OnHTML("ul#tournaments-menu-upcoming", func(h *colly.HTMLElement) {
		if !args.upcoming {
			return
		}
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
		if !args.ongoing {
			return
		}
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

	collector.OnHTML("div.fo-nttax-infobox-wrapper", func(h *colly.HTMLElement) {
		tournament := Tournament{GameName: key}
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
			case "Liquipedia Tier:":
				tournament.Tier = el.ChildText("a")
			}
		})

		if filterByTier(args.tier, tournament) {
			link := h.Request.AbsoluteURL(h.Request.URL.Path)
			matchesCollector.Visit(link)
			tournaments[key] = append(tournaments[key], tournament)

		}
	})

	collector.Visit(link)
	// wg.Wait()

	collector.Wait()
	matchesCollector.Wait()
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

func renderToTable(gameName string, tournaments []Tournament, matches []Match) {
	t := createTableWriter(gameName)
	m := createTableWriter(gameName)
	addAlignCenterForColumns(m, "Versus", "Time", "Links")

	m.AppendHeader(table.Row{"Tournament", "Versus", "Time", "Links"})

	t.AppendHeader(table.Row{"Tournament", "Start Date", "End Date", "Tier"})

	for _, tournament := range tournaments {
		t.AppendRow(table.Row{tournament.Name, tournament.StartDate, tournament.EndDate, tournament.Tier})
	}

	for _, match := range matches {
		m.AppendRow(table.Row{match.TournamentName, match.TeamLeft + " vs " + match.TeamRight, match.StartTime.Format("15:04 02/01/2006"), match.Links})
	}

	t.Render()
	m.Render()
}

func createTableWriter(gameName string) table.Writer {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetAutoIndex(true)
	t.SetTitle(names[gameName])
	// Use t.Style().something.something to style a specific thing
	// Use t.SetStyle() to style for a whole table
	t.Style().Title.Align = text.AlignCenter
	t.SetColumnConfigs([]table.ColumnConfig{
		{
			// Specify name of column to apply config to
			Name:        "Tournament",
			AlignHeader: text.AlignCenter,
		},
	})

	return t
}

func addAlignCenterForColumns(t table.Writer, cols ...string) {
	// cols is just a slice
	cfg := []table.ColumnConfig{}
	for _, col := range cols {
		cfg = append(cfg, table.ColumnConfig{
			Name:        col,
			AlignHeader: text.AlignCenter,
		})
	}

	t.SetColumnConfigs(cfg)
}

func sortByStartDate() {
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
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
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, match := range matches {
			slices.SortFunc(match, func(a, b Match) int {
				if a.StartTime.Before(b.StartTime) {
					return -1
				} else {
					return 1
				}
			})
		}
	}()

	wg.Wait()
}

func parseCommandLineArgs() CommandLineArgs {
	// Using csv to parse chosen games
	// chosenGamesFlag := flag.String("games", "", "select which games want to crawl")
	// flag.Parse()
	// chosenGames := strings.Split(*chosenGamesFlag, ",")
	// fmt.Println(chosenGames)

	// Using flag.Var and custom implementation of flag.Value to parse
	var chosenGames gamesSlice
	flag.Var(&chosenGames, "games", "select which games want to crawl")
	tier := flag.String("tier", "b", "slect which tiers to crawl")
	ongoing := flag.Bool("ongoing", true, "select ongoing tournaments")
	upcoming := flag.Bool("upcoming", true, "select upcoming tournaments")
	flag.Parse()

	if len(chosenGames) == 0 {
		for key := range names {
			chosenGames = append(chosenGames, key)
		}
	}

	args := CommandLineArgs{
		chosenGames: chosenGames,
		tier:        *tier,
		ongoing:     *ongoing,
		upcoming:    *upcoming,
	}

	return args
}

// filterByTier filters tournaments based on their tier, return true if tournament matches the tier
func filterByTier(tier string, tournament Tournament) bool {
	cur_tier := strings.ToLower(strings.ReplaceAll(tournament.Tier, "-Tier", ""))

	if cur_tier != "s" {
		if tier == "s" {
			return false
		} else {
			if cur_tier > tier {
				return false
			}
		}
	}

	return true
}

func exportJSON() {
	f, err := os.Create("data/tournaments.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	f2, err := os.Create("data/matches.json")
	if err != nil {
		panic(err)
	}
	defer f2.Close()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		buffer, err := json.Marshal(tournaments)
		if err != nil {
			panic(err)
		}

		f.Write(buffer)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		json.NewEncoder(f2).Encode(matches)
	}()

	wg.Wait()
}

func exportCSV() {
	f, err := os.Create("data/tournaments.csv")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	f2, err := os.Create("data/matches.csv")
	if err != nil {
		panic(err)
	}
	defer f2.Close()

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Using gocsv
		for _, match := range matches {
			err = gocsv.MarshalFile(match, f2)
			if err != nil {
				panic(err)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		// Using encoding/csv
		w := csv.NewWriter(f)
		defer w.Flush()
		w.Write([]string{"game", "name", "start_date", "end_date", "tier"})

		for key, tournament := range tournaments {
			for _, t := range tournament {
				w.Write([]string{key, t.Name, t.StartDate, t.EndDate, t.Tier})
			}
		}
	}()

	wg.Wait()
}
