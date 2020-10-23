// +build !integration

package rss

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/mmcdole/gofeed"
)

func TestClient(t *testing.T) {
	testCases := []struct {
		name     string
		handler  func(w http.ResponseWriter, r *http.Request)
		expected []*gofeed.Item
	}{
		{
			name: "empty RSS feed",
			handler: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>\n<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom" xmlns:cc="http://web.resource.org/cc/" xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd" xmlns:media="http://search.yahoo.com/mrss/" xmlns:content="http://purl.org/rss/1.0/modules/content/" xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">\n</rss>`)
			},
			expected: []*gofeed.Item{},
		},
		{
			name: "RSS feed with 5 items",
			handler: func(w http.ResponseWriter, r *http.Request) {
				f, err := os.Open("testdata/golden.xml")
				if err != nil {
					t.Fatal(err)
				}

				if _, err := io.Copy(w, f); err != nil {
					t.Fatal(err)
				}

			},
			expected: []*gofeed.Item{
				{
					Title:       "#1547 - Colin Quinn",
					Description: `Comedian Colin Quinn is a veteran of stage and screen, with notable stints as a cast member of "Saturday Night Live", host of Comedy Central's "Tough Crowd with Colin Quinn", and star of multiple one-man shows on and off Broadway. Quinn is also the author of several books, the most recent of which is Overstated: A Coast-to-Coast Roast of the 50 States. Check out his new show "Cop Show" available now on Colin's YouTube channel: https://bit.ly/3iD0sjV`,
					Content:     `Comedian Colin Quinn is a veteran of stage and screen, with notable stints as a cast member of "Saturday Night Live", host of Comedy Central's "Tough Crowd with Colin Quinn", and star of multiple one-man shows on and off Broadway. Quinn is also the author of several books, the most recent of which is Overstated: A Coast-to-Coast Roast of the 50 States. Check out his new show "Cop Show" available now on Colin's YouTube channel: https://bit.ly/3iD0sjV`,
				},
				{
					Title:       "#1546 - Evan Hafer & Mat Best",
					Description: `Special Forces combat veterans turned entrepreneurs Mat Best and Evan Hafer are co-founders of Black Rifle Coffee Company: a veteran-owned and operated premium, small-batch coffee roastery. When they're not busy at BRCC, you can hear them with co-host Jarred "JT" Taylor on the Free Range American podcast.`,
					Content:     `Special Forces combat veterans turned entrepreneurs Mat Best and Evan Hafer are co-founders of Black Rifle Coffee Company: a veteran-owned and operated premium, small-batch coffee roastery. When they're not busy at BRCC, you can hear them with co-host Jarred "JT" Taylor on the Free Range American podcast.`,
				},
				{
					Title:       "#1545 - W. Keith Campbell",
					Description: `Social psychologist W. Keith Campbell is a recognized expert on narcissism and its influence on society at large. His latest book, The New Science of Narcissism, explores the origins of this character trait, why its presence has grown to almost epidemic proportions, and how all of us are at least a little narcissistic.`,
					Content:     `Social psychologist W. Keith Campbell is a recognized expert on narcissism and its influence on society at large. His latest book, The New Science of Narcissism, explores the origins of this character trait, why its presence has grown to almost epidemic proportions, and how all of us are at least a little narcissistic.`,
				},
				{
					Title:       "#1544 - Tim Dillon",
					Description: `Tim Dillon is a comedian, tour guide, and host. His podcast “The Tim Dillon Show” is available on Spotify.`,
					Content:     `Tim Dillon is a comedian, tour guide, and host. His podcast “The Tim Dillon Show” is available on Spotify.`,
				},
				{
					Title:       "#1543 - Brian Muraresku & Graham Hancock",
					Description: `Attorney and scholar Brian C. Muraresku is the author of The Immortality Key: The Secret History of the Religion with No Name. Featuring an introduction by Graham Hancock, The Immortality Key is a look into the psychedelic origins of the world's great spiritual practices and what those might mean for how we view ourselves and the world around us. Hancock's most recent book is America Before: The Key to Earth's Lost Civilization, now available in Paperback.`,
					Content:     `Attorney and scholar Brian C. Muraresku is the author of The Immortality Key: The Secret History of the Religion with No Name. Featuring an introduction by Graham Hancock, The Immortality Key is a look into the psychedelic origins of the world's great spiritual practices and what those might mean for how we view ourselves and the world around us. Hancock's most recent book is America Before: The Key to Earth's Lost Civilization, now available in Paperback.`,
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := NewClient()

			ts := httptest.NewServer(http.HandlerFunc(tc.handler))

			res, err := client.Fetch(ts.URL)
			if err != nil {
				t.Fatal(err)
			}

			if len(res) != len(tc.expected) {
				t.Fatalf("expected %d items, got %d\n", len(tc.expected), len(res))
			}

			for i, item := range res {
				if item.Title != tc.expected[i].Title {
					t.Fatalf("expected title %q, got %q\n", tc.expected[i].Title, item.Title)
				}
				if item.Description != tc.expected[i].Description {
					t.Fatalf("expected description %q, got %q\n", tc.expected[i].Description, item.Description)
				}
				if item.Content != tc.expected[i].Content {
					t.Fatalf("expected content %q, got %q\n", tc.expected[i].Content, item.Content)
				}
			}

			ts.Close()
		})
	}
}
