package json

import (
	"reflect"
	"unsafe"
)

type SliceHeader struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}

type StringHeader struct {
	Data unsafe.Pointer
	Len  int
}

type Value struct {
	typ  *GoType
	ptr  unsafe.Pointer
	flag uintptr
}

func reflectValueToPointer(v *reflect.Value) unsafe.Pointer {
	return (*Value)(unsafe.Pointer(v)).ptr
}
func reflectValueToValue(v *reflect.Value) *Value {
	return (*Value)(unsafe.Pointer(v))
}

func bytesString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func stringBytes(str string) []byte {
	return (*(*[]byte)(unsafe.Pointer(&str)))[:len(str):len(str)]
}
func bytesCopyToString(b []byte, str *string) {
	*str = *(*string)(unsafe.Pointer(&b))
}

// emptyInterface is the header for an interface{} value.
type emptyInterface struct {
	typ  uintptr
	word unsafe.Pointer
}

type nonEmptyInterface struct {
	itab *struct {
		ityp *GoType // static interface type
		typ  *GoType // dynamic concrete type
		hash uint32  // copy of typ.hash
		_    [4]byte
		fun  [100000]unsafe.Pointer // method table
	}
	word *GoType
}

func UnpackNonEface(p unsafe.Pointer) *GoType {
	neface := (*nonEmptyInterface)(p)
	return neface.word
}

func UnpackType(t reflect.Type) *GoType {
	return (*GoType)((*GoIface)(unsafe.Pointer(&t)).Value)
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

type GoIface struct {
	Itab  *GoItab
	Value unsafe.Pointer
}
type GoItab struct {
	it unsafe.Pointer
	vt *GoType
	hv uint32
	_  [4]byte
	fn [1]uintptr
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

type interfacetype struct {
	typ GoType
	//  /src/runtime/type.go: interfacetype
}

func PtrElem(t *GoType) *GoType {
	return (*GoPtrType)(unsafe.Pointer(t)).Elem
}

type GoPtrType struct {
	GoType
	Elem *GoType
}
