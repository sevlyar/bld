package main

import (
	"fmt"
	"log"
	"os"
)

// throw вызывает panic с указанным сообщением, помещенным в error.
func throw(f string, a ...interface{}) {
	panic(fmt.Errorf(f, a...))
}

// rethrow прекращает panic, извлекает сообщение, добавляет к нему префикс
// и снова вызывает panic(). Функция должна вызываться только с помощью defer.
func rethrow(f string, a ...interface{}) {
	if err := recover(); err != nil {
		throw(f+": "+fmt.Sprint(err), a...)
	}
}

// exit отлавливает panic, выводит сообщение в stdout и stderr? и завершает
// работу программы с кодом 1. Если не было panic, то ничего не делает.
// Должна вызываться только с помощью defer и из main().
func exit() {
	if err := recover(); err != nil {
		log.Println(err)
		fmt.Println(err)
		os.Exit(1)
	}
}
