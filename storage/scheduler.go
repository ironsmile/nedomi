package storage

import (
	"container/heap"
	"sync"
	"time"
)

type elem struct {
	Key      string
	Callback func()
}

type expireTime struct {
	Key     string
	Expires time.Time
}

type expireHeap []expireTime

func (h expireHeap) Len() int {
	return len(h)
}

func (h expireHeap) Less(i, j int) bool {
	return h[i].Expires.Before(h[j].Expires)
}

func (h expireHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *expireHeap) Push(x interface{}) {
	*h = append(*h, x.(expireTime))
}

func (h *expireHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// Scheduler efficiently manages and executes callbacks at specified times.
type Scheduler struct {
	stopChan chan struct{}
	wg       sync.WaitGroup

	setRequest       chan *elem
	deleteRequest    chan string
	containsRequest  chan string
	containsResponse chan bool
	cleanupRequest   chan struct{}

	newExpireTime         chan expireTime
	cleanupExpiresRequest chan struct{}
}

// NewScheduler initializes and returns a newly created Scheduler instance.
func NewScheduler() (em *Scheduler) {
	em = &Scheduler{}

	em.stopChan = make(chan struct{})
	em.setRequest = make(chan *elem)
	em.deleteRequest = make(chan string)
	em.containsRequest = make(chan string)
	em.containsResponse = make(chan bool)
	em.cleanupRequest = make(chan struct{})

	em.newExpireTime = make(chan expireTime)
	em.cleanupExpiresRequest = make(chan struct{})

	em.wg.Add(1)
	go em.storageHandler()
	em.wg.Add(1)
	go em.expiresHandler()

	return
}

func (em *Scheduler) storageHandler() {
	defer em.wg.Done()
	cache := make(map[string]func())

	for {
		select {
		case <-em.stopChan:
			return

		case elem := <-em.setRequest:
			cache[elem.Key] = elem.Callback

		case <-em.cleanupRequest:
			cache = make(map[string]func())

		case key := <-em.containsRequest:
			_, ok := cache[key]
			em.containsResponse <- ok

		case key := <-em.deleteRequest:
			if f, ok := cache[key]; ok {
				go f()
			}

			delete(cache, key)
		}
	}
}

func (em *Scheduler) expiresHandler() {
	defer em.wg.Done()

	expiresDict := make(map[string]time.Time)
	expires := &expireHeap{}
	heap.Init(expires)

	for {
		var nextExpire *expireTime
		nextExpireDuration := time.Hour

		if expires.Len() > 0 {
			nextExpire = &((*expires)[0])
			nextExpireDuration = nextExpire.Expires.Sub(time.Now())
		}

		select {
		case <-em.stopChan:
			return

		case elem := <-em.newExpireTime:
			heap.Push(expires, elem)
			expiresDict[elem.Key] = elem.Expires

		case <-em.cleanupExpiresRequest:
			expiresDict = make(map[string]time.Time)
			expires = &expireHeap{}
			heap.Init(expires)

		case <-time.After(nextExpireDuration):
			if nextExpire == nil {
				continue
			}
			em.deleteRequest <- nextExpire.Key
			delete(expiresDict, nextExpire.Key)

			heap.Remove(expires, 0)
		}
	}
}

// AddEvent schedules the passed callback to be executed at the supplied time.
func (em *Scheduler) AddEvent(key string, callback func(), expire time.Duration) {
	em.newExpireTime <- expireTime{Key: key, Expires: time.Now().Add(expire)}
	em.setRequest <- &elem{Key: key, Callback: callback}
}

// Contains checks whether an event with the supplied key is scheduled.
func (em *Scheduler) Contains(key string) bool {
	em.containsRequest <- key
	return <-em.containsResponse
}

// Cleanup removes all scheduled events
func (em *Scheduler) Cleanup() {
	em.cleanupRequest <- struct{}{}
	em.cleanupExpiresRequest <- struct{}{}
}

// Destroy stops and destroys the scheduler
func (em *Scheduler) Destroy() {
	close(em.stopChan)
	em.wg.Wait()

	close(em.setRequest)
	close(em.deleteRequest)
	close(em.containsRequest)
	close(em.containsResponse)
	close(em.cleanupRequest)
	close(em.newExpireTime)
	close(em.cleanupExpiresRequest)
}
