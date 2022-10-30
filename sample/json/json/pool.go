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
	strEface    = UnpackEface("")
	isliceEface = UnpackEface([]interface{}{})
	hmapimp     = func() *hmap {
		m := make(map[string]interface{})
		return *(**hmap)(unsafe.Pointer(&m))
	}()
	mapGoType = func() *maptype {
		m := make(map[string]interface{})
		typ := reflect.TypeOf(m)
		return (*maptype)(unsafe.Pointer(UnpackType(typ)))
	}()

	chbsPool = func() (ch chan *[]byte) {
		ch = make(chan *[]byte, 4)
		go newBytes(ch)
		return
	}()
)

func newBytes(ch chan *[]byte) {
	for {
		s := make([]byte, 0, bsPoolN)
		ch <- &s
	}
}
func newMapArray(ch chan *[]byte) {
	N := 1 << 20
	size := int(mapGoType.bucket.Size)
	N = N / size
	cap := N * size
	for {
		p := unsafe_NewArray(mapGoType.bucket, N)
		s := &SliceHeader{
			Data: p,
			Len:  cap,
			Cap:  cap,
		}
		ch <- (*[]byte)(unsafe.Pointer(s))
	}
}

var (
	poolSliceInterface = sync.Pool{New: func() any {
		return make([]interface{}, 1024)
	}}

	pairPool = sync.Pool{
		New: func() any {
			s := make([]pair, 0, 128)
			return &s
		},
	}
	bsPool = sync.Pool{New: func() any {
		return <-chbsPool
		// s := make([]byte, 0, bsPoolN)
		// return &s
	}}
	poolMapArrayInterface = func() sync.Pool {
		ch := make(chan *[]byte, 4) // runtime.GOMAXPROCS(0))
		// go func() {
		// 	for {
		// 		N := 1 << 20
		// 		p := unsafe_NewArray(mapGoType.bucket, N)
		// 		s := &SliceHeader{
		// 			Data: p,
		// 			Len:  N * int(mapGoType.bucket.Size),
		// 			Cap:  N * int(mapGoType.bucket.Size),
		// 		}
		// 		ch <- (*[]byte)(unsafe.Pointer(s))
		// 	}
		// }()
		go newMapArray(ch)
		return sync.Pool{New: func() any {
			return <-ch
		}}
	}()

	cacheStructTagInfo = NewRCU[uint32, *TagInfo]()
	strPool            = NewBatch[string]()
	islicePool         = NewBatch[[]interface{}]()
	imapPool           = NewBatch[hmap]()
)

const (
	bsPoolN = 1 << 20
	batchN  = 1 << 12
)

type pair struct {
	k string
	v interface{}
}

// 获取 string 的起始地址
func strToUintptr(p string) uintptr {
	return *(*uintptr)(unsafe.Pointer(&p))
}

func LoadTagNode(v reflect.Value, hash uint32) (*TagInfo, error) {
	tag, ok := cacheStructTagInfo.Get(hash)
	if ok {
		return tag, nil
	}
	return LoadTagNodeSlow(v, hash)
}
func LoadTagNodeSlow(v reflect.Value, hash uint32) (*TagInfo, error) {
	typ := v.Type()
	ti, err := NewStructTagInfo(typ, false /*noBuildmap*/, nil)
	if err != nil {
		return nil, err
	}
	ti.Builder.Init()
	cacheStructTagInfo.Set(hash, ti)
	return ti, nil
}

//RCU 依据 Read Copy Update 原理实现
type RCU[T uintptr | uint32 | string | int, V any] struct {
	m unsafe.Pointer
}

func NewRCU[T uintptr | uint32 | string | int, V any]() (c RCU[T, V]) {
	m := make(map[T]V, 1)
	c.m = unsafe.Pointer(&m)
	return
}

func (c *RCU[T, V]) Get(key T) (v V, ok bool) {
	m := *(*map[T]V)(atomic.LoadPointer(&c.m))
	v, ok = m[key]
	return
}

func (c *RCU[T, V]) Set(key T, v V) {
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

func (c *RCU[T, V]) GetOrSet(key T, load func() (v V)) (v V) {
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

type sliceNode[T any] struct {
	s   []T
	idx uint32 // atomic
}
type Batch[T any] struct {
	pool unsafe.Pointer // *sliceNode[T]
	sync.Mutex
}

func NewBatch[T any]() *Batch[T] {
	sn := &sliceNode[T]{
		s:   nil, // make([]T, batchN),
		idx: 0,
	}
	return &Batch[T]{
		pool: unsafe.Pointer(sn),
	}
}

func BatchGet[T any](b *Batch[T]) *T {
	sn := (*sliceNode[T])(atomic.LoadPointer(&b.pool))
	idx := atomic.AddUint32(&sn.idx, 1)
	if int(idx) <= len(sn.s) {
		return &sn.s[idx-1]
	}
	return b.Make()
}

func (b *Batch[T]) Get() *T {
	sn := (*sliceNode[T])(atomic.LoadPointer(&b.pool))
	idx := atomic.AddUint32(&sn.idx, 1)
	if int(idx) <= len(sn.s) {
		return &sn.s[idx-1]
	}
	return b.Make()
}

func (b *Batch[T]) GetN(n int) *T {
	sn := (*sliceNode[T])(atomic.LoadPointer(&b.pool))
	idx := atomic.AddUint32(&sn.idx, uint32(n))
	if int(idx) <= len(sn.s) {
		return &sn.s[int(idx)-n]
	}
	return b.MakeN(n)

}

func (b *Batch[T]) Make() (p *T) {
	b.Lock()
	defer b.Unlock()
	sn := (*sliceNode[T])(atomic.LoadPointer(&b.pool))
	idx := atomic.AddUint32(&sn.idx, 1)
	if int(idx) <= len(sn.s) {
		p = &sn.s[idx-1]
		return
	}
	sn = &sliceNode[T]{
		s:   make([]T, batchN),
		idx: 1,
	}
	atomic.StorePointer(&b.pool, unsafe.Pointer(sn))
	p = &sn.s[0]
	return
}

func (b *Batch[T]) MakeN(n int) (p *T) {
	b.Lock()
	defer b.Unlock()
	sn := (*sliceNode[T])(atomic.LoadPointer(&b.pool))
	idx := atomic.AddUint32(&sn.idx, 1)
	if int(idx) <= len(sn.s) {
		p = &sn.s[idx-1]
		return
	}
	sn = &sliceNode[T]{
		s:   make([]T, batchN),
		idx: uint32(n),
	}
	atomic.StorePointer(&b.pool, unsafe.Pointer(sn))
	p = &sn.s[0]
	return
}
