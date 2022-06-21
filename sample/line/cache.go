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
	"sync/atomic"
	"unsafe"
)

type cache[T uintptr | string] struct {
	m unsafe.Pointer

	load func(key T) (v string)
}

func newCache[T uintptr | string](n int, load func(key T) (v string)) (c cache[T]) {
	m := make(map[T]string, n)
	c.m = unsafe.Pointer(&m)
	c.load = load
	return
}

func (c *cache[T]) Get(key T) (line string) {
	m := *(*map[T]string)(atomic.LoadPointer(&c.m))
	line, ok := m[key]
	if !ok {
		line = c.load(key)
		m2 := make(map[T]string, len(m)+10)
		m2[key] = line
		for {
			p := atomic.LoadPointer(&c.m)
			m = *(*map[T]string)(p)
			for k, v := range m {
				m2[k] = v
			}
			swapped := atomic.CompareAndSwapPointer(&c.m, p, unsafe.Pointer(&m2))
			if swapped {
				break
			}
		}
	}
	return
}
