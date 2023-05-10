package hashmap

import (
	"encoding/json"
	"fmt"
	"math/bits"
	"math/rand"
	"runtime"
	"testing"
	"time"
	"unsafe"
)

func Test_FirstBitIdx(t *testing.T) {
	idx := FirstBitIdx(0b010001000)
	t.Logf("idx:%v", idx)

	idx = runtime.FirstBitIdx(0b010001000)
	t.Logf("runtime-idx:%v", idx)

	idx = Ctz64(0b010001000)
	t.Logf("idx:%v", idx)

	idx = bits.Len64(0b010001000)
	t.Logf("Len64:%v", idx)

	idx = Len64_2(0b010001000)
	t.Logf("Len64_2:%v", idx)

	idx = sovTest(0b010001000)
	t.Logf("sovTest:%v", idx)

	idx = bits.LeadingZeros(0b010001000)
	t.Logf("LeadingZeros:%v", idx)

	t.Logf("deBruijn64ctz:%b", deBruijn64ctz)
}

func Benchmark_FirstBitIdx(b *testing.B) {

	b.Run("FirstBitIdx", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := int64(0); j < 1000; j++ {
				FirstBitIdx(int64(j))
			}
		}
	})

	b.Run("runtime.FirstBitIdx", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := int64(0); j < 1000; j++ {
				runtime.FirstBitIdx(int64(j))
			}
		}
	})

	b.Run("Len64_2", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := uint64(0); j < 1000; j++ {
				Len64_2(uint64(j))
			}
		}
	})
	b.Run("bits.Len64", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := uint64(0); j < 1000; j++ {
				bits.Len64(uint64(j))
			}
		}
	})

	b.Run("Ctz64", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := uint64(0); j < 1000; j++ {
				Ctz64(uint64(j))
			}
		}
	})

	x := 0
	_ = x
	b.Run("x ^= i", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := 0; j < 1000; j++ {
				x = j
			}
		}
	})

	b.Run("gogoprotobuf-sovTest", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := uint64(0); j < 1000; j++ {
				sovTest(uint64(j))
			}
		}
	})
	b.Run("gogoprotobuf-sovTest-noinline", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := uint64(0); j < 1000; j++ {
				sovTest1(j)
			}
		}
	})
}

type OperationHTTPResponse struct {
	HTTPCode *int32 `thrift:"HttpCode,1" json:"HttpCode,omitempty"`

	ContentType *string `thrift:"ContentType,2" json:"ContentType,omitempty"`

	ContentDisposition *string `thrift:"ContentDisposition,3" json:"ContentDisposition,omitempty"`

	ResponseBody []byte `thrift:"ResponseBody,4" json:"ResponseBody,omitempty"`

	Location *string `thrift:"Location,5" json:"Location,omitempty"`

	Cookie *string `thrift:"Cookie,6" json:"Cookie,omitempty"`

	BaseResp map[string]interface{} `thrift:"BaseResp,255" json:"BaseResp"`
}

func Test_JSON(t *testing.T) {
	str := `{
		"BaseResp": {
			"StatusCode": 0,
			"StatusMessage": "success"
		},
		"ContentType": "application/json",
		"HttpCode": 200,
		"ResponseBody": "eyJjb2RlIjowLCJkYXRhIjpbeyJJRCI6IjcyMDI1ODQ0NjExODIwMDkzNzYiLCJjb250YWN0TmFtZSI6Im1vIiwiY3JlYXRlVGltZSI6IjIwMjMtMDItMjEiLCJuYW1lIjoie1wiemgtQ05cIjpcIm1vIG1vIHByb1wiLFwiZW4tVVNcIjpcIlwifSIsInBob25lIjpudWxsfV0sInRva2VuIjoiIn0="
	}
	`
	m := OperationHTTPResponse{}
	err := json.Unmarshal([]byte(str), &m)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v", string(m.ResponseBody))
}

func getRandStr(l int) (str string) {
	const encodeStd = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	bs := make([]byte, l)
	for i := 0; i < l; i++ {
		idx := rand.Intn(len(encodeStd))
		bs[i] = encodeStd[idx]
	}
	return string(bs)
}

func Test_getPivotMask(t *testing.T) {
	keys := []string{}
	for i := 0; i < 256; i++ {
		keys = append(keys, getRandStr(rand.Intn(18)+2))
	}

	lens := []int{8, 16, 32, 64, 128, 256}
	for _, l := range lens {
		name := fmt.Sprintf("getPivotMask-%d", l)
		t.Run(name, func(t *testing.T) {
			m := getPivotMask(keys[:l])
			// t.Logf("l:%d, len:%d, m: %+v", l, len(m), m)
			t.Logf("l:%d, len:%d", l, len(m))
		})
	}
}

func Test_getPivotMask2(t *testing.T) {
	keys := []string{}
	// keys = []string{`"avatar"`, `"avatar72"`, `"avatar240"`, `"avatar640"`}
	for i := 0; i < 160; i++ {
		keys = append(keys, getRandStr(rand.Intn(10)+5))
	}

	lens := []int{
		8,
		16, 32, 64, 128,
	}
	for _, l := range lens {
		name := fmt.Sprintf("getPivotMask-%d", l)
		t.Run(name, func(t *testing.T) {
			m := getPivotMask(keys[:l])
			t.Logf("l:%d, len:%d, m: %+v", l, len(m), m)
		})
	}
}

/*
=== RUN   Test_logicalHash_new/getPivotMask-8
    /Users/bytedance/go/src/github.com/lxt1045/blog/sample/json/json/hash_map_test.go:87: l:8, len:3, m: [{iByte:0 mask:1 iBit:1 ratio:0} {iByte:1 mask:2 iBit:2 ratio:0} {iByte:2 mask:1 iBit:4 ratio:0}]
=== RUN   Test_logicalHash_new/getPivotMask-16
    /Users/bytedance/go/src/github.com/lxt1045/blog/sample/json/json/hash_map_test.go:87: l:16, len:5, m: [{iByte:0 mask:72 iBit:1 ratio:0} {iByte:0 mask:32 iBit:8 ratio:0} {iByte:0 mask:2 iBit:16 ratio:0} {iByte:1 mask:8 iBit:4 ratio:0} {iByte:3 mask:72 iBit:2 ratio:0}]
=== RUN   Test_logicalHash_new/getPivotMask-32
    /Users/bytedance/go/src/github.com/lxt1045/blog/sample/json/json/hash_map_test.go:87: l:32, len:6, m: [{iByte:0 mask:1 iBit:1 ratio:0} {iByte:0 mask:8 iBit:4 ratio:0} {iByte:0 mask:2 iBit:16 ratio:0} {iByte:3 mask:65 iBit:2 ratio:0} {iByte:4 mask:1 iBit:32 ratio:0} {iByte:5 mask:1 iBit:8 ratio:0}]
=== RUN   Test_logicalHash_new/getPivotMask-64
    /Users/bytedance/go/src/github.com/lxt1045/blog/sample/json/json/hash_map_test.go:87: l:64, len:9, m: [{iByte:0 mask:72 iBit:4 ratio:0} {iByte:0 mask:1 iBit:32 ratio:0} {iByte:1 mask:16 iBit:2 ratio:0} {iByte:1 mask:32 iBit:256 ratio:0} {iByte:2 mask:2 iBit:128 ratio:0} {iByte:3 mask:3 iBit:64 ratio:0} {iByte:4 mask:96 iBit:8 ratio:0} {iByte:4 mask:4 iBit:16 ratio:0} {iByte:10 mask:64 iBit:1 ratio:0}]
=== RUN   Test_logicalHash_new/getPivotMask-128
    /Users/bytedance/go/src/github.com/lxt1045/blog/sample/json/json/hash_map_test.go:87: l:128, len:11, m: [{iByte:0 mask:65 iBit:16 ratio:0} {iByte:0 mask:8 iBit:32 ratio:0} {iByte:0 mask:2 iBit:1024 ratio:0} {iByte:1 mask:1 iBit:1 ratio:0} {iByte:1 mask:2 iBit:256 ratio:0} {iByte:1 mask:32 iBit:8 ratio:0} {iByte:2 mask:16 iBit:2 ratio:0} {iByte:2 mask:2 iBit:512 ratio:0} {iByte:3 mask:32 iBit:64 ratio:0} {iByte:4 mask:2 iBit:128 ratio:0} {iByte:4 mask:32 iBit:4 ratio:0}]
*/

func Test_logicalHash_new(t *testing.T) {
	keys := []string{}
	// keys = []string{`"avatar"`, `"avatar72"`, `"avatar240"`, `"avatar640"`}
	for i := 0; i < 512; i++ {
		keys = append(keys, getRandStr(rand.Intn(10)+5))
	}

	lens := []int{
		8,
		16, 32, 64, 128,
	}
	for _, l := range lens {
		name := fmt.Sprintf("getPivotMask-%d", l)
		t.Run(name, func(t *testing.T) {
			m, _ := logicalHash(keys[:l])
			t.Logf("l:%d, len:%d, m: %+v", l, len(m), m)
			// t.Logf("l:%d, len:%d", l, len(m))

			exist := map[int]string{}
			for _, bs := range keys[:l] {
				idx := hash(bs, m)
				if _, ok := exist[idx]; !ok {
					exist[idx] = string(bs)
				} else {
					PrintKeys(keys[:l])
					t.Fatalf("insert_key:%s, idx:%d, exist:%+v", string(bs), idx, exist)
				}
			}
		})
	}
}

/*
go test -benchmem -run=^$ -bench ^Benchmark_getPivotMask$ github.com/lxt1045/blog/sample/json/json -count=1 -v -cpuprofile cpu.prof -c
go test -benchmem -run=^$ -bench ^Benchmark_getPivotMask$ github.com/lxt1045/blog/sample/json/json -count=1 -v -memprofile cpu.prof -c
go tool pprof ./json.test cpu.prof
*/
func Benchmark_getPivotMask(b *testing.B) {
	keys := []string{
		`"id"`,
		`"name"`,
		`"avatar"`,
		`"department"`,
		`"email"`,
		`"mobile"`,
		`"status"`,
		`"employeeType"`,
		`"isAdmin"`,
		`"isLeader"`,
		`"isManager"`,
		`"isAppManager"`,
		`"departmentList"`,
	}
	keys = nil
	// keys = []string{`"avatar"`, `"avatar72"`, `"avatar240"`, `"avatar640"`}
	for i := 0; i < 257; i++ {
		keys = append(keys, getRandStr(rand.Intn(18)+2))
	}
	// lens := []int{8, 16, 32, 64, 128, 180}
	lens := []int{230}
	// lens = []int{8, 64, 128}
	// lens = []int{128}
	for _, l := range lens {
		name := fmt.Sprintf("getPivotMask-%d", l)
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				getPivotMask(keys[:l])
			}
			b.StopTimer()
			b.SetBytes(int64(b.N))
		})
	}
}

func Test_divide(t *testing.T) {
	keys := []string{
		// `"id"`,
		// `"name"`,
		// `"avatar"`,
		// `"department"`,
		// `"email"`,
		// `"mobile"`,
		// `"status"`,
		// `"employeeType"`,
		// `"isAdmin"`,
		`"isLeader"`,
		`"isManager"`,
		`"isAppManager"`,
		`"departmentList"`,
	}
	// keys = []string{`"avatar"`, `"avatar72"`, `"avatar240"`, `"avatar640"`}
	ns := []mapNode{}
	for _, k := range keys {
		ns = append(ns, mapNode{
			K: k,
			V: &TagInfo{},
		})
	}

	m := buildTagMap(ns)
	t.Logf("m: %+v", m.String())

}
func Test_logicalHash(t *testing.T) {
	bsList := []string{
		"111",
		"122",
		"333",
		"433",
	}
	idxList, _ := logicalHash(bsList)
	t.Logf("%+v", idxList)

	PrintKeys(bsList)
}

// 01234567891123456789212345678931234567894123456789512345678961234567897123456789812345678991234567891012345678911123456789121234567
// 0010001001100001011101100110000101110100011000010111001000100010:"avatar"
// 00100010011000010111011001100001011101000110000101110010001101110011001000100010:"avatar72"
// 0010001001100001011101100110000101110100011000010111001000110010001101000011000000100010:"avatar240"
// 0010001001100001011101100110000101110100011000010111001000110110001101000011000000100010:"avatar640"
// [[69 61] [61 69] [63 69] [66 69]]
// [[61 69] [63 69] [66 69] [61 75]]
// [[63 69] [66 69] [61 75] [63 75]] //err
func Test_buildMap(t *testing.T) {
	// 2022/09/09 14:15:23 idxsRet---:[[8] [9] [10] [11] [12] [17] [18] [33]]
	// 2022/09/09 14:15:23 idxRet:[{iByte:1 mask:1 iBit:1}]
	// 2022/09/09 14:15:23 idxsRet---:[[8] [9] [10] [11] [12] [17] [18] [33]]
	// 2022/09/09 14:15:23 idxRet:[{iByte:1 mask:1 iBit:1}]
	// 2022/09/09 14:15:23 idxsRet---:[[60 76] [65 76] [68 76] [69 76] [73 76] [77 76] [58 81] [60 81]]
	// 2022/09/09 14:15:23 idxRet:[{iByte:7 mask:16 iBit:1} {iByte:9 mask:16 iBit:2}]
	ns := []mapNode{}
	keys := []string{
		// "111", "122", "333", "433",
		`"avatar"`, `"avatar72"`, `"avatar240"`, `"avatar640"`,
	}
	for _, k := range keys {
		ns = append(ns, mapNode{
			K: k,
			V: &TagInfo{},
		})
	}
	m := buildTagMap(ns)
	// t.Logf("%+v", m)
	t.Logf("%+v", m.String())

	for _, n := range ns {
		v := m.Get(n.K)
		if v == nil {
			t.Fatalf("[%s] not found", string(n.K))
		}
	}
}

/*
go test -benchmem -run=^$ -bench ^Benchmark_buildMap$ github.com/lxt1045/blog/sample/json/json -count=1 -v -cpuprofile cpu.prof -c
go test -benchmem -run=^$ -bench ^Benchmark_buildMap$ github.com/lxt1045/blog/sample/json/json -count=1 -v -memprofile cpu.prof -c
go tool pprof ./json.test cpu.prof
web
go build -gcflags=-m ./     2> ./gc.log
//   */
func Benchmark_buildMap(b *testing.B) {
	keys := []string{}
	for i := 0; i < 8; i++ {
		keys = append(keys, getRandStr(rand.Intn(6)+4))
	}
	ns := []mapNode{}
	for _, k := range keys {
		ns = append(ns, mapNode{
			K: k,
			V: &TagInfo{},
		})
	}
	m := buildTagMap(ns)
	b.Logf("%+v", m.String())
	for _, n := range ns {
		v := m.Get(n.K)
		if v == nil {
			b.Fatalf("[%s] not found", string(n.K))
		}
	}

	runtime.GC()

	b.Run("TagMapGetV4", func(b *testing.B) {
		NN := (b.N * 100) / len(keys)
		p := &m
		for i := 0; i < NN; i++ {
			for _, k := range keys {
				v := TagMapGetV4(p, k)
				if v == nil {
					b.Fatalf("[%s] not found", string(k))
				}
			}
		}
	})
	// return

	mm := make(map[string]mapNode)
	for _, n := range ns {
		mm[string(n.K)] = n
	}
	n := ns[len(ns)/2]
	for x := 0; x < 2; x++ {
		b.Run("Get", func(b *testing.B) {
			NN := (b.N * 100) / len(keys)
			for i := 0; i < NN; i++ {
				for _, k := range keys {
					v := m.Get(k)
					if v == nil {
						b.Fatalf("[%s] not found", string(n.K))
					}
				}
			}
		})
		b.Run("Get2", func(b *testing.B) {
			NN := (b.N * 100) / len(keys)
			for i := 0; i < NN; i++ {
				for _, k := range keys {
					v := m.Get2(k)
					if v == nil {
						b.Fatalf("[%s] not found", string(n.K))
					}
				}
			}
		})
		b.Run("Get4", func(b *testing.B) {
			NN := (b.N * 100) / len(keys)
			for i := 0; i < NN; i++ {
				for _, k := range keys {
					v := m.Get4(k)
					if v == nil {
						b.Fatalf("[%s] not found", string(n.K))
					}
				}
			}
		})
		b.Run("TagMapGetV", func(b *testing.B) {
			NN := (b.N * 100) / len(keys)
			p := &m
			for i := 0; i < NN; i++ {
				for _, k := range keys {
					v := TagMapGetV(p, k)
					if v == nil {
						b.Fatalf("[%s] not found", string(n.K))
					}
				}
			}
		})
		b.Run("TagMapGetV4", func(b *testing.B) {
			NN := (b.N * 100) / len(keys)
			p := &m
			for i := 0; i < NN; i++ {
				for _, k := range keys {
					v := TagMapGetV4(p, k)
					if v == nil {
						b.Fatalf("[%s] not found", string(n.K))
					}
				}
			}
		})
		// return
		b.Run("map", func(b *testing.B) {
			NN := (b.N * 100) / len(keys)
			for i := 0; i < NN; i++ {
				for _, k := range keys {
					v := mm[k]
					if v.V == nil {
						b.Fatalf("[%s] not found", string(n.K))
					}
				}
			}
		})
	}
}

func Benchmark_buildMap_one(b *testing.B) {
	keys := []string{}
	for i := 0; i < 16; i++ {
		keys = append(keys, getRandStr(rand.Intn(6)+4))
	}
	ns := []mapNode{}
	for _, k := range keys {
		ns = append(ns, mapNode{
			K: k,
			V: &TagInfo{},
		})
	}
	m := buildTagMap(ns)
	b.Logf("%+v", m.String())
	for _, n := range ns {
		v := m.Get(n.K)
		if v == nil {
			b.Fatalf("[%s] not found", string(n.K))
		}
	}
	mm := make(map[string]mapNode)
	mm2 := make(map[string]mapNode, 2*len(ns))
	mm4 := make(map[string]mapNode, 4*len(ns))
	mm8 := make(map[string]mapNode, 8*len(ns))
	for _, n := range ns {
		mm[string(n.K)] = n
		mm2[string(n.K)] = n
		mm4[string(n.K)] = n
		mm8[string(n.K)] = n
	}

	runtime.GC()

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })

	for _, k := range keys[:1] {
		_ = k
		b.Run("TagMapGetV", func(b *testing.B) {
			NN := (b.N * 100)
			p := &m
			for i := 0; i < NN; i++ {
				for _, k := range keys {
					v := TagMapGetV(p, k)
					if v == nil {
						b.Fatalf("[%s] not found", string(k))
					}
				}
			}
		})
		// b.Run("TagMapGetV4", func(b *testing.B) {
		// 	NN := (b.N * 100)
		// 	p := &m
		// 	for i := 0; i < NN; i++ {
		// 		v := TagMapGetV4(p, k)
		// 		if v == nil {
		// 			b.Fatalf("[%s] not found", string(k))
		// 		}
		// 	}
		// })
		b.Run("map", func(b *testing.B) {
			NN := (b.N * 100)
			for i := 0; i < NN; i++ {
				for _, k := range keys {
					v := mm[k]
					if v.V == nil {
						b.Fatalf("[%s] not found", string(k))
					}
				}
			}
		})
		b.Run("map2", func(b *testing.B) {
			NN := (b.N * 100)
			for i := 0; i < NN; i++ {
				for _, k := range keys {
					v := mm2[k]
					if v.V == nil {
						b.Fatalf("[%s] not found", string(k))
					}
				}
			}
		})
		b.Run("map4", func(b *testing.B) {
			NN := (b.N * 100)
			for i := 0; i < NN; i++ {
				for _, k := range keys {
					v := mm4[k]
					if v.V == nil {
						b.Fatalf("[%s] not found", string(k))
					}
				}
			}
		})
		b.Run("map8", func(b *testing.B) {
			NN := (b.N * 100)
			for i := 0; i < NN; i++ {
				for _, k := range keys {
					v := mm8[k]
					if v.V == nil {
						b.Fatalf("[%s] not found", string(k))
					}
				}
			}
		})
	}
}

func Benchmark_map_hash(b *testing.B) {
	keys := []string{}
	for i := 0; i < 16; i++ {
		keys = append(keys, getRandStr(rand.Intn(5)+5))
	}
	bsList := [][]byte{}
	for _, k := range keys {
		bsList = append(bsList, []byte(k))
	}
	// mapGoType := func() *maptype {
	// 	m := make(map[string]mapNode)
	// 	typ := reflect.TypeOf(m)
	// 	return (*maptype)(unsafe.Pointer(UnpackType(typ)))
	// }()
	pm := *(**hmap)(unsafe.Pointer(&map[string]mapNode{}))

	b.Run("map.hasher", func(b *testing.B) {
		NN := b.N * 100
		for i := 0; i < NN; i++ {
			for _, k := range keys {
				_ = strhash(noescape(unsafe.Pointer(&k)), uintptr(pm.hash0))
			}
		}
	})
}

//go:nosplit
func noescape(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0)
}

func toHashx(n int) (fout func(k string) int) {
	if n == 1 {
		fout = func(k string) int { return 1 }
		return
	}
	f := toHashx(n - 1)
	return func(key string) (idx int) {
		x := f(key)
		x++
		return
	}
}

func toHashx2(n int) (fout func(k string) int) {
	fout = func(k string) int { return 1 }
	for i := 0; i < n; i++ {
		f1 := fout
		f := func(key string) (idx int) {
			x := f1(key)
			x++
			return
		}
		fout = f
	}

	return
}
func Benchmark_toHashx(b *testing.B) {
	key := "11111111"
	// f5 := toHashx(5)
	// f1 := toHashx(1)
	f5 := toHashx2(5)
	f1 := toHashx2(1)
	for range make([]struct{}, 2) {
		runtime.GC()
		b.Run("toHashx(1)", func(b *testing.B) {
			NN := b.N * 100
			for i := 0; i < NN; i++ {
				f1(key)
			}
		})
		b.Run("toHashx(5)", func(b *testing.B) {
			NN := b.N * 100
			for i := 0; i < NN; i++ {
				f5(key)
			}
		})
	}
}

// hash 参数太多了，影响性能
func Benchmark_hash(b *testing.B) {
	keys := []string{}
	for i := 0; i < 16; i++ {
		keys = append(keys, getRandStr(rand.Intn(10)+5))
	}
	ns := []mapNode{}
	for _, k := range keys {
		ns = append(ns, mapNode{
			K: k,
			V: &TagInfo{},
		})
	}
	m := buildTagMap(ns)
	b.Logf("%+v", m.String())
	for _, n := range ns {
		v := m.Get(n.K)
		if v == nil {
			b.Fatalf("[%s] not found", string(n.K))
		}
	}

	idxN2 := make([]iN2, 0, 100)
	idxN2 = append(idxN2, m.idxN2...)
	idxN2 = idxN2[:cap(idxN2)]
	mm := make(map[string]mapNode)
	for _, n := range ns {
		mm[string(n.K)] = n
	}

	for range make([]struct{}, 2) {
		runtime.GC()
		b.Run("range", func(b *testing.B) {
			NN := b.N * 100
			for i := 0; i < NN; i++ {
				for _, k := range keys {
					for range []byte(k) {
					}
				}
			}
		})
		runtime.GC()
		b.Run("hash", func(b *testing.B) {
			NN := b.N * 100
			for i := 0; i < NN; i++ {
				for _, k := range keys {
					_ = hash(k, m.idxN)
				}
			}
		})
		runtime.GC()
		b.Run("hash2", func(b *testing.B) {
			NN := b.N * 100
			for i := 0; i < NN; i++ {
				for _, k := range keys {
					_ = hash2(k, m.idxNTable, m.idxN)
				}
			}
		})
		runtime.GC()
		b.Run("hash2x", func(b *testing.B) {
			NN := b.N * 100
			for i := 0; i < NN; i++ {
				for _, k := range keys {
					_ = hash2x(k, m.idxN[:m.idxNTable[len(k)]])
				}
			}
		})
		runtime.GC()
		b.Run("hash21", func(b *testing.B) {
			NN := b.N * 100
			for i := 0; i < NN; i++ {
				for _, k := range keys {
					_ = hash21(k, m.idxNTable, m.idxN)
				}
			}
		})
		runtime.GC()
		b.Run("hash41", func(b *testing.B) {
			NN := b.N * 100
			for i := 0; i < NN; i++ {
				for _, k := range keys {
					_ = hash41(k, m.idxNTable, m.idxN2)
				}
			}
		})
		runtime.GC()
		b.Run("hash4", func(b *testing.B) {
			NN := b.N * 100
			for i := 0; i < NN; i++ {
				for _, k := range keys {
					_ = hash4(k, m.idxNTable, m.idxN2)
				}
			}
		})
		runtime.GC()
		b.Run("hash4x", func(b *testing.B) {
			NN := b.N * 100
			for i := 0; i < NN; i++ {
				for _, k := range keys {
					_ = hash4x(k, m.idxN2[:m.idxNTable[len(k)]])
				}
			}
		})
		runtime.GC()
		b.Run("hash41", func(b *testing.B) {
			NN := b.N * 100
			for i := 0; i < NN; i++ {
				for _, k := range keys {
					_ = hash41(k, m.idxNTable, m.idxN2)
				}
			}
		})
		runtime.GC()
		b.Run("hash51", func(b *testing.B) {
			NN := b.N * 100
			for i := 0; i < NN; i++ {
				for _, k := range keys {
					_ = hash51(k, m.idxNTable, m.idxN2)
				}
			}
		})
		runtime.GC()
		b.Run("hash52", func(b *testing.B) {
			NN := b.N * 100
			for i := 0; i < NN; i++ {
				for _, k := range keys {
					_ = hash52(k, m.idxNTable, m.idxN2)
				}
			}
		})
		runtime.GC()
		b.Run("hash3", func(b *testing.B) {
			NN := b.N * 100
			for i := 0; i < NN; i++ {
				for _, k := range keys {
					_ = hash3(k, m.idxNTable, m.idxN2)
				}
			}
		})
		runtime.GC()
		b.Run("hash31", func(b *testing.B) {
			NN := b.N * 100
			for i := 0; i < NN; i++ {
				for _, k := range keys {
					_ = hash31(k, m.idxNTable, m.idxN2)
				}
			}
		})
		runtime.GC()
		b.Run("hash32", func(b *testing.B) {
			NN := b.N * 100
			for i := 0; i < NN; i++ {
				for _, k := range keys {
					_ = hash32(k, m.idxNTable, idxN2)
				}
			}
		})
		runtime.GC()
		b.Run("hash33", func(b *testing.B) {
			NN := b.N * 100
			for i := 0; i < NN; i++ {
				for _, k := range keys {
					_ = hash33(k, m.idxNTable, idxN2)
				}
			}
		})
		runtime.GC()
		b.Run("hash34", func(b *testing.B) {
			NN := b.N * 100
			for i := 0; i < NN; i++ {
				for _, k := range keys {
					_ = hash34(k, m.idxNTable, idxN2)
				}
			}
		})
		runtime.GC()
		k0 := make([]byte, 1024) //[]byte(keys[0])
		b.Run("simd.Hash", func(b *testing.B) {
			NN := b.N * 100
			for i := 0; i < NN; i++ {
				for _, k := range keys {
					_ = Hash(stringBytes(k), k0)
				}
			}
		})
		cs := [1024]N{}
		b.Run("simd.Hashx", func(b *testing.B) {
			NN := b.N * 100
			for i := 0; i < NN; i++ {
				for _, k := range keys {
					_ = Hashx(stringBytes(k), cs[:])
				}
			}
		})
		rcs := [1024]runtime.N{}
		b.Run("runtime.simd.Hashx", func(b *testing.B) {
			NN := b.N * 100
			for i := 0; i < NN; i++ {
				for _, k := range keys {
					_ = runtime.Hashx(stringBytes(k), rcs[:])
				}
			}
		})
		runtime.GC()
		k0[0] = keys[0][0]
		k0[1] = keys[0][1]
		b.Run("simd.Hash", func(b *testing.B) {
			NN := b.N * 100
			for i := 0; i < NN; i++ {
				for _, k := range keys {
					_ = Hash(stringBytes(k), k0)
				}
			}
		})
		runtime.GC()
		k0 = []byte(keys[0])
		b.Run("simd.Hash", func(b *testing.B) {
			NN := b.N * 100
			for i := 0; i < NN; i++ {
				for _, k := range keys {
					_ = Hash(stringBytes(k), k0)
				}
			}
		})
		runtime.GC()
		pm := *(**hmap)(unsafe.Pointer(&map[string]mapNode{}))
		b.Run("map.hasher", func(b *testing.B) {
			NN := b.N * 100
			for i := 0; i < NN; i++ {
				for _, k := range keys {
					_ = strhasher(noescape(unsafe.Pointer(&k)), uintptr(pm.hash0))
				}
			}
		})

		runtime.GC()
		b.Run("TagMapGetV", func(b *testing.B) {
			NN := b.N * 100
			p := &m
			for i := 0; i < NN; i++ {
				for _, k := range keys {
					v := TagMapGetV(p, k)
					if v == nil {
						b.Fatalf("[%s] not found", string(k))
					}
				}
			}
		})
		runtime.GC()
		b.Run("map", func(b *testing.B) {
			NN := b.N * 100
			for i := 0; i < NN; i++ {
				for _, k := range keys {
					v := mm[k]
					if v.V == nil {
						b.Fatalf("[%s] not found", k)
					}
				}
			}
		})
	}
}

var char byte

func Benchmark_range(b *testing.B) {
	keys := []string{}
	for i := 0; i < 16; i++ {
		keys = append(keys, getRandStr(rand.Intn(20)+5))
	}
	b.Run("range", func(b *testing.B) {
		NN := b.N * 100
		for i := 0; i < NN; i++ {
			for _, k := range keys {
				bs := *(*[]byte)(unsafe.Pointer(&k))
				for _, c := range bs {
					char = c
				}
			}
		}
	})
	b.Run("range2", func(b *testing.B) {
		NN := b.N * 100
		for i := 0; i < NN; i++ {
			for _, k := range keys {
				for j := 0; j < len(k); j++ {
					char = k[j]
				}
			}
		}
	})
	pm := *(**hmap)(unsafe.Pointer(&map[string]mapNode{}))
	b.Run("map.hasher", func(b *testing.B) {
		NN := b.N * 100
		for i := 0; i < NN; i++ {
			for _, k := range keys {
				_ = strhash(noescape(unsafe.Pointer(&k)), uintptr(pm.hash0))
			}
		}
	})
}

func Test_Map1(t *testing.T) {
	t.Run("bucketShift", func(t *testing.T) {
		for _, b := range []uint8{0, 2, 4, 8, 16, 17, 63, 64, 128, 255} {
			t.Logf("%d:%d", b, bucketShift(b))
		}

	})

}
