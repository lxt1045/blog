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

package cache

import (
	"runtime"
	"sync/atomic"
	"unsafe"
)

func shiftLeft(x, y int) int {
	x = x << y
	return x
}

var True = true
var False = false

func InSpaceQ(b byte, bb bool) bool {
	False = bb
	return True
}

//go:linkname InSpaceQ1 InSpaceQ
func InSpaceQ1(b byte) bool

func test(x, y int) (a [16]byte, b, f bool) {
	return SpaceBytes[x], True, False
}

func fillBytes16(b byte) (bs [16]byte) {
	for i := 0; i < 16; i++ {
		bs[i] = b
	}
	return
}

var SpaceBytes = [8][16]byte{
	fillBytes16('\t'),
	fillBytes16('\n'),
	fillBytes16('\v'),
	fillBytes16('\f'),
	fillBytes16('\r'),
	fillBytes16(' '),
	fillBytes16(0x85),
	fillBytes16(0xA0),
}

//go:linkname runtime_procPin runtime.procPin
func runtime_procPin() int

//go:linkname runtime_procUnpin runtime.procUnpin
func runtime_procUnpin()

type Cache[K comparable, V any] struct {
	noCopy noCopy

	local     unsafe.Pointer // local fixed-size per-P pool, actual type is [P]poolLocal
	localSize uint32         // size of the local array
	New       func(K) V
}

func indexLocal[K comparable, V any](l unsafe.Pointer, i int) *map[K]V {
	size := unsafe.Sizeof([1]V{})
	lp := unsafe.Pointer(uintptr(l) + uintptr(i)*(size))
	return (*map[K]V)(lp)
}

func (c *Cache[K, V]) Get(key K) V {
	pid := runtime_procPin()
again:
	size := atomic.LoadUint32(&c.localSize)
	l := atomic.LoadPointer(&c.local) // load-consume
	if uintptr(pid) < uintptr(size) {
		m := indexLocal[K, V](l, pid) // pid 按顺序增长
		v, ok := (*m)[key]
		if !ok {
			v = c.New(key)
			(*m)[key] = v
		}
		runtime_procUnpin()
		return v
	}

	// If GOMAXPROCS changes between GCs, we re-allocate the array and lose the old one.
	sizeNew := uint32(runtime.GOMAXPROCS(0))
	localNew := make([]map[K]V, sizeNew)
	if size > 0 {
		copy(localNew[:size], (*[1 << 31]map[K]V)(l)[:])
	}
	if c.New != nil {
		for i := size; i < sizeNew; i++ {
			localNew[i] = make(map[K]V)
		}
	}

	swapped := atomic.CompareAndSwapPointer(&c.local, l, unsafe.Pointer(&localNew[0]))
	if swapped {
		atomic.StoreUint32(&c.localSize, sizeNew)
	}
	goto again
}

type noCopy struct{}

// Lock is a no-op used by -copylocks checker from `go vet`.
func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}
