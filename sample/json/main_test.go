package main_test

import (
	"encoding/json"
	"reflect"
	"testing"
	"unsafe"
)

func getField(field reflect.StructField, pStruct unsafe.Pointer, fGet func(unsafe.Pointer)) {
	pValue := unsafe.Pointer(uintptr(pStruct) + uintptr(field.Offset))
	typ := field.Type

	//多层string的话,只保留最里面一层
	if typ.Kind() != reflect.Ptr {
		fGet(pValue)
		return
	}
	for ; typ.Kind() == reflect.Ptr && uintptr(pValue) != 0; typ = typ.Elem() {
		ppValue := (*unsafe.Pointer)(pValue)
		pValue = *ppValue
	}
	if uintptr(pValue) != 0 {
		fGet(pValue)
	}
	return
}

func getField1(field reflect.StructField, pStruct unsafe.Pointer, pOut unsafe.Pointer) {
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

func setField(field reflect.StructField, pStruct unsafe.Pointer, fSet func(unsafe.Pointer)) {
	pValue := unsafe.Pointer(uintptr(pStruct) + uintptr(field.Offset))
	typ := field.Type

	//多层string的话,只保留最里面一层
	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Ptr {
		fSet(pValue)
		return
	}
	for ; typ.Elem().Kind() == reflect.Ptr; typ = typ.Elem() {
		var p *unsafe.Pointer
		ppValue := (*unsafe.Pointer)(pValue)
		pValue = unsafe.Pointer(&p)
		*ppValue = pValue
	}
	fSet(pValue)
	return
}

func setField1(field reflect.StructField, pStruct unsafe.Pointer, pIn unsafe.Pointer) {
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
		copy(*(*[]byte)(unsafe.Pointer(&to)), *(*[]byte)(unsafe.Pointer(&from)))
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
	*(**unsafe.Pointer)(pValue) = (*unsafe.Pointer)(pIn)
	return
}
func Test1(t *testing.T) {
	st := struct {
		F      string
		SDeep  **string
		SDeep2 *******string
		I      int
		B      bool
	}{}
	{
		typ := reflect.TypeOf(&st)
		field := typ.Elem().Field(0)
		str := "hello, world"
		setField1(field, unsafe.Pointer(&st), unsafe.Pointer(&str))
	}
	{
		typ := reflect.TypeOf(&st)
		field := typ.Elem().Field(3)
		str := 888
		setField1(field, unsafe.Pointer(&st), unsafe.Pointer(&str))
	}
	{
		typ := reflect.TypeOf(&st)
		field := typ.Elem().Field(4)
		str := true
		setField1(field, unsafe.Pointer(&st), unsafe.Pointer(&str))
	}
	{
		typ := reflect.TypeOf(&st)
		field := typ.Elem().Field(1)
		str := "hello, 世界"
		setPointerField(field, unsafe.Pointer(&st), unsafe.Pointer(&str))
	}
	{
		typ := reflect.TypeOf(&st)
		field := typ.Elem().Field(0)
		getField(field, unsafe.Pointer(&st), func(p unsafe.Pointer) {
			pStr := (*string)(p)
			t.Logf("0:%+v", *pStr)
		})
	}
	{
		typ := reflect.TypeOf(&st)
		field := typ.Elem().Field(1)
		str := ""
		getField1(field, unsafe.Pointer(&st), unsafe.Pointer(&str))
		t.Logf("111:%+v", str)
	}
	{
		typ := reflect.TypeOf(&st)
		field := typ.Elem().Field(2)
		s := "22"
		getField(field, unsafe.Pointer(&st), func(p unsafe.Pointer) {
			s = *(*string)(p)
		})
		t.Logf("2:%+v", s)
	}
	bs, err := json.Marshal(&st)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("psh:%+v", string(bs))

}

func Benchmark_getFieldUnderlyingPointer(b *testing.B) {
	type Struct0 struct {
		String  string `json:"Type"`
		SDeep1  *string
		SDeep10 **********string
	}
	type Benchmark struct {
		name string
		f    func()
	}
	runs := []Benchmark{}
	st := Struct0{}
	typ := reflect.TypeOf(st)
	{
		field := typ.Field(0) //string
		runs = append(runs, Benchmark{
			name: "string",
			f: func() {
				setField(field, unsafe.Pointer(&st), func(p unsafe.Pointer) {
					pv := (*string)(p)
					*pv = "hello, 世界0"
				})
			},
		})
	}
	{
		field := typ.Field(0) //string
		str := "hello, 世界0"
		runs = append(runs, Benchmark{
			name: "string",
			f: func() {
				setField1(field, unsafe.Pointer(&st), unsafe.Pointer(&str))
			},
		})
	}
	{
		field := typ.Field(1) //string
		runs = append(runs, Benchmark{
			name: "*string",
			f: func() {
				setField(field, unsafe.Pointer(&st), func(p unsafe.Pointer) {
					pStr := (**string)(p)
					*pStr = &[]string{"hello, 世界1"}[0]
				})
			},
		})
	}
	{
		field := typ.Field(2) //string
		runs = append(runs, Benchmark{
			name: "**********string",
			f: func() {
				setField(field, unsafe.Pointer(&st), func(p unsafe.Pointer) {
					pStr := (**string)(p)
					*pStr = &[]string{"hello, 世界2"}[0]
				})
			},
		})
	}

	for _, r := range runs[:] {
		b.Run(r.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				r.f()
			}
			b.SetBytes(int64(b.N))
			b.StopTimer()
		})
	}
}

//*/
