package json

import (
	"strconv"
	"strings"
	"sync"
	"unsafe"

	lxterrs "github.com/lxt1045/errors"
)

type unmFunc = func(idxSlash int, store PoolStore, stream string) (i, iSlash int)
type mFunc = func(store Store, in []byte) (out []byte)

func boolMFuncs() (fUnm unmFunc, fM mFunc) {
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		iSlash = idxSlash
		if stream[0] == 't' && stream[i+1] == 'r' && stream[i+2] == 'u' && stream[i+3] == 'e' {
			i = 4
		} else if stream[0] == 'f' && stream[1] == 'a' && stream[2] == 'l' && stream[3] == 's' && stream[4] == 'e' {
			i = 5
		} else if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
			i = 4
			return
		} else {
			err := lxterrs.New("should be \"false\" or , not [%s]", ErrStream(stream))
			panic(err)
		}

		store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.tag.fSet(store, stream[0:i])
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		out = store.tag.fGet(store, in)
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

func numMFuncs() (fUnm unmFunc, fM mFunc) {
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		iSlash = idxSlash
		for ; i < len(stream); i++ {
			c := stream[i]
			if spaceTable[c] || c == ']' || c == '}' || c == ',' {
				break
			}
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.tag.fSet(store, stream[:i])
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		out = store.tag.fGet(store, in)
		return
	}
	return
}

func structMFuncsStatus1(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
	i++
	store.obj = pointerOffset(store.obj, store.tag.Offset)
	if store.tag.fSet != nil {
		store.obj = store.tag.fSet(store, stream[i:])
	}
	n, iSlash := parseObj(idxSlash-i, stream[i:], store)
	i += n
	return
}

func structMFuncs() (fUnm unmFunc, fM mFunc) {
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
			i = 4
			iSlash = idxSlash
			return
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		if store.tag.fSet != nil {
			store.obj = store.tag.fSet(store, stream[1:])
		}
		n, iSlash := parseObj(idxSlash-1, stream[1:], store)
		iSlash++
		i += n + 1
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		out = store.tag.fGet(store, in)
		return
	}
	return
}

func sliceMFuncs() (fUnm unmFunc, fM mFunc) {
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
			i = 4
			iSlash = idxSlash
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			pBase := store.tag.fSet(store, stream[i:])
			pHeader := (*SliceHeader)(pBase)
			pHeader.Data = store.obj
			return
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset) //
		store.obj = store.tag.fSet(store, stream[i:])          // 会处理指针分配

		n, iSlash := parseSlice(idxSlash-1, stream[1:], store)
		iSlash++
		i += n + 1
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		out = store.tag.fGet(store, in)
		return
	}
	return
}
func slicePointerMFuncs() (fUnm unmFunc, fM mFunc) {
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
			i = 4
			iSlash = idxSlash
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			pBase := store.tag.fSet(store, stream[i:])
			pHeader := (*SliceHeader)(pBase)
			pHeader.Data = store.obj
			return
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset) //
		store.obj = store.tag.fSet(store, stream[i:])          // 会处理指针分配

		n, iSlash := parseSlice(idxSlash-1, stream[1:], store)
		iSlash++
		i += n + 1
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		out = store.tag.fGet(store, in)
		return
	}
	return
}
func sliceStringsMFuncs() (fUnm unmFunc, fM mFunc) {
	size := int(unsafe.Sizeof(""))
	SPoolN := (1 << 20) / size
	strsPool := sync.Pool{
		New: func() any {
			strs := make([]string, 0, SPoolN)
			return &strs
		},
	}

	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
			i = 4
			iSlash = idxSlash
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			pBase := store.tag.fSet(store, stream[i:])
			pHeader := (*SliceHeader)(pBase)
			pHeader.Data = store.obj
			return
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset) //
		store.obj = store.tag.fSet(store, stream[i:])          // 会处理指针分配

		n, iSlash := parseSliceString(idxSlash-1, stream[1:], store, SPoolN, &strsPool)
		iSlash++
		i += n + 1
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		out = store.tag.fGet(store, in)
		return
	}
	return
}

func stringMFuncs() (fUnm unmFunc, fM mFunc) {
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
			i = 4
			iSlash = idxSlash
			return
		}
		var raw string
		{
			i = strings.IndexByte(stream[1:], '"')
			if idxSlash > i+1 {
				i++
				raw = stream[1:i]
				i++
				iSlash = idxSlash
			} else {
				i++
				raw, i, iSlash = parseUnescapeStr(stream, i, idxSlash)
			}
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.tag.fSet(store, raw)
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		out = store.tag.fGet(store, in)
		return
	}
	return
}

func interfaceMFuncs() (fUnm unmFunc, fM mFunc) {
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
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		out = store.tag.fGet(store, in)
		return
	}
	return
}

func mapMFuncs() (fUnm unmFunc, fM mFunc) {
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
			i = 4
			iSlash = idxSlash
			return
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		if store.tag.fSet != nil {
			store.obj = store.tag.fSet(store, stream[1:])
		}
		m, i, iSlash := parseMapInterface(idxSlash-1, stream[1:])
		iSlash++
		i++
		p := (*map[string]interface{})(store.obj)
		*p = m
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		out = store.tag.fGet(store, in)
		return
	}
	return
}
