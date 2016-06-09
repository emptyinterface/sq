package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/emptyinterface/sq"
)

func main() {

	if len(os.Args) != 2 {
		fmt.Println(os.Args[0], "<twitter username>")
		os.Exit(1)
	}

	const site = "https://twitter.com/"

	var page struct {
		Title  string `sq:"title | text"`
		Tweets []struct {
			Text       string    `sq:"p.tweet-text | text"`
			AuthorName string    `sq:"strong.fullname | text"`
			Username   string    `sq:"span.username | text"`
			Link       *url.URL  `sq:"a.js-permalink | attr(href)"`
			Created    time.Time `sq:"a.tweet-timestamp | attr(title) | time(3:04 PM - _2 Jan 2006)"`
			Retweets   int       `sq:"span.ProfileTweet-action--retweet > span | attr(data-tweet-stat-count)"`
			Likes      int       `sq:"span.ProfileTweet-action--favorite > span | attr(data-tweet-stat-count)"`
		} `sq:"div.content"`
	}

	resp, err := http.Get(site + os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	for _, err := range sq.Scrape(&page, resp.Body) {
		fmt.Println(err)
	}

	fmt.Printf("%s\n\n", page.Title)

	for _, tweet := range page.Tweets {
		fmt.Printf("%s (%s)\n", tweet.Username, tweet.Created.Format("2006/01/02"))
		fmt.Println(tweet.Text)
		fmt.Printf("(Likes: %d, Retweets: %d)\n\n", tweet.Likes, tweet.Retweets)
	}

}
