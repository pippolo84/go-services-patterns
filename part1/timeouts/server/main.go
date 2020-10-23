package main

import (
	"log"
	"time"

	//"github.com/Pippolo84/go-services-patterns/part1/timeouts/server/selective"
	"github.com/Pippolo84/go-services-patterns/part1/timeouts/server/streaming"
)

// Use the server in the "streaming" package to show that the WriteTimeout is
// not compatible with a streaming server:
// the "/streaming" endpoint will be closed before finishing the streaming

func main() {
	// Despite the "Connection":"keep-alive" header, due to WriteTimeout,
	// the connection will be closed.
	srv := streaming.NewServer(":8080", streaming.Timeouts{
		Write: 2 * time.Second,
	})

	for err := range srv.Run() {
		log.Fatal(err)
	}
}

// Take a look at the server in the "selective" package: it is able to extend the
// Write timeout selectively
// the "/" endpoint connection WON'T be closed after 2 seconds
// the "/timeout" endpoint WILL be closed after 2 seconds

// func main() {
// 	// Despite the "Connection":"keep-alive" header, due to WriteTimeout,
// 	// the connection will be closed.
// 	srv := selective.NewServer(":8080", selective.Timeouts{
// 		Write: 2 * time.Second,
// 	})

// 	for err := range srv.Run() {
// 		log.Fatal(err)
// 	}
// }
