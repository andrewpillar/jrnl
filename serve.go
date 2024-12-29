package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"
)

const serveAddr = "localhost:8080"

var ServeCmd = &Command{
	Usage: "serve",
	Short: "serve the published jrnl",
	Run:   serveCmd,
}

func serveCmd(cmd *Command, args []string) error {
	if err := Initialized("."); err != nil {
		return err
	}

	srv := http.Server{
		Addr:         serveAddr,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
		Handler:      http.FileServer(http.Dir(siteDir)),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				fmt.Fprintf(os.Stderr, "%s: %s\n", cmd.Argv0, err)
			}
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	<-ch

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	srv.Shutdown(ctx)

	return nil
}
