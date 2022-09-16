package json

import (
	"reflect"
	"sync/atomic"
	"unsafe"
)

const N = 1024

// /*

var (
// bsCache        = NewSliceCache[[]byte](N)
// strCache       = NewSliceCache[string](N)
// interfaceCache = NewSliceCache[interface{}](N)
// mapCache       = NewSliceCache[map[string]interface{}](N)
)

type sliceCache[T any] struct {
	c []T
	i int32
}

type SliceCache[T any] struct {
	c      unsafe.Pointer
	n      int32
	ch     chan *sliceCache[T]
	makeTs func() (ts []T)
}

//go:inline
func (s *SliceCache[T]) Get() (p *T) {
	c := (*sliceCache[T])(atomic.LoadPointer(&s.c))
	i := atomic.AddInt32(&c.i, 1)
	if i <= s.n {
		p = &c.c[i-1]
		return
	}
	c = <-s.ch
	// c = &sliceCache[T]{
	// 	c: s.makeTs(),
	// 	i: 1,
	// }
	p = &c.c[0]
	atomic.StorePointer(&s.c, unsafe.Pointer(c))
	return
}
func (s *SliceCache[T]) GetT() (p T) {
	c := (*sliceCache[T])(atomic.LoadPointer(&s.c))
	i := atomic.AddInt32(&c.i, 1)
	if i <= s.n {
		p = c.c[i-1]
		return
	}
	c = <-s.ch
	p = c.c[0]
	atomic.StorePointer(&s.c, unsafe.Pointer(c))
	return
}
func (s *SliceCache[T]) GetP() (p unsafe.Pointer) {
	c := (*sliceCache[T])(atomic.LoadPointer(&s.c))
	i := atomic.AddInt32(&c.i, 1)
	if i <= s.n {
		p = unsafe.Pointer(&c.c[i-1])
		return
	}
	c = <-s.ch
	p = unsafe.Pointer(&c.c[0])
	atomic.StorePointer(&s.c, unsafe.Pointer(c))
	return
}
func NewSliceCache[T any](n int32) (s *SliceCache[T]) {
	var t T
	var value interface{} = t
	_, bMap := value.(map[string]interface{})

	makeTs := func() (ts []T) {
		ts = make([]T, n)
		return
	}
	if bMap {
		makeTs = func() (ts []T) {
			m := make([]map[string]interface{}, n)
			for i := range m {
				m[i] = make(map[string]interface{})
			}
			p := (*[]map[string]interface{})(unsafe.Pointer(&ts))
			*p = m
			return
		}
	}

	s = &SliceCache[T]{
		c: unsafe.Pointer(&sliceCache[T]{
			c: makeTs(),
		}),
		n: int32(n),
	}
	s.makeTs = makeTs
	s.ch = make(chan *sliceCache[T], 8)

	{
		sc := make([]sliceCache[T], cap(s.ch))
		for i := range sc {
			sc[i].i = 1
			sc[i].c = makeTs()
			s.ch <- &sc[i]
		}
	}
	N := 1
	if bMap {
		//单独测试 cache 时，单协程生产速度太慢
		// N = 48
	}
	for i := 0; i < N; i++ {
		go func() {
			for {
				sc := make([]sliceCache[T], cap(s.ch))
				for i := range sc {
					sc[i].i = 1
					sc[i].c = makeTs()
					s.ch <- &sc[i]
				}
			}
		}()
	}
	return
}

// Type is Result type
type Type int

const (
	// Null is a null json value
	Null Type = iota
	// False is a json false boolean
	False
	// Number is json number
	Number
	// String is a json string
	String
	// True is a json true boolean
	True
	// JSON is a raw block of JSON
	JSON

	Slice
	// False/True is a json boolean
	Bool

	Bytes
	Struct
	Map
	Interface

	MaxType
)

func (t Type) IsNull() bool {
	return t <= Null
}

// String returns a string representation of the type.
func (t Type) String() string {
	switch t {
	default:
		return ""
	case Null:
		return "Null"
	case False:
		return "False"
	case Number:
		return "Number"
	case String:
		return "String"
	case True:
		return "True"
	case JSON:
		return "JSON"
	}
}

var (
	cacheStructTagInfo = newCache[uint32, *TagInfo]()
)

// 获取 string 的起始地址
func strToUintptr(p string) uintptr {
	return *(*uintptr)(unsafe.Pointer(&p))
}

func LoadTagNode(typ reflect.Type, hash uint32) (n *TagInfo, err error) {
	n, ok := cacheStructTagInfo.Get(hash)
	if ok {
		return n, nil
	}
	// log.Printf("type:%s", typ.String())
	ti, err := NewStructTagInfo(typ, false)
	if err != nil {
		return nil, err
	}
	n = (*TagInfo)(ti)
	cacheStructTagInfo.Set(hash, n)
	return
}

type Value struct {
	typ  uintptr
	ptr  unsafe.Pointer
	flag uintptr
}

func reflectValueToPointer(v *reflect.Value) unsafe.Pointer {
	return (*Value)(unsafe.Pointer(v)).ptr
}

func bytesString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// emptyInterface is the header for an interface{} value.
type emptyInterface struct {
	typ  uintptr
	word unsafe.Pointer
}

func unpackEface(v interface{}) *emptyInterface {
	empty := (*emptyInterface)(unsafe.Pointer(&v))
	return empty
}
func UnpackEface(v interface{}) GoEface {
	return *(*GoEface)(unsafe.Pointer(&v))
}

type GoEface struct {
	Type  *GoType
	Value unsafe.Pointer
}

type GoType struct {
	Size       uintptr
	PtrData    uintptr
	Hash       uint32
	Flags      uint8
	Align      uint8
	FieldAlign uint8
	KindFlags  uint8
	Traits     unsafe.Pointer
	GCData     *byte
	Str        int32
	PtrToSelf  int32
}

func PtrElem(t *GoType) *GoType {
	return (*GoPtrType)(unsafe.Pointer(t)).Elem
}

type GoPtrType struct {
	GoType
	Elem *GoType
}
