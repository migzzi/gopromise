package gopromise

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

var promiseError = errors.New("Promise Error")

func isNil(i interface{}) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}

func assert(t *testing.T, expr bool, msgs ...string) {
	if expr != true {
		t.Error(strings.Join(msgs, " "))
	}
}

func assertEqual(t *testing.T, expected, got any, msgs ...string) {
	if msgs == nil {
		errorMsg := fmt.Sprintf("Expected: %v\tGot: %v\n", expected, got)
		msgs = append(msgs, errorMsg)
	}
	assert(t, expected == got, msgs...)
}

func assertNotNil(t *testing.T, expected any, msgs ...string) {
	if msgs == nil {
		errorMsg := fmt.Sprintf("Expected: %v not to be nil", expected)
		msgs = append(msgs, errorMsg)
	}
	assert(t, !isNil(expected), msgs...)
}

func assertNil(t *testing.T, expected any, msgs ...string) {
	if msgs == nil {
		errorMsg := fmt.Sprintf("Expected: %v to be nil", expected)
		msgs = append(msgs, errorMsg)
	}
	assert(t, isNil(expected), msgs...)
}

func assertNotErr(t *testing.T, expected any, msgs ...string) {
	if msgs == nil {
		errorMsg := fmt.Sprintf("Expected: %v not to be an error", expected)
		msgs = append(msgs, errorMsg)
	}
	_, ok := expected.(error)
	assert(t, !ok, msgs...)
}

func assertErr(t *testing.T, expected any, msgs ...string) {
	if msgs == nil {
		errorMsg := fmt.Sprintf("Expected: %v to be an error", expected)
		msgs = append(msgs, errorMsg)
	}
	_, ok := expected.(error)
	assert(t, ok, msgs...)
}

func TestNew(t *testing.T) {
	p := New(func(resolve func(any), reject func(error)) {
		resolve(42)
	})
	assertNotNil(t, p)

	res, err := p.Await()
	assertEqual(t, res, 42)
	assertNotErr(t, err)
}

func TestPromise_Then(t *testing.T) {
	p1 := New(func(resolve func(int), reject func(error)) {
		resolve(42)
	})

	res, err := p1.Await()
	assertEqual(t, res, 42)
	assertNotErr(t, err)

	p2 := Then(p1, func(v int) int {
		return v + 1
	})

	res, err = p2.Await()
	assertEqual(t, res, 43)
	assertNotErr(t, err)
}

func TestPromise_Catch(t *testing.T) {
	p1 := New(func(resolve func(any), reject func(error)) {
		reject(promiseError)
	})
	p2 := Then(p1, func(v any) any {
		t.Fatal("should not execute Then")
		return nil
	})
	p3 := Catch(p1, func(v error) any {
		return "Tadaa"
	})

	res, err := p1.Await()
	assertNil(t, res)
	assertErr(t, err)
	assertEqual(t, err, promiseError)

	p2.Await()

	res, err = p3.Await()
	assertNotErr(t, err)
	assertEqual(t, res, "Tadaa")
}

func TestPromise_Panic(t *testing.T) {
	p1 := New(func(resolve func(any), reject func(error)) {
		panic(nil)
	})
	p2 := New(func(resolve func(any), reject func(error)) {
		panic("random error")
	})
	p3 := New(func(resolve func(any), reject func(error)) {
		panic(promiseError)
	})

	val, err := p1.Await()
	assertErr(t, err)
	assertEqual(t, "<nil>", err.Error())
	assertNil(t, val)

	val, err = p2.Await()
	assertErr(t, err)
	assertEqual(t, "random error", err.Error())
	assertNil(t, val)

	val, err = p3.Await()
	assertErr(t, err)
	assertEqual(t, err, promiseError)
	assertNil(t, val)
}

func TestAll_AllSuccess(t *testing.T) {
	p1 := New(func(resolve func(int), reject func(error)) {
		resolve(1)
	})
	p2 := New(func(resolve func(int), reject func(error)) {
		resolve(2)
	})
	p3 := New(func(resolve func(int), reject func(error)) {
		resolve(3)
	})

	p := All(p1, p2, p3)

	res, err := p.Await()
	assertNotErr(t, err)
	for idx, r := range res {
		assertEqual(t, idx+1, r)
	}
}

func TestAll_WithRejection(t *testing.T) {
	p1 := New(func(resolve func(int), reject func(error)) {
		resolve(1)
	})
	p2 := New(func(resolve func(int), reject func(error)) {
		reject(promiseError)
	})
	p3 := New(func(resolve func(int), reject func(error)) {
		resolve(3)
	})

	p := All(p1, p2, p3)
	res, err := p.Await()

	assertErr(t, err)
	assertNil(t, res)
	assertEqual(t, promiseError, err)
}

func TestAll_AllRejection(t *testing.T) {
	p1 := New(func(resolve func(int), reject func(error)) {
		reject(promiseError)
	})
	p2 := New(func(resolve func(int), reject func(error)) {
		reject(promiseError)
	})
	p3 := New(func(resolve func(int), reject func(error)) {
		reject(promiseError)
	})

	p := All(p1, p2, p3)
	res, err := p.Await()

	assertErr(t, err)
	assertNil(t, res)
	assertEqual(t, promiseError, err)
}

func TestAll_EmptyList(t *testing.T) {
	var empty []*Promise[any]
	p := All(empty...)
	assertNil(t, p)
}

func TestRace_AllSuccess(t *testing.T) {
	p1 := New(func(resolve func(int), reject func(error)) {
		time.Sleep(100 * time.Millisecond)
		resolve(1)
	})
	p2 := New(func(resolve func(int), reject func(error)) {
		time.Sleep(300 * time.Millisecond)
		resolve(2)
	})
	p3 := New(func(resolve func(int), reject func(error)) {
		time.Sleep(500 * time.Millisecond)
		resolve(3)
	})

	p := Race(p1, p2, p3)

	res, err := p.Await()
	assertNotErr(t, err)
	assertEqual(t, 1, res)

}

func TestRace_WithRejection(t *testing.T) {
	p1 := New(func(resolve func(any), reject func(error)) {
		time.Sleep(300 * time.Millisecond)
		resolve(1)
	})
	p2 := New(func(resolve func(any), reject func(error)) {
		time.Sleep(100 * time.Millisecond)
		reject(promiseError)
	})
	p3 := New(func(resolve func(any), reject func(error)) {
		time.Sleep(500 * time.Millisecond)
		resolve(3)
	})

	p := Race(p1, p2, p3)
	res, err := p.Await()

	assertErr(t, err)
	assertNil(t, res)
	assertEqual(t, promiseError, err)
}

func TestRace_AllRejection(t *testing.T) {
	err1 := errors.New("Err 1")
	err2 := errors.New("Err 2")
	p1 := New(func(resolve func(any), reject func(error)) {
		time.Sleep(100 * time.Millisecond)
		reject(err1)
	})
	p2 := New(func(resolve func(any), reject func(error)) {
		time.Sleep(300 * time.Millisecond)
		reject(err2)
	})

	p := Race(p1, p2)
	res, err := p.Await()

	assertErr(t, err)
	assertNil(t, res)
	assertEqual(t, err1, err)
}

func TestRace_EmptyList(t *testing.T) {
	var empty []*Promise[any]
	p := Race(empty...)
	assertNil(t, p)
}
