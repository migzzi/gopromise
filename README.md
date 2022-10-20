# gopromise
[![License](https://img.shields.io/github/license/migzzi/gopromise)]()
[![Release](https://img.shields.io/github/v/release/migzzi/gopromise)](https://goreportcard.com/report/github.com/chebyrash/promise)
[![Build Status](https://img.shields.io/github/workflow/status/migzzi/gopromise/Test?label=tests)](https://github.com/chebyrash/promise/actions)


An implementation of JavaScript/A+ like promises in golang inspired by a couple of open source implementations.

## Install
```bash
    $ go get -u github.com/migzzi/gopromise
```

## Usage
```go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
    "ioutil"

	"github.com/migzzi/gopromise"
)

func main() {
	p1 := gopromise.New(func(resolve func(int), reject func(error)) {
		factorial := getFactorial(20)
		resolve(factorial)
	})
	p2 := gopromise.New(func(resolve func(string), reject func(error)) {
		users, err := getUsers()
		if err != nil {
			reject(err)
			return
		}
		resolve(users)
	})

	factorial, _ := p1.Await()
	fmt.Println(factorial)

	users, _ := p2.Await()
	fmt.Println(users)

    p3 := gopromise.Then(p2, func(users []GithubUser) int {
        return len(users)
    })

    gopromise.Catch(p2, func(err error) any {
        fmt.Println("Error")
        return nil
    })

    usersCount, _ := p3.Await()
	fmt.Println(usersCount)

}

func getFactorial(n int) int {
	if n == 1 {
		return 1
	}
	return n * findFactorial(n-1)
}

func getUsers() ([]GithubUser, error) {
	resp, err := http.Get("https://api.github.com/users")
	if err != nil {
		return "", err
	}
    defer reso.Body.Close()

    body, err := ioutil.ReadAll(req.Body)
    if err != nil {
        panic(err.Error())
    }

    type GithubUser {
        id int `json:"id"`
        username string `json:"login"`
        avatarUrl string `json:"avatar_url"`
    }

	var users []GithubUser
	err = json.NewDecoder(body).Decode(&users)
	return users, err
}
```