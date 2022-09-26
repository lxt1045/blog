package json

import (
	"fmt"
	"math/rand"
	"runtime"
	"testing"
	"unsafe"
)

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
	bsList := [][]byte{}
	for _, k := range keys {
		bsList = append(bsList, []byte(k))
	}

	lens := []int{8, 16, 32, 64, 128, 256}
	for _, l := range lens {
		name := fmt.Sprintf("getPivotMask-%d", l)
		t.Run(name, func(t *testing.T) {
			m := getPivotMask(bsList[:l])
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
	bsList := [][]byte{}
	for _, k := range keys {
		bsList = append(bsList, []byte(k))
	}

	lens := []int{
		8,
		16, 32, 64, 128,
	}
	for _, l := range lens {
		name := fmt.Sprintf("getPivotMask-%d", l)
		t.Run(name, func(t *testing.T) {
			m := getPivotMask(bsList[:l])
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
	for i := 0; i < 160; i++ {
		keys = append(keys, getRandStr(rand.Intn(10)+5))
	}
	bsList := [][]byte{}
	for _, k := range keys {
		bsList = append(bsList, []byte(k))
	}

	lens := []int{
		8,
		16, 32, 64, 128,
	}
	for _, l := range lens {
		name := fmt.Sprintf("getPivotMask-%d", l)
		t.Run(name, func(t *testing.T) {
			m, _ := logicalHash(bsList[:l])
			t.Logf("l:%d, len:%d, m: %+v", l, len(m), m)
			// t.Logf("l:%d, len:%d", l, len(m))

			exist := map[int]string{}
			for _, bs := range bsList[:l] {
				idx := hash(bs, m)
				if _, ok := exist[idx]; !ok {
					exist[idx] = string(bs)
				} else {
					PrintKeys(bsList[:l])
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
	for i := 0; i < 160; i++ {
		keys = append(keys, getRandStr(rand.Intn(18)+2))
	}
	bsList := [][]byte{}
	for _, k := range keys {
		bsList = append(bsList, []byte(k))
	}

	lens := []int{8, 16, 32, 64, 128}
	// lens = []int{8, 64, 128}
	// lens = []int{128}
	for _, l := range lens {
		name := fmt.Sprintf("getPivotMask-%d", l)
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				getPivotMask(bsList[:l])
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
			K: []byte(k),
			V: &TagInfo{},
		})
	}

	m := buildTagMap(ns)
	t.Logf("m: %+v", m.String())

}
func Test_logicalHash(t *testing.T) {
	bsList := [][]byte{
		[]byte("111"),
		[]byte("122"),
		[]byte("333"),
		[]byte("433"),
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
			K: []byte(k),
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
	for i := 0; i < 16; i++ {
		keys = append(keys, getRandStr(rand.Intn(20)+5))
	}
	bsList := [][]byte{}
	for _, k := range keys {
		bsList = append(bsList, []byte(k))
	}
	ns := []mapNode{}
	for _, k := range keys {
		ns = append(ns, mapNode{
			K: []byte(k),
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

	b.Run("TagMapGetV", func(b *testing.B) {
		NN := (b.N * 100) / len(bsList)
		p := &m
		for i := 0; i < NN; i++ {
			for _, k := range bsList {
				v := TagMapGetV3(p, k)
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
			NN := (b.N * 100) / len(bsList)
			for i := 0; i < NN; i++ {
				for _, k := range bsList {
					v := m.Get(k)
					if v == nil {
						b.Fatalf("[%s] not found", string(n.K))
					}
				}
			}
		})
		b.Run("Get2", func(b *testing.B) {
			NN := (b.N * 100) / len(bsList)
			for i := 0; i < NN; i++ {
				for _, k := range bsList {
					v := m.Get2(k)
					if v == nil {
						b.Fatalf("[%s] not found", string(n.K))
					}
				}
			}
		})
		b.Run("Get3", func(b *testing.B) {
			NN := (b.N * 100) / len(bsList)
			for i := 0; i < NN; i++ {
				for _, k := range bsList {
					v := m.Get3(k)
					if v == nil {
						b.Fatalf("[%s] not found", string(n.K))
					}
				}
			}
		})
		b.Run("TagMapGetV", func(b *testing.B) {
			NN := (b.N * 100) / len(bsList)
			p := &m
			for i := 0; i < NN; i++ {
				for _, k := range bsList {
					v := TagMapGetV(p, k)
					if v == nil {
						b.Fatalf("[%s] not found", string(n.K))
					}
				}
			}
		})
		b.Run("TagMapGetV3", func(b *testing.B) {
			NN := (b.N * 100) / len(bsList)
			p := &m
			for i := 0; i < NN; i++ {
				for _, k := range bsList {
					v := TagMapGetV3(p, k)
					if v == nil {
						b.Fatalf("[%s] not found", string(n.K))
					}
				}
			}
		})
		// return
		b.Run("map", func(b *testing.B) {
			NN := (b.N * 100) / len(bsList)
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

func Benchmark_map_hash(b *testing.B) {
	keys := []string{}
	for i := 0; i < 16; i++ {
		keys = append(keys, getRandStr(rand.Intn(20)+5))
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

func Benchmark_hash(b *testing.B) {
	keys := []string{}
	for i := 0; i < 10; i++ {
		keys = append(keys, getRandStr(rand.Intn(20)+5))
	}
	bsList := [][]byte{}
	for _, k := range keys {
		bsList = append(bsList, []byte(k))
	}
	ns := []mapNode{}
	for _, k := range keys {
		ns = append(ns, mapNode{
			K: []byte(k),
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

	for range make([]struct{}, 2) {
		b.Run("hash", func(b *testing.B) {
			NN := b.N * 100
			for i := 0; i < NN; i++ {
				for _, k := range bsList {
					_ = hash(k, m.idxN)
				}
			}
		})
		b.Run("hash2", func(b *testing.B) {
			NN := b.N * 100
			for i := 0; i < NN; i++ {
				for _, k := range bsList {
					_ = hash2(k, m.idxNTable, m.idxN)
				}
			}
		})
		b.Run("hash3", func(b *testing.B) {
			NN := b.N * 100
			for i := 0; i < NN; i++ {
				for _, k := range bsList {
					_ = hash3(k, m.idxNTable, m.idxN2)
				}
			}
		})
		b.Run("hash4", func(b *testing.B) {
			NN := b.N * 100
			for i := 0; i < NN; i++ {
				for _, k := range bsList {
					_ = hash4(k, m.idxNTable, m.idxN2)
				}
			}
		})
		b.Run("hash5", func(b *testing.B) {
			NN := b.N * 100
			for i := 0; i < NN; i++ {
				for _, k := range bsList {
					_ = hash5(k, m.idxNTable, m.idxN2)
				}
			}
		})
		b.Run("range", func(b *testing.B) {
			NN := b.N * 100
			for i := 0; i < NN; i++ {
				for _, k := range bsList {
					for range k {
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

		b.Run("TagMapGetV3", func(b *testing.B) {
			NN := b.N * 100
			p := &m
			for i := 0; i < NN; i++ {
				for _, k := range bsList {
					v := TagMapGetV3(p, k)
					if v == nil {
						b.Fatalf("[%s] not found", string(k))
					}
				}
			}
		})
		mm := make(map[string]mapNode)
		for _, n := range ns {
			mm[string(n.K)] = n
		}
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

func Test_Map1(t *testing.T) {
	t.Run("bucketShift", func(t *testing.T) {
		for _, b := range []uint8{0, 2, 4, 8, 16, 17, 63, 64, 128, 255} {
			t.Logf("%d:%d", b, bucketShift(b))
		}

	})

}
