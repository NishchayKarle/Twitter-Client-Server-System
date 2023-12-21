package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"time"
)

const usage = "Usage: benchmark version testSize threads\n" +
	" version =  (p) - parallel version, (s) sequential version \n" +
	" testSize = the test size \n" +
	"\t xsmall = Run the extra small test size\n" +
	"\t small = Run the small test size\n" +
	"\t medium = Run the  medium test size\n" +
	"\t large = Run the large test size\n" +
	"\t xlarge = Run the extra large test size\n" +
	" threads (required for  p version only) = the number of threads to pass to twitter.go\n"

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
func runAllRequests(threads, version string, postInfo []int) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	var cmd *exec.Cmd

	if version == "p" {
		cmd = exec.CommandContext(ctx, "go", "run", "proj1/twitter", threads)
	} else {
		cmd = exec.CommandContext(ctx, "go", "run", "proj1/twitter")
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Errorf("<runTwitter>: error in getting stdout pipe: Contact Professor Samuels, if see this message.")
		os.Exit(1)
	}
	stdin, errIn := cmd.StdinPipe()
	if errIn != nil {
		fmt.Errorf("<runTwitter>: error in getting stdin pipe: Contact Professor Samuels, if see this message.")
		os.Exit(1)
	}

	if err := cmd.Start(); err != nil {
		fmt.Errorf("<cmd.Start> error in executing Test: Contact Professor Samuels, if see this message.")
		os.Exit(1)
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
				fmt.Errorf("<AllRequests> add cmd.encode error in executing Test: Contact Professor Samuels, if see this message.")
				os.Exit(1)
			}
		}
		for idx := addIdx; idx < (addIdx + len(postInfo)); idx++ {
			requestContains := requestsContains[idx]
			if err := encoder.Encode(&requestContains); err != nil {
				fmt.Errorf("<AllRequests> request cmd.encode error in executing Test: Contact Professor Samuels, if see this message.")
				os.Exit(1)
			}
		}
		<-wave1Done
		for idx := containsIdx; idx < (containsIdx + len(postInfo)/2); idx++ {
			requestRemove := requestsRemoves[idx]
			if err := encoder.Encode(&requestRemove); err != nil {
				fmt.Errorf("<AllRequests> remove cmd.encode error in executing Test: Contact Professor Samuels, if see this message.")
				os.Exit(1)
			}
		}
		<-wave2Done

		for idx := removeIdx; idx < (removeIdx + len(postInfo)/2); idx++ {
			requestContains := requestsContains2[idx]
			if err := encoder.Encode(&requestContains); err != nil {
				fmt.Errorf("<AllRequests> contains cmd.encode error in executing Test: Contact Professor Samuels, if see this message.")
				os.Exit(1)
			}
		}
		for idx := containsIdx2; idx < (containsIdx2 + len(postInfo)/2); idx++ {
			requestRemove := requestsRemoves2[idx]
			if err := encoder.Encode(&requestRemove); err != nil {
				fmt.Errorf("<AllRequests> remove cmd.encode error in executing Test: Contact Professor Samuels, if see this message.")
				os.Exit(1)
			}
		}
		<-wave3Done
		if err := encoder.Encode(&requestFeed); err != nil {
			fmt.Errorf("<AllRequests> request cmd.encode error in executing Test: Contact Professor Samuels, if see this message.")
			os.Exit(1)
		}
		<-wave4Done
		if err := encoder.Encode(&doneRequest); err != nil {
			fmt.Errorf("<AllRequests> done cmd.encode error in executing Test: Contact Professor Samuels, if see this message.")
			os.Exit(1)
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
						fmt.Errorf("Add Request & Response id and success fields do not match. Got(Id=%v,Success=%v), Expected(Id=%v,Success=%v)",
							response.Id, response.Success, value.Id, value.Success)
						os.Exit(1)
					}
					count++
				} else if value, ok := responsesContains[int(response.Id)]; ok {
					if value.Id != response.Id {
						fmt.Errorf("Contains Request & Response id fields do not match. Got(%v), Expected(%v)",
							response.Id, value.Id)
						os.Exit(1)
					}
					count++
				} else {
					fmt.Errorf("Received an invalid id back from twitter.go. We only added ids [0,%v) but got(id):%v", len(postInfo), response.Id)
					os.Exit(1)
				}
				if count == (len(postInfo) * 2) {
					wave1Done <- true
				}
			} else if count >= (len(postInfo)*2) && count < ((len(postInfo)*2)+(len(postInfo)/2)) {
				if value, ok := responsesRemoves[int(response.Id)]; ok {
					if value.Id != response.Id || value.Success != response.Success {
						fmt.Errorf("Remove Request & Response id and success fields do not match. Got(Id=%v,Success=%v), Expected(Id=%v,Success=%v)",
							response.Id, response.Success, value.Id, value.Success)
						os.Exit(1)
					}
					count++
				} else {
					fmt.Errorf("Received an invalid id back from twitter.go. We only removed the even ids but got(id):%v", response.Id)
					os.Exit(1)
				}
				if count == ((len(postInfo) * 2) + (len(postInfo) / 2)) {
					wave2Done <- true
				}
			} else if count >= (len(postInfo)*2+(len(postInfo)/2)) && count < (len(postInfo)*2+(len(postInfo)/2)*3) {
				if value, ok := responsesContains2[int(response.Id)]; ok {
					if value.Id != response.Id || value.Success != response.Success {
						fmt.Errorf("Contains Request & Response id and success fields do not match. Got(Id=%v,Success=%v), Expected(Id=%v,Success=%v)",
							response.Id, response.Success, value.Id, value.Success)
						os.Exit(1)
					}
					count++
				} else if value, ok := responsesRemoves2[int(response.Id)]; ok {
					if value.Id != response.Id || value.Success != response.Success {
						fmt.Errorf("Remove Request & Response id and success fields do not match. Got(Id=%v,Success=%v), Expected(Id=%v,Success=%v)",
							response.Id, response.Success, value.Id, value.Success)
						os.Exit(1)
					}
					count++
				} else {
					fmt.Errorf("Received an invalid id back from twitter.go. We only removed the odd ids but got(id):%v", response.Id)
					os.Exit(1)
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
					fmt.Errorf("Feed Request & Response id fields do not match. Got(%v), Expected(%v)",
						responseFeed.Id, responseFeedExpected.Id)
					os.Exit(1)
				} else {
					if len(responseFeedExpected.Feed) != len(responseFeed.Feed) {
						fmt.Errorf("Feed Response number of posts not equal to each other. Got(%v), Expected(%v)",
							len(responseFeed.Feed), len(responseFeedExpected.Feed))
					}
					count++
				}
				wave4Done <- true
				break
			}
		}
		if count != (len(postInfo)*2+(len(postInfo)/2)*3)+1 {
			fmt.Errorf("Did not receive the right amount of Add&Feed acknowledgements. Got:%v, Expected:%v", count, (len(postInfo)*2+(len(postInfo)/2)*3)+1)
			os.Exit(1)
		}
		outDone <- true
	}()
	<-inDone
	<-outDone
	if err := cmd.Wait(); err != nil {
		fmt.Errorf("The automated test timed out. You may have a deadlock, starvation issue and/or you did not implement" +
			" the necessary code for passing this test.")
		os.Exit(1)
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
func AllRequestsXtraSmall(threads string, version string) {
	posts := generateSlice(20)
	rand.Shuffle(len(posts), func(i, j int) { posts[i], posts[j] = posts[j], posts[i] })
	runAllRequests(threads, version, posts)
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
func AllRequestsSmall(threads string, version string) {
	posts := generateSlice(100)
	rand.Shuffle(len(posts), func(i, j int) { posts[i], posts[j] = posts[j], posts[i] })
	runAllRequests(threads,version, posts)
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
func AllRequestsMedium(threads string, version string) {
	posts := generateSlice(10000)
	rand.Shuffle(len(posts), func(i, j int) { posts[i], posts[j] = posts[j], posts[i] })
	runAllRequests(threads, version, posts)
}

// AllRequestsLarge
// Action(s):
// 1. Sends 25000 Add requests into the feed
// 2. Sends Contains requests right after sending the add requests.
// 3. Check the responses back from the contains request. We don't care about the
// Contains requests but rather just ensuring that they get returned by the twitter.go.
// 4. Sends Remove requests by removing only the odd ids
// 5. Checks to make sure the evens are still there and all the odds are gone by sending contains requests
// 6. Sends a Done request and waits for the server to exit.
func AllRequestsLarge(threads string, version string) {
	posts := generateSlice(25000)
	rand.Shuffle(len(posts), func(i, j int) { posts[i], posts[j] = posts[j], posts[i] })
	runAllRequests(threads, version, posts)
}

// AllRequestsXtraLarge
// Action(s):
// 1. Sends 75000 Add requests into the feed
// 2. Sends Contains requests right after sending the add requests.
// 3. Check the responses back from the contains request. We don't care about the
// Contains requests but rather just ensuring that they get returned by the twitter.go.
// 4. Sends Remove requests by removing only the odd ids
// 5. Checks to make sure the evens are still there and all the odds are gone by sending contains requests
// 6. Sends a Done request and waits for the server to exit.
func AllRequestsXtraLarge(threads string, version string) {
	posts := generateSlice(75000)
	rand.Shuffle(len(posts), func(i, j int) { posts[i], posts[j] = posts[j], posts[i] })
	runAllRequests(threads, version, posts)
}

func main() {

	if len(os.Args) < 3 {
		fmt.Println(usage)
	} else {
		version := os.Args[1]
		test := os.Args[2]
		var threads string
		if version == "p" {
			threads = os.Args[3]
		}

		start := time.Now()

		if test == "xsmall" {
			AllRequestsXtraSmall(threads,  version)
		} else if test == "small" {
			AllRequestsSmall(threads, version)
		} else if test == "medium" {
			AllRequestsMedium(threads,  version)
		} else if test == "large" {
			AllRequestsLarge(threads,  version)
		} else if test == "xlarge" {
			AllRequestsXtraLarge(threads, version)
		} else {
			fmt.Printf("Invalid argument:%v", test)
			fmt.Println(usage)
		}
		fmt.Printf("%.2f\n", time.Since(start).Seconds())
	}

}
