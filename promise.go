package gopromise

import (
	"fmt"
	"sync"
)

type promiseStatus uint16

const (
	PENDING promiseStatus = iota
	FULFILLED
	REJECTED
)

type Promise[T any] struct {
	value  T
	reason error
	status promiseStatus
	mutex  *sync.Mutex
	wg     *sync.WaitGroup
}

func New[T any](exec func(resolve func(T), reject func(error))) *Promise[T] {
	if exec == nil {
		panic("executor cannot be nil")
	}

	p := &Promise[T]{
		status: PENDING,
		mutex:  &sync.Mutex{},
		wg:     &sync.WaitGroup{},
	}

	p.wg.Add(1)

	go func() {
		// catch exception error happen in the executor
		defer func() {
			r := recover()
			if err, ok := r.(error); ok {
				p.reject(err)
			} else {
				p.reject(fmt.Errorf("%+v", r))
			}
		}()
		exec(p.resolve, p.reject)
	}()

	return p
}

func (p *Promise[T]) resolve(val T) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.status != PENDING {
		return
	}

	p.status = FULFILLED
	p.value = val
	p.wg.Done()
}

func (p *Promise[T]) reject(err error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.status != PENDING {
		return
	}

	p.status = REJECTED
	p.reason = err
	p.wg.Done()
}

func (p *Promise[T]) Await() (T, error) {
	p.wg.Wait()
	return p.value, p.reason
}

func Then[T, R any](src *Promise[T], cb func(val T) R) *Promise[R] {
	if src == nil {
		panic("must provide valid promise")
	}
	return New(func(resolve func(R), reject func(error)) {
		val, err := src.Await()
		if err != nil {
			reject(err)
			return
		}
		resOrProm := cb(val)
		if rp, ok := interface{}(resOrProm).(*Promise[R]); ok {
			Then(rp, func(val R) R { resolve(val); return val })
			Catch(rp, func(err error) any { reject(err); return nil })
			return
		}
		resolve(resOrProm)
	})
}

func Catch[T, R any](src *Promise[T], cb func(err error) R) *Promise[R] {
	return New(func(resolve func(R), reject func(error)) {
		_, err := src.Await()
		if err != nil {
			resOrProm := cb(err)
			if rp, ok := interface{}(resOrProm).(*Promise[R]); ok {
				Then(rp, func(val R) R { resolve(val); return val })
				Catch(rp, func(err error) any { reject(err); return nil })
				return
			}
			resolve(resOrProm)
			return
		}
	})
}

func Resolve[T any](value T) *Promise[T] {
	return &Promise[T]{
		value:  value,
		status: FULFILLED,
		mutex:  new(sync.Mutex),
		wg:     new(sync.WaitGroup),
	}
}

// Reject returns a Promise that has been rejected with a given error.
func Reject[T any](err error) *Promise[T] {
	return &Promise[T]{
		reason: err,
		status: REJECTED,
		mutex:  new(sync.Mutex),
		wg:     new(sync.WaitGroup),
	}
}

type pair[T, R any] struct {
	first  T
	second R
}

func All[T any](promises ...*Promise[T]) *Promise[[]T] {
	if len(promises) == 0 {
		return nil
	}
	return New(func(resolve func([]T), reject func(error)) {
		doneChan := make(chan bool, len(promises))
		errChan := make(chan error, 1)
		values := make([]T, len(promises))
		for idx, p := range promises {
			idx := idx
			_ = Then(p, func(val T) T {
				values[idx] = val
				doneChan <- true
				return val
			})
			_ = Catch(p, func(err error) error {
				errChan <- err
				return err
			})
		}

		for idx := 0; idx < len(promises); idx++ {
			select {
			case <-doneChan:
			case err := <-errChan:
				reject(err)
				return
			}
		}
		resolve(values)
	})
}

func Race[T any](promises ...*Promise[T]) *Promise[T] {
	if len(promises) == 0 {
		return nil
	}
	return New(func(resolve func(T), reject func(error)) {
		valueChan := make(chan T, 1)
		errChan := make(chan error, 1)
		for _, p := range promises {
			_ = Then(p, func(val T) T {
				valueChan <- val
				return val
			})
			_ = Catch(p, func(err error) error {
				errChan <- err
				return err
			})
		}

		select {
		case val := <-valueChan:
			resolve(val)
		case err := <-errChan:
			reject(err)
		}
	})

}
