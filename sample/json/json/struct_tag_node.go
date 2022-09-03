package json

import (
	"reflect"
	"unsafe"
)

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
	ti, err := NewTagInfo(typ)
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
	ti, err := NewTagInfo(typ)
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
