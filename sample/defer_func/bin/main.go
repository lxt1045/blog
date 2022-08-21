//go:build (amd64 || amd64p32) && gc && go1.5
// +build amd64 amd64p32
// +build gc
// +build go1.5

package main

import (
	"log"
	_ "unsafe"
)

func main() {
	// My()

	TryTag()
}

func Deferx() {
	log.Println("inner defer 1")
}

func TryTag() (err error) {
	defer Deferx()
	defer Deferx()

	return
}
