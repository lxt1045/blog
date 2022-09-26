package main

import (
	"fmt"
)

// go tool compile -N  -S main.go >main.s
func main() {
	x := "qq"

	var i interface{}

	i = x

	fmt.Printf("%+v", i)
}

func toEface() {

}
