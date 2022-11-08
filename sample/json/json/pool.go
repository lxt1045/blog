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
	ti, err := NewStructTagInfo(typ, nil /*stackBuilder*/, nil, nil /* ancestors*/)
	if err != nil {
		return nil, err
	}

	ti.buildChildMap()

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

/*
  Pool 和 store.Pool 一起，一次 Unmashal 调用，补充一次 pool 填满，此次执行期间，不会其他进程争抢；
  结束后再归还，剩下的下次还可以继续使用
*/
type sliceNode[T any] struct {
	s   []T
	idx uint32 // atomic
}
type Batch[T any] struct {
	pool  unsafe.Pointer // *sliceNode[T]
	gStrs [][]T
	cond  *sync.Cond // 唤醒通知
	sync.Mutex
}

func NewBatch[T any]() *Batch[T] {
	sn := &sliceNode[T]{
		s:   nil, // make([]T, batchN),
		idx: 0,
	}
	ret := &Batch[T]{
		pool:  unsafe.Pointer(sn),
		gStrs: make([][]T, 4),
	}
	ret.cond = sync.NewCond(&ret.Mutex)
	go func() {
		count := 0
		for {
			count = 0
			for i := range ret.gStrs {
				ret.cond.L.Lock()
				if len(ret.gStrs[i]) == 0 {
					count++
					ret.gStrs[i] = make([]T, batchN)
				}
				ret.cond.L.Unlock()
			}
			if count == 0 {
				ret.cond.L.Lock()
				ret.cond.Wait()
				ret.cond.L.Unlock()
			}
		}
	}()

	return ret
}

func BatchGet[T any](b *Batch[T]) *T {
	sn := (*sliceNode[T])(atomic.LoadPointer(&b.pool))
	idx := atomic.AddUint32(&sn.idx, 1)
	if int(idx) <= len(sn.s) {
		return &sn.s[idx-1]
	}
	return b.Make1()
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
	b.cond.L.Lock()
	defer b.cond.L.Unlock()

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
func (b *Batch[T]) Make1() (p *T) {
	b.cond.L.Lock()
	defer func() {
		b.cond.L.Unlock()
		b.cond.Broadcast()
	}()
	sn := (*sliceNode[T])(atomic.LoadPointer(&b.pool))
	idx := atomic.AddUint32(&sn.idx, 1)
	if int(idx) <= len(sn.s) {
		p = &sn.s[idx-1]
		return
	}
	strs := []T{}
	for i := range b.gStrs {
		if len(b.gStrs) != 0 {
			strs = b.gStrs[i]
			b.gStrs[i] = nil
			break
		}
	}
	if len(strs) == 0 {
		strs = make([]T, batchN)
	}
	sn = &sliceNode[T]{
		s:   strs,
		idx: 1,
	}
	atomic.StorePointer(&b.pool, unsafe.Pointer(sn))
	p = &sn.s[0]
	return
}

func (b *Batch[T]) MakeN(n int) (p *T) {
	b.cond.L.Lock()
	defer func() {
		b.cond.L.Unlock()
		b.cond.Broadcast()
	}()
	sn := (*sliceNode[T])(atomic.LoadPointer(&b.pool))
	idx := atomic.AddUint32(&sn.idx, 1)
	if int(idx) <= len(sn.s) {
		p = &sn.s[idx-1]
		return
	}
	if n > batchN {
		strs := make([]T, n)
		return &strs[0]
	}
	strs := []T{}
	for i := range b.gStrs {
		if len(b.gStrs) != 0 {
			strs = b.gStrs[i]
			b.gStrs[i] = nil
			break
		}
	}
	if len(strs) == 0 {
		strs = make([]T, batchN)
	}
	sn = &sliceNode[T]{
		s:   make([]T, batchN), // <-b.ch, // make([]T, batchN),
		idx: uint32(n),
	}
	atomic.StorePointer(&b.pool, unsafe.Pointer(sn))
	p = &sn.s[0]
	return
}

type Store struct {
	obj unsafe.Pointer
	tag *TagInfo
}
type PoolStore struct {
	obj  unsafe.Pointer
	pool *dynamicPool
	tag  *TagInfo
}

type dynamicPool struct {
	structPool  unsafe.Pointer
	noscanPool  []byte   // 不含指针的
	intsPool    []int    // 不含指针的
	stringPool  []string // 不含指针的
	pointerPool []unsafe.Pointer
	mapPool     *sync.Pool
	// scanPool    []Struct{} // 含指针的，GC 需要标注、扫描
	// 含指针的 struct 需要单独处理
	/*
		type X struct{
			dynamicPool
			structSlice []Struct // reflcet 动态生成对象
		}
	*/
}

func (ps PoolStore) Idx(idx uintptr) (p unsafe.Pointer) {
	p = pointerOffset(ps.pool.structPool, idx)
	*(*unsafe.Pointer)(ps.obj) = p
	return
}

func (ps PoolStore) GetNoscan() []byte {
	pool := ps.pool.noscanPool
	ps.pool.noscanPool = nil
	if cap(pool)-len(pool) > 0 {
		return pool
	}
	l := 8 * 1024
	return make([]byte, 0, l)
}

func (ps PoolStore) SetNoscan(pool []byte) {
	ps.pool.noscanPool = pool
	return
}

func GrowBytes(in []byte, need int) []byte {
	l := need + len(in)
	if l <= cap(in) {
		return in[:l]
	}
	if l < 8*1024 {
		l = 8 * 1024
	} else {
		l *= 2
	}
	out := make([]byte, 0, l)
	out = append(out, in...)
	return out[:l]
}

func (ps PoolStore) GetStrings() []string {
	pool := ps.pool.stringPool
	ps.pool.stringPool = nil
	if cap(pool)-len(pool) > 0 {
		return pool
	}
	l := 1024
	return make([]string, 0, l)
}

func (ps PoolStore) SetStrings(strs []string) {
	ps.pool.stringPool = strs
	return
}

func GrowStrings(in []string, need int) []string {
	l := need + len(in)
	if l <= cap(in) {
		return in[:l]
	}
	if l < 1024 {
		l = 1024
	} else {
		l *= 2
	}
	out := make([]string, 0, l)
	out = append(out, in...)
	return out[:l]
}

func (ps PoolStore) GetInts() []int {
	pool := ps.pool.intsPool
	ps.pool.intsPool = nil
	if cap(pool)-len(pool) > 0 {
		return pool
	}
	l := 1024
	return make([]int, 0, l)
}

func (ps PoolStore) SetInts(strs []int) {
	if cap(strs)-len(strs) > 4 {
		ps.pool.intsPool = strs
		return
	}
}

func GrowInts(in []int) []int {
	l := 1 + len(in)
	if l <= cap(in) {
		return in[:l]
	}
	if l < 1024 {
		l = 1024
	} else {
		l *= 2
	}
	out := make([]int, 0, l)
	out = append(out, in...)
	return out[:l]
}
