// MIT License
//
// Copyright (c) 2021 Xiantu Li
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"fmt"
	"log"
	"reflect"
	"runtime"
	"unsafe"
)

func init() {
	log.SetFlags(log.Flags() | log.Lmicroseconds | log.Lshortfile) //log.Llongfile
}

func main() {

	for _ = range [1]struct{}{} {
		GetPc(func() {})
		GetPc(func() {})
	}

	log.SetFlags(log.Lshortfile | log.LstdFlags)
	log.Println("test")

	FuncForPC0()

	Printf()

}

func GetPc(f func()) {
	pc := reflect.ValueOf(f).Pointer()
	log.Println(runtime.FuncForPC(pc).FileLine(pc))

	p := (*uintptr)(unsafe.Pointer(&f))
	log.Printf("%d:%d", uintptr(*p), pc)
}

func InlineFunc() (pcs []uintptr) {
	func() {
		func() {
			func() {
				pcs = make([]uintptr, 32)
				n := runtime.Callers(0, pcs)
				pcs = pcs[:n]
			}()
		}()
	}()
	return
}

func parse(pcs []uintptr) (cs []string) {
	traces, more, f := runtime.CallersFrames(pcs), true, runtime.Frame{}
	for more {
		f, more = traces.Next()
		cs = append(cs,
			fmt.Sprintf("%s:%d\t%s\n", f.File, f.Line, f.Function),
		)
	}
	return
}

func parse2(pcs []uintptr) (cs []string) {
	for i := range pcs {
		f, _ := runtime.CallersFrames(pcs[i : i+1]).Next()
		cs = append(cs,
			fmt.Sprintf("%s:%d\t%s\n", f.File, f.Line, f.Function),
		)
	}
	return
}

func parse3(pcs []uintptr) (cs []string) {
	for _, pc := range pcs {
		f := runtime.FuncForPC(pc)
		file, line := f.FileLine(pc)
		cs = append(cs,
			fmt.Sprintf("%s:%d\t%s\n", file, line, f.Name()),
		)
	}
	return
}

func parse4(pcs []uintptr) (cs []string) {
	for _, pc := range pcs {
		f := runtime.FuncForPC(pc)
		file, line := f.FileLine(pc)
		cs = append(cs,
			fmt.Sprintf("%s:%d\t%s\n", file, line, f.Name()),
		)
	}
	return
}
func Printf() {
	pcs := InlineFunc()
	log.Printf("len:%d,\n%v\n", len(parse(pcs)), parse(pcs))
	log.Printf("len:%d,\n%v\n", len(parse2(pcs)), parse2(pcs))
	log.Printf("len:%d,\n%v\n", len(parse3(pcs)), parse3(pcs))
	log.Printf("len:%d,\n%v\n", len(parse4(pcs)), parse4(pcs))
	return
}

func FuncForPC0() {
	pc, file, line, ok := runtime.Caller(0)
	if ok {
		log.Println("Func Name=" + runtime.FuncForPC(pc).Name())
		log.Println(runtime.FuncForPC(pc).FileLine(pc))
		log.Printf("file: %s    line=%d\n", file, line)
	}
}
