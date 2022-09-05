package json

import (
	"strconv"
	"unsafe"

	lxterrs "github.com/lxt1045/errors"
)

type unmFunc = func(pObj unsafe.Pointer, stream []byte, tag *TagInfo) (i int, err error)
type mFunc = func(pObj unsafe.Pointer, in []byte, tag *TagInfo) (out []byte)

func boolMFuncs() (fUnm unmFunc, fM mFunc) {
	fUnm = func(pObj unsafe.Pointer, stream []byte, tag *TagInfo) (i int, err error) {
		if stream[0] == 't' && stream[i+1] == 'r' && stream[i+2] == 'u' && stream[i+3] == 'e' {
			i = 4
		} else if stream[0] == 'f' && stream[1] == 'a' && stream[2] == 'l' && stream[3] == 's' && stream[4] == 'e' {
			i = 5
		} else if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
			i = 4
			return
		} else {
			err = lxterrs.New("should be \"false\" or , not [%s]", ErrStream(stream))
			return
		}
		_, err = tag.fSet(pointerOffset(pObj, tag.Offset), stream[0:i])
		if err != nil {
			err = lxterrs.New("error type:%s", ErrStream(stream[:]))
			return
		}
		return
	}
	fM = func(pObj unsafe.Pointer, in []byte, tag *TagInfo) (out []byte) {
		_, out = tag.fGet(pointerOffset(pObj, tag.Offset), in)
		return
	}
	return
}

func float64UnmFuncs(stream []byte) (f float64, i int, err error) {
	for ; i < len(stream); i++ {
		c := stream[i]
		if spaceTable[c] || c == ']' || c == '}' || c == ',' {
			break
		}
	}
	f, err = strconv.ParseFloat(bytesString(stream[:i]), 64)
	if err != nil {
		err = lxterrs.Wrap(err, ErrStream(stream[:i]))
		return
	}
	return
}

func numMFuncs() (fUnm unmFunc, fM mFunc) {
	fUnm = func(pObj unsafe.Pointer, stream []byte, tag *TagInfo) (i int, err error) {
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

func structMFuncs() (fUnm unmFunc, fM mFunc) {
	fUnm = func(pObj unsafe.Pointer, stream []byte, tag *TagInfo) (i int, err error) {
		i++
		pObj = pointerOffset(pObj, tag.Offset)
		if tag.fSet != nil {
			pObj, err = tag.fSet(pObj, stream[i:])
			if err != nil {
				err = lxterrs.Wrap(err, "error fSet, tag:%s", tag)
				return
			}
		}
		n, err := parseObj(stream[i:], pObj, tag)
		if err != nil {
			err = lxterrs.Wrap(err, "error type:%s", ErrStream(stream[i:]))
			return
		}
		i += n
		return
	}
	fM = func(pObj unsafe.Pointer, in []byte, tag *TagInfo) (out []byte) {
		_, out = tag.fGet(pointerOffset(pObj, tag.Offset), in)
		return
	}
	return
}

func sliceMFuncs() (fUnm unmFunc, fM mFunc) {
	fUnm = func(pObj unsafe.Pointer, stream []byte, tag *TagInfo) (i int, err error) {
		i++
		n, err := parseSlice(stream[i:], pObj, tag)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(stream[i:]))
			return
		}
		i += n
		return
	}
	fM = func(pObj unsafe.Pointer, in []byte, tag *TagInfo) (out []byte) {
		_, out = tag.fGet(pointerOffset(pObj, tag.Offset), in)
		return
	}
	return
}

func stringMFuncs() (fUnm unmFunc, fM mFunc) {
	fUnm = func(pObj unsafe.Pointer, stream []byte, tag *TagInfo) (i int, err error) {
		i++
		raw, n := parseStr(stream[i:])
		i += n
		_, err = tag.fSet(pointerOffset(pObj, tag.Offset), raw)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(stream[i:]))
			return
		}
		return
	}
	fM = func(pObj unsafe.Pointer, in []byte, tag *TagInfo) (out []byte) {
		_, out = tag.fGet(pointerOffset(pObj, tag.Offset), in)
		return
	}
	return
}

func interfaceMFuncs() (fUnm unmFunc, fM mFunc) {
	fUnm = func(pObj unsafe.Pointer, stream []byte, tag *TagInfo) (i int, err error) {
		n := trimSpace(stream[i:])
		i += n
		iv := (*interface{})(pointerOffset(pObj, tag.Offset))
		n, err = parseInterface(stream[i:], iv)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(stream[i:]))
			return
		}
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
