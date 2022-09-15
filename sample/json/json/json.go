package json

import (
	"bytes"
	"errors"
	"math"
	"reflect"
	"strconv"
	"unicode/utf16"
	"unicode/utf8"
	"unsafe"

	lxterrs "github.com/lxt1045/errors"
)

func ErrStream(stream []byte) string {
	if len(stream[:]) > 128 {
		stream = stream[:128]
	}
	str := string(stream)
	return str
}

var spaceTable = [256]bool{
	'\t': true, '\n': true, '\v': true, '\f': true, '\r': true, ' ': true, 0x85: true, 0xA0: true,
}

func trimSpace(stream []byte) (i int) {
	for ; spaceTable[stream[i]]; i++ {
	}
	return
}

// 为了 inline 部门共用逻辑让调用者完成; 逻辑 解析：冒号 和 逗号 等单字符
func parseByte(stream []byte, b byte) (i, n int) {
	for ; ; i++ {
		if stream[i] == b {
			n++
			continue
		}
		if !spaceTable[stream[i]] {
			return
		}
	}
}

func parseObjToSlice(stream []byte, s []interface{}) (i int) {
	return 0
}

func bsToStr(bs []byte) string {
	return *(*string)(unsafe.Pointer(&bs))
}

// 解析 {}
// func parseObj(sts status, stream []byte, pObj unsafe.Pointer, tag *TagInfo) (i int) {
func parseObj(idxSlash int, stream []byte, pObj unsafe.Pointer, tag *TagInfo) (i, iSlash int) {
	iSlash = idxSlash
	n, nB := 0, 0
	key := []byte{}
	i += trimSpace(stream[i:])
	if stream[i] == '}' {
		i++
		return
	}
	for {
		// 解析 key: key不需要转义，因为在解析 struct 的时候会提供转义和不转义两种形式
		{
			// 手动内联
			start := i
			n = bytes.IndexByte(stream[i+1:], '"')
			if n >= 0 {
				i += n + 2
				key = stream[start:i]
			}
		}
		// 解析 冒号
		n, nB = parseByte(stream[i:], ':')
		i += n
		if nB != 1 {
			panic(lxterrs.New(ErrStream(stream[i:])))
		}
		// 解析 value
		son := tag.GetChild(key)

		if son != nil {
			n, iSlash = son.fUnm(iSlash-i, pObj, stream[i:], son)
			iSlash += i
		} else {
			var iface interface{}
			n, iSlash = parseInterface(iSlash-i, stream[i:], &iface)
			iSlash += i
		}
		i += n
		// 解析 逗号
		n, nB = parseByte(stream[i:], ',')
		i += n
		if nB != 1 {
			if nB == 0 && '}' == stream[i] {
				i++
				return
			}
			panic(lxterrs.New(ErrStream(stream[i:])))
		}
	}
}

func parseMapInterface(idxSlash int, stream []byte) (m map[string]interface{}, i, iSlash int) {
	iSlash = idxSlash
	n, nB := 0, 0
	key := []byte{}
	m = *mapCache.Get() //map 和 interface 一起获取，合并两次损耗为一次
	value := interfaceCache.Get()
	// m = make(map[string]interface{})
	for {
		i += trimSpace(stream[i:])
		{
			i++
			n = bytes.IndexByte(stream[i:], '"')
			if n >= 0 {
				n += i
				key = stream[i:n]
				i = n + 1
			}
		}
		n, nB = parseByte(stream[i:], ':')
		if nB != 1 {
			panic(lxterrs.New(ErrStream(stream[i:])))
		}
		i += n
		n, iSlash = parseInterface(iSlash-i, stream[i:], value)
		iSlash += i
		i += n
		m[string(key)] = *value
		n, nB = parseByte(stream[i:], ',')
		i += n
		if nB != 1 {
			if nB == 0 && '}' == stream[i] {
				i++
				return
			}
			panic(lxterrs.New(ErrStream(stream[i:])))
		}
	}
}

func parseSliceInterface(idxSlash int, stream []byte) (s []interface{}, i, iSlash int) {
	iSlash = idxSlash
	i = trimSpace(stream[i:])
	var value interface{}
	// pS := <-chSlice
	// s = *pS.(*[]interface{})
	for n, nB := 0, 0; ; {
		n, iSlash = parseInterface(iSlash-i, stream[i:], &value)
		iSlash += i
		i += n
		s = append(s, value)
		n, nB = parseByte(stream[i:], ',')
		i += n
		if nB != 1 {
			if nB == 0 && ']' == stream[i] {
				i++
				return
			}
			panic(lxterrs.New(ErrStream(stream[i:])))
		}
	}
}
func parseSlice(idxSlash int, stream []byte, pObj unsafe.Pointer, tag *TagInfo) (i, iSlash int) {
	iSlash = idxSlash
	i = trimSpace(stream[i:])
	if stream[i] == ']' {
		i++
		pBase := tag.fSet(pointerOffset(pObj, tag.Offset), stream[i:])
		// pSlice := (*[]uint8)(pBase)
		// *pSlice = make([]uint8, 0)
		pHeader := (*reflect.SliceHeader)(pBase)
		pHeader.Data = uintptr(pObj)
		return
	}
	pBase := tag.fSet(pointerOffset(pObj, tag.Offset), stream[i:])
	son := tag.ChildList[0]
	size := int(son.Type.Size())
	// size := (int(son.Type.Size()) + int(unsafe.Sizeof(&[]uint8{1}[0])) - 1) / int(unsafe.Sizeof((*uint8)(nil)))
	pSlice := (*[]uint8)(pBase) // 此处会被 GC 回收，因为申请的时候没有指示包含指针，所以不会被 GC 标志含有指针
	pHeader := (*reflect.SliceHeader)(pBase)
	for n, nB := 0, 0; ; {
		if pHeader.Len+size >= pHeader.Cap {
			// *pSlice = append(*pSlice, make([]uint8, size)...) // 此处会被 GC 回收，因为申请的时候没有指示包含指针，所以不会被 GC 标志含有指针
			// 只能用 reflect.MakeSlice
			l := pHeader.Len / size
			c := l * 2
			if l == 0 {
				c = 4
			}
			v := reflect.MakeSlice(tag.BaseType, l, c)
			p := reflectValueToPointer(&v)
			pS := (*[]uint8)(p)
			// pHeader = (*reflect.SliceHeader)(p)
			// pHeader.Len = (pHeader.Len + 1) * size
			// pHeader.Cap = pHeader.Cap * size
			// copy(*pSlice, *pS)

			pH := (*reflect.SliceHeader)(p)
			pH.Len = pH.Len * size
			pH.Cap = pH.Cap * size
			copy(*pS, *pSlice)

			pHeader.Len += size
			pHeader.Data = pH.Data
		} else {
			pHeader.Len += size
		}
		p := pointerOffset(unsafe.Pointer(pHeader.Data), uintptr(pHeader.Len-size))
		if son != nil && son.fUnm != nil {
			n, iSlash = son.fUnm(iSlash-i, p, stream[i:], son)
			iSlash += i
			if n == 0 {
				pHeader.Len -= size
			}
		} else {
			var iface interface{}
			n, iSlash = parseInterface(iSlash-i, stream[i:], &iface)
			iSlash += i
		}
		i += n
		n, nB = parseByte(stream[i:], ',')
		i += n
		if nB != 1 {
			if nB == 0 && ']' == stream[i] {
				i++
				pHeader.Len, pHeader.Cap = pHeader.Len/size, pHeader.Cap/size
				return
			}
			panic(lxterrs.New(ErrStream(stream[i:])))
		}
	}
}

// key 后面的单元: Num, str, bool, slice, obj, null
func parseInterface(idxSlash int, stream []byte, p *interface{}) (i, iSlash int) {
	iSlash = idxSlash
	// i = trimSpace(stream)
	switch stream[0] {
	default: // num
		var f float64
		f, i = float64UnmFuncs(stream)
		*p = f
		return
	case '{': // obj
		var m map[string]interface{}
		m, i, iSlash = parseMapInterface(iSlash-1, stream[1:])
		iSlash++
		i++
		*p = m
		return
	case '}':
		return
	case '[': // slice
		var s []interface{}
		s, i, iSlash = parseSliceInterface(iSlash-1, stream[1:])
		iSlash++
		i++
		*p = s
		return
	case ']':
		return
	case 'n':
		if stream[i+1] != 'u' || stream[i+2] != 'l' || stream[i+3] != 'l' {
			err := lxterrs.New("should be \"null\", not [%s]", ErrStream(stream))
			panic(err)
		}
		i = 4
		return
	case 't':
		if stream[i+1] != 'r' || stream[i+2] != 'u' || stream[i+3] != 'e' {
			err := lxterrs.New("should be \"true\", not [%s]", ErrStream(stream))
			panic(err)
		}
		i = 4
		*p = true
		return
	case 'f':
		if stream[i+1] != 'a' || stream[i+2] != 'l' || stream[i+3] != 's' || stream[i+4] != 'e' {
			err := lxterrs.New("should be \"false\", not [%s]", ErrStream(stream))
			panic(err)
		}
		i = 5
		*p = false
		return
	case '"':
		var raw []byte
		//
		raw, i, iSlash = parseStr(stream, iSlash)
		pStr := strCache.Get()
		*pStr = bytesString(raw)
		*p = pStr
		// *p = bytesString(raw)
		return
	}
	return
}

type status struct {
	idxUnescape int
}

//解析 obj: {}, 或 []
func parseRoot(stream []byte, pObj unsafe.Pointer, tag *TagInfo) (err error) {
	idxSlash := bytes.IndexByte(stream[1:], '\\')
	if idxSlash < 0 {
		idxSlash = math.MaxInt
	}
	// sts := status{}
	if stream[0] == '{' {
		parseObj(idxSlash, stream[1:], pObj, tag)
		return
	}
	if stream[0] == '[' {
		parseSlice(idxSlash, stream[1:], pObj, tag)
		return
	}
	return
}

//stream: "fgshw1321"...
func parseStr2(stream []byte) (raw []byte, i int) {
	lastIdx, i := 1, 1
	for {
		if i >= len(stream) || stream[i] == '"' || stream[i] == '\\' {
			if stream[i] == '\\' {
				word, wordSize := unescapeStr(stream[i:])
				if len(raw) <= 0 {
					raw = stream[1:i:i] //新建 []byte 避免修改员 stream
					// raw = stream[1:i] //新建 []byte 避免修改员 stream
				} else if lastIdx < i {
					raw = append(raw, stream[lastIdx:i]...)
				}
				raw = append(raw, word...)
				i += wordSize
				lastIdx = i
				continue
			}
			break
		}
		i++
	}
	if len(raw) == 0 {
		raw = stream[1:i]
	} else if lastIdx < i {
		raw = append(raw, stream[lastIdx:i]...)
	}
	return raw, i + 1
}

func parseStr(stream []byte, nextSlashIdx int) (raw []byte, i, nextSlashIdxOut int) {
	// TODO: 专业抄，把 '"' 提前准备好，少了个几步，也能快很多了
	i = bytes.IndexByte(stream[1:], '"')
	if i >= 0 && nextSlashIdx > i+1 {
		i++
		raw = stream[1:i]
		i++
		nextSlashIdxOut = nextSlashIdx
		return
	}
	i++
	return parseUnescapeStr(stream, i, nextSlashIdx)
}

// 可以可以在 struct 那边写两个 key，解析的时候就可以不用管了
// stream:`"<a href=\"//itunes.apple.com/us/app/twitter/id409789998?mt=12%5C%22\" rel=\"\\\"nofollow\\\"\">Twitter for Mac</a>"`
func parseUnescapeStr(stream []byte, nextQuotesIdx, nextSlashIdx int) (raw []byte, i, nextSlashIdxOut int) {
	if nextSlashIdx < 0 {
		nextSlashIdx = bytes.IndexByte(stream[1:], '\\')
		if nextSlashIdx < 0 {
			nextSlashIdx = math.MaxInt
		} else {
			nextSlashIdx++
		}

	}
	if nextQuotesIdx < 0 {
		panic(string(stream[i:]))
	}
	if nextSlashIdx > nextQuotesIdx {
		i = nextQuotesIdx + 1
		raw = stream[:i]
		return
	}
	lastIdx := 0
	for {
		i = nextSlashIdx
		word, wordSize := unescapeStr(stream[i:])
		if len(raw) == 0 {
			raw = stream[1:i:i] //新建 []byte 避免修改员 stream
		} else if lastIdx < i {
			raw = append(raw, stream[lastIdx:i]...)
		}
		raw = append(raw, word...)
		i += wordSize
		lastIdx = i
		if word[0] == '"' {
			nextQuotesIdx = bytes.IndexByte(stream[i:], '"')
			if nextQuotesIdx < 0 {
				panic(string(stream[i:]))
			}
			nextQuotesIdx += i
		}

		nextSlashIdx = bytes.IndexByte(stream[i:], '\\')
		if nextSlashIdx < 0 {
			nextSlashIdx = math.MaxInt
			break
		}
		nextSlashIdx += i
		if nextSlashIdx > nextQuotesIdx {
			break
		}
	}
	raw = append(raw, stream[lastIdx:nextQuotesIdx]...)
	return raw, nextQuotesIdx + 1, nextSlashIdx
}

// unescape unescapes a string
//“\\”、“\"”、“\/”、“\b”、“\f”、“\n”、“\r”、“\t”
// \u后面跟随4位16进制数字: "\uD83D\uDE02"
func unescapeStr(raw []byte) (word []byte, size int) {
	// i==0是 '\\', 所以从1开始
	switch raw[1] {
	case '\\':
		word, size = []byte{'\\'}, 2
	case '/':
		word, size = []byte{'/'}, 2
	case 'b':
		word, size = []byte{'\b'}, 2
	case 'f':
		word, size = []byte{'\f'}, 2
	case 'n':
		word, size = []byte{'\n'}, 2
	case 'r':
		word, size = []byte{'\r'}, 2
	case 't':
		word, size = []byte{'\t'}, 2
	case '"':
		word, size = []byte{'"'}, 2
	case 'u':
		//\uD83D
		if len(raw) < 6 {
			panic(errors.New("incorrect format: \\" + string(raw)))
		}
		last := raw[:6]
		r0 := unescapeToRune(last[2:])
		size, raw = 6, raw[6:]
		if utf16.IsSurrogate(r0) { // 如果utf-6还有后续(不完整)
			if len(raw) < 6 || raw[0] != '\\' || raw[1] != 'u' {
				l := 6
				if l > len(raw) {
					l = len(raw)
				}
				panic(errors.New("incorrect format: \\" + string(last) + string(raw[:l])))
			}
			r1 := unescapeToRune(raw[:6])
			// we expect it to be correct so just consume it
			r0 = utf16.DecodeRune(r0, r1)
			size += 6
		}
		// provide enough space to encode the largest utf8 possible
		word = make([]byte, 4)
		n := utf8.EncodeRune(word, r0)
		word = word[:n]
	default:
		panic(errors.New("incorrect format: " + ErrStream(raw)))
	}
	return
}

// runeit returns the rune from the the \uXXXX
func unescapeToRune(raw []byte) rune {
	n, err := strconv.ParseUint(string(raw), 16, 64)
	if err != nil {
		panic(errors.New("err:" + err.Error() + ",ncorrect format: " + string(raw)))
	}
	return rune(n)
}

//go:noescape
func IndexByte(bs []byte, c byte) int

//go:noescape
func IndexBytes(bs []byte, cs []byte) int

//go:noescape
func IndexBytes1(bs []byte, cs []byte) int

//go:noescape
func IndexBytes2(bs []byte, cs []byte) int

func Test1(x, y int) (a, b int)
func Test2(a int, xs []byte) (n int)

var SpaceBytes = [8][16]byte{
	fillBytes16('\t'),
	fillBytes16('\n'),
	fillBytes16('\v'),
	fillBytes16('\f'),
	fillBytes16('\r'),
	fillBytes16(' '),
	fillBytes16(0x85),
	fillBytes16(0xA0),
}

func fillBytes16(b byte) (bs [16]byte) {
	for i := 0; i < 16; i++ {
		bs[i] = b
	}
	return
}

// asm 中读入 X0 寄存器
var SpaceQ = [8]byte{0x85, 0xA0, '\t', '\n', '\v', '\f', '\r', ' '}

// 在 asm 中实现
func InSpaceQ(b byte) bool
