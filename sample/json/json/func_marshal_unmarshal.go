package json

import (
	"encoding/base64"
	"strconv"
	"strings"
	"unsafe"

	lxterrs "github.com/lxt1045/errors"
)

type unmFunc = func(idxSlash int, store PoolStore, stream string) (i, iSlash int)
type mFunc = func(store Store, in []byte) (out []byte)

func pointerOffset(p unsafe.Pointer, offset uintptr) (pOut unsafe.Pointer) {
	return unsafe.Pointer(uintptr(p) + uintptr(offset))
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

func boolMFuncs2(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			iSlash = idxSlash
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			if stream[0] == 't' && stream[i+1] == 'r' && stream[i+2] == 'u' && stream[i+3] == 'e' {
				i = 4
				*(*bool)(store.obj) = true
			} else if stream[0] == 'f' && stream[1] == 'a' && stream[2] == 'l' && stream[3] == 's' && stream[4] == 'e' {
				i = 5
				*(*bool)(store.obj) = false
			} else {
				err := lxterrs.New("should be \"false\" or \"true\", not [%s]", ErrStream(stream))
				panic(err)
			}
			return
		}
		// fM = boolGet
		fM = func(store Store, in []byte) (out []byte) {
			// store.obj = pointerOffset(store.obj, store.tag.Offset)
			if *(*bool)(store.obj) {
				out = append(in, "true"...)
			} else {
				out = append(in, "false"...)
			}
			return
		}
		return
	}

	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		iSlash = idxSlash
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		if stream[0] == 't' && stream[i+1] == 'r' && stream[i+2] == 'u' && stream[i+3] == 'e' {
			i = 4
			store.obj = store.Idx(*pidx)
			*(*bool)(store.obj) = true
		} else if stream[0] == 'f' && stream[1] == 'a' && stream[2] == 'l' && stream[3] == 's' && stream[4] == 'e' {
			i = 5
			store.obj = store.Idx(*pidx)
			*(*bool)(store.obj) = false
		} else if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
			i = 4
			return
		} else {
			err := lxterrs.New("should be \"false\" or \"true\", not [%s]", ErrStream(stream))
			panic(err)
		}
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		pObj := *(*unsafe.Pointer)(store.obj)
		if pObj == nil {
			out = append(in, "null"...)
		} else if *(*bool)(pObj) {
			out = append(in, "true"...)
		} else {
			out = append(in, "false"...)
		}
		return
	}
	return
}

func float64UnmFuncs(stream string) (f float64, i int) {
	for ; i < len(stream); i++ {
		c := stream[i]
		if spaceTable[c] || c == ']' || c == '}' || c == ',' {
			break
		}
	}
	f, err := strconv.ParseFloat(stream[:i], 64)
	if err != nil {
		err = lxterrs.Wrap(err, ErrStream(stream[:i]))
		panic(err)
	}
	return
}

func uint64MFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			iSlash = idxSlash
			for ; i < len(stream); i++ {
				c := stream[i]
				if spaceTable[c] || c == ']' || c == '}' || c == ',' {
					break
				}
			}
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			num, err := strconv.ParseUint(stream[:i], 10, 64)
			if err != nil {
				err = lxterrs.Wrap(err, ErrStream(stream[:i]))
				panic(err)
			}
			*(*uint64)(store.obj) = num
			return
		}
		fM = func(store Store, in []byte) (out []byte) {
			// store.obj = pointerOffset(store.obj, store.tag.Offset)
			num := *(*uint64)(store.obj)
			out = strconv.AppendUint(in, num, 10)
			return
		}
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		iSlash = idxSlash
		for ; i < len(stream); i++ {
			c := stream[i]
			if spaceTable[c] || c == ']' || c == '}' || c == ',' {
				break
			}
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = store.Idx(*pidx)
		num, err := strconv.ParseUint(stream[:i], 10, 64)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(stream[:i]))
			panic(err)
		}
		*(*uint64)(store.obj) = num
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		num := *(*uint64)(store.obj)
		out = strconv.AppendUint(in, num, 10)
		return
	}
	return
}

func int64MFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			iSlash = idxSlash
			if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
				i = 4
				return
			}
			for ; i < len(stream); i++ {
				c := stream[i]
				if spaceTable[c] || c == ']' || c == '}' || c == ',' {
					break
				}
			}
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			num, err := strconv.ParseInt(stream[:i], 10, 64)
			if err != nil {
				err = lxterrs.Wrap(err, ErrStream(stream[:i]))
				panic(err)
			}
			*(*int64)(store.obj) = num
			return
		}
		fM = func(store Store, in []byte) (out []byte) {
			// store.obj = pointerOffset(store.obj, store.tag.Offset)
			num := *(*int64)(store.obj)
			out = strconv.AppendInt(in, num, 10)
			return
		}
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		iSlash = idxSlash
		for ; i < len(stream); i++ {
			c := stream[i]
			if spaceTable[c] || c == ']' || c == '}' || c == ',' {
				break
			}
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = store.Idx(*pidx)
		num, err := strconv.ParseInt(stream[:i], 10, 64)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(stream[:i]))
			panic(err)
		}
		*(*int64)(store.obj) = num
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		num := *(*int64)(store.obj)
		out = strconv.AppendInt(in, num, 10)
		return
	}
	return
}

func uint32MFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			iSlash = idxSlash
			for ; i < len(stream); i++ {
				c := stream[i]
				if spaceTable[c] || c == ']' || c == '}' || c == ',' {
					break
				}
			}
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			num, err := strconv.ParseUint(stream[:i], 10, 64)
			if err != nil {
				err = lxterrs.Wrap(err, ErrStream(stream[:i]))
				panic(err)
			}
			*(*uint32)(store.obj) = uint32(num)
			return
		}
		fM = func(store Store, in []byte) (out []byte) {
			// store.obj = pointerOffset(store.obj, store.tag.Offset)
			num := *(*uint32)(store.obj)
			out = strconv.AppendUint(in, uint64(num), 10)
			return
		}
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		iSlash = idxSlash
		for ; i < len(stream); i++ {
			c := stream[i]
			if spaceTable[c] || c == ']' || c == '}' || c == ',' {
				break
			}
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = store.Idx(*pidx)
		num, err := strconv.ParseUint(stream[:i], 10, 64)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(stream[:i]))
			panic(err)
		}
		*(*uint32)(store.obj) = uint32(num)
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		num := *(*uint32)(store.obj)
		out = strconv.AppendUint(in, uint64(num), 10)
		return
	}
	return
}

func int32MFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			iSlash = idxSlash
			for ; i < len(stream); i++ {
				c := stream[i]
				if spaceTable[c] || c == ']' || c == '}' || c == ',' {
					break
				}
			}
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			num, err := strconv.ParseInt(stream[:i], 10, 64)
			if err != nil {
				err = lxterrs.Wrap(err, ErrStream(stream[:i]))
				panic(err)
			}
			*(*int32)(store.obj) = int32(num)
			return
		}
		fM = func(store Store, in []byte) (out []byte) {
			// store.obj = pointerOffset(store.obj, store.tag.Offset)
			num := *(*int32)(store.obj)
			out = strconv.AppendInt(in, int64(num), 10)
			return
		}
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		iSlash = idxSlash
		for ; i < len(stream); i++ {
			c := stream[i]
			if spaceTable[c] || c == ']' || c == '}' || c == ',' {
				break
			}
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = store.Idx(*pidx)
		num, err := strconv.ParseInt(stream[:i], 10, 64)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(stream[:i]))
			panic(err)
		}
		*(*int32)(store.obj) = int32(num)
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		num := *(*int64)(store.obj)
		out = strconv.AppendInt(in, num, 10)
		return
	}
	return
}

func uint16MFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			iSlash = idxSlash
			for ; i < len(stream); i++ {
				c := stream[i]
				if spaceTable[c] || c == ']' || c == '}' || c == ',' {
					break
				}
			}
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			num, err := strconv.ParseUint(stream[:i], 10, 64)
			if err != nil {
				err = lxterrs.Wrap(err, ErrStream(stream[:i]))
				panic(err)
			}
			*(*uint16)(store.obj) = uint16(num)
			return
		}
		fM = func(store Store, in []byte) (out []byte) {
			// store.obj = pointerOffset(store.obj, store.tag.Offset)
			num := *(*uint16)(store.obj)
			out = strconv.AppendUint(in, uint64(num), 10)
			return
		}
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		iSlash = idxSlash
		for ; i < len(stream); i++ {
			c := stream[i]
			if spaceTable[c] || c == ']' || c == '}' || c == ',' {
				break
			}
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = store.Idx(*pidx)
		num, err := strconv.ParseUint(stream[:i], 10, 64)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(stream[:i]))
			panic(err)
		}
		*(*uint16)(store.obj) = uint16(num)
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		num := *(*uint16)(store.obj)
		out = strconv.AppendUint(in, uint64(num), 10)
		return
	}
	return
}

func int16MFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			iSlash = idxSlash
			for ; i < len(stream); i++ {
				c := stream[i]
				if spaceTable[c] || c == ']' || c == '}' || c == ',' {
					break
				}
			}
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			num, err := strconv.ParseInt(stream[:i], 10, 64)
			if err != nil {
				err = lxterrs.Wrap(err, ErrStream(stream[:i]))
				panic(err)
			}
			*(*int16)(store.obj) = int16(num)
			return
		}
		fM = func(store Store, in []byte) (out []byte) {
			// store.obj = pointerOffset(store.obj, store.tag.Offset)
			num := *(*int16)(store.obj)
			out = strconv.AppendInt(in, int64(num), 10)
			return
		}
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		iSlash = idxSlash
		for ; i < len(stream); i++ {
			c := stream[i]
			if spaceTable[c] || c == ']' || c == '}' || c == ',' {
				break
			}
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = store.Idx(*pidx)
		num, err := strconv.ParseInt(stream[:i], 10, 64)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(stream[:i]))
			panic(err)
		}
		*(*int16)(store.obj) = int16(num)
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		num := *(*int16)(store.obj)
		out = strconv.AppendInt(in, int64(num), 10)
		return
	}
	return
}

func uint8MFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			iSlash = idxSlash
			for ; i < len(stream); i++ {
				c := stream[i]
				if spaceTable[c] || c == ']' || c == '}' || c == ',' {
					break
				}
			}
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			num, err := strconv.ParseUint(stream[:i], 10, 64)
			if err != nil {
				err = lxterrs.Wrap(err, ErrStream(stream[:i]))
				panic(err)
			}
			*(*uint8)(store.obj) = uint8(num)
			return
		}
		fM = func(store Store, in []byte) (out []byte) {
			// store.obj = pointerOffset(store.obj, store.tag.Offset)
			num := *(*uint8)(store.obj)
			out = strconv.AppendUint(in, uint64(num), 10)
			return
		}
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		iSlash = idxSlash
		for ; i < len(stream); i++ {
			c := stream[i]
			if spaceTable[c] || c == ']' || c == '}' || c == ',' {
				break
			}
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = store.Idx(*pidx)
		num, err := strconv.ParseUint(stream[:i], 10, 64)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(stream[:i]))
			panic(err)
		}
		*(*uint8)(store.obj) = uint8(num)
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		num := *(*uint8)(store.obj)
		out = strconv.AppendUint(in, uint64(num), 10)
		return
	}
	return
}

func int8MFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			iSlash = idxSlash
			for ; i < len(stream); i++ {
				c := stream[i]
				if spaceTable[c] || c == ']' || c == '}' || c == ',' {
					break
				}
			}
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			num, err := strconv.ParseInt(stream[:i], 10, 64)
			if err != nil {
				err = lxterrs.Wrap(err, ErrStream(stream[:i]))
				panic(err)
			}
			*(*int8)(store.obj) = int8(num)
			return
		}
		fM = func(store Store, in []byte) (out []byte) {
			// store.obj = pointerOffset(store.obj, store.tag.Offset)
			num := *(*int64)(store.obj)
			out = strconv.AppendInt(in, num, 10)
			return
		}
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		iSlash = idxSlash
		for ; i < len(stream); i++ {
			c := stream[i]
			if spaceTable[c] || c == ']' || c == '}' || c == ',' {
				break
			}
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = store.Idx(*pidx)
		num, err := strconv.ParseInt(stream[:i], 10, 64)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(stream[:i]))
			panic(err)
		}
		*(*int8)(store.obj) = int8(num)
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		num := *(*int8)(store.obj)
		out = strconv.AppendInt(in, int64(num), 10)
		return
	}
	return
}

func float64MFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			iSlash = idxSlash
			for ; i < len(stream); i++ {
				c := stream[i]
				if spaceTable[c] || c == ']' || c == '}' || c == ',' {
					break
				}
			}
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			f, err := strconv.ParseFloat(stream[:i], 64)
			if err != nil {
				err = lxterrs.Wrap(err, ErrStream(stream[:i]))
				panic(err)
			}
			*(*float64)(store.obj) = f
			return
		}
		fM = func(store Store, in []byte) (out []byte) {
			// store.obj = pointerOffset(store.obj, store.tag.Offset)
			num := *(*float64)(store.obj)
			out = strconv.AppendFloat(in, float64(num), 'f', -1, 64)
			return
		}
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		iSlash = idxSlash
		for ; i < len(stream); i++ {
			c := stream[i]
			if spaceTable[c] || c == ']' || c == '}' || c == ',' {
				break
			}
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = store.Idx(*pidx)
		f, err := strconv.ParseFloat(stream[:i], 64)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(stream[:i]))
			panic(err)
		}
		*(*float64)(store.obj) = f
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		num := *(*float64)(store.obj)
		out = strconv.AppendFloat(in, float64(num), 'f', -1, 64)
		return
	}
	return
}
func float32MFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			iSlash = idxSlash
			for ; i < len(stream); i++ {
				c := stream[i]
				if spaceTable[c] || c == ']' || c == '}' || c == ',' {
					break
				}
			}
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			f, err := strconv.ParseFloat(stream[:i], 64)
			if err != nil {
				err = lxterrs.Wrap(err, ErrStream(stream[:i]))
				panic(err)
			}
			*(*float32)(store.obj) = float32(f)
			return
		}
		fM = func(store Store, in []byte) (out []byte) {
			// store.obj = pointerOffset(store.obj, store.tag.Offset)
			num := *(*float32)(store.obj)
			out = strconv.AppendFloat(in, float64(num), 'f', -1, 64)
			return
		}
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		iSlash = idxSlash
		for ; i < len(stream); i++ {
			c := stream[i]
			if spaceTable[c] || c == ']' || c == '}' || c == ',' {
				break
			}
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = store.Idx(*pidx)
		f, err := strconv.ParseFloat(stream[:i], 64)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(stream[:i]))
			panic(err)
		}
		*(*float32)(store.obj) = float32(f)
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		num := *(*float32)(store.obj)
		out = strconv.AppendFloat(in, float64(num), 'f', -1, 64)
		return
	}
	return
}

func structMFuncs(pidx, sonPidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
				i = 4
				iSlash = idxSlash
				return
			}
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			store.pointerPool = pointerOffset(store.pointerPool, *sonPidx) //这里有问题，这个 pool 导致 slicePool 的偏移
			store.slicePool = pointerOffset(store.slicePool, store.tag.idxSliceObjPool)
			n, iSlash := parseObj(idxSlash-1, stream[1:], store)
			iSlash++
			i += n + 1
			return
		}
		fM = func(store Store, in []byte) (out []byte) {
			// store.obj = pointerOffset(store.obj, store.tag.Offset)
			out = marshalStruct(store, in)
			return
		}
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
			i = 4
			iSlash = idxSlash
			return
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.pointerPool = pointerOffset(store.pointerPool, *sonPidx) //这里有问题，这个 pool 导致 slicePool 的偏移
		store.slicePool = pointerOffset(store.slicePool, store.tag.idxSliceObjPool)
		p := *(*unsafe.Pointer)(store.obj)
		if p == nil {
			store.obj = store.Idx(*pidx)
		}
		n, iSlash := parseObj(idxSlash-1, stream[1:], store)
		iSlash++
		i += n + 1
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = *(*unsafe.Pointer)(store.obj)
		out = marshalStruct(store, in)
		return
	}
	return
}

func sliceIntsMFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
				i = 4
				iSlash = idxSlash
				store.obj = pointerOffset(store.obj, store.tag.Offset)
				pHeader := (*SliceHeader)(store.obj)
				pHeader.Data = store.obj
				return
			}
			store.obj = pointerOffset(store.obj, store.tag.Offset) //
			n, iSlash := parseIntSlice(idxSlash-1, stream[1:], store)
			iSlash++
			i += n + 1
			return
		}
		fM = func(store Store, in []byte) (out []byte) {
			// store.obj = pointerOffset(store.obj, store.tag.Offset)
			pHeader := (*SliceHeader)(store.obj)
			son := store.tag.ChildList[0]
			out = marshalSlice(in, Store{obj: pHeader.Data, tag: son}, pHeader.Len)
			return
		}
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
			i = 4
			iSlash = idxSlash
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			store.obj = store.Idx(*pidx)
			pHeader := (*SliceHeader)(store.obj)
			pHeader.Data = store.obj
			return
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset) //
		p := *(*unsafe.Pointer)(store.obj)
		if p == nil {
			store.obj = store.Idx(*pidx)
		}
		n, iSlash := parseIntSlice(idxSlash-1, stream[1:], store)
		iSlash++
		i += n + 1
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		pHeader := (*SliceHeader)(store.obj)
		son := store.tag.ChildList[0]
		out = marshalSlice(in, Store{obj: pHeader.Data, tag: son}, pHeader.Len)
		return
	}
	return
}
func sliceNoscanMFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
				i = 4
				iSlash = idxSlash
				store.obj = pointerOffset(store.obj, store.tag.Offset)
				pHeader := (*SliceHeader)(store.obj)
				pHeader.Data = store.obj
				return
			}
			store.obj = pointerOffset(store.obj, store.tag.Offset) //
			n, iSlash := parseNoscanSlice(idxSlash-1, stream[1:], store)
			iSlash++
			i += n + 1
			return
		}
		fM = func(store Store, in []byte) (out []byte) {
			// store.obj = pointerOffset(store.obj, store.tag.Offset)
			pHeader := (*SliceHeader)(store.obj)
			son := store.tag.ChildList[0]
			out = marshalSlice(in, Store{obj: pHeader.Data, tag: son}, pHeader.Len)
			return
		}
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
			i = 4
			iSlash = idxSlash
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			store.obj = store.Idx(*pidx)
			pHeader := (*SliceHeader)(store.obj)
			pHeader.Data = store.obj
			return
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset) //
		p := *(*unsafe.Pointer)(store.obj)
		if p == nil {
			store.obj = store.Idx(*pidx)
		}
		n, iSlash := parseNoscanSlice(idxSlash-1, stream[1:], store)
		iSlash++
		i += n + 1
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		pHeader := (*SliceHeader)(store.obj)
		son := store.tag.ChildList[0]
		out = marshalSlice(in, Store{obj: pHeader.Data, tag: son}, pHeader.Len)
		return
	}
	return
}

func sliceNoscanMFuncs2(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
				i = 4
				iSlash = idxSlash
				store.obj = pointerOffset(store.obj, store.tag.Offset)
				pHeader := (*SliceHeader)(store.obj)
				pHeader.Data = store.obj
				return
			}
			store.obj = pointerOffset(store.obj, store.tag.Offset) //
			n, iSlash := parseNoscanSlice(idxSlash-1, stream[1:], store)
			iSlash++
			i += n + 1
			return
		}
		fM = func(store Store, in []byte) (out []byte) {
			// store.obj = pointerOffset(store.obj, store.tag.Offset)
			pHeader := (*SliceHeader)(store.obj)
			son := store.tag.ChildList[0]
			out = marshalSlice(in, Store{obj: pHeader.Data, tag: son}, pHeader.Len)
			return
		}
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
			i = 4
			iSlash = idxSlash
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			store.obj = store.Idx(*pidx)
			pHeader := (*SliceHeader)(store.obj)
			pHeader.Data = store.obj
			return
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset) //
		p := *(*unsafe.Pointer)(store.obj)
		if p == nil {
			store.obj = store.Idx(*pidx)
		}
		n, iSlash := parseNoscanSlice(idxSlash-1, stream[1:], store)
		iSlash++
		i += n + 1
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		pHeader := (*SliceHeader)(store.obj)
		son := store.tag.ChildList[0]
		out = marshalSlice(in, Store{obj: pHeader.Data, tag: son}, pHeader.Len)
		return
	}
	return
}

func sliceMFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
				i = 4
				iSlash = idxSlash
				store.obj = pointerOffset(store.obj, store.tag.Offset)
				pHeader := (*SliceHeader)(store.obj)
				pHeader.Data = store.obj
				return
			}
			store.obj = pointerOffset(store.obj, store.tag.Offset) //
			store.slicePool = pointerOffset(store.slicePool, store.tag.idxSliceObjPool)
			n, iSlash := parseSlice(idxSlash-1, stream[1:], store)
			iSlash++
			i += n + 1
			return
		}
		fM = func(store Store, in []byte) (out []byte) {
			// store.obj = pointerOffset(store.obj, store.tag.Offset)
			pHeader := (*SliceHeader)(store.obj)
			son := store.tag.ChildList[0]
			out = marshalSlice(in, Store{obj: pHeader.Data, tag: son}, pHeader.Len)
			return
		}
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
			i = 4
			iSlash = idxSlash
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			store.obj = store.Idx(*pidx)
			pHeader := (*SliceHeader)(store.obj)
			pHeader.Data = store.obj
			return
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset) //
		store.slicePool = pointerOffset(store.slicePool, store.tag.idxSliceObjPool)
		p := *(*unsafe.Pointer)(store.obj)
		if p == nil {
			store.obj = store.Idx(*pidx) // TODO 这个可以 pidx==nil 合并? 这时 *pidx==0？
		}
		n, iSlash := parseSlice(idxSlash-1, stream[1:], store)
		iSlash++
		i += n + 1
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		pHeader := (*SliceHeader)(store.obj)
		son := store.tag.ChildList[0]
		out = marshalSlice(in, Store{obj: pHeader.Data, tag: son}, pHeader.Len)
		return
	}
	return
}

func bytesMFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
				i = 4
				iSlash = idxSlash
				store.obj = pointerOffset(store.obj, store.tag.Offset)
				pHeader := (*SliceHeader)(store.obj)
				pHeader.Data = store.obj
				return
			}
			store.obj = pointerOffset(store.obj, store.tag.Offset) //

			//TODO : 解析长度（,]}），hex.Decode
			bytesSet(store, stream[i:])
			return
		}
		fM = bytesGet
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
			i = 4
			iSlash = idxSlash
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			pHeader := (*SliceHeader)(store.obj)
			pHeader.Data = store.obj
			return
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset) //
		p := *(*unsafe.Pointer)(store.obj)
		if p == nil {
			store.obj = store.Idx(*pidx)
		}
		n, iSlash := parseSlice(idxSlash-1, stream[1:], store)
		iSlash++
		i += n + 1
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = *(*unsafe.Pointer)(store.obj)
		bytesGet(store, in)
		return
	}
	return
}

func slicePointerMFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
			i = 4
			iSlash = idxSlash
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			store.obj = *(*unsafe.Pointer)(store.obj)
			if store.obj == nil {
				store.obj = store.Idx(*pidx)
			}
			pHeader := (*SliceHeader)(store.obj)
			pHeader.Data = store.obj
			return
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset) //
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			store.obj = store.Idx(*pidx)
		}

		n, iSlash := parseSlice(idxSlash-1, stream[1:], store)
		iSlash++
		i += n + 1
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		// pHeader := (*SliceHeader)(store.obj)
		// son := store.tag.ChildList[0]
		// out = marshalSlice(in, Store{obj: pHeader.Data, tag: son}, pHeader.Len)
		return
	}
	return
}
func sliceStringsMFuncs() (fUnm unmFunc, fM mFunc) {
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
			i = 4
			iSlash = idxSlash
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			pHeader := (*SliceHeader)(store.obj)
			pHeader.Data = store.obj
			return
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset) //
		n, iSlash := parseSliceString(idxSlash-1, stream[1:], store)
		iSlash++
		i += n + 1
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		pHeader := (*SliceHeader)(store.obj)
		son := store.tag.ChildList[0]
		out = marshalSlice(in, Store{obj: pHeader.Data, tag: son}, pHeader.Len)
		return
	}
	return
}

func stringUnm(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
	if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
		i = 4
		iSlash = idxSlash
		return
	}
	store.obj = pointerOffset(store.obj, store.tag.Offset)
	pstr := (*string)(store.obj)
	{
		i = strings.IndexByte(stream[1:], '"')
		if idxSlash > i+1 {
			i++
			*pstr = stream[1:i]
			i++
			iSlash = idxSlash
		} else {
			i++
			*pstr, i, iSlash = parseUnescapeStr(stream, i, idxSlash)
		}
	}
	return
}
func stringM(store Store, in []byte) (out []byte) {
	str := *(*string)(store.obj)
	out = append(in, '"')
	out = append(out, str...)
	out = append(out, '"')
	return
}
func stringMFuncs2(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
				i = 4
				iSlash = idxSlash
				return
			}
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			pstr := (*string)(store.obj)
			{
				i = strings.IndexByte(stream[1:], '"')
				if idxSlash > i+1 {
					i++
					*pstr = stream[1:i]
					i++
					iSlash = idxSlash
				} else {
					i++
					*pstr, i, iSlash = parseUnescapeStr(stream, i, idxSlash)
				}
			}
			return
		}
		fM = stringM
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = store.Idx(*pidx)
		if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
			i = 4
			iSlash = idxSlash
			return
		}
		pstr := (*string)(store.obj)
		{
			i = strings.IndexByte(stream[1:], '"')
			if idxSlash > i+1 {
				i++
				*pstr = stream[1:i]
				i++
				iSlash = idxSlash
			} else {
				i++
				*pstr, i, iSlash = parseUnescapeStr(stream, i, idxSlash)
			}
		}
		return
	}

	fM = func(store Store, in []byte) (out []byte) {
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		return stringM(store, in)
	}
	return
}

func interfaceMFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
				i = 4
				iSlash = idxSlash
				return
			}
			iSlash = idxSlash
			n := trimSpace(stream[i:])
			i += n
			iv := (*interface{})(pointerOffset(store.obj, store.tag.Offset))
			n, iSlash = parseInterface(idxSlash-i, stream[i:], iv)
			idxSlash += i
			i += n
			// *iv = iface
			return
		}
		fM = func(store Store, in []byte) (out []byte) {
			iface := *(*interface{})(store.obj)
			out = marshalInterface(in, iface)
			return
		}
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
			i = 4
			iSlash = idxSlash
			return
		}
		iSlash = idxSlash
		n := trimSpace(stream[i:])
		i += n
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = store.Idx(*pidx)
		iv := (*interface{})(store.obj)
		n, iSlash = parseInterface(idxSlash-i, stream[i:], iv)
		idxSlash += i
		i += n
		// *iv = iface
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		iface := *(*interface{})(store.obj)
		out = marshalInterface(in, iface)
		return
	}
	return
}

func mapMFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
				i = 4
				iSlash = idxSlash
				return
			}
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			m, i, iSlash := parseMapInterface(idxSlash-1, stream[1:])
			iSlash++
			i++
			p := (*map[string]interface{})(store.obj)
			*p = m
			return
		}
		fM = func(store Store, in []byte) (out []byte) {
			// store.obj = pointerOffset(store.obj, store.tag.Offset)
			m := *(*map[string]interface{})(store.obj)
			out = marshalMapInterface(in, m)
			return
		}
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
			i = 4
			iSlash = idxSlash
			return
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = store.Idx(*pidx)
		m, i, iSlash := parseMapInterface(idxSlash-1, stream[1:])
		iSlash++
		i++
		p := (*map[string]interface{})(store.obj)
		*p = m
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		m := *(*map[string]interface{})(store.obj)
		out = marshalMapInterface(in, m)
		return
	}
	return
}
