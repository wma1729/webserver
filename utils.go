package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"sync"
)

const DEFAULT_CHANNEL_SIZE = 256
const DEFAULT_NUM_OF_WORKERS = 10

//
// This is a thread-safe version of a map with request ID
// as the key and the corresponding hashed password as the
// value.
// 
// When the request is first seen, it is added to the map
// with an empty hashed password. This is used as an indication
// of the fact that the request is received.
//
// When the request is later processed, the key is updated
// with the hashed password.
//
type RequestMap struct {
	requests	map[int64]string
	lock		sync.RWMutex
}

func (rm *RequestMap) Add(key int64, value string) {
	rm.lock.Lock()
	defer rm.lock.Unlock()
	if rm.requests == nil {
		rm.requests = make(map[int64]string)
	}
	rm.requests[key] = value
}

func (rm *RequestMap) Get(key int64) (string, bool) {
	rm.lock.RLock()
	defer rm.lock.RUnlock()
	value, found := rm.requests[key]
	return value, found
}

func (rm *RequestMap) Dump() {
	rm.lock.RLock()
	defer rm.lock.RUnlock()
	log.Printf("Number of requests in map = %d", len(rm.requests))
	for key, value := range rm.requests {
		log.Printf("%d => %s\n", key, value)
	}
}

//
// This is a thread-safe server statistics info.
// The average is computed as information is added.
// We could've kept the total time and calculated the
// average as requested and that would be faster if
// there are a handful of stats requests.
//
type ServerStatistics struct {
	Total   int64           `json:"total"`
        Average int64           `json:"average"`
	lock    sync.Mutex      `json:"-"`
}

func (ss *ServerStatistics) Add(microsec int64) {
	ss.lock.Lock()
	defer ss.lock.Unlock()

	if ss.Total == 0 {
		ss.Total = 1
		ss.Average = microsec
	} else {
		ss.Total += 1
		ss.Average = (ss.Average + microsec) / 2
	}
}

func (ss *ServerStatistics) Marshal() (string, error) {
	ss.lock.Lock()
	defer ss.lock.Unlock()

	buf, err := json.Marshal(ss)
	if err != nil {
		log.Fatalf("Marshaling failed: %v", err)
	}
	return string(buf), err
}

func (ss *ServerStatistics) Dump() {
	ss.lock.Lock()
	defer ss.lock.Unlock()

	log.Printf("Total Requests = %d, Average time = %d microseconds", ss.Total, ss.Average)
}

var channelsize int = 0
var numofworker int = 0

func GetChannelSize() (int) {
	if channelsize != 0 {
		return channelsize
	}

	csz := os.Getenv("CHANNEL_SIZE")
	if csz != "" {
		i, err := strconv.ParseInt(csz, 10, 32)
		if err != nil {
			log.Printf("Invalid value of CHANNEL_SIZE %s", csz)
			channelsize = DEFAULT_CHANNEL_SIZE
		} else {
			if i < 256 {
				channelsize = 256
			} else if (i > 16384) {
				channelsize = 16384
			} else {
				channelsize = int(i)
			}
		}
	} else {
		channelsize = DEFAULT_CHANNEL_SIZE
	}

	return channelsize
}

func GetNumOfWorkers() (int) {
	if numofworker != 0 {
		return numofworker
	}

	nw := os.Getenv("NUM_OF_WORKERS")
	if nw != "" {
		i, err := strconv.ParseInt(nw, 10, 32)
		if err != nil {
			log.Printf("Invalid value of NUM_OF_WORKERS %s", nw)
			numofworker = DEFAULT_NUM_OF_WORKERS
		} else {
			if i < 10 {
				numofworker = 10
			} else if (i > 100) {
				numofworker = 100
			} else {
				numofworker = int(i)
			}
		}
	} else {
		numofworker = DEFAULT_NUM_OF_WORKERS
	}

	return numofworker
}
