# CrawlifyEsports
`CrawlifyEsports` is a lightweight command-line tool designed to scrape [Liquipedia](https://liquipedia.net/) for tournament data and upcoming match information across a variety of esports titles.

## Motivation
Esport has rapidly grown into a massive global industry, with countless tournaments and matches taking place across various games every day. Keeping up with all this information, especially for fans and professionals alike, can be challenging.

Liquipedia is one of the most comprehensive resources for esports information, but manually browsing for tournament and match data can be time-consuming. `CrawlifyEsports` was created to streamline this process by automating the extraction of tournament details and upcoming matches directly from Liquipedia.

## 🚀 Quick Start

### Install CrawlifyEsports using Git

```bash
git clone https://github.com/FrostJ143/GamesTournamentsScraper.git
```

### Run CrawlifyEsports to crawl all tournaments with a+ tier 

```bash
./crawlify -tier=a
```

## Usage

Available flags:

* `-tier` - choose which tiers of tournaments above to crawl (s, a, b, ..., default: b)
* `-upcoming` - choose to crawl upcoming tournaments or not (default: true)
* `-ongoing` - choose to crawl ongoing tournaments or not (default: true)
* `-games` - choose which games to crawl (default: all)

Example:

```bash
./crawlify -games=cs2,lol,val -tier=a -upcoming=false
```

## Supported Games:
* Counter Strike 2 (cs2)
* Valorant (val)
* League of Legend (lol)
* PUBG (pubg)


