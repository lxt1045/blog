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

package errors

import (
	"runtime"
	"strconv"
	"sync"
	"unsafe"

	"github.com/lxt1045/errors"
)

const DefaultDepth = 32

var (
	cacheStack = errors.AtomicCache[string, []string]{}
	pool       = sync.Pool{
		New: func() any { return &[DefaultDepth]uintptr{} },
	}
)

func toString(p []uintptr) string {
	bs := (*[DefaultDepth * 8]byte)(unsafe.Pointer(&p[0]))[:len(p)*8]
	return *(*string)(unsafe.Pointer(&bs))
}

func NewStack(skip int) (stack []string) {
	pcs := pool.Get().(*[DefaultDepth]uintptr)
	n := runtime.Callers(skip+2, pcs[:DefaultDepth-skip])
	key := toString(pcs[:n])

	stack = cacheStack.Get(key)
	if len(stack) == 0 {
		stack = parseSlow(pcs[:n])
		cacheStack.Set(key, stack)
	}
	pool.Put(pcs)
	return
}
func parseSlow(pcs []uintptr) (cs []string) {
	traces, more, f := runtime.CallersFrames(pcs), true, runtime.Frame{}
	for more {
		f, more = traces.Next()
		cs = append(cs, f.File+":"+strconv.Itoa(f.Line))
	}
	return
}

func NewStack2(skip int) (stack []string) {
	pcs := pool.Get().(*[DefaultDepth]uintptr)
	n := buildStack(pcs[:])
	key := toString(pcs[:n])

	//
	stack = cacheStack.Get(key)
	if len(stack) == 0 {
		pcs1 := make([]uintptr, DefaultDepth)
		npc1 := runtime.Callers(skip+2, pcs1[:DefaultDepth-skip])

		stack = parseSlow(pcs1[:npc1])
		cacheStack.Set(key, stack)
	}
	pool.Put(pcs)
	return
}
