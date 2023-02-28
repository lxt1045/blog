package hashmap

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"unsafe"
)

func TestHash(t *testing.T) {
	keys := []string{}
	for i := 0; i < 2; i++ {
		keys = append(keys, getRandStr(rand.Intn(32)))
	}
	bs := []byte(keys[0])
	cs := []byte(keys[1])

	y := Hash(bs, cs)
	t.Logf("y:%d", y)
}
func TestHashx(t *testing.T) {
	keys := []string{}
	for i := 0; i < 2; i++ {
		keys = append(keys, getRandStr(rand.Intn(32)))
	}
	bs := []byte(keys[0])
	cs := [1024]N{}

	y := Hashx(bs, cs[:])
	t.Logf("y:%d", y)
}

func Tt(value reflect.Value) {
	if value.Kind() != reflect.Ptr {
		panic("ss")
	}
	value = value.Elem()
	if value.Kind() != reflect.Struct {
		panic("ss")
	}
	field := value.FieldByName("SS")
	if !field.CanSet() {
		return
	}
	// str := "888"
	// field.Set(reflect.ValueOf(&str))
	field = field.Elem()
	if !field.CanSet() {
		return
	}
	field.SetString("999")
	// field.SetIterValue()

	m := value.FieldByName("M")
	if !m.CanSet() {
		panic("m")
	}
	iter := m.MapRange()
	if iter == nil {
		return
	}
	for iter.Next() {
		v := iter.Value().String()
		v += "-xxx"
		m.SetMapIndex(iter.Key(), reflect.ValueOf(v))
	}
}

func Test_String(t *testing.T) {

	runtime.GC()
	s := struct {
		SS *string
		M  map[string]string
	}{}
	x := ""
	s.SS = &x
	s.M = map[string]string{
		"a":  "b",
		"a1": "b11",
	}
	t.Run("1", func(t *testing.T) {
		Tt(reflect.ValueOf(&s))
		t.Logf("%+v", s)
		t.Logf("%+v", *s.SS)
	})
}

func TestCtz64(t *testing.T) {
	// xs := make([]byte, 16)
	// xs[4] = 'x'
	xs := [32]byte{}
	y := Test2('a', xs[:])
	t.Logf("Test1(),as:%+v,y:%d", xs, y)
	t.Logf("Test1(),as:%+v", fillBytes16('x'))
}

func BenchmarkTest1(b *testing.B) {
	b.Run("0", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Test1(0b100, 5)
		}
	})
	b.Run("2", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Test1(0b1000000000000000000, 5)
		}
	})
}

var bsG string
var bsG1 []byte

func BenchmarkTrimSpace(b *testing.B) {
	bs0 := make([]byte, 16+len(j0))
	pbs := (*[1 << 31]byte)(unsafe.Pointer((uintptr(unsafe.Pointer(&bs0[0])) + 0xf) & (^uintptr(0xf))))
	bs := pbs[:len(j0x)]
	d := J0{}
	json.Unmarshal([]byte(j0), &d)
	bs, _ = json.Marshal(&d)
	b.Log(string(bs))
	// copy(bs, j0)
	countG := [7]int{}
	table := [256]bool{'\t': true, '\n': true, '\v': true, '\f': true, '\r': true, ' ': true, 0x85: true, 0xA0: true}

	xs1 := "12345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"
	for i := 0; i < 3; i++ {
		xs1 += xs1
	}
	b.Run("bytes.IndexByte", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			strings.Index(xs1, "asdfgjlnd")
		}
		b.SetBytes(int64(b.N))
		b.StopTimer()
	})
	b.Run("7", func(b *testing.B) {
		xss := [32]byte{}
		p := (*[8]byte)(unsafe.Pointer((uintptr(unsafe.Pointer(&xss)) + 0xf) & (^uintptr(0xf))))
		*p = [8]byte{0x85, 0xA0, '\t', '\n', '\v', '\f', '\r', ' '}
		xs := (*p)[:]

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			max := len(bs)
			max = max / 8 * 8
			count := IndexBytes2(xs, bs[0:max])
			for ; max < len(bs); max++ {
				if table[bs[max]] {
					count++
				}
			}
			// count := IndexBytes2(xs, bs)
			if countG[6] == 0 {
				countG[6] = count
			}
		}
		b.SetBytes(int64(b.N))
		b.StopTimer()
	})
	// b.Logf("countG:%+v", countG)
	// return
	b.Run("5", func(b *testing.B) {
		xss := [32]byte{}
		p := (*[8]byte)(unsafe.Pointer(uintptr(unsafe.Pointer(&xss)) & (^uintptr(0xf))))
		*p = [8]byte{0x85, 0xA0, '\t', '\n', '\v', '\f', '\r', ' '}
		xs := (*p)[:]

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			count := IndexBytes1(xs, bs[0:])
			if countG[4] == 0 {
				countG[4] = count
			}
		}
		b.SetBytes(int64(b.N))
		b.StopTimer()
	})
	b.Run("6", func(b *testing.B) {
		xss := [32]byte{}
		p := (*[8]byte)(unsafe.Pointer(uintptr(unsafe.Pointer(&xss)) & (^uintptr(0xf))))
		*p = [8]byte{0x85, 0xA0, '\t', '\n', '\v', '\f', '\r', ' '}
		xs := (*p)[:]

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			count := 0
			Ns := [8]int{}
			for j := 0; j < len(bs); {
				last := len(bs)
				for idx, x := range xs[:] {
					if Ns[idx] < j {
						c := bytes.IndexByte(bs[j:], x)
						if c < 0 {
							Ns[idx] = len(bs)
						} else {
							Ns[idx] = j + c
						}
					}
					if Ns[idx] < last {
						last = Ns[idx]
					}
				}
				j = last

				idx := IndexBytes(xs, bs[j:])
				if idx <= 0 {
					j++
					continue
				}
				j += idx
				count += idx
			}
			if countG[5] == 0 {
				countG[5] = count
			}
		}
		b.SetBytes(int64(b.N))
		b.StopTimer()
	})

	b.Run("1", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			count := 0
			for _, bb := range bs {
				if table[bb] {
					count++
				}
			}
			if countG[0] == 0 {
				countG[0] = count
			}
		}
		b.SetBytes(int64(b.N))
		b.StopTimer()
	})

	b.Run("1-2", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			count := 0
			for _, bb := range bs {
				if InSpaceQ(bb) {
					count++
				}
			}
			if countG[0] == 0 {
				countG[0] = count
			}
		}
		b.SetBytes(int64(b.N))
		b.StopTimer()
	})

	b.Run("2", func(b *testing.B) {
		xs := []byte{0x85, 0xA0, '\t', '\n', '\v', '\f', '\r', ' '}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			count := 0
			for _, x := range xs[:] {
				for i, c := 0, 0; ; {
					c = bytes.IndexByte(bs[i:], x)
					if c < 0 {
						break
					}
					count++
					i = i + c + 1
					for ; bs[i] == x; i++ {
						count++
					}
				}
			}
			if countG[1] == 0 {
				countG[1] = count
			}
		}
		b.SetBytes(int64(b.N))
		b.StopTimer()
	})
	b.Run("2-1", func(b *testing.B) {
		xs := []byte{0x85, 0xA0, '\t', '\n', '\v', '\f', '\r'}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			count := 0
			for _, x := range xs[:] {
				for i, c := 0, 0; ; {
					c = bytes.IndexByte(bs[i:], x)
					if c < 0 {
						break
					}
					count++
					i = i + c + 1
					for ; bs[i] == x; i++ {
						count++
					}
				}
			}
			// for _, b := range bs {
			// 	if b == ' ' {
			// 		count++
			// 	}
			// }
			if countG[1] == 0 {
				countG[1] = count
			}
		}
		b.SetBytes(int64(b.N))
		b.StopTimer()
	})
	b.Run("2-3", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			count := 0
			for i, c := 0, 0; ; {
				c = bytes.IndexByte(bs[i:], '\r')
				if c < 0 {
					break
				}
				count++
				i = i + c + 1
			}

		}
		b.SetBytes(int64(b.N))
		b.StopTimer()
	})
	b.Run("2-2", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			count := 0
			for i := 0; i < len(bs); i++ {
				if bs[i] == ' ' {
					count++
				}
			}
			if countG[1] == 0 {
				countG[1] = count
			}
		}
		b.SetBytes(int64(b.N))
		b.StopTimer()
	})

	b.Run("3", func(b *testing.B) {
		table := [256]bool{'\t': true, '\n': true, '\v': true, '\f': true, '\r': true, ' ': true, 0x85: true, 0xA0: true}

		xs := [4]uint64{}

		for i := 0; i < 256; i++ {
			if table[i] {
				xs[i/64] |= 1 << (uint64(i) % 64)
			}
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var count uint64
			for _, x := range bs {
				// x := uint64(bb)
				// if (xs[x/64] & (1 << (x % 64))) > 0 {
				// 	count++
				// }
				count += (xs[x/64] >> (x % 64)) & 1
			}
			if countG[2] == 0 {
				countG[2] = int(count)
			}
		}
		b.SetBytes(int64(b.N))
		b.StopTimer()
	})

	b.Run("4", func(b *testing.B) {
		xss := [16]byte{}
		p := (*[8]byte)(unsafe.Pointer(uintptr(unsafe.Pointer(&xss)) & (^uintptr(0xf))))
		*p = [8]byte{0x85, 0xA0, '\t', '\n', '\v', '\f', '\r', ' '}
		xs := (*p)[:]

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			count := 0
			for _, x := range bs {
				if IndexByte(xs, x) >= 0 {
					count++
				}
			}
			// strings.Index()
			if countG[3] == 0 {
				countG[3] = count
			}
		}
		b.SetBytes(int64(b.N))
		b.StopTimer()
	})

	if countG[0] != countG[1] || countG[0] != countG[2] || countG[0] != countG[3] ||
		countG[0] != countG[4] || countG[0] != countG[5] || countG[0] != countG[6] {
		b.Fatalf("countG:%+v", countG)
	} else {
		b.Logf("countG:%+v", countG)
	}
}

func BenchmarkMapAcess(b *testing.B) {
	m := make(map[string][]byte)
	keys := make([]string, 200)
	for i := range keys {
		key := make([]byte, 20)
		for i := range key {
			key[i] = byte(rand.Uint32())
		}
		m[string(key)] = key
		keys[i] = string(key)
	}
	b.Run("1", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = m[keys[i%len(keys)]]
		}
		b.SetBytes(int64(b.N))
		b.StopTimer()
	})
}

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

var j0 = `{"id":"7028151259660092936","name":{"zh-CN":"陈三","en-US":""},"avatar":{"avatar":"https://7b90749e-a6be-4a86-9f9b-7dce3d2ecf5g~?image_size=noop&cut_type=&quality=&format=png&sticker_format=.webp","avatar72":"https://7b90749e-a6be-4a86-9f9b-7dce3d2ecf5g~?image_size=72x72&cut_type=&quality=&format=png&sticker_format=.webp","avatar240":"https://7b90749e-a6be-4a86-9f9b-7dce3d2ecf5g~?image_size=240x240&cut_type=&quality=&format=png&sticker_format=.webp","avatar640":"https://7b90749e-a6be-4a86-9f9b-7dce3d2ecf5g~?image_size=640x640&cut_type=&quality=&format=png&sticker_format=.webp"},"department":{"id":"6826585686905406989","name":{"zh-CN":"研发部","en-US":"RD Department"}},"email":"","mobile":{"phone":"18030838810","code":"86"},"status":{"accountStatus":false,"employmentStatus":false,"registerStatus":false},"employeeType":{"id":1,"name":{"zh-CN":"正式","en-US":"Regular"},"active":true},"isAdmin":false,"isLeader":false,"isManager":false,"isAppManager":false,"departmentList":{"id":"6826585686905406989","name":{"zh-CN":"研发部","en-US":"RD Department"}}}`

type J0 struct {
	ID   string `json:"id"`
	Name struct {
		ZhCN string `json:"zh-CN"`
		EnUS string `json:"en-US"`
	} `json:"name"`
	Avatar struct {
		Avatar    string `json:"avatar"`
		Avatar72  string `json:"avatar72"`
		Avatar240 string `json:"avatar240"`
		Avatar640 string `json:"avatar640"`
	} `json:"avatar"`
	Department struct {
		ID   string `json:"id"`
		Name struct {
			ZhCN string `json:"zh-CN"`
			EnUS string `json:"en-US"`
		} `json:"name"`
	} `json:"department"`
	Email  string `json:"email"`
	Mobile struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	} `json:"mobile"`
	Status struct {
		AccountStatus    bool `json:"accountStatus"`
		EmploymentStatus bool `json:"employmentStatus"`
		RegisterStatus   bool `json:"registerStatus"`
	} `json:"status"`
	EmployeeType struct {
		ID   int `json:"id"`
		Name struct {
			ZhCN string `json:"zh-CN"`
			EnUS string `json:"en-US"`
		} `json:"name"`
		Active bool `json:"active"`
	} `json:"employeeType"`
	IsAdmin        bool `json:"isAdmin"`
	IsLeader       bool `json:"isLeader"`
	IsManager      bool `json:"isManager"`
	IsAppManager   bool `json:"isAppManager"`
	DepartmentList struct {
		ID   string `json:"id"`
		Name struct {
			ZhCN string `json:"zh-CN"`
			EnUS string `json:"en-US"`
		} `json:"name"`
	} `json:"departmentList"`
}
