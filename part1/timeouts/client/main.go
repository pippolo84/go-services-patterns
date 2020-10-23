package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"time"
)

// func main() {
// 	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		time.Sleep(2 * time.Second)
// 		fmt.Fprint(w, "Hello, client!")
// 	}))

// 	client := &http.Client{
// 		// Timeout specifies a time limit for requests made by this
// 		// Client. The timeout includes connection time, any
// 		// redirects, and reading the response body. The timer remains
// 		// running after Get, Head, Post, or Do return and will
// 		// interrupt reading of the Response.Body.
// 		Timeout: time.Second,
// 	}

// 	resp, err := client.Get(ts.URL)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer resp.Body.Close()

// 	buf, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println(string(buf))
// }

func main() {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		fmt.Fprint(w, "Hello, client!")
	}))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL, nil)
	if err != nil {
		panic(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(buf))
}
