package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
)

var commands = map[string]func(context.Context, *flag.FlagSet, []string) error{
	"scrape": cliScrape,
	"serve": cliServe,
}

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		log.Fatal("usage: subcommand [args]")
	}

	cmd := flag.Arg(0)
	flagset := flag.NewFlagSet("", flag.ContinueOnError)
	fn, ok := commands[cmd]
	if !ok {
		log.Fatalf("unknown command %s", cmd)
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		sig := <-sigCh
		log.Printf("got signal %s", sig)
		cancel()
	}()

	args := flag.Args()
	err := fn(ctx, flagset, args[1:])
	if err != nil {
		log.Fatal(err)
	}
}
