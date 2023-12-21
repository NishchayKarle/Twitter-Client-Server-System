package feed

//Feed represents a user's twitter feed
// You will add to this interface the implementations as you complete them.

import (
	"proj1/lock"
)

type Feed interface {
	Add(body string, timestamp float64)
	Remove(timestamp float64) bool
	Contains(timestamp float64) bool
	Length() int
	GetFeed() []*JsonPost
}

// You CAN add to this structure but you cannot remove any of the original fields. You must use
// the original fields in your implementation. You can assume the feed will not have duplicate posts
type feed struct {
	start  *post // a pointer to the beginning post
	length int
	lock   *lock.MyRWLock
}

//post is the internal representation of a post on a user's twitter feed (hidden from outside packages)
// You CAN add to this structure but you cannot remove any of the original fields. You must use
// the original fields in your implementation.
type post struct {
	body      string  // the text of the post
	timestamp float64 // Unix timestamp of the post
	next      *post   // the next post in the feed
}

type JsonPost struct {
	Body      string  `json:"body"`
	Timestamp float64 `json:"timestamp"`
}

//NewPost creates and returns a new post value given its body and timestamp
func newPost(body string, timestamp float64, next *post) *post {
	return &post{body, timestamp, next}
}

//NewFeed creates a empy user feed
func NewFeed() Feed {
	return &feed{start: nil, length: 0, lock: lock.NewMyRWLock()}
}

// Add inserts a new post to the feed. The feed is always ordered by the timestamp where
// the most recent timestamp is at the beginning of the feed followed by the second most
// recent timestamp, etc. You may need to insert a new post somewhere in the feed because
// the given timestamp may not be the most recent.
func (f *feed) Add(body string, timestamp float64) {
	f.lock.Lock()
	f.length++
	post := newPost(body, timestamp, nil)
	if f.start == nil {
		f.start = post
		f.lock.Unlock()
		return
	}

	if post.timestamp >= f.start.timestamp {
		post.next = f.start
		f.start = post
		f.lock.Unlock()
		return
	}

	prev := f.start
	curr := prev.next

	for curr != nil && curr.timestamp > post.timestamp {
		prev = curr
		curr = curr.next
	}
	post.next = curr
	prev.next = post
	f.lock.Unlock()
}

// Remove deletes the post with the given timestamp. If the timestamp
// is not included in a post of the feed then the feed remains
// unchanged. Return true if the deletion was a success, otherwise return false
func (f *feed) Remove(timestamp float64) bool {
	f.lock.Lock()
	if f.start == nil {
		f.lock.Unlock()
		return false
	}

	if f.start.timestamp == timestamp {
		f.start = f.start.next
		f.length--
		f.lock.Unlock()
		return true
	}

	prev := f.start
	curr := prev.next

	for curr != nil {
		if curr.timestamp == timestamp {
			prev.next = curr.next
			f.length--
			f.lock.Unlock()
			return true
		}
		prev = curr
		curr = curr.next
	}
	f.lock.Unlock()
	return false
}

// Contains determines whether a post with the given timestamp is
// inside a feed. The function returns true if there is a post
// with the timestamp, otherwise, false.
func (f *feed) Contains(timestamp float64) bool {
	f.lock.RLock()
	curr := f.start

	for curr != nil {
		if curr.timestamp == timestamp {
			f.lock.RUnlock()
			return true
		}
		curr = curr.next
	}
	f.lock.RUnlock()
	return false
}

func (f *feed) Length() int {
	return f.length
}

func (f *feed) GetFeed() []*JsonPost {
	var userFeed []*JsonPost
	f.lock.RLock()
	head := f.start
	for i := 0; i < f.Length(); i++ {
		userFeed = append(userFeed, &JsonPost{head.body, head.timestamp})
		head = head.next
	}
	f.lock.RUnlock()
	return userFeed
}
