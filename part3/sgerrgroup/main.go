package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/Pippolo84/go-services-patterns/part3/sgerrgroup/client"
	"github.com/montanaflynn/stats"
)

func main() {
	const nrun int = 500

	rand.Seed(time.Now().UnixNano())

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		x := rand.Float32()
		switch {
		case x >= 0.999:
			time.Sleep(1000 * time.Millisecond)
		case x >= 0.99:
			time.Sleep(800 * time.Millisecond)
		case x >= 0.95:
			time.Sleep(80 * time.Millisecond)
		case x >= 0.5:
			time.Sleep(10 * time.Millisecond)
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}))
	defer ts.Close()

	for _, concurrency := range []int{1, 2, 5, 10} {
		fmt.Printf("concurrency level %d\n", concurrency)
		c := client.NewClient(http.DefaultClient, concurrency)

		latency := make([]time.Duration, nrun)
		for i := 0; i < nrun; i++ {
			start := time.Now()

			req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
			if err != nil {
				panic(err)
			}

			resp, err := c.Do(req)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()

			if _, err := io.Copy(ioutil.Discard, resp.Body); err != nil {
				panic(err)
			}

			latency[i] = time.Since(start)
		}

		printStats(latency)
	}
}

func printStats(latency []time.Duration) {
	data := stats.LoadRawData(latency)

	p50, err := stats.Percentile(data, 50)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\tp50\t= %v\n", time.Duration(p50))

	p95, err := stats.Percentile(data, 95)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\tp95\t= %v\n", time.Duration(p95))

	p99, err := stats.Percentile(data, 99)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\tp99\t= %v\n", time.Duration(p99))

	p999, err := stats.Percentile(data, 99.9)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\tp99.9\t= %v\n", time.Duration(p999))

	p9999, err := stats.Percentile(data, 99.99)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\tp99.99\t= %v\n", time.Duration(p9999))

	p99999, err := stats.Percentile(data, 99.999)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\tp99.999\t= %v\n", time.Duration(p99999))

	max, err := stats.Max(data)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\tmax\t= %v\n", time.Duration(max))
	fmt.Println()
}
