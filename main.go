package main

import (
	"log"
	"net/http"
	"sync"
)

var TheDumbo    http.Server
var StopIt      chan struct {}
var WorkerSync  sync.WaitGroup

func main() {
	// Create a ServeMux
	mux := http.NewServeMux();

	// Register the endpoints
	mux.HandleFunc("/hash", CalculateHash)
	mux.HandleFunc("/hash/", GetHash)
	mux.HandleFunc("/stats", GetStatistics)
	mux.HandleFunc("/shutdown", Shutdown)

	// Create a channel for graceful shutdown; nothing will ever be written to it!
	StopIt = make(chan struct {}, 1)

	// Create the http server
	TheDumbo := &http.Server { Addr: ":8080", Handler: mux }

	// Initalize the wait group for worker thread synchronization
	WorkerSync.Add(GetNumOfWorkers())

	// Start the http server in a separate thread
	go func() {
		if err := TheDumbo.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	} ()

	// Wait for gracious shutdown
	<- StopIt

	// All done!
	close(RequestChannel)
	WorkerSync.Wait()
	TheDumbo.Close()
}
