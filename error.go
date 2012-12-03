package main

import (
	"errors"
	"fmt"
)

func throw(f string, a ...interface{}) {
	panic(errors.New(fmt.Sprintf(f, a...)))
}

func catch() {
	if err := recover(); err != nil {
		fmt.Println(err)
	}
}

func rethrow(f string, a ...interface{}) {
	if err := recover(); err != nil {
		throw(fmt.Sprint(err)+": "+f, a...)
	}
}
