package json

import (
	"bytes"
	"reflect"
	"strconv"
	"unsafe"

	lxterrs "github.com/lxt1045/errors"
)

type unmFunc = func(idxSlash int, pObj unsafe.Pointer, stream []byte, tag *TagInfo) (i, iSlash int)
type mFunc = func(pObj unsafe.Pointer, in []byte, tag *TagInfo) (out []byte)

func boolMFuncs() (fUnm unmFunc, fM mFunc) {
	fUnm = func(idxSlash int, pObj unsafe.Pointer, stream []byte, tag *TagInfo) (i, iSlash int) {
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
		tag.fSet(pointerOffset(pObj, tag.Offset), stream[0:i])
		return
	}
	fM = func(pObj unsafe.Pointer, in []byte, tag *TagInfo) (out []byte) {
		_, out = tag.fGet(pointerOffset(pObj, tag.Offset), in)
		return
	}
	return
}

func float64UnmFuncs(stream []byte) (f float64, i int) {
	for ; i < len(stream); i++ {
		c := stream[i]
		if spaceTable[c] || c == ']' || c == '}' || c == ',' {
			break
		}
	}
	f, err := strconv.ParseFloat(bytesString(stream[:i]), 64)
	if err != nil {
		err = lxterrs.Wrap(err, ErrStream(stream[:i]))
		panic(err)
	}
	return
}

func numMFuncs() (fUnm unmFunc, fM mFunc) {
	fUnm = func(idxSlash int, pObj unsafe.Pointer, stream []byte, tag *TagInfo) (i, iSlash int) {
		iSlash = idxSlash
		for ; i < len(stream); i++ {
			c := stream[i]
			if spaceTable[c] || c == ']' || c == '}' || c == ',' {
				break
			}
		}
		tag.fSet(pointerOffset(pObj, tag.Offset), stream[:i])
		return
	}
	fM = func(pObj unsafe.Pointer, in []byte, tag *TagInfo) (out []byte) {
		_, out = tag.fGet(pointerOffset(pObj, tag.Offset), in)
		return
	}
	return
}

func structMFuncsStatus1(idxSlash int, pObj unsafe.Pointer, stream []byte, tag *TagInfo) (i, iSlash int) {
	i++
	pObj = pointerOffset(pObj, tag.Offset)
	if tag.fSet != nil {
		pObj = tag.fSet(pObj, stream[i:])
	}
	n, iSlash := parseObj(idxSlash-i, stream[i:], pObj, tag)
	i += n
	return
}

func structMFuncs() (fUnm unmFunc, fM mFunc) {
	fUnm = func(idxSlash int, pObj unsafe.Pointer, stream []byte, tag *TagInfo) (i, iSlash int) {
		if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
			i = 4
			iSlash = idxSlash
			return
		}
		pObj = pointerOffset(pObj, tag.Offset)
		if tag.fSet != nil {
			pObj = tag.fSet(pObj, stream[1:])
		}
		n, iSlash := parseObj(idxSlash-1, stream[1:], pObj, tag)
		iSlash++
		i += n + 1
		return
	}
	fM = func(pObj unsafe.Pointer, in []byte, tag *TagInfo) (out []byte) {
		_, out = tag.fGet(pointerOffset(pObj, tag.Offset), in)
		return
	}
	return
}

func sliceMFuncs() (fUnm unmFunc, fM mFunc) {
	fUnm = func(idxSlash int, pObj unsafe.Pointer, stream []byte, tag *TagInfo) (i, iSlash int) {
		if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
			i = 4
			iSlash = idxSlash
			pBase := tag.fSet(pointerOffset(pObj, tag.Offset), stream[i:])
			// pSlice := (*[]uint8)(pBase)
			// *pSlice = make([]uint8, 0)
			pHeader := (*reflect.SliceHeader)(pBase)
			pHeader.Data = uintptr(pObj)
			return
		}
		n, iSlash := parseSlice(idxSlash-1, stream[1:], pObj, tag)
		iSlash++
		i += n + 1
		return
	}
	fM = func(pObj unsafe.Pointer, in []byte, tag *TagInfo) (out []byte) {
		_, out = tag.fGet(pointerOffset(pObj, tag.Offset), in)
		return
	}
	return
}

func stringMFuncs() (fUnm unmFunc, fM mFunc) {
	fUnm = func(idxSlash int, pObj unsafe.Pointer, stream []byte, tag *TagInfo) (i, iSlash int) {
		if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
			i = 4
			iSlash = idxSlash
			return
		}
		var raw []byte
		{
			i = bytes.IndexByte(stream[1:], '"')
			if i >= 0 && idxSlash > i+1 {
				i++
				raw = stream[1:i]
				i++
				iSlash = idxSlash
			} else {
				i++
				raw, i, iSlash = parseUnescapeStr(stream, i, idxSlash)
			}
		}
		tag.fSet(pointerOffset(pObj, tag.Offset), raw)
		return
	}
	fM = func(pObj unsafe.Pointer, in []byte, tag *TagInfo) (out []byte) {
		_, out = tag.fGet(pointerOffset(pObj, tag.Offset), in)
		return
	}
	return
}

func interfaceMFuncs() (fUnm unmFunc, fM mFunc) {
	fUnm = func(idxSlash int, pObj unsafe.Pointer, stream []byte, tag *TagInfo) (i, iSlash int) {
		if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
			i = 4
			iSlash = idxSlash
			return
		}
		iSlash = idxSlash
		n := trimSpace(stream[i:])
		i += n
		iv := (*interface{})(pointerOffset(pObj, tag.Offset))
		n, iSlash = parseInterface(idxSlash-i, stream[i:], iv)
		idxSlash += i
		i += n
		// *iv = iface
		return
	}
	fM = func(pObj unsafe.Pointer, in []byte, tag *TagInfo) (out []byte) {
		_, out = tag.fGet(pointerOffset(pObj, tag.Offset), in)
		return
	}
	return
}
