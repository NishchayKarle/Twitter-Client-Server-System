package queue

import (
	"sync/atomic"
	"unsafe"
)

type Request struct {
	Command   string  `json:"command"`
	Id        int     `json:"id"`
	Body      string  `json:"body"`
	Timestamp float64 `json:"timestamp"`
}

type QueueNode struct {
	task *Request
	next unsafe.Pointer
}

type LockFreeQueue struct {
	head unsafe.Pointer
	tail unsafe.Pointer
}

func NewLockFreeQueue() *LockFreeQueue {
	queueNode := unsafe.Pointer(&QueueNode{})
	return &LockFreeQueue{head: queueNode, tail: queueNode}
}

func (lockFreeQueue *LockFreeQueue) Enqueue(task *Request) {
	queueNode := &QueueNode{task: task}
	for {
		tail := Load(&lockFreeQueue.tail)
		next := Load(&tail.next)
		if tail == Load(&lockFreeQueue.tail) {
			if next == nil {
				if CompareAndSwap(&tail.next, next, queueNode) {
					CompareAndSwap(&lockFreeQueue.tail, tail, queueNode)
					return
				}
			} else {
				CompareAndSwap(&lockFreeQueue.tail, tail, next)
			}
		}
	}
}

func (lockFreeQueue *LockFreeQueue) Dequeue() *Request {
	for {
		head := Load(&lockFreeQueue.head)
		tail := Load(&lockFreeQueue.tail)
		next := Load(&head.next)
		if head == Load(&lockFreeQueue.head) {
			if head == tail {
				if next == nil {
					return nil
				}
				CompareAndSwap(&lockFreeQueue.tail, tail, next)
			} else {
				task := next.task
				if CompareAndSwap(&lockFreeQueue.head, head, next) {
					return task
				}
			}
		}
	}
}

func (lockFreeQueue *LockFreeQueue) Empty() bool {
	head := Load(&lockFreeQueue.head)
	tail := Load(&lockFreeQueue.tail)
	next := Load(&head.next)
	if head == Load(&lockFreeQueue.head) {
		if head == tail {
			if next == nil {
				return true
			}
		}
	}
	return true
}

func Load(pointer *unsafe.Pointer) *QueueNode {
	return (*QueueNode)(atomic.LoadPointer(pointer))
}

func CompareAndSwap(pointer *unsafe.Pointer, old *QueueNode, new *QueueNode) bool {
	return atomic.CompareAndSwapPointer(
		pointer, unsafe.Pointer(old), unsafe.Pointer(new))
}
