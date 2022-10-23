package json

import (
	"encoding/base64"
	"strconv"
	"unsafe"

	lxterrs "github.com/lxt1045/errors"
)

type setFunc = func(store PoolStore, bs string) (pBase unsafe.Pointer)
type getFunc = func(store Store, in []byte) (out []byte)

func pointerOffset(p unsafe.Pointer, offset uintptr) (pOut unsafe.Pointer) {
	return unsafe.Pointer(uintptr(p) + uintptr(offset))
}

func boolSet(store PoolStore, bs string) (pBase unsafe.Pointer) {
	pBase = store.obj
	if bs[0] == 't' {
		*(*bool)(store.obj) = true
	} else {
		*(*bool)(store.obj) = false
	}
	return
}
func boolGet(store Store, in []byte) (out []byte) {
	pObj := store.obj
	if *(*bool)(pObj) {
		out = append(in, "true"...)
	} else {
		out = append(in, "false"...)
	}
	return
}
func boolFuncs(pidx *uintptr) (fSet setFunc, fGet getFunc) {
	if pidx == nil {
		return boolSet, boolGet
	}
	fSet = func(store PoolStore, bs string) (pBase unsafe.Pointer) {
		store.obj = store.Idx(*pidx)
		return boolSet(store, bs)
	}
	fGet = func(store Store, in []byte) (out []byte) {
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		return boolGet(store, in)
	}
	return
}

func uint64Set(store PoolStore, bs string) (pBase unsafe.Pointer) {
	pBase = store.obj
	num, err := strconv.ParseUint(bs, 10, 64)
	if err != nil {
		err = lxterrs.Wrap(err, ErrStream(bs))
		return
	}
	*(*uint64)(store.obj) = num
	return
}
func uint64Get(store Store, in []byte) (out []byte) {
	pObj := store.obj
	num := *(*uint64)(pObj)
	str := strconv.FormatUint(num, 10)
	out = append(in, str...)
	return
}
func uint64Funcs(pidx *uintptr) (fSet setFunc, fGet getFunc) {
	if pidx == nil {
		return uint64Set, uint64Get
	}
	fSet = func(store PoolStore, bs string) (pBase unsafe.Pointer) {
		store.obj = store.Idx(*pidx)
		return uint64Set(store, bs)
	}
	fGet = func(store Store, in []byte) (out []byte) {
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		return uint64Get(store, in)
	}
	return
}

func int64Set(store PoolStore, bs string) (pBase unsafe.Pointer) {
	pBase = store.obj
	num, err := strconv.ParseInt(bs, 10, 64)
	if err != nil {
		err = lxterrs.Wrap(err, ErrStream(bs))
		return
	}
	*(*int64)(store.obj) = num
	return
}
func int64Get(store Store, in []byte) (out []byte) {
	pObj := store.obj
	num := *(*int64)(pObj)
	str := strconv.FormatInt(num, 10)
	out = append(in, str...)
	return
}
func int64Funcs(pidx *uintptr) (fSet setFunc, fGet getFunc) {
	if pidx == nil {
		return int64Set, int64Get
	}
	fSet = func(store PoolStore, bs string) (pBase unsafe.Pointer) {
		store.obj = store.Idx(*pidx)
		return int64Set(store, bs)
	}
	fGet = func(store Store, in []byte) (out []byte) {
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		return int64Get(store, in)
	}
	return
}

func uint32Set(store PoolStore, bs string) (pBase unsafe.Pointer) {
	pBase = store.obj
	num, err := strconv.ParseUint(bs, 10, 64)
	if err != nil {
		err = lxterrs.Wrap(err, ErrStream(bs))
		return
	}
	*(*uint32)(store.obj) = uint32(num)
	return
}
func uint32Get(store Store, in []byte) (out []byte) {
	pObj := store.obj
	num := *(*uint32)(pObj)
	str := strconv.FormatUint(uint64(num), 10)
	out = append(in, str...)
	return
}
func uint32Funcs(pidx *uintptr) (fSet setFunc, fGet getFunc) {
	if pidx == nil {
		return uint32Set, uint32Get
	}
	fSet = func(store PoolStore, bs string) (pBase unsafe.Pointer) {
		store.obj = store.Idx(*pidx)
		return uint32Set(store, bs)
	}
	fGet = func(store Store, in []byte) (out []byte) {
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		return uint32Get(store, in)
	}
	return
}

func int32Set(store PoolStore, bs string) (pBase unsafe.Pointer) {
	pBase = store.obj
	num, err := strconv.ParseInt(bs, 10, 64)
	if err != nil {
		err = lxterrs.Wrap(err, ErrStream(bs))
		return
	}
	*(*int32)(store.obj) = int32(num)
	return
}
func int32Get(store Store, in []byte) (out []byte) {
	pObj := store.obj
	num := *(*int64)(pObj)
	str := strconv.FormatInt(num, 10)
	out = append(in, str...)
	return
}
func int32Funcs(pidx *uintptr) (fSet setFunc, fGet getFunc) {
	if pidx == nil {
		return int32Set, int32Get
	}
	fSet = func(store PoolStore, bs string) (pBase unsafe.Pointer) {
		store.obj = store.Idx(*pidx)
		return int32Set(store, bs)
	}
	fGet = func(store Store, in []byte) (out []byte) {
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		return int32Get(store, in)
	}
	return
}

func uint16Set(store PoolStore, bs string) (pBase unsafe.Pointer) {
	pBase = store.obj
	num, err := strconv.ParseUint(bs, 10, 64)
	if err != nil {
		err = lxterrs.Wrap(err, ErrStream(bs))
		return
	}
	*(*uint16)(store.obj) = uint16(num)
	return
}
func uint16Get(store Store, in []byte) (out []byte) {
	pObj := store.obj
	num := *(*uint16)(pObj)
	str := strconv.FormatUint(uint64(num), 10)
	out = append(in, str...)
	return
}
func uint16Funcs(pidx *uintptr) (fSet setFunc, fGet getFunc) {
	if pidx == nil {
		return uint16Set, uint16Get
	}
	fSet = func(store PoolStore, bs string) (pBase unsafe.Pointer) {
		store.obj = store.Idx(*pidx)
		return uint16Set(store, bs)
	}
	fGet = func(store Store, in []byte) (out []byte) {
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		return uint16Get(store, in)
	}
	return
}

func int16Set(store PoolStore, bs string) (pBase unsafe.Pointer) {
	pBase = store.obj
	num, err := strconv.ParseInt(bs, 10, 64)
	if err != nil {
		err = lxterrs.Wrap(err, ErrStream(bs))
		return
	}
	*(*int16)(store.obj) = int16(num)
	return
}
func int16Get(store Store, in []byte) (out []byte) {
	pObj := store.obj
	num := *(*int64)(pObj)
	str := strconv.FormatInt(num, 10)
	out = append(in, str...)
	return
}
func int16Funcs(pidx *uintptr) (fSet setFunc, fGet getFunc) {
	if pidx == nil {
		return int16Set, int16Get
	}
	fSet = func(store PoolStore, bs string) (pBase unsafe.Pointer) {
		store.obj = store.Idx(*pidx)
		return int16Set(store, bs)
	}
	fGet = func(store Store, in []byte) (out []byte) {
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		return int16Get(store, in)
	}
	return
}

func uint8Set(store PoolStore, bs string) (pBase unsafe.Pointer) {
	pBase = store.obj
	num, err := strconv.ParseUint(bs, 10, 64)
	if err != nil {
		err = lxterrs.Wrap(err, ErrStream(bs))
		return
	}
	*(*uint8)(store.obj) = uint8(num)
	return
}
func uint8Get(store Store, in []byte) (out []byte) {
	pObj := store.obj
	num := *(*uint8)(pObj)
	str := strconv.FormatUint(uint64(num), 10)
	out = append(in, str...)
	return
}
func uint8Funcs(pidx *uintptr) (fSet setFunc, fGet getFunc) {
	if pidx == nil {
		return uint8Set, uint8Get
	}
	fSet = func(store PoolStore, bs string) (pBase unsafe.Pointer) {
		store.obj = store.Idx(*pidx)
		return uint8Set(store, bs)
	}
	fGet = func(store Store, in []byte) (out []byte) {
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		return uint8Get(store, in)
	}
	return
}

func int8Set(store PoolStore, bs string) (pBase unsafe.Pointer) {
	pBase = store.obj
	num, err := strconv.ParseInt(bs, 10, 64)
	if err != nil {
		err = lxterrs.Wrap(err, ErrStream(bs))
		return
	}
	*(*int8)(store.obj) = int8(num)
	return
}
func int8Get(store Store, in []byte) (out []byte) {
	pObj := store.obj
	num := *(*int64)(pObj)
	str := strconv.FormatInt(num, 10)
	out = append(in, str...)
	return
}
func int8Funcs(pidx *uintptr) (fSet setFunc, fGet getFunc) {
	if pidx == nil {
		return int8Set, int8Get
	}
	fSet = func(store PoolStore, bs string) (pBase unsafe.Pointer) {
		store.obj = store.Idx(*pidx)
		return int8Set(store, bs)
	}
	fGet = func(store Store, in []byte) (out []byte) {
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		return int8Get(store, in)
	}
	return
}

func float64Set(store PoolStore, bs string) (pBase unsafe.Pointer) {
	pBase = store.obj
	f, err := strconv.ParseFloat(bs, 64)
	if err != nil {
		err = lxterrs.Wrap(err, ErrStream(bs))
		return
	}
	*(*float64)(store.obj) = f
	return
}
func float64Get(store Store, in []byte) (out []byte) {
	pObj := store.obj
	num := *(*float64)(pObj)
	out = strconv.AppendFloat(in, float64(num), 'f', -1, 64)
	return
}
func float64Funcs(pidx *uintptr) (fSet setFunc, fGet getFunc) {
	if pidx == nil {
		return float64Set, float64Get
	}
	fSet = func(store PoolStore, bs string) (pBase unsafe.Pointer) {
		store.obj = store.Idx(*pidx)
		return float64Set(store, bs)
	}
	fGet = func(store Store, in []byte) (out []byte) {
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		return float64Get(store, in)
	}
	return
}

func float32Set(store PoolStore, bs string) (pBase unsafe.Pointer) {
	pBase = store.obj
	f, err := strconv.ParseFloat(bs, 64)
	if err != nil {
		err = lxterrs.Wrap(err, ErrStream(bs))
		return
	}
	*(*float32)(store.obj) = float32(f)
	return
}
func float32Get(store Store, in []byte) (out []byte) {
	pObj := store.obj
	num := *(*float64)(pObj)
	out = strconv.AppendFloat(in, float64(num), 'f', -1, 64)
	return
}
func float32Funcs(pidx *uintptr) (fSet setFunc, fGet getFunc) {
	if pidx == nil {
		return float32Set, float32Get
	}
	fSet = func(store PoolStore, bs string) (pBase unsafe.Pointer) {
		store.obj = store.Idx(*pidx)
		return float32Set(store, bs)
	}
	fGet = func(store Store, in []byte) (out []byte) {
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		return float32Get(store, in)
	}
	return
}

func stringSet(store PoolStore, bs string) (pBase unsafe.Pointer) {
	pBase = store.obj
	*(*string)(store.obj) = *(*string)(unsafe.Pointer(&bs))
	return
}
func stringGet(store Store, in []byte) (out []byte) {
	pObj := store.obj
	str := *(*string)(pObj)
	out = append(in, '"')
	out = append(out, str...)
	out = append(out, '"')
	return
}
func stringFuncs(pidx *uintptr) (fSet setFunc, fGet getFunc) {
	if pidx == nil {
		return stringSet, stringGet
	}
	fSet = func(store PoolStore, bs string) (pBase unsafe.Pointer) {
		store.obj = store.Idx(*pidx)
		return stringSet(store, bs)
	}
	fGet = func(store Store, in []byte) (out []byte) {
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		return stringGet(store, in)
	}
	return
}

func bytesSet(store PoolStore, bs string) (pBase unsafe.Pointer) {
	pBase = store.obj
	pbs := (*[]byte)(store.obj)
	// *pbs = make([]byte, len(bs)*2)
	// n, err := base64.StdEncoding.Decode(*pbs, stringBytes(bs))
	var err error
	*pbs, err = base64.StdEncoding.DecodeString(bs)
	if err != nil {
		err = lxterrs.Wrap(err, ErrStream(bs))
		return
	}
	// *pbs = (*pbs)[:n]
	return
}
func bytesGet(store Store, in []byte) (out []byte) {
	pObj := store.obj
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
func bytesFuncs(pidx *uintptr) (fSet setFunc, fGet getFunc) {
	if pidx == nil {
		return bytesSet, bytesGet
	}
	fSet = func(store PoolStore, bs string) (pBase unsafe.Pointer) {
		store.obj = store.Idx(*pidx)
		return bytesSet(store, bs)
	}
	fGet = func(store Store, in []byte) (out []byte) {
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		return bytesGet(store, in)
	}
	return
}

func sliceFuncs(pidx *uintptr) (fSet setFunc, fGet getFunc) {
	if pidx == nil {
		fSet = func(store PoolStore, bs string) (pBase unsafe.Pointer) {
			pBase = store.obj
			return
		}
		return
	}
	fSet = func(store PoolStore, bs string) (pBase unsafe.Pointer) {
		store.obj = store.Idx(*pidx)
		pBase = store.obj
		return
	}
	return
}

func structChildFuncs(pidx *uintptr) (fSet setFunc, fGet getFunc) {
	if pidx != nil {
		fSet = func(store PoolStore, bs string) (pBase unsafe.Pointer) {
			p := *(*unsafe.Pointer)(store.obj)
			if p == nil {
				store.obj = store.Idx(*pidx)
			}
			return p
		}
		fGet = func(store Store, in []byte) (out []byte) {
			return in
		}
	}
	return
}

// 匿名嵌入
func anonymousStructFuncs(pidx *uintptr, offset uintptr, fSet0 setFunc, fGet0 getFunc) (fSet setFunc, fGet getFunc) {
	if pidx == nil {
		fSet = func(store PoolStore, bs string) (pBase unsafe.Pointer) {
			pSon := pointerOffset(store.obj, offset)
			store.obj = pSon
			return fSet0(store, bs)
		}
		fGet = func(store Store, in []byte) (out []byte) {
			store.obj = pointerOffset(store.obj, offset)
			return fGet0(store, in)
		}
		return
	}
	fSet = func(store PoolStore, bs string) (pBase unsafe.Pointer) {
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			store.obj = store.Idx(*pidx)
		}
		return fSet0(store, bs)
	}
	fGet = func(store Store, in []byte) (out []byte) {
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj != nil {
			return fGet0(store, in)
		}
		return in
	}
	return
}

func iterfaceFuncs(pidx *uintptr) (fSet setFunc, fGet getFunc) {
	fSet = func(store PoolStore, bs string) (pBase unsafe.Pointer) {
		return store.obj
	}
	fGet = func(store Store, in []byte) (out []byte) {
		pObj := store.obj
		iface := *(*interface{})(pObj)
		out = marshalInterface(in, iface)
		return
	}
	if pidx != nil {
		fSet = func(store PoolStore, bs string) (pBase unsafe.Pointer) {
			store.obj = store.Idx(*pidx)
			return store.obj
		}
		fGet = func(store Store, in []byte) (out []byte) {
			store.obj = *(*unsafe.Pointer)(store.obj)
			if store.obj == nil {
				out = append(in, "null"...)
				return
			}
			pObj := store.obj
			iface := *(*interface{})(pObj)
			out = marshalInterface(in, iface)
			return
		}
	}
	return
}

func mapFuncs(pidx *uintptr) (fSet setFunc, fGet getFunc) {
	fSet = func(store PoolStore, bs string) (pBase unsafe.Pointer) {
		// p := (*map[string]interface{})(store.obj)
		// *p = make(map[string]interface{})
		return store.obj
	}

	if pidx != nil {
		fSet = func(store PoolStore, bs string) (pBase unsafe.Pointer) {
			store.obj = store.Idx(*pidx)
			// p := (*map[string]interface{})(store.obj)
			// *p = make(map[string]interface{})
			return store.obj
		}

		fGet = func(store Store, in []byte) (out []byte) {
			store.obj = *(*unsafe.Pointer)(store.obj)
			if store.obj == nil {
				out = append(in, "null"...)
				return
			}
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			return
		}
	}
	return
}
