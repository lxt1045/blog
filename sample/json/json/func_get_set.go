package json

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strconv"
	"unsafe"

	lxterrs "github.com/lxt1045/errors"
)

type setFunc = func(store PoolStore, bs []byte) (pBase unsafe.Pointer)
type getFunc = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte)

func pointerOffset(p unsafe.Pointer, offset uintptr) (pOut unsafe.Pointer) {
	return unsafe.Pointer(uintptr(p) + uintptr(offset))
}

func ptrTypeFuncs[T any](builder *TypeBuilder, name string, ptrDeep int, fSet setFunc, fGet getFunc) (fSet1 setFunc, fGet1 getFunc) {
	var idx *uintptr = &[]uintptr{0}[0]
	var x T
	typ := reflect.TypeOf(x)
	builder.AppendTagField(name, typ, idx)
	fSet1 = func(store PoolStore, bs []byte) (pBase unsafe.Pointer) {
		p := pointerOffset(store.pool, *idx)
		*(**T)(store.obj) = (*T)(p)
		store.obj = p
		return fSet(store, bs)
	}
	fGet1 = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
		p := *(*unsafe.Pointer)(pObj)
		return fGet(p, in)
	}
	for i := 1; i < ptrDeep; i++ {
		var idxP *uintptr = &[]uintptr{0}[0]
		builder.AppendPointer(fmt.Sprintf("%s_%d", name, i), idxP)
		fSet0, fGet0 := fSet1, fGet1
		fSet1 = func(store PoolStore, bs []byte) (pBase unsafe.Pointer) {
			p := pointerOffset(store.pool, *idxP)
			*(**unsafe.Pointer)(store.obj) = (*unsafe.Pointer)(p)
			store.obj = p
			return fSet0(store, bs)
		}
		fGet1 = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
			p := *(*unsafe.Pointer)(pObj)
			return fGet0(p, in)
		}
	}
	return
}
func boolFuncs(builder *TypeBuilder, name string, ptrDeep int) (fSet setFunc, fGet getFunc) {
	fSet = func(store PoolStore, bs []byte) (pBase unsafe.Pointer) {
		pBase = store.obj
		if bs[0] == 't' {
			*(*bool)(store.obj) = true
		} else {
			*(*bool)(store.obj) = false
		}
		return
	}
	fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
		pBase = pObj
		if *(*bool)(pObj) {
			out = append(in, []byte("false")...)
		} else {
			out = append(in, []byte("true")...)
		}
		return
	}
	if ptrDeep > 0 {
		fSet, fGet = ptrTypeFuncs[bool](builder, name, ptrDeep, fSet, fGet)
	}
	return
}

func uint64Funcs(builder *TypeBuilder, name string, ptrDeep int) (fSet setFunc, fGet getFunc) {
	fSet = func(store PoolStore, bs []byte) (pBase unsafe.Pointer) {
		pBase = store.obj
		num, err := strconv.ParseUint(bytesString(bs), 10, 64)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(bs))
			return
		}
		*(*uint64)(store.obj) = num
		return
	}
	fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
		pBase = pObj
		num := *(*uint64)(pObj)
		str := strconv.FormatUint(num, 10)
		out = append(in, str...)
		return
	}
	if ptrDeep > 0 {
		fSet, fGet = ptrTypeFuncs[uint64](builder, name, ptrDeep, fSet, fGet)
	}
	return
}

func int64Funcs(builder *TypeBuilder, name string, ptrDeep int) (fSet setFunc, fGet getFunc) {
	fSet = func(store PoolStore, bs []byte) (pBase unsafe.Pointer) {
		pBase = store.obj
		num, err := strconv.ParseInt(bytesString(bs), 10, 64)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(bs))
			return
		}
		*(*int64)(store.obj) = num
		return
	}
	fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
		pBase = pObj
		num := *(*int64)(pObj)
		str := strconv.FormatInt(num, 10)
		out = append(in, str...)
		return
	}
	if ptrDeep > 0 {
		fSet, fGet = ptrTypeFuncs[int64](builder, name, ptrDeep, fSet, fGet)
	}
	return
}
func uint32Funcs(builder *TypeBuilder, name string, ptrDeep int) (fSet setFunc, fGet getFunc) {
	fSet = func(store PoolStore, bs []byte) (pBase unsafe.Pointer) {
		pBase = store.obj
		num, err := strconv.ParseUint(bytesString(bs), 10, 32)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(bs))
			return
		}
		*(*uint32)(store.obj) = uint32(num)
		return
	}
	fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
		pBase = pObj
		num := *(*uint32)(pObj)
		str := strconv.FormatUint(uint64(num), 10)
		out = append(in, str...)
		return
	}
	if ptrDeep > 0 {
		fSet, fGet = ptrTypeFuncs[uint32](builder, name, ptrDeep, fSet, fGet)
	}
	return
}
func int32Funcs(builder *TypeBuilder, name string, ptrDeep int) (fSet setFunc, fGet getFunc) {
	fSet = func(store PoolStore, bs []byte) (pBase unsafe.Pointer) {
		pBase = store.obj
		num, err := strconv.ParseInt(bytesString(bs), 10, 32)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(bs))
			return
		}
		*(*int32)(store.obj) = int32(num)
		return
	}
	fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
		pBase = pObj
		num := *(*int32)(pObj)
		str := strconv.FormatInt(int64(num), 10)
		out = append(in, str...)
		return
	}
	if ptrDeep > 0 {
		fSet, fGet = ptrTypeFuncs[int32](builder, name, ptrDeep, fSet, fGet)
	}
	return
}
func uint16Funcs(builder *TypeBuilder, name string, ptrDeep int) (fSet setFunc, fGet getFunc) {
	fSet = func(store PoolStore, bs []byte) (pBase unsafe.Pointer) {
		pBase = store.obj
		num, err := strconv.ParseUint(bytesString(bs), 10, 32)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(bs))
			return
		}
		*(*uint16)(store.obj) = uint16(num)
		return
	}
	fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
		pBase = pObj
		num := *(*uint16)(pObj)
		str := strconv.FormatUint(uint64(num), 10)
		out = append(in, str...)
		return
	}
	if ptrDeep > 0 {
		fSet, fGet = ptrTypeFuncs[uint16](builder, name, ptrDeep, fSet, fGet)
	}
	return
}
func int16Funcs(builder *TypeBuilder, name string, ptrDeep int) (fSet setFunc, fGet getFunc) {
	fSet = func(store PoolStore, bs []byte) (pBase unsafe.Pointer) {
		pBase = store.obj
		num, err := strconv.ParseInt(bytesString(bs), 10, 32)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(bs))
			return
		}
		*(*int16)(store.obj) = int16(num)
		return
	}
	fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
		pBase = pObj
		num := *(*int16)(pObj)
		str := strconv.FormatInt(int64(num), 10)
		out = append(in, str...)
		return
	}
	if ptrDeep > 0 {
		fSet, fGet = ptrTypeFuncs[int16](builder, name, ptrDeep, fSet, fGet)
	}
	return
}
func uint8Funcs(builder *TypeBuilder, name string, ptrDeep int) (fSet setFunc, fGet getFunc) {
	fSet = func(store PoolStore, bs []byte) (pBase unsafe.Pointer) {
		pBase = store.obj
		num, err := strconv.ParseUint(bytesString(bs), 10, 32)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(bs))
			return
		}
		*(*uint8)(store.obj) = uint8(num)
		return
	}
	fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
		pBase = pObj
		num := *(*uint8)(pObj)
		str := strconv.FormatUint(uint64(num), 10)
		out = append(in, str...)
		return
	}
	if ptrDeep > 0 {
		fSet, fGet = ptrTypeFuncs[uint8](builder, name, ptrDeep, fSet, fGet)
	}
	return
}
func int8Funcs(builder *TypeBuilder, name string, ptrDeep int) (fSet setFunc, fGet getFunc) {
	fSet = func(store PoolStore, bs []byte) (pBase unsafe.Pointer) {
		pBase = store.obj
		num, err := strconv.ParseInt(bytesString(bs), 10, 32)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(bs))
			return
		}
		*(*int8)(store.obj) = int8(num)
		return
	}
	fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
		pBase = pObj
		num := *(*int8)(pObj)
		str := strconv.FormatInt(int64(num), 10)
		out = append(in, str...)
		return
	}
	if ptrDeep > 0 {
		fSet, fGet = ptrTypeFuncs[int8](builder, name, ptrDeep, fSet, fGet)
	}
	return
}
func float64Funcs(builder *TypeBuilder, name string, ptrDeep int) (fSet setFunc, fGet getFunc) {
	fSet = func(store PoolStore, bs []byte) (pBase unsafe.Pointer) {
		pBase = store.obj
		f, err := strconv.ParseFloat(bytesString(bs), 64)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(bs))
			return
		}
		*(*float64)(store.obj) = f
		return
	}
	fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
		pBase = pObj
		num := *(*float64)(pObj)
		out = strconv.AppendFloat(in, float64(num), 'f', -1, 64)
		return
	}
	if ptrDeep > 0 {
		fSet, fGet = ptrTypeFuncs[float64](builder, name, ptrDeep, fSet, fGet)
	}
	return
}
func float32Funcs(builder *TypeBuilder, name string, ptrDeep int) (fSet setFunc, fGet getFunc) {
	fSet = func(store PoolStore, bs []byte) (pBase unsafe.Pointer) {
		pBase = store.obj
		f, err := strconv.ParseFloat(bytesString(bs), 64)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(bs))
			return
		}
		*(*float64)(store.obj) = f
		return
	}
	fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
		pBase = pObj
		num := *(*float64)(pObj)
		out = strconv.AppendFloat(in, float64(num), 'f', -1, 64)
		return
	}
	if ptrDeep > 0 {
		fSet, fGet = ptrTypeFuncs[float32](builder, name, ptrDeep, fSet, fGet)
	}
	return
}
func stringFuncs(builder *TypeBuilder, name string, ptrDeep int) (fSet setFunc, fGet getFunc) {
	fSet = func(store PoolStore, bs []byte) (pBase unsafe.Pointer) {
		pBase = store.obj
		*(*string)(store.obj) = *(*string)(unsafe.Pointer(&bs))
		return
	}
	fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
		pBase = pObj
		str := *(*string)(pObj)
		out = append(in, str...)
		return
	}
	if ptrDeep > 0 {
		fSet, fGet = ptrTypeFuncs[string](builder, name, ptrDeep, fSet, fGet)
	}
	return
}
func bytesFuncs(builder *TypeBuilder, name string, ptrDeep int) (fSet setFunc, fGet getFunc) {
	// []byte 是一种特殊的底层数据类型，需要 base64 编码
	fSet = func(store PoolStore, bs []byte) (pBase unsafe.Pointer) {
		pBase = store.obj
		pbs := (*[]byte)(store.obj)
		*pbs = make([]byte, len(bs)*2)
		n, err := base64.StdEncoding.Decode(*pbs, bs)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(bs))
			return
		}
		*pbs = (*pbs)[:n]
		return
	}
	fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
		pBase = pObj
		bs := *(*[]byte)(pObj)
		l, need := len(in), base64.StdEncoding.EncodedLen(len(bs))
		if l+need > cap(in) {
			//没有足够空间
			in = append(in, make([]byte, need)...)
		}
		base64.StdEncoding.Encode(in[l:l+need], bs)
		out = in[:l+need]
		return
	}
	if ptrDeep > 0 {
		fSet, fGet = ptrTypeFuncs[[]byte](builder, name, ptrDeep, fSet, fGet)
	}
	return
}
func sliceFuncs(builder *TypeBuilder, name string, ptrDeep int) (fSet setFunc, fGet getFunc) {
	fSet = func(store PoolStore, bs []byte) (pBase unsafe.Pointer) {
		pBase = store.obj
		return
	}
	fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
		p := *(*unsafe.Pointer)(pObj)
		return p, in
	}
	if ptrDeep > 0 {
		fSet, fGet = ptrTypeFuncs[[]uint8](builder, name, ptrDeep, fSet, fGet)
	}
	return
}

func structChildFuncs(builder *TypeBuilder, name string, ptrDeep int, typ reflect.Type) (fSet setFunc, fGet getFunc) {
	if ptrDeep > 0 {
		var idx *uintptr = &[]uintptr{0}[0]
		builder.AppendTagField(name, typ, idx)
		fSet = func(store PoolStore, bs []byte) (pBase unsafe.Pointer) {
			p := *(*unsafe.Pointer)(store.obj)
			if p == nil {
				/*
					此处可以使用缓存池 sync.Pool(不合适，没有合适的回收时机，，，)， 或者自己实现生成池
				*/
				// p = fNew()
				p = pointerOffset(store.pool, *idx)
				*(*unsafe.Pointer)(store.obj) = p
			}
			return unsafe.Pointer(p)
		}
		fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
			p := *(*unsafe.Pointer)(pObj)
			return p, in
		}

		for i := 0; i < ptrDeep; i++ {
			var idxP *uintptr = &[]uintptr{0}[0]
			builder.AppendPointer(fmt.Sprintf("%s_%d", name, i), idxP)
			fSet0, fGet0 := fSet, fGet
			fSet1 := func(store PoolStore, bs []byte) (pBase unsafe.Pointer) {
				p := pointerOffset(store.pool, *idxP)
				*(**unsafe.Pointer)(store.obj) = (*unsafe.Pointer)(p)
				store.obj = p
				return fSet0(store, bs)
			}
			fGet1 := func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
				p := *(*unsafe.Pointer)(pObj)
				return fGet0(p, in)
			}
			fSet, fGet = fSet1, fGet1
		}
	}
	return
}

// 匿名嵌入
func anonymousStructFuncs(builder *TypeBuilder, name string, ptrDeep int, typ reflect.Type, offset uintptr, fSet0 setFunc, fGet0 getFunc) (fSet setFunc, fGet getFunc) {
	if ptrDeep <= 0 {
		fSet = func(store PoolStore, bs []byte) (pBase unsafe.Pointer) {
			pBase = store.obj
			pSon := pointerOffset(store.obj, offset)
			store.obj = pSon
			return fSet0(store, bs)
		}
		fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
			pBase = pObj
			pSon := pointerOffset(pObj, offset)
			return fGet0(pSon, in)
		}
	} else {
		var idx *uintptr = &[]uintptr{0}[0]
		builder.AppendTagField(name, typ, idx)
		fSet = func(store PoolStore, bs []byte) (pBase unsafe.Pointer) {
			p := *(*unsafe.Pointer)(store.obj)
			if p == nil {
				// p = fNew()
				p = pointerOffset(store.pool, *idx)
				*(*unsafe.Pointer)(store.obj) = p
			}
			store.obj = p
			return fSet0(store, bs)
		}
		fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
			p := *(*unsafe.Pointer)(pObj)
			if p != nil {
				return fGet0(p, in)
			}
			return nil, in
		}

		for i := 0; i < ptrDeep; i++ {
			var idxP *uintptr = &[]uintptr{0}[0]
			builder.AppendPointer(fmt.Sprintf("%s_%d", name, i), idxP)
			fSet0, fGet0 := fSet, fGet
			fSet1 := func(store PoolStore, bs []byte) (pBase unsafe.Pointer) {
				p := pointerOffset(store.pool, *idxP)
				*(**unsafe.Pointer)(store.obj) = (*unsafe.Pointer)(p)
				store.obj = p
				return fSet0(store, bs)
			}
			fGet1 := func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
				p := *(*unsafe.Pointer)(pObj)
				return fGet0(p, in)
			}
			fSet, fGet = fSet1, fGet1
		}
	}
	return
}
func iterfaceFuncs(builder *TypeBuilder, name string, ptrDeep int) (fSet setFunc, fGet getFunc) {
	fSet = func(store PoolStore, bs []byte) (pBase unsafe.Pointer) {
		return store.obj
	}
	if ptrDeep > 0 {
		fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
			p := *(*unsafe.Pointer)(pObj)
			return p, in
		}
		fSet, fGet = ptrTypeFuncs[bool](builder, name, ptrDeep, fSet, fGet)
	}
	return
}

func mapFuncs(builder *TypeBuilder, name string, ptrDeep int) (fSet setFunc, fGet getFunc) {
	fSet = func(store PoolStore, bs []byte) (pBase unsafe.Pointer) {
		p := (*map[string]interface{})(store.obj)
		*p = make(map[string]interface{})
		return store.obj
	}

	if ptrDeep > 0 {
		fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
			p := *(*unsafe.Pointer)(pObj)
			return p, in
		}
		fSet, fGet = ptrTypeFuncs[bool](builder, name, ptrDeep, fSet, fGet)
	}
	return
}
