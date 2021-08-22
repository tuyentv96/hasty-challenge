package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/tuyentv96/hasty-challenge/cmd"
)

func main() {
	ctx := context.Background()
	cli, cleanup, err := cmd.InitApplication(ctx)
	if err != nil {
		log.Fatalln(err.Error())
	}

	handleSigterm(cleanup)

	err = cli.Commands().Run(os.Args)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

// handleSigterm -- Handles Ctrl+C or most other means of "controlled" shutdown gracefully.
// Invokes the supplied func before exiting.
func handleSigterm(handleExit func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		handleExit()
		os.Exit(1)
	}()
}
