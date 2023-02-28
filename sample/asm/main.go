package main

import (
	"fmt"
)

// go tool compile -N  -S main.go >main.s
var XXX = 8

func main() {
	a := 1
	shlx := a << XXX
	x := "qq"

	var i interface{}

	i = x
	i = shlx

	fmt.Printf("%+v", i)
}

func toEface() {

}
