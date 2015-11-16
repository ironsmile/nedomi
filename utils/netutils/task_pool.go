package netutils

import (
	"fmt"
	"runtime"
	"sync"
)

type task interface {
	Execute()
}

type pool struct {
	mu    sync.Mutex
	size  int
	tasks chan task
	kill  chan struct{}
	wg    sync.WaitGroup
}

func newPool(size int) *pool {
	pool := &pool{
		tasks: make(chan task, 128),
		kill:  make(chan struct{}),
	}
	fmt.Println(size)
	pool.resize(size)
	return pool
}

func (p *pool) worker() {
	defer p.wg.Done()
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	for {
		select {
		case task, ok := <-p.tasks:
			if !ok {
				return
			}
			task.Execute()
		case <-p.kill:
			return
		}
	}
}

func (p *pool) resize(n int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for p.size < n {
		p.size++
		p.wg.Add(1)
		go p.worker()
	}
	for p.size > n {
		p.size--
		p.kill <- struct{}{}
	}
}

func (p *pool) Close() {
	close(p.tasks)
}

func (p *pool) Wait() {
	p.wg.Wait()
}

func (p *pool) Exec(task task) {
	p.tasks <- task
}

type funcExecute func()

func (f funcExecute) Execute() {
	f()
}
