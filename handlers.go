package main

import (
	"context"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

const RQST_SERVICE_TIME = 5

type hashRequest struct {
	id		int64
	startTime       int64
	clearpwd	string
}

var RequestChannel      chan hashRequest
var rqstMap             RequestMap
var stats               ServerStatistics = ServerStatistics { Total: 0, Average: 0 }
var nextId              int64 = 0

func worker() {
	// The loop executes as long as the channel is not closed
	for hr := range RequestChannel {
		log.Printf("Process request ID: %v", hr);

		// Wait if needed
		now := time.Now().Unix()
		if now < (hr.startTime + RQST_SERVICE_TIME) {
			time.Sleep(time.Duration(hr.startTime + RQST_SERVICE_TIME - now) * time.Second)
		}

		begin := time.Now()

		// Process the request
		hasher := sha512.New()
		hasher.Write([]byte(hr.clearpwd))
		rqstMap.Add(hr.id, base64.URLEncoding.EncodeToString(hasher.Sum(nil)))

		end := time.Now();
		elapsed := end.Sub(begin);

		stats.Add(elapsed.Microseconds())

		log.Println(hr)
		rqstMap.Dump()
		stats.Dump()
	}

	log.Println("Worker thread terminating")
	WorkerSync.Done()
}

func init() {
	RequestChannel = make(chan hashRequest, GetChannelSize())
	for i := 0; i < GetNumOfWorkers(); i++ {
		go worker()
	}
}

func CalculateHash (w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/hash" {
		http.NotFound(w, r)
		return
	}

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest);
			return
		}

		pwd := r.PostFormValue("password")
		if pwd == "" {
			http.Error(w, "Required form field not found", http.StatusBadRequest);
			return
		}

		hr := hashRequest {
			id:             atomic.AddInt64(&nextId, 1),
			startTime:      time.Now().Unix(),
			clearpwd:       pwd,
		}

		rqstMap.Add(hr.id, "")

		RequestChannel <- hr

		fmt.Fprintln(w, hr.id)
	} else {
		http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
	}
}

func GetHash (w http.ResponseWriter, r *http.Request) {
	// Split is weird; it is what it is!
	pathSegments := strings.Split(r.URL.Path, "/");

	// Make sure there are two and only two path segments
	if 3 != len(pathSegments) {
		http.NotFound(w, r)
		return
	}

	// Make sure the first segment is hash
	if pathSegments[1] != "hash" {
		http.NotFound(w, r)
		return
	}

	// Make sure the next segment is a number
	var id int64
	id, err := strconv.ParseInt(pathSegments[2], 10, 64)
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest);
		return
	}

	if id <= 0 {
		http.Error(w, "Invalid request ID", http.StatusBadRequest);
		return
	}

	// Find the ID in the map
	hashstr, found := rqstMap.Get(id)
	if !found {
		http.Error(w, "Request not found", http.StatusNotFound);
		return;
	}

	if hashstr == "" {
		http.Error(w, "Request not ready", http.StatusNotFound);
	}

	fmt.Fprintln(w, hashstr)
}

func GetStatistics (w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/stats" {
		http.NotFound(w, r)
		return
	}

	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		jsonstr, err := stats.Marshal()
		if err != nil {
			http.Error(w, "JSON marshalling failed", http.StatusInternalServerError)
		} else {
			fmt.Fprintln(w, jsonstr)
		}
	} else {
		http.Error(w, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
	}
}

func Shutdown (w http.ResponseWriter, r *http.Request) {
	if err := TheDumbo.Shutdown(context.Background()); err != nil {
		log.Fatalf("HTTP server Shutdown: %v", err)
	}

	log.Printf("HTTP server shutdown graciously")
	close(StopIt)
}
