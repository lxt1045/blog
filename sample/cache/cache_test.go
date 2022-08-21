package cache

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

var j0x = `{
	"id": "7028151259660092936",
    "name": {
        "zh-CN": "陈三",
        "en-US": ""
    },
    "avatar": {
        "avatar": "https://7b90749e-a6be-4a86-9f9b-7dce3d2ecf5g~?image_size=noop&cut_type=&quality=&format=png&sticker_format=.webp",
        "avatar72": "https://7b90749e-a6be-4a86-9f9b-7dce3d2ecf5g~?image_size=72x72&cut_type=&quality=&format=png&sticker_format=.webp",
        "avatar240": "https://7b90749e-a6be-4a86-9f9b-7dce3d2ecf5g~?image_size=240x240&cut_type=&quality=&format=png&sticker_format=.webp",
        "avatar640": "https://7b90749e-a6be-4a86-9f9b-7dce3d2ecf5g~?image_size=640x640&cut_type=&quality=&format=png&sticker_format=.webp"
    },
    "department": {
        "id": "6826585686905406989",
        "name": {
            "zh-CN": "研发部",
            "en-US": "RD Department"
        }
    },
    "email": "",
    "mobile": {
        "phone": "18030838810",
        "code": "86"
    },
    "status": {
        "accountStatus": false,
        "employmentStatus": false,
        "registerStatus": false
    },
    "employeeType": {
        "id": 1,
        "name": {
            "zh-CN": "正式",
            "en-US": "Regular"
        },
        "active": true
    },
    "isAdmin": false,
    "isLeader": false,
    "isManager": false,
    "isAppManager": false,
    "departmentList": {
        "id": "6826585686905406989",
        "name": {
            "zh-CN": "研发部",
            "en-US": "RD Department"
        }
    }
}`

func TestBytesIdx(t *testing.T) {
	keys := []string{
		"mobile", "phone", "code", "isLeader", "isManager", "isAppManager", "departmentList", "id", "name", "id", "avatar", "avatar", "avatar72", "avatar240", "avatar640", "email", "employeeType", "name", "active", "id", "isAdmin", "name", "zh-CN", "en-US", "department", "id", "name", "status", "accountStatus", "employmentStatus", "registerStatus",
	}
	bytes := make([]byte, 0, 1024)
	nodes := make([]*Node, 0, 1024)
	for i, k := range keys {
		if j := strings.Index(string(bytes), k); j >= 0 {
			continue
		}
		idx := len(bytes)
		if len(nodes) <= idx {
			nodes = append(nodes, make([]*Node, 1+idx*2-len(nodes))...)
		}
		nodes[idx] = &Node{
			k: fmt.Sprintf("src/github.com/lxt1045/blog/sample:%s", k),
			v: fmt.Sprintf("json.Test:%v", i),
		}
		bytes = append(bytes, k...)
	}

	t.Run("index", func(t *testing.T) {
		for _, k := range keys {
			if j := strings.Index(string(bytes), k); j >= 0 {
				t.Logf("key:%v, node:%+v", k, *nodes[j])
			}
		}
	})

	t.Run("p", func(t *testing.T) {

		m := map[string]interface{}{}
		err := json.Unmarshal([]byte(j0x), &m)
		if err != nil {
			t.Fatal(err)
		}
		keys := []string{}
		for k, v := range m {
			for _, kk := range keys {
				if kk == k {
					continue
				}
			}
			keys = append(keys, k)
			if mm, ok := v.(map[string]interface{}); ok {
				for k := range mm {
					for _, kk := range keys {
						if kk == k {
							continue
						}
					}
					keys = append(keys, k)
				}
			}
		}
		for _, k := range keys {
			fmt.Printf(` "%s",`, k)
		}
		t.Logf("keys:%v", keys)

	})
}

func BenchmarkIdx(b *testing.B) {
	keys := []string{
		"mobile", "phone", "code", "isLeader", "isManager", "isAppManager", "departmentList", "id", "name", "avatar", "avatar72", "avatar240", "avatar640", "email", "employeeType", "name", "active", "isAdmin", "zh-CN", "en-US", "department", "status", "accountStatus", "employmentStatus", "registerStatus",
	}
	bytes := make([]byte, 0, 1024)
	nodes := make([]*Node, 0, 1024)
	for i, k := range keys {
		if j := strings.Index(string(bytes), k); j >= 0 {
			continue
		}
		idx := len(bytes)
		if len(nodes) <= idx {
			nodes = append(nodes, make([]*Node, 1+idx*2-len(nodes))...)
		}
		nodes[idx] = &Node{
			k: fmt.Sprintf("src/github.com/lxt1045/blog/sample:%s", k),
			v: fmt.Sprintf("json.Test:%v", i),
		}
		bytes = append(bytes, k...)
	}
	strs := string(bytes)
	b.Logf("len(bytes):%d", len(bytes))
	b.Logf("registerStatus:%d", strings.Index(strs, "isManager"))
	b.Logf("id:%d", strings.Index(strs, "id"))
	b.Logf("registerStatus:%d", strings.Index(strs, "registerStatus"))

	b.Run("1", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if j := strings.Index(strs, "mobile"); j < 0 {
				b.Fatal("errors!")
			}
		}
	})
	b.Run("2", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if j := strings.Index(strs, "isManager"); j < 0 {
				b.Fatal("errors!")
			}
		}
	})
	b.Run("3", func(b *testing.B) {
		strs := string(bytes)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if j := strings.Index(strs, "id"); j < 0 {
				b.Fatal("errors!")
			}
		}
	})
	b.Run("4", func(b *testing.B) {
		strs := string(bytes)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if j := strings.Index(strs, "registerStatus"); j < 0 {
				b.Fatal("errors!")
			}
		}
	})
	// return
}

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
