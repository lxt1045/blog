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

var (
	cacheStructTagInfo = newCache[uint32, *TagInfo]()
)

// 获取 string 的起始地址
func strToUintptr(p string) uintptr {
	return *(*uintptr)(unsafe.Pointer(&p))
}

func LoadTagNode(typ reflect.Type, hash uint32) (*TagInfo, error) {
	tag, ok := cacheStructTagInfo.Get(hash)
	if ok {
		return tag, nil
	}
	return LoadTagNodeSlow(typ, hash)
}
func LoadTagNodeSlow(typ reflect.Type, hash uint32) (*TagInfo, error) {
	ti, err := NewStructTagInfo(typ, false, nil)
	if err != nil {
		return nil, err
	}
	n := (*TagInfo)(ti)
	n.Builder.Build()

	N := (8 * 1024 / n.TypeSize) + 1
	l := N * int(n.TypeSize)
	n.BPool.New = func() any {
		p := unsafe_NewArray(n.Builder.goType, N)
		// pH := &reflect.SliceHeader{
		pH := &SliceHeader{
			Data: p,
			Len:  l,
			Cap:  l,
		}
		return (*[]uint8)(unsafe.Pointer(pH))
	}
	cacheStructTagInfo.Set(hash, n)
	return n, nil
}

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

func (s *SlicePool) Grow(pHeader *SliceHeader) {
	l := pHeader.Cap / s.size
	// c := l * 2 //
	c := l + l/2
	if c == 0 {
		c = 4
	}
	s.SetMakeN(c)

	v := reflect.MakeSlice(s.typ, l, c)
	p := reflectValueToPointer(&v)
	pH := (*SliceHeader)(p)
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
		pH := (*SliceHeader)(p)
		pH.Len = pH.Len * s.size
		pH.Cap = pH.Cap * s.size
		return p
	}
	return
}

/*
 先试试这个；
 然后试试 &map[string]interface{} 的时候, 先 sync.pool 获取一个：
 type pool struct{
	strs []string
	efaces []interface{}
	maps []map[string]interface{} 的底层 array 的列表
	floats []float64
	...
 }
 执行过程中直接 make，不需要加锁；
 完成后再 put 回去
//*/

type sliceNode[T any] struct {
	s   []T
	idx uint32 // atomic
}
type Batch[T any] struct {
	pool unsafe.Pointer // *sliceNode[T]
}

func NewBatch[T any]() *Batch[T] {
	sn := &sliceNode[T]{
		s:   make([]T, 1024*1024),
		idx: 0,
	}
	return &Batch[T]{
		pool: unsafe.Pointer(sn),
	}
}

func GetStr(b *Batch[string]) (p *string) {
	sn := (*sliceNode[string])(atomic.LoadPointer(&b.pool))
	idx := atomic.AddUint32(&sn.idx, 1)
	if int(idx) >= len(sn.s) {
		return b.Make()
	}
	p = &sn.s[idx-1]
	return
}

func (b *Batch[T]) Get() (p *T) {
	sn := (*sliceNode[T])(atomic.LoadPointer(&b.pool))
	// sn := (*sliceNode[T])(b.pool)
	idx := atomic.AddUint32(&sn.idx, 1)
	if int(idx) > len(sn.s) {
		sn = &sliceNode[T]{
			s:   make([]T, 1024*1024),
			idx: 1,
		}
		atomic.StorePointer(&b.pool, unsafe.Pointer(sn))
		p = &sn.s[0]
		return
	}
	p = &sn.s[idx-1]
	return
}
func (b *Batch[T]) Make() (p *T) {
	sn := &sliceNode[T]{
		s:   make([]T, 1024*1024),
		idx: 1,
	}
	atomic.StorePointer(&b.pool, unsafe.Pointer(sn))
	p = &sn.s[0]
	return
}

const N = 1024

// /*

var (
// bsCache        = NewSliceCache[[]byte](N)
// strCache       = NewSliceCache[string](N)
// interfaceCache = NewSliceCache[interface{}](N)
// mapCache       = NewSliceCache[map[string]interface{}](N)
)

type sliceCache[T any] struct {
	c []T
	i int32
}

type SliceCache[T any] struct {
	c      unsafe.Pointer
	n      int32
	ch     chan *sliceCache[T]
	makeTs func() (ts []T)
}

//go:inline
func (s *SliceCache[T]) Get() (p *T) {
	c := (*sliceCache[T])(atomic.LoadPointer(&s.c))
	i := atomic.AddInt32(&c.i, 1)
	if i <= s.n {
		p = &c.c[i-1]
		return
	}
	c = <-s.ch
	// c = &sliceCache[T]{
	// 	c: s.makeTs(),
	// 	i: 1,
	// }
	p = &c.c[0]
	atomic.StorePointer(&s.c, unsafe.Pointer(c))
	return
}
func (s *SliceCache[T]) GetT() (p T) {
	c := (*sliceCache[T])(atomic.LoadPointer(&s.c))
	i := atomic.AddInt32(&c.i, 1)
	if i <= s.n {
		p = c.c[i-1]
		return
	}
	c = <-s.ch
	p = c.c[0]
	atomic.StorePointer(&s.c, unsafe.Pointer(c))
	return
}
func (s *SliceCache[T]) GetP() (p unsafe.Pointer) {
	c := (*sliceCache[T])(atomic.LoadPointer(&s.c))
	i := atomic.AddInt32(&c.i, 1)
	if i <= s.n {
		p = unsafe.Pointer(&c.c[i-1])
		return
	}
	c = <-s.ch
	p = unsafe.Pointer(&c.c[0])
	atomic.StorePointer(&s.c, unsafe.Pointer(c))
	return
}
func NewSliceCache[T any](n int32) (s *SliceCache[T]) {
	var t T
	var value interface{} = t
	_, bMap := value.(map[string]interface{})

	makeTs := func() (ts []T) {
		ts = make([]T, n)
		return
	}
	if bMap {
		makeTs = func() (ts []T) {
			m := make([]map[string]interface{}, n)
			for i := range m {
				m[i] = make(map[string]interface{})
			}
			p := (*[]map[string]interface{})(unsafe.Pointer(&ts))
			*p = m
			return
		}
	}

	s = &SliceCache[T]{
		c: unsafe.Pointer(&sliceCache[T]{
			c: makeTs(),
		}),
		n: int32(n),
	}
	s.makeTs = makeTs
	s.ch = make(chan *sliceCache[T], 8)

	{
		sc := make([]sliceCache[T], cap(s.ch))
		for i := range sc {
			sc[i].i = 1
			sc[i].c = makeTs()
			s.ch <- &sc[i]
		}
	}
	N := 1
	if bMap {
		//单独测试 cache 时，单协程生产速度太慢
		// N = 48
	}
	for i := 0; i < N; i++ {
		go func() {
			for {
				sc := make([]sliceCache[T], cap(s.ch))
				for i := range sc {
					sc[i].i = 1
					sc[i].c = makeTs()
					s.ch <- &sc[i]
				}
			}
		}()
	}
	return
}
