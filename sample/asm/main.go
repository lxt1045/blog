package main

import (
	"fmt"

	"github.com/petermattis/goid"
)

// go tool compile -N  -S main.go >main.s
var XXX = 8

var gID int64

func getG() {
	gID = goid.Get()
}

// func main1() {
// 	ForCheckptr0()

// 	closure := NewClosure
// 	closure()
// 	closure()

// 	a := 1
// 	shlx := a << XXX
// 	x := "qq"

// 	var i interface{}

// 	i = x
// 	i = shlx

// 	fmt.Printf("%+v", i)
// }

// var closure = NewClosure()

func main() {
	getG()

	var closure func()
	closure = NewClosure()
	closure()
	closure()
	// closure = NewClosure()
	closure()

	fmt.Printf("type of closure: %T\n", closure)

	ForCheckptr0()

	i := 1
	func() {
		// XXX = 0
		i += 1
	}()
}

//go:nocheckptr
func ForCheckptr0() {
	// type Int int
	i := 0
	func() {
		// XXX = 0
		i += 1
	}()
	// XXX = 0
	// ForCheckptr()
	// XXX = 0
}

func NewClosure() func() {
	// type Int int
	i := 0
	return func() {
		// XXX = 0
		i += 1
	}
}

func ForCheckptr() {
	XXX = 9
	a := 1
	shlx := a << XXX
	x := "qq"

	var i interface{}

	i = x
	i = shlx
	fmt.Printf("%+v", i)
}
