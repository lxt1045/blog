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

//go:build amd64
// +build amd64

package errors

import (
	"fmt"
	"runtime"
	"strconv"
	"sync/atomic"
	"unsafe"
	_ "unsafe"
)

func GotoTag(err error)
func Tag() (err error)

type cacheType = map[uintptr]uintptr

var (
	tagCache  = func() unsafe.Pointer { m := make(cacheType, 1024); return unsafe.Pointer(&m) }()
	tryTagErr func(error)
)

// storeTag 由 asm函数调用，存储 Tag
func storeTag(pc uintptr) {
	cache := *(*cacheType)(atomic.LoadPointer(&tagCache))
	_, ok := cache[pc]
	if !ok {
		// funcEntry := runtime.FuncForPC(pc).Entry() // 函数入口
		f, _ := runtime.CallersFrames([]uintptr{pc}).Next()
		funcEntry := f.Entry

		cache2 := make(cacheType, len(cache)+10)
		cache2[pc] = funcEntry
		cache2[funcEntry] = pc

		for {
			p := atomic.LoadPointer(&tagCache)
			cache = *(*cacheType)(p)
			for k, v := range cache {
				cache2[k] = v
			}
			swapped := atomic.CompareAndSwapPointer(&tagCache, p, unsafe.Pointer(&cache2))
			if swapped {
				break
			}
		}
	}
	return
}

// storeTag 由 asm函数调用，查询 Tag
func loadTag(pc uintptr) uintptr {
	cache := *(*cacheType)(atomic.LoadPointer(&tagCache))
	pcTag, ok := cache[pc]
	if ok {
		return pcTag
	}
	// funcEntry := runtime.FuncForPC(pc).Entry() // 函数入口
	f, _ := runtime.CallersFrames([]uintptr{pc}).Next()
	funcEntry := f.Entry

	pcTag, ok = cache[funcEntry]
	if !ok {
		panic(fmt.Sprintf("has no Tag point, %s:%d", f.File, f.Line))
	}

	cache2 := make(cacheType, len(cache)+10)
	cache2[pc] = pcTag

	for {
		p := atomic.LoadPointer(&tagCache)
		cache = *(*cacheType)(p)
		for k, v := range cache {
			cache2[k] = v
		}
		swapped := atomic.CompareAndSwapPointer(&tagCache, p, unsafe.Pointer(&cache2))
		if swapped {
			break
		}
	}

	return pcTag
}

func Jump(pc uintptr, err error)
func Jump1(pc uintptr, err error)
func Jump2(pc, parent uintptr, err error) uintptr
func NewTag() (func(error), error)
func NewTag2() (tag, error)

// storeTag 由 asm函数调用，存储 Tag
func newTag(pc uintptr) (f func(error)) {
	return func(err error) {
		Jump1(pc, err)
	}
}

//go:noinline
// TryJump 由 asm函数调用，查询 Tag
func TryJump(err error) {
	pc := GetPC()[0]
	cache := *(*cacheType)(atomic.LoadPointer(&tagCache))
	pcTag, ok := cache[pc]
	if ok && err == nil {
		return
	}
	// funcEntry := runtime.FuncForPC(pc).Entry() // 函数入口
	f, _ := runtime.CallersFrames([]uintptr{pc}).Next()
	funcEntry := f.Entry

	pcTag, ok = cache[funcEntry]
	if !ok {
		panic(fmt.Sprintf("has no Tag point, %s:%d", f.File, f.Line))
	}

	cache2 := make(cacheType, len(cache)+10)
	cache2[pc] = pcTag

	for {
		p := atomic.LoadPointer(&tagCache)
		cache = *(*cacheType)(p)
		for k, v := range cache {
			cache2[k] = v
		}
		swapped := atomic.CompareAndSwapPointer(&tagCache, p, unsafe.Pointer(&cache2))
		if swapped {
			break
		}
	}
	Jump(pcTag, err)

	return
}

type tag struct {
	pc     uintptr
	parent uintptr
}

func (t tag) Try(err error) {
	//还是要加上检查，否则报错信息太难看
	// 但是检查时只要检查 更上一级的 PC 是否相等即可，不需要复杂的 map 存储了！！！
	parent := Jump2(t.pc, t.parent, err)
	if parent != t.parent {
		cs := toCallers([]uintptr{parent, t.parent, GetPC()[0]})
		e := fmt.Errorf("tag.Try() should be called in [%s] not in [%s]; line:%s",
			cs[1].name, cs[0].name, cs[2].line)
		if err != nil {
			e = fmt.Errorf("%w; %+v", err, e)
		}
		if tryTagErr != nil {
			tryTagErr(e)
			return
		}
		panic(err)
	}
}

type Caller struct {
	pc   uintptr
	line string
	name string
}

func toCallers(pcs []uintptr) (callers []Caller) {
	for f, i := runtime.CallersFrames(pcs), 0; ; i++ {
		ff, more := f.Next()
		line := ff.File + ":" + strconv.Itoa(ff.Line)
		callers = append(callers, Caller{
			pc:   pcs[i],
			line: line,
			name: ff.Function,
		})
		if !more {
			break
		}
	}
	return
}
