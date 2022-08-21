//go:build (amd64 || amd64p32) && gc && go1.5
// +build amd64 amd64p32
// +build gc
// +build go1.5

package main

import (
	"log"
	_ "unsafe"

	"github.com/lxt1045/blog/sample/defer_func/stack"
)

func init() {
	log.SetFlags(log.Flags() | log.Lmicroseconds | log.Lshortfile)
}

func main() {
	// My()

	err := TryTag()
	log.Printf("err:%+v", err)
}

var x int

//go:noinline
func DeferX() {
	log.Println("inner defer: DeferX")
	x = 1100
}
func TryTag() (err error) {
Tagxxx:
	if x != 0 {
		defer DeferX()
	}
	defer func() {
		log.Println("inner defer 1")
	}()
	d := stack.GetDefer()
	if d == nil {
		log.Printf("GetDefer:<nil>")
	}
	for ; d != nil; d = d.Next() {
		log.Printf("GetDefer:%d", d.PC())
	}
	return
	goto Tagxxx
}

func TryTag1() (err error) {
	defer func() {
		log.Println("defer 1")
	}()
	d := stack.GetDefer()
	for ; d != nil; d = d.Next() {
		log.Printf("defer pc:%d", d.PC())
	}
	return
}
