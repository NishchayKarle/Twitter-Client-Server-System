package server

import (
	"encoding/json"
	"os"
	"proj1/feed"
	"proj1/queue"
	"proj1/semaphore"
	"runtime"
	"sync"
)

type ActionRespone struct {
	Success bool `json:"success"`
	Id      int  `json:"id"`
}

type FeedResponse struct {
	Id       int              `json:"id"`
	UserFeed []*feed.JsonPost `json:"feed"`
}

type Config struct {
	Encoder *json.Encoder // Represents the buffer to encode Responses
	Decoder *json.Decoder // Represents the buffer to decode Requests
	Mode    string        // Represents whether the server should execute
	// sequentially or in parallel
	// If Mode == "s"  then run the sequential version
	// If Mode == "p"  then run the parallel version
	// These are the only values for Version
	ConsumersCount int // Represents the number of consumers to spawn
}

func NewConfig(m string, consumers int) *Config {
	return &Config{json.NewEncoder(os.Stdout), json.NewDecoder(os.Stdin), m, consumers}
}

func producer(config Config, lockFreeQueue *queue.LockFreeQueue, sem *semaphore.Semaphore, done *bool) {
	var request queue.Request
	e := config.Decoder.Decode(&request)
	if e != nil {
		return
	}

	if sem != nil {
		sem.Up()
	}

	lockFreeQueue.Enqueue(&request)

	if request.Command == "DONE" {
		*done = true
		return
	}
}

func consumer(config Config, lockFreeQueue *queue.LockFreeQueue,
	userFeed feed.Feed, sem *semaphore.Semaphore, wg *sync.WaitGroup, done *bool) {

	for {
		request := lockFreeQueue.Dequeue()
		if request == nil && config.Mode == "s" {
			break
		}

		if request == nil && *done {
			break
		}

		if request != nil {
			switch request.Command {
			case "ADD":
				var response ActionRespone
				userFeed.Add(request.Body, request.Timestamp)
				response.Success = true
				response.Id = request.Id
				config.Encoder.Encode(&response)

			case "REMOVE":
				var response ActionRespone
				response.Success = false
				if userFeed.Contains(request.Timestamp) {
					userFeed.Remove(request.Timestamp)
					response.Success = true
				}
				response.Id = request.Id
				config.Encoder.Encode(&response)

			case "CONTAINS":
				var response ActionRespone
				response.Success = false
				if userFeed.Contains(request.Timestamp) {
					response.Success = true
				}
				response.Id = request.Id
				config.Encoder.Encode(&response)

			case "FEED":
				var response FeedResponse
				response.Id = request.Id
				response.UserFeed = userFeed.GetFeed()
				config.Encoder.Encode(&response)

			case "DONE":
				break
			}
		} else {
			if sem != nil {
				sem.Down()
			}
		}
	}

	if config.Mode == "p" {
		wg.Done()
	}

}

//Run starts up the twitter server based on the configuration
//information provided and only returns when the server is fully
// shutdown.
func Run(config Config, userFeed feed.Feed) {
	lockFreeQueue := queue.NewLockFreeQueue()
	done := false
	if config.Mode == "s" {
		for !done {
			producer(config, lockFreeQueue, nil, &done)
			consumer(config, lockFreeQueue, userFeed, nil, nil, &done)
		}

	} else {
		var wg sync.WaitGroup
		sem := semaphore.NewSemaphore(0)

		for i := 0; i < config.ConsumersCount; i++ {
			go consumer(config, lockFreeQueue, userFeed, sem, &wg, &done)
			wg.Add(1)
		}

		for !done {
			producer(config, lockFreeQueue, sem, &done)
		}

		for runtime.NumGoroutine() > 1 {
			sem.Up()
		}
	}
}
