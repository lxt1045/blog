package json

import (
	"fmt"
	"reflect"
	"unsafe"
)

/*
	Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Uintptr
	Float32
	Float64
	Complex64
	Complex128
	Array
	Chan
	Func
	Interface
	Map
	Ptr
	Slice
	String
	Struct
	UnsafePointer
*/

func tryThrowErr(err error) {
	if err != nil {
		panic(err)
	}
}
func tryThrowMemoryLengthErr(pTag *TagInfo, bs []byte) {
	t := pTag.StructField.Type
	if len(bs) != int(t.Size()) {
		err := fmt.Errorf("%v类型[%v]的输入参数长度有误,应该是%d,而不是:%d",
			pTag.StructField.Name, t, t.Size(), len(bs))
		tryThrowErr(err)
	}
}

func getField(field reflect.StructField, pStruct unsafe.Pointer, pOut unsafe.Pointer) {
	pValue := unsafe.Pointer(uintptr(pStruct) + uintptr(field.Offset))
	typ := field.Type

	for ; typ.Kind() == reflect.Ptr && uintptr(pValue) != 0; typ = typ.Elem() {
		ppValue := (*unsafe.Pointer)(pValue)
		pValue = *ppValue
	}
	if uintptr(pValue) != 0 {
		from := reflect.SliceHeader{
			Data: uintptr(pValue),
			Len:  int(typ.Size()),
			Cap:  int(typ.Size()),
		}
		to := reflect.SliceHeader{
			Data: uintptr(pOut),
			Len:  int(typ.Size()),
			Cap:  int(typ.Size()),
		}
		copy(*(*[]byte)(unsafe.Pointer(&to)), *(*[]byte)(unsafe.Pointer(&from)))
	}
	return
}

func setField(field reflect.StructField, pStruct unsafe.Pointer, pIn unsafe.Pointer) {
	pValue := unsafe.Pointer(uintptr(pStruct) + uintptr(field.Offset))
	typ := field.Type
	if typ.Kind() != reflect.Ptr {
		from := reflect.SliceHeader{
			Data: uintptr(pIn),
			Len:  int(typ.Size()),
			Cap:  int(typ.Size()),
		}
		to := reflect.SliceHeader{
			Data: uintptr(pValue),
			Len:  int(typ.Size()),
			Cap:  int(typ.Size()),
		}
		sFrom, sTo := *(*[]byte)(unsafe.Pointer(&from)), *(*[]byte)(unsafe.Pointer(&to))
		copy(sTo, sFrom)
		return
	}
	setPointerField(field, pStruct, pIn)
	return
}

func setPointerField(field reflect.StructField, pStruct unsafe.Pointer, pIn unsafe.Pointer) {
	pValue := unsafe.Pointer(uintptr(pStruct) + uintptr(field.Offset))
	typ := field.Type
	for ; typ.Elem().Kind() == reflect.Ptr; typ = typ.Elem() {
		var p *unsafe.Pointer
		ppValue := (*unsafe.Pointer)(pValue)
		pValue = unsafe.Pointer(&p)
		*ppValue = pValue
	}
	*(*unsafe.Pointer)(pValue) = *(*unsafe.Pointer)(pIn)
	return
}

func setFieldStringPointer(field reflect.StructField, pStruct unsafe.Pointer, pIn unsafe.Pointer) {
	pValue := unsafe.Pointer(uintptr(pStruct) + uintptr(field.Offset))
	typ := field.Type
	if typ.Kind() != reflect.Ptr {
		*(*string)(pValue) = *(*string)(pIn)
		return
	}
	setPointerField(field, pStruct, pIn)
	return
}

func setFieldString(pValue unsafe.Pointer, str string) {
	*(*string)(pValue) = str
	return
}

func setBool(pValue unsafe.Pointer, b bool) {
	*(*bool)(pValue) = b
	return
}
func setFieldInt(field reflect.StructField, pStruct unsafe.Pointer, pIn unsafe.Pointer) {
	pValue := unsafe.Pointer(uintptr(pStruct) + uintptr(field.Offset))
	*(*int)(pValue) = *(*int)(pIn)
	return
}

func setFieldSlice(pValue unsafe.Pointer, pIn unsafe.Pointer) {
	*(*reflect.SliceHeader)(pValue) = *(*reflect.SliceHeader)(pIn)
	return
}

func pointerOffset(p unsafe.Pointer, offset uintptr) (pOut unsafe.Pointer) {
	return unsafe.Pointer(uintptr(p) + uintptr(offset))
}

func getBaseTypeFuncs[T any](ptrDeep int,
	fSet func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer, err error),
	fGet func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte),
) (
	fSet1 func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer, err error),
	fGet1 func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte),
) {
	fSet1 = func(pObj unsafe.Pointer, bs []byte) (pBase unsafe.Pointer, err error) {
		var obj T
		*(**T)(pObj) = &obj
		return fSet(unsafe.Pointer(&obj), bs)
	}
	fGet1 = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
		p := *(*unsafe.Pointer)(pObj)
		return fGet(p, in)
	}
	fSet, fGet = fSet1, fGet1
	for i := 1; i < ptrDeep; i++ {
		fSet1 := func(pObj unsafe.Pointer, bs []byte) (pBase unsafe.Pointer, err error) {
			var p unsafe.Pointer
			*(**unsafe.Pointer)(pObj) = &p
			return fSet(unsafe.Pointer(&p), bs)
		}
		fGet1 := func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
			p := *(*unsafe.Pointer)(pObj)
			return fGet(p, in)
		}
		fSet, fGet = fSet1, fGet1
	}
	return
}
