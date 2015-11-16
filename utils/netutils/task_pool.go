package netutils

import (
	"fmt"
	"runtime"
	"sync"
)

type task interface {
	Execute()
}

type workerPool struct {
	mu    sync.Mutex
	size  int
	tasks chan task
	kill  chan struct{}
	wg    sync.WaitGroup
}

func newPool(size int) *workerPool {
	p := &workerPool{
		tasks: make(chan task, 128),
		kill:  make(chan struct{}),
	}
	fmt.Println(size)
	p.resize(size)
	return p
}

func (p *workerPool) worker() {
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

func (p *workerPool) resize(n int) {
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

func (p *workerPool) Close() {
	close(p.tasks)
}

func (p *workerPool) Wait() {
	p.wg.Wait()
}

func (p *workerPool) Exec(task task) {
	p.tasks <- task
}

type funcExecute func()

func (f funcExecute) Execute() {
	f()
}
