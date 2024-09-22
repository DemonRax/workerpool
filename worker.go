package worker_pool

import (
	"errors"
	"sync"
)

var (
	ErrSkip = errors.New("skip operation, no error")
)

type worker struct {
	size      int
	wg, eg    *sync.WaitGroup
	dataChan  chan any
	errChan   chan error
	errors    []error
	operation func(any) error
	cSuccess  int
}

func New(size int, op func(any) error) *worker {
	return &worker{
		size:      size,
		wg:        &sync.WaitGroup{},
		eg:        &sync.WaitGroup{},
		dataChan:  make(chan any, size*2),
		errChan:   make(chan error, size*2),
		errors:    make([]error, 0, size),
		operation: op,
	}
}

func (w *worker) Operation(op func(any) error) {
	w.operation = op
}

func (w *worker) Add(data any) {
	w.dataChan <- data
}

func (w *worker) Start() {
	w.eg.Add(1)
	go func() {
		for err := range w.errChan {
			if err != nil {
				if !errors.Is(err, ErrSkip) {
					w.errors = append(w.errors, err)
				}
			} else {
				w.cSuccess++
			}
		}
		w.eg.Done()
	}()

	for i := 0; i < w.size; i++ {
		w.wg.Add(1)
		go func() {
			for data := range w.dataChan {
				w.errChan <- w.operation(data)
			}
			w.wg.Done()
		}()
	}
}

type Errs []error

func (e Errs) First() error {
	if len(e) == 0 {
		return nil
	}
	return e[0]
}

func (w *worker) Wait() Errs {
	close(w.dataChan)
	w.wg.Wait()
	close(w.errChan)
	w.eg.Wait()
	if len(w.errors) == 0 {
		return nil
	}
	return w.errors
}

func (w *worker) Success() int {
	return w.cSuccess
}
