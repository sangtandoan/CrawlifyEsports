package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/gocolly/colly"
	"github.com/ipinfo/go/v2/ipinfo"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"

	_ "time/tzdata"
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

	tiers = make(map[string]string)

	// channels = map[string]chan Tournament{
	// 	lol:  make(chan Tournament),
	// 	cs2:  make(chan Tournament),
	// 	val:  make(chan Tournament),
	// 	pubg: make(chan Tournament),
	// }

	tournaments = make(map[string][]Tournament) // map[string][]tournament{}, init empty map
	matches     = make(map[string][]Match)
	timeZone    = "Local"
)

// gameSlice implement flag.Value interface
type gamesSlice []string

// flag.Value interface method
func (gs *gamesSlice) String() string {
	return strings.Join(*gs, ",")
}

// flag.Value interface method
func (gs *gamesSlice) Set(value string) error {
	// *gs = append(*gs, value) uses this if want to have same multiple flag to set this value
	*gs = strings.Split(value, ",") // uses this if just want to have one flag to set this value using csv
	return nil
}

const token = "5004afb124de6d"

func main() {
	start := time.Now()
	// closeChan := make(chan bool)
	// done := make(chan bool)

	// create ipClient to fetch data from ip
	ipClient := ipinfo.NewClient(nil, nil, token)
	// fetch global IP
	ip, err := getGlobalIP()
	if err != nil {
		fmt.Println(err)
	}
	info, err := ipClient.GetIPInfo(net.ParseIP(ip))
	if err != nil {
		fmt.Println(err)
	}
	timeZone = info.Timezone
	fmt.Println(timeZone)

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

	crawlMatchesForLOL()

	// Sort tournaments, matches based on StartDate
	sortByStartDate()

	for key := range tournaments {
		renderToTable(key, tournaments[key], matches[key])
		fmt.Println()
	}

	// Check if dir exist, if not create new one
	_, err = os.Stat("data")
	if os.IsNotExist(err) {
		err := os.Mkdir("data", 0755)
		if err != nil {
			panic(err)
		}
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		exportJSON()
	}()

	wg.Add(1)
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

	ch1 := make(chan Tournament)
	ch2 := make(chan Match)

	go func() {
		for tournament := range ch1 {
			tournaments[key] = append(tournaments[key], tournament)
		}
	}()

	go func() {
		for match := range ch2 {
			matches[key] = append(matches[key], match)
		}
	}()

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
			t, _ := time.Parse("January 02, 2006 - 15:04 MST", startTime)
			// Load loocal time
			loc, _ := time.LoadLocation(timeZone)
			fmt.Println(loc)
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
				ch2 <- match
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
			ch1 <- tournament
			tiers[link] = tournament.Tier
			matchesCollector.Visit(link)
		}
	})

	collector.Visit(link)
	// wg.Wait()

	collector.Wait()
	matchesCollector.Wait()

	// Ensure not leak goroutines
	close(ch1)
	close(ch2)
}

func crawlMatchesForLOL() {
	collector := colly.NewCollector()

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

	collector.OnHTML("div.panel-box div:nth-child(2) div:nth-child(1) div", func(h *colly.HTMLElement) {
		wg := new(sync.WaitGroup)
		mutex := &sync.Mutex{}

		h.ForEach("table", func(_ int, el *colly.HTMLElement) {
			wg.Add(1)
			go func() {
				defer wg.Done()

				team_left := el.ChildText("td.team-left")
				team_right := el.ChildText("td.team-right")
				if team_left != "TBD" && team_right != "TBD" {
					link := el.ChildAttr("div.tournament-text-flex a", "href")
					link = el.Request.AbsoluteURL(link)

					if isTournamentInTiers(link) {
						// crawl startTime
						// When parsing a time with a zone abbreviation like MST, if the zone abbreviation has a defined offset in the current location, then that offset is used.
						// The zone abbreviation "UTC" is recognized as UTC regardless of location.
						// If the zone abbreviation is unknown, Parse records the time as being in a fabricated location with the given zone abbreviation and a zero offset.
						// This choice means that such a time can be parsed and reformatted with the same layout losslessly, but the exact instant used in the representation will differ by the actual zone offset.
						// To avoid such problems, prefer time layouts that use a numeric zone offset, or use ParseInLocation.
						// -> If time has timezone use time.LoadLocation to load that timezone to parse and load local to parse again
						startTime := el.ChildText("span.match-countdown span.timer-object")
						cest, _ := time.LoadLocation("CET")
						t, err := time.ParseInLocation("January 2, 2006 - 15:04 MST", startTime, cest)
						if err != nil {
							fmt.Println(err)
						}
						// Load loocal time
						loc, _ := time.LoadLocation(timeZone)
						fmt.Println(loc)
						// timeInUTCPlus7 := time.FixedZone("UTC+7", 7*60*60)
						// Change time to local time
						t = t.In(loc)

						link := el.Request.AbsoluteURL(el.ChildAttr("span.match-countdown a", "href"))

						match := Match{
							TournamentName: el.ChildText("div.tournament-text-flex a"),
							TeamLeft:       team_left,
							TeamRight:      team_right,
							StartTime:      t,
							Links:          []string{link},
						}

						mutex.Lock()
						matches["lol"] = append(matches["lol"], match)
						mutex.Unlock()
					}
				}
			}()
		})

		wg.Wait()
	})

	collector.Visit("https://liquipedia.net/leagueoflegends/Liquipedia:Matches")
}

func isTournamentInTiers(link string) bool {
	if _, ok := tiers[link]; !ok {
		temp := strings.Split(link, "/")
		link = strings.Join(temp[:len(temp)-1], "/")

		if _, ok := tiers[link]; !ok {
			return false
		}
	}

	return true
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

// sortByStartDate sorts Matches and Tournaments in increasing order
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
	tier := flag.String("tier", "b", "select which tiers to crawl")
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

// getGlobalIP send a GET request to ipinfo to get the global IP address
func getGlobalIP() (string, error) {
	res, err := http.Get("https://ipinfo.io/ip")
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	// Read the response body
	ip, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(ip), nil
}
