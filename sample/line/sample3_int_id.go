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
	"runtime"
	"strconv"
	"sync/atomic"
	"unsafe"
)

var intIDCache *[]*string = func() *[]*string {
	s := make([]*string, 1024)
	return &s
}()

func GetLineNO3(id int) string {
	cache := intIDCache
	if len(*cache) <= id {
		s := make([]*string, id*2)
		copy(s, *cache)
		// intIDCache = &s
		atomic.StoreUintptr((*uintptr)(unsafe.Pointer(&intIDCache)), (uintptr)(unsafe.Pointer(&s)))
		cache = &s
	}
	p := (*cache)[id]
	if p != nil && *p != "" {
		return *p
	}
	_, file, l, ok := runtime.Caller(1)
	if ok {
		line := file + ":" + strconv.Itoa(l)
		// (*cache)[id] = &line
		atomic.StoreUintptr((*uintptr)(unsafe.Pointer(&(*cache)[id])), (uintptr)(unsafe.Pointer(&line)))
		(*cache)[id] = &line // 这句让编译器确保line逃逸,因为line未逃逸的话,会导致引用未初始化内存.
		return line
	}
	return ""
}
