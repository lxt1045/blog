package json

import (
	"fmt"
	"math/rand"
	"testing"
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

/*

    /Users/bytedance/go/src/github.com/lxt1045/blog/sample/json/json/bin_tree_test.go:96: l:8, len:4, m: [{iByte:1 mask:1 iBit:8} {iByte:2 mask:5 iBit:1} {iByte:6 mask:16 iBit:2} {iByte:6 mask:1 iBit:4}]
=== RUN   Test_logicalHash_new/getPivotMask-16
    /Users/bytedance/go/src/github.com/lxt1045/blog/sample/json/json/bin_tree_test.go:96: l:16, len:6, m: [{iByte:1 mask:4 iBit:1} {iByte:1 mask:8 iBit:16} {iByte:5 mask:4 iBit:4} {iByte:5 mask:65 iBit:8} {iByte:7 mask:64 iBit:2} {iByte:11 mask:2 iBit:32}]
=== RUN   Test_logicalHash_new/getPivotMask-32
    /Users/bytedance/go/src/github.com/lxt1045/blog/sample/json/json/bin_tree_test.go:96: l:32, len:8, m: [{iByte:0 mask:16 iBit:64} {iByte:0 mask:64 iBit:8} {iByte:1 mask:2 iBit:32} {iByte:1 mask:8 iBit:1} {iByte:4 mask:33 iBit:16} {iByte:6 mask:1 iBit:4} {iByte:8 mask:64 iBit:2} {iByte:11 mask:2 iBit:128}]
=== RUN   Test_logicalHash_new/getPivotMask-64
    /Users/bytedance/go/src/github.com/lxt1045/blog/sample/json/json/bin_tree_test.go:96: l:64, len:9, m: [{iByte:0 mask:16 iBit:128} {iByte:0 mask:34 iBit:8} {iByte:2 mask:16 iBit:1} {iByte:2 mask:33 iBit:4} {iByte:3 mask:8 iBit:64} {iByte:6 mask:65 iBit:16} {iByte:6 mask:32 iBit:2} {iByte:8 mask:1 iBit:32} {iByte:11 mask:2 iBit:256}]
=== RUN   Test_logicalHash_new/getPivotMask-128
    /Users/bytedance/go/src/github.com/lxt1045/blog/sample/json/json/bin_tree_test.go:96: l:128, len:11, m: [{iByte:0 mask:1 iBit:1} {iByte:1 mask:16 iBit:128} {iByte:1 mask:4 iBit:512} {iByte:1 mask:1 iBit:64} {iByte:2 mask:16 iBit:4} {iByte:2 mask:4 iBit:2} {iByte:3 mask:1 iBit:1024} {iByte:3 mask:2 iBit:8} {iByte:4 mask:4 iBit:256} {iByte:4 mask:96 iBit:16} {iByte:11 mask:64 iBit:32}]
--- PASS: Test_logicalHash_new (0.78s)
*/

func Test_logicalHash_new(t *testing.T) {
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
		// `"isLeader"`,
		// `"isManager"`,
		// `"isAppManager"`,
		// `"departmentList"`,
	}
	// keys = []string{`"avatar"`, `"avatar72"`, `"avatar240"`, `"avatar640"`}
	for i := 0; i < 160; i++ {
		keys = append(keys, getRandStr(rand.Intn(18)+2))
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
				idx := hash(bs, nil, m)
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
	// keys = []string{`"avatar"`, `"avatar72"`, `"avatar240"`, `"avatar640"`}
	for i := 0; i < 160; i++ {
		keys = append(keys, getRandStr(rand.Intn(18)+2))
	}
	bsList := [][]byte{}
	for _, k := range keys {
		bsList = append(bsList, []byte(k))
	}

	lens := []int{8, 16, 32, 64, 128}
	lens = []int{8, 64}
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
		// `"isLeader"`,
		// `"isManager"`,
		// `"isAppManager"`,
		// `"departmentList"`,
	}
	// keys = []string{`"avatar"`, `"avatar72"`, `"avatar240"`, `"avatar640"`}
	for i := 0; i < 16; i++ {
		keys = append(keys, getRandStr(rand.Intn(18)+2))
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
	// for i := 0; i < 22; i++ {
	// 	ns = append(ns,
	// 		mapNode{
	// 			K: []byte(fmt.Sprintf("%d", rand.Intn(100000))),
	// 			V: &TagInfo{},
	// 		})
	// }
	m := buildTagMap(ns)
	b.Logf("%+v", m.String())
	for _, n := range ns {
		v := m.Get(n.K)
		if v == nil {
			b.Fatalf("[%s] not found", string(n.K))
		}
	}

	mm := make(map[string]mapNode)
	for _, n := range ns {
		mm[string(n.K)] = n
	}
	n := ns[len(ns)/2]

	b.Run("buildTagMap", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// n := ns[i%len(ns)]
			v := m.Get(n.K)
			for i := 0; i < 99; i++ {
				v = m.Get(n.K)
			}
			if v == nil {
				b.Fatalf("[%s] not found", string(n.K))
			}
		}
		b.StopTimer()
		b.SetBytes(int64(b.N))
	})
	b.Run("buildTagMap-2", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// n := ns[i%len(ns)]
			v := m.Get2(n.K)
			for i := 0; i < 99; i++ {
				v = m.Get2(n.K)
			}
			if v == nil {
				b.Fatalf("[%s] not found", string(n.K))
			}
		}
		b.StopTimer()
		b.SetBytes(int64(b.N))
	})
	b.Run("buildTagMap", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// n := ns[i%len(ns)]
			v := m.Get(n.K)
			for i := 0; i < 99; i++ {
				v = m.Get(n.K)
			}
			if v == nil {
				b.Fatalf("[%s] not found", string(n.K))
			}
		}
		b.StopTimer()
		b.SetBytes(int64(b.N))
	})
	b.Run("buildTagMap-2", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// n := ns[i%len(ns)]
			v := m.Get2(n.K)
			for i := 0; i < 99; i++ {
				v = m.Get2(n.K)
			}
			if v == nil {
				b.Fatalf("[%s] not found", string(n.K))
			}
		}
		b.StopTimer()
		b.SetBytes(int64(b.N))
	})
	// return
	b.Run("map", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// n := ns[i%len(ns)]
			v, ok := mm[string(n.K)]
			for i := 0; i < 99; i++ {
				v = mm[string(n.K)]
			}
			if !ok {
				b.Fatalf("%s:%s", string(n.K), string(v.K))
			}
		}
		b.StopTimer()
		b.SetBytes(int64(b.N))
	})
}
