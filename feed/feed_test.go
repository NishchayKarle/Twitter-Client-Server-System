package feed

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"
)

func addGoroutine(amount int, feed Feed, localCount int, wg *sync.WaitGroup) {
	for i := 0; i < localCount; i++ {
		num := amount + i
		body := strconv.Itoa(num)
		feed.Add(body, float64(num))
	}
	wg.Done()
}
func addGoroutine2(shouldAddEven bool, amount int, feed Feed, localCount int, wg *sync.WaitGroup) {
	for i := 0; i < localCount; i++ {
		num := amount + i
		body := strconv.Itoa(num)
		if shouldAddEven && num%2 == 0 {
			feed.Add(body, float64(num))
		} else if !shouldAddEven && num%2 != 0 {
			feed.Add(body, float64(num))
		}
	}
	wg.Done()
}

func removeGoroutine(t *testing.T, amount int, feed Feed, localCount int, wg *sync.WaitGroup) {
	//First try to remove all the even numbers from the local block of work
	for i := 0; i < localCount; i++ {
		num := amount + i
		if num%2 == 0 {
			if !feed.Remove(float64(num)) {
				t.Errorf("FAILED: Feed should contain timestamp (%v) but did not\n", i)
			}
		}
	}
	//Second try to remove all the odd numbers from the local block of work
	for i := 0; i < localCount; i++ {
		num := amount + i
		if num%2 != 0 {
			if !feed.Remove(float64(num)) {
				t.Errorf("FAILED: Feed should contain timestamp (%v) but did not\n", i)
			}
		}
	}
	wg.Done()
}
func containsGoroutine(t *testing.T, amount int, feed Feed, localCount int, wg *sync.WaitGroup) {

	for i := 0; i < localCount; i++ {
		num := amount + i
		if num%2 == 0 && feed.Contains(float64(i)) {
			t.Errorf("FAILED: Feed should not contain timestamp (%v)\n", i)
		}
	}
	wg.Done()
}
func removeGoroutine2(shouldRemoveEven bool, t *testing.T, amount int, feed Feed, localCount int, wg *sync.WaitGroup) {

	for i := 0; i < localCount; i++ {
		num := amount + i
		if shouldRemoveEven && num%2 == 0 {
			if !feed.Remove(float64(num)) {
				t.Errorf("FAILED: Feed should contain timestamp (%v) but did not\n", i)
			}
		} else if !shouldRemoveEven && num%2 != 0 {
			if !feed.Remove(float64(num)) {
				t.Errorf("FAILED: Feed should contain timestamp (%v) but did not\n", i)
			}
		}
	}
	wg.Done()
}
func randomReads(feed Feed, localCount int, wg *sync.WaitGroup) {
	for i := 0; i < localCount; i++ {
		feed.Contains(float64(i))
	}
	wg.Done()
}

func TestSimpleSeq(t *testing.T) {

	feed := NewFeed()

	//Check to make sure Contains returns False on empty feed
	for i := 1; i <= rand.Intn(100); i++ {
		if feed.Contains(float64(i)) {
			t.Errorf("Feed is empty. Contains = %v", i)
		}
	}

	//Check to make sure Remove returns False on empty feed
	for i := 1; i <= rand.Intn(100); i++ {
		if feed.Remove(float64(i)) {
			t.Errorf("Feed is empty but removed = %v", i)
		}
	}

	num := 2
	body := strconv.Itoa(num)
	feed.Add(body, float64(num))

	//Check to make sure contains returns True for 2
	if !feed.Contains(float64(num)) {
		t.Errorf("Feed is empty. Found = %v", num)
	}
}
func TestAdd(t *testing.T) {

	postInfo := [20]int{1, 2, 18, 9, 8, 20, 16, 10, 6, 14, 17, 15, 19, 5, 13, 11, 7, 4, 3, 12}
	feed := NewFeed()

	//Add 20 posts to the feed
	for _, num := range postInfo {
		body := strconv.Itoa(num)
		feed.Add(body, float64(num))
	}
	//Order of the Timestamps
	order := []float64{20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	//Check to make sure the order is correct for the feed
	for i := 0; i < len(order); i++ {
		if !feed.Contains(order[i]) {
			t.Errorf("Added:%v but did not find it in feed.", order[i])
		}
	}

}
func TestContains(t *testing.T) {

	postInfo := [20]int{1, 2, 18, 9, 8, 20, 16, 10, 6, 14, 17, 15, 19, 5, 13, 11, 7, 4, 3, 12}
	feed := NewFeed()

	//Add 20 posts to the feed
	for _, num := range postInfo {
		body := strconv.Itoa(num)
		feed.Add(body, float64(num))
	}
	//Check to make sure all the numbers are inside the feed
	for i := 1; i <= 20; i++ {
		if !feed.Contains(float64(i)) {
			t.Errorf("Missing Posts after  post:%d", i)
		}
	}
	//Order of the timestamps
	order := []float64{20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	for i := 0; i < len(order); i++ {
		if !feed.Contains(order[i]) {
			t.Errorf("Added:%v but did not find it in feed.", order[i])
		}
	}
}
func TestRemove(t *testing.T) {

	postInfo := [20]int{1, 2, 18, 9, 8, 20, 16, 10, 6, 14, 17, 15, 19, 5, 13, 11, 7, 4, 3, 12}
	feed := NewFeed()

	//Add 20 posts to the feed
	for _, num := range postInfo {
		body := strconv.Itoa(num)
		feed.Add(body, float64(num))
	}
	//Remove the even posts
	for i := 1; i <= 20; i++ {
		if i%2 == 0 {
			if !feed.Remove(float64(i)) {
				t.Errorf("Tried to remove even timestamp:%v but it was not found", i)
			}
		}
	}
	//Order of the timestamps
	//order := []float64{19, 17, 15, 13, 11, 9, 7, 5, 3, 1}

	//Check to make sure the order is correct for the feed after removing evens
	for i := 0; i < len(postInfo); i++ {
		if postInfo[i]%2 == 0 && feed.Contains(float64(postInfo[i])) {
			t.Errorf("Removed:%v, from the list but it's still there.", postInfo[i])
		} else if postInfo[i]%2 != 0 {
			if !feed.Contains(float64(postInfo[i])) {
				t.Errorf("Added:%v but did not find it in feed.", postInfo[i])
			}
		}
	}

	//Remove the odd posts
	for i := 1; i <= 20; i++ {
		if i%2 != 0 {
			if !feed.Remove(float64(i)) {
				t.Errorf("Tried to remove odd timestamp:%v but it was not found", i)
			}
		}
	}
	//order = []float64{}
	//Check to make sure that nothing is inside the feed after removing everything
	for i := 0; i < len(postInfo); i++ {
		if feed.Remove(float64(postInfo[i])) || feed.Contains(float64(postInfo[i])) {
			t.Errorf("Removed all items but not all were removed:\n"+"(Got):%v\n", i)
		}
	}
}
func TestParallelAdd(t *testing.T) {

	const totalSize = 5000
	const threadCount = 100
	const localCount = totalSize / threadCount
	feed := NewFeed()

	var wg sync.WaitGroup

	//Have each goroutine spawned go add in a few posts
	for i := 0; i < threadCount; i++ {
		wg.Add(1)
		go addGoroutine(i*localCount, feed, localCount, &wg)
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go randomReads(feed, totalSize, &wg) //Throw in some readers while adding
		}
		
	}
	wg.Wait()
	
		//Verify that all 1000 posts are contained in the feed
		for i := 0; i < totalSize; i++ {
			if !feed.Contains(float64(i)) {
				t.Errorf("FAILED: Feed should contain timestamp (%v)\n", i)
			}
		}

		//Verify that you can remove all 1000 posts
		for i := 0; i < totalSize; i++ {
			if !feed.Remove(float64(i)) {
				t.Errorf("FAILED: Did not remove (%v)\n", i)
			}
		}
}
func TestParallelRemoveAndAdd(t *testing.T) {

	const totalSize = 5000
	const threadCount = 100
	const localCount = totalSize / threadCount
	feed := NewFeed()

	//Sequentially add in all the posts
	for i := 0; i < totalSize; i++ {
		body := strconv.Itoa(i)
		feed.Add(body, float64(i))
	}
	var wg sync.WaitGroup
	// Now remove all the posts
	for i := 0; i < threadCount; i++ {
		wg.Add(1)
		go removeGoroutine(t, i*localCount, feed, localCount, &wg)
		for i := 0; i < 15; i++ {
			wg.Add(1)
			go randomReads(feed, totalSize, &wg) //Throw in some readers while removing
		}
	}
	wg.Wait()
	// Check to make sure feed does not contain any of the posts added (checking contains)
	for i := 0; i < totalSize; i++ {
		if feed.Contains(float64(i)) {
			t.Errorf("FAILED: Feed should not contain timestamp (%v)\n", i)
		}
	}
	//Check to make sure that nothing is inside the feed after removing everything
	for i := 0; i < threadCount; i++ {
		if feed.Remove(float64(i)) || feed.Contains(float64(i)) {
			t.Errorf("Removed all items but not all were removed:\n"+"(Got):%v\n", i)
		}
	}
}
func TestParallelAll(t *testing.T) {

	const totalSize = 5000
	const threadCount = 50
	const localCount = totalSize / threadCount
	feed := NewFeed()

	//First: add the even timestamps
	var wg sync.WaitGroup
	for i := 0; i < threadCount; i++ {
		wg.Add(1)
		go addGoroutine2(true, i*localCount, feed, localCount, &wg)
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go randomReads(feed, totalSize, &wg) //Through in some readers while adding
		}
	}
	wg.Wait()

	//Second: add the odd timestamps but also remove even timestamps
	for i := 0; i < threadCount; i++ {
		wg.Add(2)
		go addGoroutine2(false, i*localCount, feed, localCount, &wg)
		go removeGoroutine2(true, t, i*localCount, feed, localCount, &wg)
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go randomReads(feed, totalSize, &wg) //Throw in some readers while adding
		}
	}
	wg.Wait()
	//Third: check to make sure evens were removed and remove odds
	for i := 0; i < threadCount; i++ {
		wg.Add(2)
		go removeGoroutine2(false, t, i*localCount, feed, localCount, &wg)
		go containsGoroutine(t, i*localCount, feed, localCount, &wg)
		for i := 0; i < 15; i++ {
			wg.Add(1)
			go randomReads(feed, totalSize, &wg) //Throw in some readers while removing
		}
	}
	wg.Wait()
	//Fourth Check to make sure that nothing is inside the feed after removing everything
	for i := 0; i < threadCount; i++ {
		if feed.Remove(float64(i)) || feed.Contains(float64(i)) {
			t.Errorf("Removed all items but not all were removed:\n"+"(Got):%v\n", i)
		}
	}
}