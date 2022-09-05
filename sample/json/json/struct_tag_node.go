package json

import (
	"reflect"
	"sync/atomic"
	"unsafe"
)

const N = 1024

// /*

var (
	bsCache        = NewSliceCache[[]byte](N)
	strCache       = NewSliceCache[string](N)
	interfaceCache = NewSliceCache[interface{}](N)
	mapCache       = NewSliceCache[map[string]interface{}](N)
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
	cacheStructTagInfoP   = newCache[uintptr, *tagNode]()
	cacheStructTagInfoStr = newCache[string, *tagNodeStr]()
)

// 获取 string 的起始地址
func strToUintptr(p string) uintptr {
	return *(*uintptr)(unsafe.Pointer(&p))
}
func LoadTagNode(typ reflect.Type) (n *tagNode, err error) {
	pname := strToUintptr(typ.String()) //typ.Name() 在匿名struct时，是空的
	ppkg := strToUintptr(typ.PkgPath())
	n, ok := cacheStructTagInfoP.Get(pname)
	if ok {
		if n.pkgPath == ppkg {
			return
		}
		if n, ok := n.pkgCache.Get(ppkg); ok {
			return n, nil
		}
	}
	ti, err := NewStructTagInfo(typ)
	if err != nil {
		return nil, err
	}
	n = &tagNode{
		pkgPath:  ppkg,
		tagInfo:  ti,
		pkgCache: newCache[uintptr, *tagNode](),
	}
	if !ok {
		cacheStructTagInfoP.Set(pname, n)
	} else {
		n.pkgCache.Set(ppkg, n)
	}
	return
}
func LoadTagNodeStr(typ reflect.Type) (n *tagNodeStr) {
	pname := typ.String()
	ppkg := typ.PkgPath()
	n, ok := cacheStructTagInfoStr.Get(pname)
	if ok {
		if n.pkgPath == ppkg {
			return
		}
		if n, ok := n.pkgCache.Get(ppkg); ok {
			return n
		}
	}
	ti, err := NewStructTagInfo(typ)
	if err != nil {
		panic(err)
	}
	n = &tagNodeStr{
		pkgPath:  ppkg,
		tagInfo:  ti,
		pkgCache: newCache[string, *tagNodeStr](),
	}
	if !ok {
		cacheStructTagInfoStr.Set(pname, n)
	} else {
		n.pkgCache.Set(ppkg, n)
	}
	return
}

type tagNodeStr struct {
	pkgPath  string
	tagInfo  *TagInfo
	pkgCache cache[string, *tagNodeStr] //如果 name 相等，则从这个缓存中获取
}

type tagNode struct {
	pkgPath  uintptr
	tagInfo  *TagInfo
	pkgCache cache[uintptr, *tagNode] //如果 name 相等，则从这个缓存中获取
}

const PANIC = true

func tryPanic(e any) {
	if PANIC {
		panic(e)
	}
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
