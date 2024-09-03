package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Defining database as a global, not best practice solution but works for this case.
var database DB

func main() {
	var err error
	database, err = NewFileDB("db.json")
	if err != nil {
		fmt.Printf("Error initializing database: %s\n", err)
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/cmd", handleCommand)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println("Server is running on http://0.0.0.0:8080")
	err = RunServer(server)
	if err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}

// RunServer runs ListenAndServe on the http.Server with graceful shutdown.
func RunServer(srv *http.Server) error {
	shutdownError := make(chan error)

	// Start a background goroutine.
	go func() {
		// Create a quit channel which carries os.Signal values.
		quit := make(chan os.Signal, 1)

		// Use signal.Notify() to listen for incoming SIGINT and SIGTERM signals and  relay them to the quit channel.
		// Any other signals will not be caught by signal.Notify() and will retain their default behavior.
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		// Read the signal from the quit channel.
		// This code will block until a signal is received.
		sig := <-quit

		log.Printf("caught signal: %s\n", sig.String())

		// Create a context with a 5-second timeout.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Shutdown() will return nil if the graceful shutdown was successful,
		// or an error (which may happen because of a problem closing the listeners,
		// or because the shutdown didn't complete before the 5-second context deadline is hit).
		// We relay this return value to the shutdownError channel.
		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}
		shutdownError <- nil
	}()

	// Calling Shutdown() on our api will cause ListenAndServe() to immediately
	// return a http.ErrServerClosed error. So if we see this error, it is actually a
	// good thing and an indication that the graceful shutdown has started. So we check
	// specifically for this, only returning the error if it is NOT http.ErrServerClosed.
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// Otherwise, we wait to receive the return value from Shutdown() on the shutdownError channel.
	// If return value is an error, we know that there was a problem with the graceful shutdown.
	err = <-shutdownError
	if err != nil {
		return err
	}

	return nil
}
