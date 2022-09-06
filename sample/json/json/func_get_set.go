package json

import (
	"encoding/base64"
	"strconv"
	"unsafe"

	lxterrs "github.com/lxt1045/errors"
)

type setFunc = func(pObj unsafe.Pointer, bs []byte) (pBase unsafe.Pointer)
type getFunc = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte)

func pointerOffset(p unsafe.Pointer, offset uintptr) (pOut unsafe.Pointer) {
	return unsafe.Pointer(uintptr(p) + uintptr(offset))
}

func ptrTypeFuncs[T any](ptrDeep int, fSet setFunc, fGet getFunc) (fSet1 setFunc, fGet1 getFunc) {
	fSet1 = func(pObj unsafe.Pointer, bs []byte) (pBase unsafe.Pointer) {
		var obj T
		*(**T)(pObj) = &obj
		return fSet(unsafe.Pointer(&obj), bs)
	}
	fGet1 = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
		p := *(*unsafe.Pointer)(pObj)
		return fGet(p, in)
	}
	for i := 1; i < ptrDeep; i++ {
		fSet0, fGet0 := fSet1, fGet1
		fSet1 = func(pObj unsafe.Pointer, bs []byte) (pBase unsafe.Pointer) {
			var p unsafe.Pointer
			*(**unsafe.Pointer)(pObj) = &p
			return fSet0(unsafe.Pointer(&p), bs)
		}
		fGet1 = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
			p := *(*unsafe.Pointer)(pObj)
			return fGet0(p, in)
		}
	}
	return
}

func boolFuncs(ptrDeep int) (fSet setFunc, fGet getFunc) {
	fSet = func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer) {
		pBase = pObj
		if raw[0] == 't' {
			*(*bool)(pObj) = true
		} else {
			*(*bool)(pObj) = false
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
		fSet, fGet = ptrTypeFuncs[bool](ptrDeep, fSet, fGet)
	}
	return
}

func uint64Funcs(ptrDeep int) (fSet setFunc, fGet getFunc) {
	fSet = func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer) {
		pBase = pObj
		num, err := strconv.ParseUint(bytesString(raw), 10, 64)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(raw))
			return
		}
		*(*uint64)(pObj) = num
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
		fSet, fGet = ptrTypeFuncs[uint64](ptrDeep, fSet, fGet)
	}
	return
}

func int64Funcs(ptrDeep int) (fSet setFunc, fGet getFunc) {
	fSet = func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer) {
		pBase = pObj
		num, err := strconv.ParseInt(bytesString(raw), 10, 64)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(raw))
			return
		}
		*(*int64)(pObj) = num
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
		fSet, fGet = ptrTypeFuncs[int64](ptrDeep, fSet, fGet)
	}
	return
}
func uint32Funcs(ptrDeep int) (fSet setFunc, fGet getFunc) {
	fSet = func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer) {
		pBase = pObj
		num, err := strconv.ParseUint(bytesString(raw), 10, 32)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(raw))
			return
		}
		*(*uint32)(pObj) = uint32(num)
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
		fSet, fGet = ptrTypeFuncs[uint32](ptrDeep, fSet, fGet)
	}
	return
}
func int32Funcs(ptrDeep int) (fSet setFunc, fGet getFunc) {
	fSet = func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer) {
		pBase = pObj
		num, err := strconv.ParseInt(bytesString(raw), 10, 32)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(raw))
			return
		}
		*(*int32)(pObj) = int32(num)
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
		fSet, fGet = ptrTypeFuncs[int32](ptrDeep, fSet, fGet)
	}
	return
}
func uint16Funcs(ptrDeep int) (fSet setFunc, fGet getFunc) {
	fSet = func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer) {
		pBase = pObj
		num, err := strconv.ParseUint(bytesString(raw), 10, 32)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(raw))
			return
		}
		*(*uint16)(pObj) = uint16(num)
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
		fSet, fGet = ptrTypeFuncs[uint16](ptrDeep, fSet, fGet)
	}
	return
}
func int16Funcs(ptrDeep int) (fSet setFunc, fGet getFunc) {
	fSet = func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer) {
		pBase = pObj
		num, err := strconv.ParseInt(bytesString(raw), 10, 32)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(raw))
			return
		}
		*(*int16)(pObj) = int16(num)
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
		fSet, fGet = ptrTypeFuncs[int16](ptrDeep, fSet, fGet)
	}
	return
}
func uint8Funcs(ptrDeep int) (fSet setFunc, fGet getFunc) {
	fSet = func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer) {
		pBase = pObj
		num, err := strconv.ParseUint(bytesString(raw), 10, 32)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(raw))
			return
		}
		*(*uint8)(pObj) = uint8(num)
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
		fSet, fGet = ptrTypeFuncs[uint8](ptrDeep, fSet, fGet)
	}
	return
}
func int8Funcs(ptrDeep int) (fSet setFunc, fGet getFunc) {
	fSet = func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer) {
		pBase = pObj
		num, err := strconv.ParseInt(bytesString(raw), 10, 32)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(raw))
			return
		}
		*(*int8)(pObj) = int8(num)
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
		fSet, fGet = ptrTypeFuncs[int8](ptrDeep, fSet, fGet)
	}
	return
}
func float64Funcs(ptrDeep int) (fSet setFunc, fGet getFunc) {
	fSet = func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer) {
		pBase = pObj
		f, err := strconv.ParseFloat(bytesString(raw), 64)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(raw))
			return
		}
		*(*float64)(pObj) = f
		return
	}
	fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
		pBase = pObj
		num := *(*float64)(pObj)
		out = strconv.AppendFloat(in, float64(num), 'f', -1, 64)
		return
	}
	if ptrDeep > 0 {
		fSet, fGet = ptrTypeFuncs[float64](ptrDeep, fSet, fGet)
	}
	return
}
func float32Funcs(ptrDeep int) (fSet setFunc, fGet getFunc) {
	fSet = func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer) {
		pBase = pObj
		f, err := strconv.ParseFloat(bytesString(raw), 64)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(raw))
			return
		}
		*(*float64)(pObj) = f
		return
	}
	fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
		pBase = pObj
		num := *(*float64)(pObj)
		out = strconv.AppendFloat(in, float64(num), 'f', -1, 64)
		return
	}
	if ptrDeep > 0 {
		fSet, fGet = ptrTypeFuncs[float32](ptrDeep, fSet, fGet)
	}
	return
}
func stringFuncs(ptrDeep int) (fSet setFunc, fGet getFunc) {
	fSet = func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer) {
		pBase = pObj
		*(*string)(pObj) = *(*string)(unsafe.Pointer(&raw))
		return
	}
	fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
		pBase = pObj
		str := *(*string)(pObj)
		out = append(in, str...)
		return
	}
	if ptrDeep > 0 {
		fSet, fGet = ptrTypeFuncs[string](ptrDeep, fSet, fGet)
	}
	return
}
func bytesFuncs(ptrDeep int) (fSet setFunc, fGet getFunc) {
	// []byte 是一种特殊的底层数据类型，需要 base64 编码
	fSet = func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer) {
		pBase = pObj
		pbs := (*[]byte)(pObj)
		*pbs = make([]byte, len(raw)*2)
		n, err := base64.StdEncoding.Decode(*pbs, raw)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(raw))
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
		fSet, fGet = ptrTypeFuncs[[]byte](ptrDeep, fSet, fGet)
	}
	return
}
func sliceFuncs(ptrDeep int) (fSet setFunc, fGet getFunc) {
	fSet = func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer) {
		pBase = pObj
		return
	}
	fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
		p := *(*unsafe.Pointer)(pObj)
		return p, in
	}
	if ptrDeep > 0 {
		fSet, fGet = ptrTypeFuncs[[]uint8](ptrDeep, fSet, fGet)
	}
	return
}

func structChildFuncs(ptrDeep int, fNew func() unsafe.Pointer) (fSet setFunc, fGet getFunc) {
	if ptrDeep > 0 {
		fSet = func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer) {
			p := *(*unsafe.Pointer)(pObj)
			if p == nil {
				/*
					此处可以使用缓存池 sync.Pool(不合适，没有合适的回收时机，，，)， 或者自己实现生成池
				*/
				p = fNew()
				*(*unsafe.Pointer)(pObj) = p
			}
			return unsafe.Pointer(p)
		}
		fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
			p := *(*unsafe.Pointer)(pObj)
			return p, in
		}

		for i := 0; i < ptrDeep; i++ {
			fSet0, fGet0 := fSet, fGet
			fSet1 := func(pObj unsafe.Pointer, bs []byte) (pBase unsafe.Pointer) {
				var p unsafe.Pointer
				*(**unsafe.Pointer)(pObj) = &p
				return fSet0(unsafe.Pointer(&p), bs)
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
func anonymousStructFuncs(ptrDeep int, offset uintptr, fSet0 setFunc, fGet0 getFunc,
	fNew func() unsafe.Pointer) (fSet setFunc, fGet getFunc) {
	if ptrDeep <= 0 {
		fSet = func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer) {
			pBase = pObj
			pSon := pointerOffset(pObj, offset)
			return fSet0(pSon, raw)
		}
		fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
			pBase = pObj
			pSon := pointerOffset(pObj, offset)
			return fGet0(pSon, in)
		}
	} else {
		fSet = func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer) {
			p := *(*unsafe.Pointer)(pObj)
			if p == nil {
				p = fNew()
				*(*unsafe.Pointer)(pObj) = p
			}
			return fSet0(p, raw)
		}
		fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
			p := *(*unsafe.Pointer)(pObj)
			if p != nil {
				return fGet0(p, in)
			}
			return nil, in
		}

		for i := 0; i < ptrDeep; i++ {
			fSet0, fGet0 := fSet, fGet
			fSet1 := func(pObj unsafe.Pointer, bs []byte) (pBase unsafe.Pointer) {
				var p unsafe.Pointer
				*(**unsafe.Pointer)(pObj) = &p
				return fSet0(unsafe.Pointer(&p), bs)
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
func iterfaceFuncs(ptrDeep int) (fSet setFunc, fGet getFunc) {
	fSet = func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer) {
		return pObj
	}
	if ptrDeep > 0 {
		fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
			p := *(*unsafe.Pointer)(pObj)
			return p, in
		}
		fSet, fGet = ptrTypeFuncs[bool](ptrDeep, fSet, fGet)
	}
	return
}

func mapFuncs(ptrDeep int) (fSet setFunc, fGet getFunc) {
	fSet = func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer) {
		p := (*map[string]interface{})(pObj)
		*p = make(map[string]interface{})
		return pObj
	}

	if ptrDeep > 0 {
		fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
			p := *(*unsafe.Pointer)(pObj)
			return p, in
		}
		fSet, fGet = ptrTypeFuncs[bool](ptrDeep, fSet, fGet)
	}
	return
}
