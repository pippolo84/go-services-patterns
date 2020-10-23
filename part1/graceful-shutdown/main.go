package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Pippolo84/go-services-patterns/part1/graceful-shutdown/graceful"
)

const cooldown time.Duration = 5 * time.Second

func main() {
	srv := graceful.NewServer(":8080")

	errs := srv.Run()

	// trap incoming SIGINT and SIGKTERM
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// block until a signal or an error from the server is received
	select {
	case err := <-errs:
		log.Println(err)
	case sig := <-signalChan:
		log.Printf("got signal: %v, shutting down...\n", sig)
	}

	// graceful shutdown the server
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), cooldown)
	defer cancelShutdown()

	var wg sync.WaitGroup
	wg.Add(1)

	if err := srv.Shutdown(shutdownCtx, &wg); err != nil {
		log.Println(err)
	}

	wg.Wait()
}
