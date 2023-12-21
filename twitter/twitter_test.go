package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	//"io/ioutil"
	"math/rand"
	"os/exec"
	"strconv"
	"testing"
	"time"
)

type _TestAddRequest struct {
	Command   string  `json:"command"`
	Id        int64   `json:"id"`
	Timestamp float64 `json:"timestamp"`
	Body      string  `json:"body"`
}
type _TestRemoveRequest struct {
	Command   string  `json:"command"`
	Id        int64   `json:"id"`
	Timestamp float64 `json:"timestamp"`
}
type _TestContainsRequest struct {
	Command   string  `json:"command"`
	Id        int64   `json:"id"`
	Timestamp float64 `json:"timestamp"`
}
type _TestFeedRequest struct {
	Command string `json:"command"`
	Id      int64  `json:"id"`
}
type _TestDoneRequest struct {
	Command string `json:"command"`
}

type _TestNormalResponse struct {
	Success bool  `json:"success"`
	Id      int64 `json:"id"`
}

type _TestFeedResponse struct {
	Id   int64           `json:"id"`
	Feed []_TestPostData `json:"feed"`
}

type _TestPostData struct {
	Body      string  `json:"body"`
	Timestamp float64 `json:"timestamp"`
}

func generateSlice(size int) []int {

	slice := make([]int, size, size)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < size; i++ {
		slice[i] = i
	}
	return slice
}

///////
// Auxiliary functions needed for the tests.
//////
func createAdds(numbers []int, idx int) (map[int]_TestAddRequest, map[int]_TestNormalResponse, int) {

	requests := make(map[int]_TestAddRequest)
	responses := make(map[int]_TestNormalResponse)

	for _, number := range numbers {
		numberStr := strconv.Itoa(number)
		request := _TestAddRequest{"ADD", int64(idx), float64(number), numberStr}
		response := _TestNormalResponse{true, int64(idx)}
		requests[idx] = request
		responses[idx] = response
		idx++
	}
	return requests, responses, idx
}
func createContains(numbers []int, successes []bool, idx int) (map[int]_TestContainsRequest, map[int]_TestNormalResponse, int) {

	requests := make(map[int]_TestContainsRequest)
	responses := make(map[int]_TestNormalResponse)

	for i, number := range numbers {
		request := _TestContainsRequest{"CONTAINS", int64(idx), float64(number)}
		response := _TestNormalResponse{successes[i], int64(idx)}
		requests[idx] = request
		responses[idx] = response
		idx++
	}
	return requests, responses, idx
}
func createRemoves(numbers []int, successes []bool, idx int) (map[int]_TestRemoveRequest, map[int]_TestNormalResponse, int) {

	requests := make(map[int]_TestRemoveRequest)
	responses := make(map[int]_TestNormalResponse)

	for i, number := range numbers {
		request := _TestRemoveRequest{"REMOVE", int64(idx), float64(number)}
		response := _TestNormalResponse{successes[i], int64(idx)}
		requests[idx] = request
		responses[idx] = response
		idx++
	}
	return requests, responses, idx
}
func createFeed(numbers []int, idx int) (_TestFeedRequest, _TestFeedResponse, int) {

	postData := make([]_TestPostData, len(numbers))
	request := _TestFeedRequest{"FEED", int64(idx)}

	for i, number := range numbers {
		numberStr := strconv.Itoa(number)
		postData[i] = _TestPostData{numberStr, float64(number)}
	}
	response := _TestFeedResponse{int64(idx), postData}
	return request, response, idx + 1
}

//////
// Beginning of actual twitter tests
//////

// SimpleDone
// Action(s):
// 1. Sends a single "DONE" Request to the server,
// 2. The server should shutdown. This tests blocks until the server fully shutdowns.
func TestSimpleDone(t *testing.T) {

	numOfThreadsStr := "4"

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "run", "twitter.go", numOfThreadsStr)

	stdin, errIn := cmd.StdinPipe()
	if errIn != nil {
		t.Fatal("<TestSimpleDone>: stdin error in getting stdin pipe: Contact Professor Samuels, if see this message.")
	}

	if err := cmd.Start(); err != nil {
		t.Fatal("<TestSimpleDone> cmd.Start error in executing test: Contact Professor Samuels, if see this message.")
	}
	done := make(chan bool)

	go func() {

		encoder := json.NewEncoder(stdin)

		request := _TestDoneRequest{"DONE"}

		if err := encoder.Encode(&request); err != nil {
			t.Fatal("<TestSimpleDone> cmd.encode in executing test: Contact Professor Samuels, if see this message.")
		}
		done <- true
	}()

	<-done
	if err := cmd.Wait(); err != nil {
		t.Errorf("The automated test timed out. You may have a deadlock, starvation issue and/or you did not implement" +
			"the necessary code for passing this test.")
	}
}

// SimpleWaitDone
// Action(s):
// 1. Creates a a "DONE" Request,
// 2. Waits a few seconds before sending the Request.
// 3. TThe server should shutdown. This tests blocks until the server fully shutdowns.
func TestSimpleWaitDone(t *testing.T) {

	numOfThreadsStr := "4"

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "run", "twitter.go", numOfThreadsStr)
	stdin, errIn := cmd.StdinPipe()
	if errIn != nil {
		t.Fatal("<SimpleWaitDone>: stdin error in getting stdin pipe: Contact Professor Samuels, if see this message.")
	}

	if err := cmd.Start(); err != nil {
		t.Fatal("<SimpleWaitDone> cmd.Start error in executing test: Contact Professor Samuels, if see this message.")
	}
	done := make(chan bool)

	go func() {
		encoder := json.NewEncoder(stdin)

		request := _TestDoneRequest{"DONE"}

		time.Sleep(3 * time.Second)
		if err := encoder.Encode(&request); err != nil {
			t.Fatal("<SimpleWaitDone> cmd.encode error in executing test: Contact Professor Samuels, if see this message.")
		}
		done <- true
	}()

	<-done
	if err := cmd.Wait(); err != nil {
		t.Errorf("The automated test timed out. You may have a deadlock, starvation issue and/or you did not implement" +
			" the necessary code for passing this test.")
	}
}

// AddRequests
// Actions(s):
//  1.  Generates ~100 "ADD" requests,
//  2.  Checks that the server returns back a response acknowledgement for each add request
//  3.  Makes the program wait 30 seconds before continuing with a final Done request
func TestAddRequests(t *testing.T) {

	numOfThreadsStr := "3"

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "run", "twitter.go", numOfThreadsStr)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal("<AddRequests>: error in getting stdout pipe: Contact Professor Samuels, if see this message.")
	}
	stdin, errIn := cmd.StdinPipe()
	if errIn != nil {
		t.Fatal("<AddRequests>: error in getting stdin pipe: Contact Professor Samuels, if see this message.")
	}

	if err := cmd.Start(); err != nil {
		t.Fatal("<AddRequests> cmd.Start error in executing Test: Contact Professor Samuels, if see this message.")
	}
	inDone := make(chan bool)
	outDone := make(chan bool)

	doneRequest := _TestDoneRequest{"DONE"}
	numbers := generateSlice(100)
	requests, responses, _ := createAdds(numbers, 0)

	go func() {
		encoder := json.NewEncoder(stdin)
		for _, request := range requests {
			if err := encoder.Encode(&request); err != nil {
				t.Fatal("<cmd.encode> error in executing Test: Contact Professor Samuels, if see this message.")
			}
		}
		if err := encoder.Encode(&doneRequest); err != nil {
			t.Fatal("<cmd.encode> error in executing Test: Contact Professor Samuels, if see this message.")
		}
		inDone <- true
	}()

	go func() {
		decoder := json.NewDecoder(stdout)
		var count int
		for {
			var response _TestNormalResponse
			if err := decoder.Decode(&response); err != nil {
				break
			}
			if value, ok := responses[int(response.Id)]; ok {
				if value.Id != response.Id || value.Success != response.Success {
					t.Errorf("Id and success fields do not match. Got(Id=%v,Success=%v), Expected(Id=%v,Success=%v)",
						response.Id, response.Success, value.Id, value.Success)
				}
				count++
			} else {
				t.Errorf("Received an invalid id back from twitter.go. We only We should only have ids between 0-99 but got:%v", response.Id)
			}
			if count%5 == 0 {
				time.Sleep(1 * time.Second)
			}
		}
		if count != 100 {
			t.Errorf("Did not receive the right amount of Add acknowledgements. Got:%v, Expected:%v", count, len(requests))
		}
		outDone <- true
	}()

	<-inDone
	<-outDone
	if err := cmd.Wait(); err != nil {
		t.Errorf("The automated test timed out. You may have a deadlock, starvation issue and/or you did not implement" +
			" the necessary code for passing this test.")
	}

}

// SimpleAddRequest
// Action(s):
// 1. This tests spawns a high number of threads with a large block size
// 2. Sends a single Add Request
// 3. Makes sure the response acknowledgement is sent back
// 4. Sends a Done request and waits for the server to exit.
func TestSimpleAddRequest(t *testing.T) {

	numOfThreadsStr := "16"

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "run", "twitter.go", numOfThreadsStr)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal("<SimpleAddRequest>: error in getting stdout pipe: Contact Professor Samuels, if see this message.")
	}
	stdin, errIn := cmd.StdinPipe()
	if errIn != nil {
		t.Fatal("<SimpleAddRequest>: error in getting stdin pipe: Contact Professor Samuels, if see this message.")
	}

	if err := cmd.Start(); err != nil {
		t.Fatal("<SimpleAddRequest> cmd.Start error in executing Test: Contact Professor Samuels, if see this message.")
	}
	inDone := make(chan bool)
	outDone := make(chan bool)

	doneRequest := _TestDoneRequest{"DONE"}
	numbers := generateSlice(1)
	requests, responses, _ := createAdds(numbers, 0)

	go func() {
		encoder := json.NewEncoder(stdin)
		for _, request := range requests {
			time.Sleep(1 * time.Second) // Wait a second before sending the add request
			if err := encoder.Encode(&request); err != nil {
				t.Fatal("<cmd.encode> error in executing Test: Contact Professor Samuels, if you see this message.")
			}
		}
		if err := encoder.Encode(&doneRequest); err != nil {
			t.Fatal("<cmd.encode> error in executing Test: Contact Professor Samuels, if you see this message.")
		}
		inDone <- true
	}()

	go func() {
		decoder := json.NewDecoder(stdout)
		var count int
		for {
			var response _TestNormalResponse
			if err := decoder.Decode(&response); err != nil {
				break
			}
			if value, ok := responses[int(response.Id)]; ok {
				if value.Id != response.Id || value.Success != response.Success {
					t.Errorf("Id and success fields do not match. Got(Id=%v,Success=%v), Expected(Id=%v,Success=%v)",
						response.Id, response.Success, value.Id, value.Success)
				}
				count++
			} else {
				t.Errorf("Received an invalid id back from twitter.go. We only Added the number 'O' but got(id):%v, expected(id):0", response.Id)
			}
			if count%5 == 0 {
				time.Sleep(1 * time.Second)
			}
		}
		if count != 1 {
			t.Errorf("Did not receive the right amount of Add acknowledgements. Got:%v, Expected:%v", count, len(requests))
		}
		outDone <- true
	}()

	<-inDone
	<-outDone
	if err := cmd.Wait(); err != nil {
		t.Errorf("The automated test timed out. You may have a deadlock, starvation issue and/or you did not implement" +
			" the necessary code for passing this test.")
	}
}

// ContainsRequest
// Action(s):
// 1. Sends a rew "CONTAINS" requests before adding anything to the server
// 2. Waits to receive reponses where each should be false
// 3. Sends a Done request and waits for the server to exit.
func TestSimpleContainsRequest(t *testing.T) {

	numOfThreadsStr := "16"

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "run", "twitter.go", numOfThreadsStr)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal("<SimpleContainsRequest>: error in getting stdout pipe: Contact Professor Samuels, if see this message.")
	}
	stdin, errIn := cmd.StdinPipe()
	if errIn != nil {
		t.Fatal("<SimpleContainsRequest>: error in getting stdin pipe: Contact Professor Samuels, if see this message.")
	}

	if err := cmd.Start(); err != nil {
		t.Fatal("<SimpleContainsRequest> cmd.Start error in executing Test: Contact Professor Samuels, if see this message.")
	}
	inDone := make(chan bool)
	outDone := make(chan bool)

	doneRequest := _TestDoneRequest{"DONE"}
	successSlice := make([]bool, 30)
	numbers := generateSlice(30)
	requests, responses, _ := createContains(numbers, successSlice, 0)

	go func() {
		encoder := json.NewEncoder(stdin)
		for _, request := range requests {
			time.Sleep(1 * time.Millisecond) // Wait a second before sending the add request
			if err := encoder.Encode(&request); err != nil {
				t.Fatal("<SimpleContainsRequest> contains cmd.encode error in executing Test: Contact Professor Samuels, if see this message.")
			}
		}
		if err := encoder.Encode(&doneRequest); err != nil {
			t.Fatal("<SimpleContainsRequest> done cmd.encode error in executing Test: Contact Professor Samuels, if see this message.")
		}
		inDone <- true
	}()

	go func() {
		decoder := json.NewDecoder(stdout)
		var count int
		for {
			var response _TestNormalResponse
			if err := decoder.Decode(&response); err != nil {
				break
			}
			if value, ok := responses[int(response.Id)]; ok {
				if value.Id != response.Id || value.Success != response.Success {
					t.Errorf("Id and success fields do not match. Got(Id=%v,Success=%v), Expected(Id=%v,Success=%v)",
						response.Id, response.Success, value.Id, value.Success)
				}
				count++
			} else {
				t.Errorf("Received an invalid id back from twitter.go. We only Added the number 'O' but got(id):%v, expected(id):0", response.Id)
			}
			if count%5 == 0 {
				time.Sleep(1 * time.Second)
			}
		}
		if count != 30 {
			t.Errorf("Did not receive the right amount of Add acknowledgements. Got:%v, Expected:%v", count, len(requests))
		}
		outDone <- true
	}()

	<-inDone
	<-outDone
	if err := cmd.Wait(); err != nil {
		t.Errorf("The automated test timed out. You may have a deadlock, starvation issue and/or you did not implement" +
			" the necessary code for passing this test.")
	}
}

// ContainsRequest
// Action(s):
// 1. Sends "ADD" requests with ids within the range of [0-30] into the feed but also throws in some contains requests. // 2. Check to verify that the contains responses are sent back. We don't care about their return values but rather just ensuring that they get returned by the twitter.go
// 3. Sends a Done request and waits for the server to exit.
func TestAddWithContainsRequest(t *testing.T) {

	numOfThreadsStr := "4"

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "run", "twitter.go", numOfThreadsStr)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal("<AddWithContainsRequest>: error in getting stdout pipe: Contact Professor Samuels, if see this message.")
	}
	stdin, errIn := cmd.StdinPipe()
	if errIn != nil {
		t.Fatal("<AddWithContainsRequest>: error in getting stdin pipe: Contact Professor Samuels, if see this message.")
	}

	if err := cmd.Start(); err != nil {
		t.Fatal("<AddWithContainsRequest> cmd.Start error in executing Test: Contact Professor Samuels, if see this message.")
	}
	inDone := make(chan bool)
	outDone := make(chan bool)

	doneRequest := _TestDoneRequest{"DONE"}
	successSlice := make([]bool, 30)
	numbers := generateSlice(30)
	requestsAdd, responsesAdd, addIdx := createAdds(numbers, 0)
	requestsContains, responsesContains, _ := createContains(numbers, successSlice, addIdx)

	go func() {
		encoder := json.NewEncoder(stdin)
		for idx := 0; idx < len(requestsAdd); idx++ {
			requestAdd := requestsAdd[idx]
			if err := encoder.Encode(&requestAdd); err != nil {
				t.Fatal("<AddWithContainsRequest> ADD cmd.encode error in executing Test: Contact Professor Samuels, if see this message.")
			}
			time.Sleep(1 * time.Millisecond) // Wait milisecond before sending the add request
			requestContains := requestsContains[addIdx]
			if err := encoder.Encode(&requestContains); err != nil {
				t.Fatal("<AddWithContainsRequest> Contains cmd.encode error in executing Test: Contact Professor Samuels, if see this message.")
			}
			addIdx++
		}
		if err := encoder.Encode(&doneRequest); err != nil {
			t.Fatal("<AddWithContainsRequest> Done cmd.encode error in executing Test: Contact Professor Samuels, if see this message.")
		}
		inDone <- true
	}()

	go func() {
		decoder := json.NewDecoder(stdout)
		var count int
		for {
			var response _TestNormalResponse
			if err := decoder.Decode(&response); err != nil {
				break
			}
			if value, ok := responsesAdd[int(response.Id)]; ok {
				if value.Id != response.Id || value.Success != response.Success {
					t.Errorf("Add Request & Response id and success fields do not match. Got(ID=%v,Success=%v), Expected(ID=%v,Success=%v)",
						response.Id, response.Success, value.Id, value.Success)
				}
				count++
			} else if value, ok := responsesContains[int(response.Id)]; ok {
				if value.Id != response.Id {
					t.Errorf("Contains Request & Response id fields do not match. Got(%v), Expected(%v)",
						response.Id, value.Id)
				}
				count++
			} else {
				t.Errorf("Received an invalid id back from twitter.go. We only added ids [0,30] but got(id):%v", response.Id)
			}
			if count%5 == 0 {
				time.Sleep(1 * time.Second)
			}
		}
		if count != len(requestsAdd)+len(requestsContains) {
			t.Errorf("Did not receive the right amount of Add&Contains acknowledgements. Got:%v, Expected:%v", count, len(requestsAdd)+len(requestsContains))
		}
		outDone <- true
	}()

	<-inDone
	<-outDone
	if err := cmd.Wait(); err != nil {
		t.Errorf("The automated test timed out. You may have a deadlock, starvation issue and/or you did not implement" +
			" the necessary code for passing this test.")
	}
}

// SimpleFeedRequest
// Action(s):
// 1. Sends a FEED request on an empty feed.
// 2. Check the response items is empty because the feed is empty.
// 3. Sends a Done request and waits for the server to exit.
func TestSimpleFeedRequest(t *testing.T) {
	numOfThreadsStr := "16"

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "run", "twitter.go", numOfThreadsStr)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal("<SimpleFeedRequest>: error in getting stdout pipe: Contact Professor Samuels, if see this message.")
	}
	stdin, errIn := cmd.StdinPipe()
	if errIn != nil {
		t.Fatal("<SimpleFeedRequest>: error in getting stdin pipe: Contact Professor Samuels, if see this message.")
	}

	if err := cmd.Start(); err != nil {
		t.Fatal("<SimpleFeedRequest> cmd.Start error in executing Test: Contact Professor Samuels, if see this message.")
	}
	inDone := make(chan bool)
	outDone := make(chan bool)

	doneRequest := _TestDoneRequest{"DONE"}

	numbers := []int{}
	request, responseExpected, _ := createFeed(numbers, 0)

	go func() {
		encoder := json.NewEncoder(stdin)
		time.Sleep(1 * time.Millisecond) // Wait a second before sending the feed request
		if err := encoder.Encode(&request); err != nil {
			t.Fatal("<SimpleFeedRequest> Feed cmd.encode error in executing Test: Contact Professor Samuels, if see this message.")
		}
		time.Sleep(1 * time.Millisecond) // Wait a second before sending the add request
		if err := encoder.Encode(&doneRequest); err != nil {
			t.Fatal("<SimpleFeedRequest> Done cmd.encode error in executing Test: Contact Professor Samuels, if see this message.")
		}
		inDone <- true
	}()

	go func() {
		decoder := json.NewDecoder(stdout)
		var count int
		for {
			var response _TestFeedResponse
			if err := decoder.Decode(&response); err != nil {
				break
			}
			if response.Id != responseExpected.Id {
				t.Errorf("Feed Request & Response id fields do not match. Got(%v), Expected(%v)",
					response.Id, responseExpected.Id)
			} else {
				if len(response.Feed) != 0 {
					t.Errorf("Feed Request was sent on an empty feed but the response return posts. Got(%v), Expected(%v)",
						len(response.Feed), 0)
				}
				count++
			}
			if count%5 == 0 {
				time.Sleep(1 * time.Second)
			}
		}
		if count != 1 {
			t.Errorf("Did not receive the right amount of Feed Request acknowledgements. Got:%v, Expected:%v", count, 1)
		}
		outDone <- true
	}()

	<-inDone
	<-outDone
	if err := cmd.Wait(); err != nil {
		t.Errorf("The automated test timed out. You may have a deadlock, starvation issue and/or you did not implement" +
			" the necessary code for passing this test.")
	}
}

// SimpleAddAndFeedRequest
// Action(s):
// 1. Sends series of Add requests to be added to the feed.
// 2. Checks to verify the responses are sent back from the Add requests
// 3. Sends a FEED request.
// 4. Checks the feed response to make sure the items were added in the right order based on timestamp.
// 5. Sends a Done request and waits for the server to exit.
func TestSimpleAddAndFeedRequest(t *testing.T) {
	numOfThreadsStr := "3"
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "run", "twitter.go", numOfThreadsStr)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal("<SimpleAddAndFeedRequest>: error in getting stdout pipe: Contact Professor Samuels, if see this message.")
	}
	stdin, errIn := cmd.StdinPipe()
	if errIn != nil {
		t.Fatal("<SimpleAddAndFeedRequest>: error in getting stdin pipe: Contact Professor Samuels, if see this message.")
	}

	if err := cmd.Start(); err != nil {
		t.Fatal("<SimpleAddAndFeedRequest> cmd.Start error in executing Test: Contact Professor Samuels, if see this message.")
	}
	inDone := make(chan bool)
	outDone := make(chan bool)
	addDone := make(chan bool)

	doneRequest := _TestDoneRequest{"DONE"}
	postInfo := []int{1, 2, 18, 9, 8, 20, 16, 10, 6, 14, 17, 15, 19, 5, 13, 11, 7, 4, 3, 12}
	requestsAdd, responsesAdd, addIdx := createAdds(postInfo, 0)
	order := []int{20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	requestFeed, responseFeedExpected, _ := createFeed(order, addIdx)

	go func() {
		encoder := json.NewEncoder(stdin)
		for idx := 0; idx < len(postInfo); idx++ {
			requestAdd := requestsAdd[idx]
			if err := encoder.Encode(&requestAdd); err != nil {
				t.Fatal("<SimpleAddAndFeedRequest> add cmd.encode error in executing test: Contact Professor Samuels, if see this message.")
			}
		}
		<-addDone
		if err := encoder.Encode(&requestFeed); err != nil {
			t.Fatal("<SimpleAddAndFeedRequest> feed cmd.encode error in executing test: Contact Professor Samuels, if see this message.")
		}
		if err := encoder.Encode(&doneRequest); err != nil {
			t.Fatal("<SimpleAddAndFeedRequest> done cmd.encode errror in executing test: Contact Professor Samuels, if see this message.")
		}
		inDone <- true
	}()

	go func() {
		decoder := json.NewDecoder(stdout)
		var count int
		for {
			var response _TestNormalResponse
			if err := decoder.Decode(&response); err != nil {
				break
			}
			if value, ok := responsesAdd[int(response.Id)]; ok {
				if value.Id != response.Id || value.Success != response.Success {
					t.Errorf("Add request & response id and success fields do not match. Got(Id=%v,Success=%v), Expected(Id=%v,Success=%v)",
						response.Id, response.Success, value.Id, value.Success)
				}
				count++
			} else {
				t.Errorf("Received an invalid id back from twitter.go. We only added ids [0,30] but got(id):%v", response.Id)
			}
			if count == 20 {
				addDone <- true
				var responseFeed _TestFeedResponse
				if err := decoder.Decode(&responseFeed); err != nil {
					break
				}
				if responseFeed.Id != responseFeedExpected.Id {
					t.Errorf("Feed request & response id fields do not match. Got(%v), Expected(%v)",
						responseFeed.Id, responseFeedExpected.Id)
				} else {
					if len(responseFeedExpected.Feed) != len(responseFeed.Feed) {
						t.Errorf("Feed response number of posts not equal to each other. Got(%v), Expected(%v)",
							len(responseFeed.Feed), len(responseFeedExpected.Feed))
					}
					for idx, post := range responseFeed.Feed {
						if post.Body != responseFeedExpected.Feed[idx].Body || post.Timestamp != responseFeedExpected.Feed[idx].Timestamp {
							t.Errorf("Feed response post data does not match. This is checking that the order returned is correct. Got(Body:%v, TimeStamp:%v), Expected(Body:%v, TimeStamp:%v)",
								post.Body, post.Timestamp, responseFeedExpected.Feed[idx].Body, responseFeedExpected.Feed[idx].Timestamp)
						}
					}
					count++
				}
			}
		}
		if count != len(requestsAdd)+1 {
			t.Errorf("Did not receive the right amount of Add&Feed acknowledgements. Got:%v, Expected:%v", count, len(requestsAdd)+1)
		}
		outDone <- true
	}()

	<-inDone
	<-outDone
	if err := cmd.Wait(); err != nil {
		t.Errorf("The automated test timed out. You may have a deadlock, starvation issue and/or you did not implement" +
			" the necessary code for passing this test.")
	}
}

// SimpleRemoveRequest
// Action(s):
// 1. Creates a remove request on an empty feed
// 2. Waits for the remove response to come back and verify its false
// 3. Sends a Done request and waits for the server to exit.
func TestSimpleRemoveRequest(t *testing.T) {

	numOfThreadsStr := "4"

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "run", "twitter.go", numOfThreadsStr)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal("<SimpleRemoveRequest>: error in getting stdout pipe: Contact Professor Samuels, if see this message.")
	}
	stdin, errIn := cmd.StdinPipe()
	if errIn != nil {
		t.Fatal("<SimpleRemoveRequest>: error in getting stdin pipe: Contact Professor Samuels, if see this message.")
	}

	if err := cmd.Start(); err != nil {
		t.Fatal("<SimpleRemoveRequest> cmd.Start error in executing Test: Contact Professor Samuels, if see this message.")
	}
	inDone := make(chan bool)
	outDone := make(chan bool)

	doneRequest := _TestDoneRequest{"DONE"}
	successSlice := make([]bool, 30)
	numbers := generateSlice(30)
	requestsRemoves, responsesRemoves, _ := createRemoves(numbers, successSlice, 0)

	go func() {
		encoder := json.NewEncoder(stdin)
		for idx := 0; idx < len(numbers); idx++ {
			requestRemove := requestsRemoves[idx]
			if err := encoder.Encode(&requestRemove); err != nil {
				t.Fatal("<SimpleRemoveRequest> request cmd.encode error in executing test: Contact Professor Samuels, if see this message.")
			}
			time.Sleep(1 * time.Millisecond) // Wait milisecond before sending the add request
		}
		if err := encoder.Encode(&doneRequest); err != nil {
			t.Fatal("<SimpleRemoveRequest> done error in executing test: Contact Professor Samuels, if see this message.")
		}
		inDone <- true
	}()

	go func() {
		decoder := json.NewDecoder(stdout)
		var count int
		for {
			var response _TestNormalResponse
			if err := decoder.Decode(&response); err != nil {
				break
			}
			if value, ok := responsesRemoves[int(response.Id)]; ok {
				if value.Id != response.Id || value.Success != response.Success {
					t.Errorf("Remove request & response id and success fields do not match. Got(Id=%v,Success=%v), Expected(Id=%v,Success=%v)",
						response.Id, response.Success, value.Id, value.Success)
				}
				count++
			} else {
				t.Errorf("Received an invalid id back from twitter.go. We only added ids [0,30] but got(id):%v", response.Id)
			}
			if count%5 == 0 {
				time.Sleep(1 * time.Second)
			}
		}
		if count != len(responsesRemoves) {
			t.Errorf("Did not receive the right amount of Remove acknowledgements. Got:%v, Expected:%v", count, len(responsesRemoves))
		}
		outDone <- true
	}()

	<-inDone
	<-outDone
	if err := cmd.Wait(); err != nil {
		t.Errorf("The automated test timed out. You may have a deadlock, starvation issue and/or you did not implement" +
			" the necessary code for passing this test.")
	}
}

const evenParity = 1
const oddParity = 0

func getParity(numbers []int, parity int) []int {

	parityNums := make([]int, len(numbers)/2)
	var idx int
	for _, num := range numbers {
		if (parity == evenParity && num%2 == 0) ||
			(parity == oddParity && num%2 != 0) {
			parityNums[idx] = num
			idx++
		}
	}
	sort.Sort(sort.Reverse(sort.IntSlice(parityNums)))
	return parityNums
}
func runAllRequests(threads string, postInfo []int, t *testing.T) {

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "run", "twitter.go", threads)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal("<runTwitter>: error in getting stdout pipe: Contact Professor Samuels, if see this message.")
	}
	stdin, errIn := cmd.StdinPipe()
	if errIn != nil {
		t.Fatal("<runTwitter>: error in getting stdin pipe: Contact Professor Samuels, if see this message.")
	}

	if err := cmd.Start(); err != nil {
		t.Fatal("<cmd.Start> error in executing Test: Contact Professor Samuels, if see this message.")
	}
	inDone := make(chan bool)
	outDone := make(chan bool)

	/***** First Wave: Add all posts and random Contains ***/
	wave1Done := make(chan bool)
	doneRequest := _TestDoneRequest{"DONE"}
	requestsAdd, responsesAdd, addIdx := createAdds(postInfo, 0)
	successSlice := make([]bool, len(postInfo))
	requestsContains, responsesContains, containsIdx := createContains(postInfo, successSlice, addIdx)

	/**** Second Wave: Remove all the even numbers *****/
	evenPosts := getParity(postInfo, evenParity)
	oddPosts := getParity(postInfo, oddParity)
	wave2Done := make(chan bool)
	removePosts := evenPosts
	successSliceRemove := make([]bool, len(removePosts))
	for idx, _ := range removePosts {
		successSliceRemove[idx] = true
	}
	requestsRemoves, responsesRemoves, removeIdx := createRemoves(removePosts, successSliceRemove, containsIdx)

	/**** Third Wave: Check that evens are removed and remove all the odds posts **/
	wave3Done := make(chan bool)
	containsOrder := evenPosts
	successSliceContains := make([]bool, len(containsOrder))
	requestsContains2, responsesContains2, containsIdx2 := createContains(containsOrder, successSliceContains, removeIdx)

	removePosts2 := oddPosts
	successSliceRemove2 := make([]bool, len(removePosts2))
	for idx, _ := range removePosts2 {
		successSliceRemove2[idx] = true
	}
	requestsRemoves2, responsesRemoves2, removeIdx2 := createRemoves(removePosts2, successSliceRemove2, containsIdx2)

	/**** Fourth Wave: check the feed is empty **/
	wave4Done := make(chan bool)
	order2 := []int{}
	requestFeed, responseFeedExpected, _ := createFeed(order2, removeIdx2)

	go func() {
		encoder := json.NewEncoder(stdin)
		for idx := 0; idx < len(postInfo); idx++ {
			requestAdd := requestsAdd[idx]
			if err := encoder.Encode(&requestAdd); err != nil {
				t.Fatal("<AllRequests> add cmd.encode error in executing Test: Contact Professor Samuels, if see this message.")
			}
		}
		for idx := addIdx; idx < (addIdx + len(postInfo)); idx++ {
			requestContains := requestsContains[idx]
			if err := encoder.Encode(&requestContains); err != nil {
				t.Fatal("<AllRequests> request cmd.encode error in executing Test: Contact Professor Samuels, if see this message.")
			}
		}
		<-wave1Done
		for idx := containsIdx; idx < (containsIdx + len(postInfo)/2); idx++ {
			requestRemove := requestsRemoves[idx]
			if err := encoder.Encode(&requestRemove); err != nil {
				t.Fatal("<AllRequests> remove cmd.encode error in executing Test: Contact Professor Samuels, if see this message.")
			}
		}
		<-wave2Done

		for idx := removeIdx; idx < (removeIdx + len(postInfo)/2); idx++ {
			requestContains := requestsContains2[idx]
			if err := encoder.Encode(&requestContains); err != nil {
				t.Fatal("<AllRequests> contains cmd.encode error in executing Test: Contact Professor Samuels, if see this message.")
			}
		}
		for idx := containsIdx2; idx < (containsIdx2 + len(postInfo)/2); idx++ {
			requestRemove := requestsRemoves2[idx]
			if err := encoder.Encode(&requestRemove); err != nil {
				t.Fatal("<AllRequests> remove cmd.encode error in executing Test: Contact Professor Samuels, if see this message.")
			}
		}
		<-wave3Done
		if err := encoder.Encode(&requestFeed); err != nil {
			t.Fatal("<AllRequests> request cmd.encode error in executing Test: Contact Professor Samuels, if see this message.")
		}
		<-wave4Done
		if err := encoder.Encode(&doneRequest); err != nil {
			t.Fatal("<AllRequests> done cmd.encode error in executing Test: Contact Professor Samuels, if see this message.")
		}
		inDone <- true
	}()

	go func() {
		decoder := json.NewDecoder(stdout)
		var count int
		for {
			var response _TestNormalResponse
			if count < ((len(postInfo) * 2) + ((len(postInfo) / 2) * 3)) {
				if err := decoder.Decode(&response); err != nil {
					break
				}
			}
			if count >= 0 && count < (len(postInfo)*2) {
				if value, ok := responsesAdd[int(response.Id)]; ok {
					if value.Id != response.Id || value.Success != response.Success {
						t.Errorf("Add Request & Response id and success fields do not match. Got(Id=%v,Success=%v), Expected(Id=%v,Success=%v)",
							response.Id, response.Success, value.Id, value.Success)
					}
					count++
				} else if value, ok := responsesContains[int(response.Id)]; ok {
					if value.Id != response.Id {
						t.Errorf("Contains Request & Response id fields do not match. Got(%v), Expected(%v)",
							response.Id, value.Id)
					}
					count++
				} else {
					t.Errorf("Received an invalid id back from twitter.go. We only added ids [0,%v) but got(id):%v", len(postInfo), response.Id)
				}
				if count == (len(postInfo) * 2) {
					wave1Done <- true
				}
			} else if count >= (len(postInfo)*2) && count < ((len(postInfo)*2)+(len(postInfo)/2)) {
				if value, ok := responsesRemoves[int(response.Id)]; ok {
					if value.Id != response.Id || value.Success != response.Success {
						t.Errorf("Remove Request & Response id and success fields do not match. Got(Id=%v,Success=%v), Expected(Id=%v,Success=%v)",
							response.Id, response.Success, value.Id, value.Success)
					}
					count++
				} else {
					t.Errorf("Received an invalid id back from twitter.go. We only removed the even ids but got(id):%v", response.Id)
				}
				if count == ((len(postInfo) * 2) + (len(postInfo) / 2)) {
					wave2Done <- true
				}
			} else if count >= (len(postInfo)*2+(len(postInfo)/2)) && count < (len(postInfo)*2+(len(postInfo)/2)*3) {
				if value, ok := responsesContains2[int(response.Id)]; ok {
					if value.Id != response.Id || value.Success != response.Success {
						t.Errorf("Contains Request & Response id and success fields do not match. Got(Id=%v,Success=%v), Expected(Id=%v,Success=%v)",
							response.Id, response.Success, value.Id, value.Success)
					}
					count++
				} else if value, ok := responsesRemoves2[int(response.Id)]; ok {
					if value.Id != response.Id || value.Success != response.Success {
						t.Errorf("Remove Request & Response id and success fields do not match. Got(Id=%v,Success=%v), Expected(Id=%v,Success=%v)",
							response.Id, response.Success, value.Id, value.Success)
					}
					count++
				} else {
					t.Errorf("Received an invalid id back from twitter.go. We only removed the odd ids but got(id):%v", response.Id)
				}
				if count == (len(postInfo)*2 + (len(postInfo)/2)*3) {
					wave3Done <- true
				}
			} else {
				var responseFeed _TestFeedResponse
				if err := decoder.Decode(&responseFeed); err != nil {
					fmt.Printf("Got an Error\n")
					break
				}
				if responseFeed.Id != responseFeedExpected.Id {
					t.Errorf("Feed Request & Response id fields do not match. Got(%v), Expected(%v)",
						responseFeed.Id, responseFeedExpected.Id)
				} else {
					if len(responseFeedExpected.Feed) != len(responseFeed.Feed) {
						t.Errorf("Feed Response number of posts not equal to each other. Got(%v), Expected(%v)",
							len(responseFeed.Feed), len(responseFeedExpected.Feed))
					}
					count++
				}
				wave4Done <- true
				break
			}
		}
		if count != (len(postInfo)*2+(len(postInfo)/2)*3)+1 {
			t.Errorf("Did not receive the right amount of Add&Feed acknowledgements. Got:%v, Expected:%v", count, (len(postInfo)*2+(len(postInfo)/2)*3)+1)
		}
		outDone <- true
	}()
	<-inDone
	<-outDone
	if err := cmd.Wait(); err != nil {
		t.Errorf("The automated test timed out. You may have a deadlock, starvation issue and/or you did not implement" +
			" the necessary code for passing this test.")
	}
}

// AllRequestsXtraSmall
// Action(s):
// 1. Sends 20 Add requests into the feed
// 2. Sends Contains requests right after sending the add requests.
// 3. Check the responses back from the contains request. We don't care about the
// Contains requests but rather just ensuring that they get returned by the twitter.go.
// 4. Sends Remove requests by removing only the odd ids
// 5. Checks to make sure the evens are still there and all the odds are gone by sending contains requests
// 6. Sends a Done request and waits for the server to exit.
func TestAllRequestsXtraSmall(t *testing.T) {
	threads := "40"
	posts := generateSlice(20)
	rand.Shuffle(len(posts), func(i, j int) { posts[i], posts[j] = posts[j], posts[i] })
	runAllRequests(threads, posts, t)
}

// AllRequestsSmall
// Action(s):
// 1. Sends 100 Add requests into the feed
// 2. Sends Contains requests right after sending the add requests.
// 3. Check the responses back from the contains request. We don't care about the
// Contains requests but rather just ensuring that they get returned by the twitter.go.
// 4. Sends Remove requests by removing only the odd ids
// 5. Checks to make sure the evens are still there and all the odds are gone by sending contains requests
// 6. Sends a Done request and waits for the server to exit.
func TestAllRequestsSmall(t *testing.T) {
	threads := "48"
	posts := generateSlice(100)
	rand.Shuffle(len(posts), func(i, j int) { posts[i], posts[j] = posts[j], posts[i] })
	runAllRequests(threads, posts, t)
}

// AllRequestsMedium
// Action(s):
// 1. Sends 10000 Add requests into the feed
// 2. Sends Contains requests right after sending the add requests.
// 3. Check the responses back from the contains request. We don't care about the
// Contains requests but rather just ensuring that they get returned by the twitter.go.
// 4. Sends Remove requests by removing only the odd ids
// 5. Checks to make sure the evens are still there and all the odds are gone by sending contains requests
// 6. Sends a Done request and waits for the server to exit.
func TestAllRequestsMedium(t *testing.T) {
	threads := "16"
	posts := generateSlice(10000)
	rand.Shuffle(len(posts), func(i, j int) { posts[i], posts[j] = posts[j], posts[i] })
	runAllRequests(threads, posts, t)
}

// AllRequestsLarge
// Action(s):
// 1. Sends 10000 Add requests into the feed
// 2. Sends Contains requests right after sending the add requests.
// 3. Check the responses back from the contains request. We don't care about the
// Contains requests but rather just ensuring that they get returned by the twitter.go.
// 4. Sends Remove requests by removing only the odd ids
// 5. Checks to make sure the evens are still there and all the odds are gone by sending contains requests
// 6. Sends a Done request and waits for the server to exit.
func TestAllRequestsLarge(t *testing.T) {
	threads := "32"
	posts := generateSlice(25000)
	rand.Shuffle(len(posts), func(i, j int) { posts[i], posts[j] = posts[j], posts[i] })
	runAllRequests(threads, posts, t)
}

// AllRequestsXtraLarge
// Action(s):
// 1. Sends 10000 Add requests into the feed
// 2. Sends Contains requests right after sending the add requests.
// 3. Check the responses back from the contains request. We don't care about the
// Contains requests but rather just ensuring that they get returned by the twitter.go.
// 4. Sends Remove requests by removing only the odd ids
// 5. Checks to make sure the evens are still there and all the odds are gone by sending contains requests
// 6. Sends a Done request and waits for the server to exit.
func TestAllRequestsXtraLarge(t *testing.T) {
	threads := "32"
	posts := generateSlice(75000)
	rand.Shuffle(len(posts), func(i, j int) { posts[i], posts[j] = posts[j], posts[i] })
	runAllRequests(threads, posts, t)
}
