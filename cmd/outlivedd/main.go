package main

import (
	"context"
	"flag"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/bobg/aesite"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		sig := <-sigCh
		log.Printf("got signal %s", sig)
		cancel()
	}()

	var (
		seed       = flag.Int64("seed", time.Now().Unix(), "RNG seed")
		projectID  = flag.String("project", "outlived-163105", "project ID")
		locationID = flag.String("location", "xxx", "location ID")
		addr       = flag.String("addr", ":80", "web server listen address")
		test       = flag.Bool("test", false, "run in test mode")
	)

	flag.Parse()

	rand.Seed(*seed)

	if *test {
		err := aesite.DSTest(ctx, *projectID)
		if err != nil {
			log.Fatal(err)
		}
	}

	c, err := newController(ctx, *projectID, *locationID)
	if err != nil {
		log.Fatal(err)
	}

	// http.HandleFunc("/", c.handleHome)
	http.HandleFunc("/load", c.handleLoad)

	log.Printf("listening for requests on %s", *addr)

	srv := &http.Server{Addr: *addr}
	go srv.ListenAndServe()

	<-ctx.Done()
	srv.Shutdown(ctx)
}
