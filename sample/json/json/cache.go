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

package json

import (
	"reflect"
	"sync"
	"sync/atomic"
	"unsafe"
)

// json.go 中有很多用完就写入 map[string] interface{} 中的，可以用 sync.pool

type cache[T uintptr | uint32 | string | int, V any] struct {
	m unsafe.Pointer
}

func newCache[T uintptr | uint32 | string | int, V any]() (c cache[T, V]) {
	m := make(map[T]V, 1)
	c.m = unsafe.Pointer(&m)
	return
}

func (c *cache[T, V]) Get(key T) (v V, ok bool) {
	m := *(*map[T]V)(atomic.LoadPointer(&c.m))
	v, ok = m[key]
	return
}

func (c *cache[T, V]) Set(key T, v V) {
	m := *(*map[T]V)(atomic.LoadPointer(&c.m))
	if _, ok := m[key]; ok {
		return
	}
	m2 := make(map[T]V, len(m)+10)
	m2[key] = v
	for {
		p := atomic.LoadPointer(&c.m)
		m = *(*map[T]V)(p)
		if _, ok := m[key]; ok {
			return
		}
		for k, v := range m {
			m2[k] = v
		}
		swapped := atomic.CompareAndSwapPointer(&c.m, p, unsafe.Pointer(&m2))
		if swapped {
			break
		}
	}
}

func (c *cache[T, V]) GetOrSet(key T, load func() (v V)) (v V) {
	m := *(*map[T]V)(atomic.LoadPointer(&c.m))
	v, ok := m[key]
	if !ok {
		v = load()
		m2 := make(map[T]V, len(m)+10)
		m2[key] = v
		for {
			p := atomic.LoadPointer(&c.m)
			m = *(*map[T]V)(p)
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

//SlicePool 用于分配 Slice，避免频繁 grow
type SlicePool struct {
	sync.Pool
	MakeN int
	NewN  uint32
	size  int
	typ   reflect.Type
}

func (p *SlicePool) SetMakeN(cap int) {
	atomic.StoreUint32(&p.NewN, 1)
	if p.MakeN < cap {
		p.MakeN = cap
	}
}

func (s *SlicePool) Grow(pHeader *reflect.SliceHeader) {
	l := pHeader.Cap / s.size
	// c := l * 2 //
	c := l + l/2
	if c == 0 {
		c = 4
	}
	s.SetMakeN(c)

	v := reflect.MakeSlice(s.typ, l, c)
	p := reflectValueToPointer(&v)
	pH := (*reflect.SliceHeader)(p)
	// pH.Len = pHeader.Len
	pH.Cap = pH.Cap * s.size

	// copy(*(*[]uint8)(p), *(*[]uint8)(unsafe.Pointer(pHeader)))
	_ = append((*(*[]uint8)(unsafe.Pointer(pHeader)))[:0], *(*[]uint8)(p)...)

	pHeader.Cap = pH.Cap
	pHeader.Data = pH.Data
}

func NewSlicePool(typ, sonTyp reflect.Type) (s *SlicePool) {
	s = &SlicePool{}
	s.typ = typ
	s.size = int(sonTyp.Size())
	s.New = func() any {
		n := s.MakeN
		if n < 4 {
			n = 4
		}
		if atomic.AddUint32(&s.NewN, 1)%4000 == 1 {
			n := s.MakeN / 2
			if n < 4 {
				n = 4
			}
			s.MakeN = n
		}
		v := reflect.MakeSlice(typ, s.MakeN, s.MakeN)
		p := reflectValueToPointer(&v)
		pH := (*reflect.SliceHeader)(p)
		pH.Len = pH.Len * s.size
		pH.Cap = pH.Cap * s.size
		return p
	}
	return
}
