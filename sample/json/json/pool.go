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

//cache 依据 RCU(Read Copy Update) 原理实现
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
