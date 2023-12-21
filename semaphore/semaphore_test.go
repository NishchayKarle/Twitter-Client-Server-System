package semaphore

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func test(counter *int64, wg *sync.WaitGroup, sema *Semaphore) {

	sema.Down()

	for i := 0; i < 1e6; i++ {
		*counter = *counter + 1
	}
	sema.Up()
	wg.Done()
}
func runTest(threads int, t *testing.T) {

	var counter int64
	var wg sync.WaitGroup

	sema := NewSemaphore(1)

	for goID := 0; goID < threads; goID++ {
		wg.Add(1)
		go test(&counter, &wg, sema)
	}
	wg.Wait()
	if counter != int64(threads*1e6) {
		t.Errorf("Expected = %v Got = %v", int64(threads*1e6), counter)
	}
}
func Test1(t *testing.T) {

	var tests = []struct {
		threads int
	}{
		{2},
		{3},
		{5},
		{8},
	}
	for num, test := range tests {
		testname := fmt.Sprintf("T=%v", num)
		t.Run(testname, func(t *testing.T) {
			runTest(test.threads, t)
		})
	}
}
func worker(sema *Semaphore, group *sync.WaitGroup) {

	for i := 0; i < 10; i++ {
		sema.Down()
		time.Sleep(time.Nanosecond * 2)
		sema.Up()
	}
	group.Done()
}
func Test2(t *testing.T) {

	sema := NewSemaphore(5)
	threads := 200
	var group sync.WaitGroup
	for i := 0; i < threads; i++ {
		group.Add(1)
		go worker(sema, &group)
	}
	group.Wait()
}
