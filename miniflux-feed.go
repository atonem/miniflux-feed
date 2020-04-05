package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	feeds "github.com/gorilla/feeds"
	miniflux "miniflux.app/client"
)

var url = os.Getenv("MINIFLUX_URL")
var token = os.Getenv("MINIFLUX_TOKEN")
var port = os.Getenv("PORT")
var external_url = os.Getenv("MINIFLUX_EXTERNAL_URL")

// Authentication with an API Key:
var client = miniflux.New(url, token)

func main() {
	if len(external_url) == 0 {
		external_url = url
	}

	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Printf("Miniflux url: %v", url)
	log.Printf("Miniflux external url: %v", external_url)
	log.Printf("Listening on: %v", port)
	http.HandleFunc("/", FeedHandler)
	http.ListenAndServe(":"+port, nil)

}

// TODO: Params
func getEntries(limit int, offset int) (miniflux.Entries, error) {
	// Fetch all feeds.
	filter := miniflux.Filter{
		Limit:     limit,
		Offset:    offset,
		Status:    "unread",
		Order:     "published_at",
		Direction: "desc",
	}
	result, err := client.Entries(&filter)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return result.Entries, nil

}

func FeedHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
	var limit = 10
	var offset = 0
	query := r.URL.Query()

	// Limit
	if query.Get("limit") != "" {
		parsed, err := strconv.Atoi(query.Get("limit"))

		if err != nil {
			fmt.Println(err)
			fmt.Fprintf(w, "Error: %s", err)
			return
		}
		limit = parsed
	}

	// Offset
	if query.Get("offset") != "" {
		parsed, err := strconv.Atoi(query.Get("offset"))

		if err != nil {
			fmt.Println(err)
			fmt.Fprintf(w, "Error: %s", err)
			return
		}
		offset = parsed
	}

	entries, err := getEntries(limit, offset)

	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "Error: %s", err)
		return
	}
	feed := CreateFeedFromEntries(entries)

	rss, err := feed.ToRss()

	if err != nil {
		fmt.Println(err)
	}

	fmt.Fprint(w, rss)
}

func CreateFeedFromEntries(entries miniflux.Entries) *feeds.Feed {

	var items []*feeds.Item

	for _, entry := range entries {
		items = append(items, &feeds.Item{
			Title: entry.Feed.Title + " | " + entry.Title,
			// TODO: Allow configuration if using /undread/ or /history/
			Link:    &feeds.Link{Href: external_url + "/unread/entry/" + strconv.FormatInt(entry.ID, 10)},
			Content: entry.Content,
			Author: &feeds.Author{
				Name: entry.Author,
			},
			Created: entry.Date,
		})
	}

	feed := &feeds.Feed{
		Title:       "miniflux@" + external_url,
		Link:        &feeds.Link{Href: external_url},
		Description: "Generated feed for " + external_url,
		Created:     time.Now(),
		Items:       items,
	}

	return feed

}
