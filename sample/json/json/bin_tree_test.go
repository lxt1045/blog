package json

import (
	"testing"
)

func Test_append2(t *testing.T) {
	xs := []int{1, 2, 3, 4, 5}
	i := 4
	xs = append(xs[:i+1], xs[i:len(xs)-1]...)
	xs[i] = 88
	t.Logf("%+v", xs)
}

func Test_logicalHash2(t *testing.T) {
	bsList := [][]byte{
		// []byte(`"id"`),
		// []byte(`"name"`),
		// []byte(`"avatar"`),
		// []byte(`"department"`),
		// []byte(`"email"`),
		// []byte(`"mobile"`),
		[]byte(`"status"`),
		[]byte(`"employeeType"`),
		[]byte(`"isAdmin"`),
		[]byte(`"isLeader"`),
		[]byte(`"isManager"`),
		[]byte(`"isAppManager"`),
		[]byte(`"departmentList"`),
	}
	idxList, _ := logicalHash(bsList)
	t.Logf("idxList: \n\n%+v", idxList)

	PrintKeys2(bsList)
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
	keys = []string{`"avatar"`, `"avatar72"`, `"avatar240"`, `"avatar640"`}
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
