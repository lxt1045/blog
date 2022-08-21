package cache

import (
	"fmt"
	"sync"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

type Node struct {
	k, v string
}
type Node2 struct {
	k, v uintptr
}

func strToUintptr(p string) uintptr {
	return *(*uintptr)(unsafe.Pointer(&p))
}

var ckey string = "src/github.com/lxt1045/blog/sample"

func TestPre(t *testing.T) {
	t.Run("p", func(t *testing.T) {
		p1 := strToUintptr(ckey)
		ckey1 := ckey
		ckey2 := ckey1
		p2 := strToUintptr(ckey1)
		p3 := strToUintptr(ckey2)

		t.Logf("1:%d, 2:%d, 3:%d", p1, p2, p3)
	})
}
func BenchmarkMap(b *testing.B) {
	m1 := make(map[string]Node)
	m2 := make(map[uintptr]Node2)
	var N = 10240
	for i := 0; i < N; i++ {
		key := fmt.Sprintf("json.Map%d", i)
		value := "src/github.com/lxt1045/blog/sample"
		m1[key] = Node{
			key, value,
		}

		pkey := strToUintptr(key)
		pvalue := strToUintptr(value)
		m2[pkey] = Node2{
			pkey, pvalue,
		}
	}
	key := "src/github.com/lxt1045/blog/sample"
	value := "json.Map"
	b.Run("m1", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if nGet, ok := m1[key]; !ok {
				m1[key] = Node{
					key, value,
				}
			} else if value != nGet.v {
				if _, ok := m1[key+value]; !ok {
					m1[key+value] = Node{
						key, value,
					}
				}
			}
		}
	})
	b.Run("m1-1", func(b *testing.B) {
		m1 := make(map[string]Node)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if nGet, ok := m1[key]; !ok {
				m1[key] = Node{
					key, value,
				}
			} else if value != nGet.v {
				if _, ok := m1[key+value]; !ok {
					m1[key+value] = Node{
						key, value,
					}
				}
			}
		}
	})
	pkey := strToUintptr(key)
	pvalue := strToUintptr(value)
	b.Run("m2", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if nGet, ok := m2[pkey]; !ok {
				m2[pkey] = Node2{
					pkey, pvalue,
				}
			} else if pvalue != nGet.v {
				if _, ok := m2[pvalue]; !ok {
					m2[pvalue] = Node2{
						pkey, pvalue,
					}
				}
			}
		}
	})
	// return
}

func TestCache(t *testing.T) {
	t.Run("Cache", func(t *testing.T) {
		cache := Cache[uintptr, uintptr]{
			New: func(k uintptr) (v uintptr) {
				return k
			},
		}
		k := uintptr(100)
		v := cache.Get(100)
		assert.Equal(t, v, k)
	})
}

/*
go test -benchmem -run=^$ -bench "^(BenchmarkCache)$" github.com/lxt1045/errors -count=1 -v -cpuprofile cpu.prof -c
go tool pprof ./errors.test cpu.prof
*/

func BenchmarkCache(b *testing.B) {
	cache := Cache[int, int]{
		New: func(k int) (v int) {
			return k
		},
	}
	N := 10240
	for i := 0; i < N; i++ {
		cache.Get(i)
	}
	b.Run("cache", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			cache.Get(i % N)
		}
		b.StopTimer()
	})
	// return

	m := map[int]int{}
	for i := 0; i < N; i++ {
		m[i] = i
	}
	b.Run("map", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, ok := m[i%N]
			if !ok {
				m[i%N] = i
			}
		}
		b.StopTimer()
	})
	var lock sync.RWMutex
	b.Run("map+RWMutex", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			lock.RLock()
			_, ok := m[i%N]
			lock.RUnlock()
			if !ok {
				lock.Lock()
				m[i%N] = i
				lock.Unlock()
			}
		}
		b.StopTimer()
	})
	b.Run("RWMutex", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			lock.RLock()
			lock.RUnlock()
		}
		b.StopTimer()
	})
}
