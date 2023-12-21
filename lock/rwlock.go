// Package lock provides an implementation of a read-write lock
// that uses condition variables and mutexes.
package lock

import "sync"

type RWLock interface {
	RLock()
	RUnlock()
	Lock()
	Unlock()
}

type MyRWLock struct {
	numReaders        int
	conditionVariable *sync.Cond
}

func NewMyRWLock() *MyRWLock {
	return &MyRWLock{0, sync.NewCond(new(sync.Mutex))}
}

func (lock *MyRWLock) RLock() {
	lock.conditionVariable.L.Lock()
	if lock.numReaders == 32 {
		lock.conditionVariable.Wait()
	}
	lock.numReaders++
	lock.conditionVariable.L.Unlock()
}

func (lock *MyRWLock) RUnlock() {
	lock.conditionVariable.L.Lock()
	lock.numReaders--
	if lock.numReaders == 0 {
		lock.conditionVariable.Signal()
	}
	lock.conditionVariable.L.Unlock()
}

func (lock *MyRWLock) Lock() {
	lock.conditionVariable.L.Lock()
	for lock.numReaders > 0 {
		lock.conditionVariable.Wait()
	}
}

func (lock *MyRWLock) Unlock() {
	lock.conditionVariable.Signal()
	lock.conditionVariable.L.Unlock()
}
