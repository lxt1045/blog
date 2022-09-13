package json

import (
	"bytes"
	"errors"
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
func parseObj(stream []byte, pObj unsafe.Pointer, tag *TagInfo) (i int) {
	n, nB := 0, 0
	key := []byte{}
	i += trimSpace(stream[i:])
	for {
		// 解析 key
		// key, n = parseStr(stream[i:])
		{
			//TODO: nextSlashIdx = bytes.IndexByte(stream[1:], '\\')
			nextSlashIdx := -1

			// 手动内联
			n = bytes.IndexByte(stream[i+1:], '"')
			if n >= 0 && nextSlashIdx <= 0 {
				n += 2
				key = stream[i : i+n]
			} else {
				key, n = parseUnescapeStr(stream[i:], nextSlashIdx, n)
			}
		}
		i += n
		// 解析 冒号
		n, nB = parseByte(stream[i:], ':')
		i += n
		if nB != 1 {
			panic(lxterrs.New(ErrStream(stream[i:])))
		}
		// 解析 value
		var son *TagInfo
		if tag != nil {
			// son = tag.Children[string(key)]
			son = tag.GetChild(key)
		}
		if son != nil {
			n = son.fUnm(pObj, stream[i:], son)
		} else {
			var iface interface{}
			n = parseInterface(stream[i:], &iface)
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

func parseMapInterface(stream []byte) (m map[string]interface{}, i int) {
	n, nB := 0, 0
	key := []byte{}
	m = *mapCache.Get() //map 和 interface 一起获取，合并两次损耗为一次
	value := interfaceCache.Get()
	// m = make(map[string]interface{})
	for {
		i += trimSpace(stream[i:])
		key, n = parseStr(stream[i:])
		key = key[1 : len(key)-1]
		i += n
		n, nB = parseByte(stream[i:], ':')
		if nB != 1 {
			panic(lxterrs.New(ErrStream(stream[i:])))
		}
		i += n
		n = parseInterface(stream[i:], value)
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

func parseSliceInterface(stream []byte) (s []interface{}, i int) {
	i = trimSpace(stream[i:])
	var value interface{}
	// pS := <-chSlice
	// s = *pS.(*[]interface{})
	for n, nB := 0, 0; ; {
		n = parseInterface(stream[i:], &value)
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
func parseSlice(stream []byte, pObj unsafe.Pointer, tag *TagInfo) (i int) {
	i = trimSpace(stream[i:])
	pBase := tag.fSet(pointerOffset(pObj, tag.Offset), stream[i:])
	son := tag.ChildList[0]
	size := int(son.Type.Size())
	pSlice := (*[]uint8)(pBase)
	pHeader := (*reflect.SliceHeader)(pBase)
	for n, nB := 0, 0; ; {
		if pHeader.Len+size >= pHeader.Cap {
			*pSlice = append(*pSlice, make([]uint8, size)...)
		} else {
			pHeader.Len += size
		}
		p := pointerOffset(unsafe.Pointer(pHeader.Data), uintptr(pHeader.Len-size))
		if son != nil && son.fUnm != nil {
			n = son.fUnm(p, stream[i:], son)
		} else {
			var iface interface{}
			n = parseInterface(stream[i:], &iface)
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
func parseInterface(stream []byte, p *interface{}) (i int) {
	// i = trimSpace(stream)
	switch stream[0] {
	default: // num
		var f float64
		f, i = float64UnmFuncs(stream)
		*p = f
		return
	case '{': // obj
		var m map[string]interface{}
		m, i = parseMapInterface(stream[1:])
		i++
		*p = m
		return
	case '}':
		return
	case '[': // slice
		var s []interface{}
		s, i = parseSliceInterface(stream[1:])
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
		raw, i = parseStr(stream)
		raw = raw[1 : len(raw)-1]
		pStr := strCache.Get()
		*pStr = bytesString(raw)
		*p = pStr
		// *p = bytesString(raw)
		return
	}
	return
}

//解析 obj: {}, 或 []
func parseRoot(stream []byte, pObj unsafe.Pointer, tag *TagInfo) (err error) {
	i := 1
	if stream[0] == '{' {
		parseObj(stream[i:], pObj, tag)
		return
	}
	if stream[0] == '[' {
		i++
		n := parseSlice(stream[i:], pObj, tag)
		i += n
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

func parseStr(stream []byte) (raw []byte, i int) {
	nextSlashIdx := -1 // nextSlashIdx = bytes.IndexByte(stream[1:], '\\')

	// TODO: 专业抄，把 '"' 提前准备好，少了个几步，也能快很多了
	i = bytes.IndexByte(stream[1:], '"')
	if i >= 0 && nextSlashIdx <= 0 {
		i += 2
		raw = stream[:i]
		return
	}

	return parseUnescapeStr(stream, nextSlashIdx, i)
}
func parseUnescapeStr(stream []byte, nextSlashIdx, nextQuotesIdx int) (raw []byte, i int) {
	if nextQuotesIdx < 0 {
		panic(string(stream[i:]))
	}
	if nextSlashIdx > nextQuotesIdx {
		i = nextQuotesIdx + 2
		raw = stream[:i]
		return
	}
	lastIdx := 1
	i = 1
	for {
		word, wordSize := unescapeStr(stream[i:])
		if len(raw) <= 0 {
			raw = stream[i : nextSlashIdx+i : nextSlashIdx+i] //新建 []byte 避免修改员 stream
			// raw = stream[1:i] //新建 []byte 避免修改员 stream
		} else if lastIdx < i {
			raw = append(raw, stream[lastIdx:nextSlashIdx+i]...)
		}
		raw = append(raw, word...)
		i += wordSize
		lastIdx = i

		nextSlashIdx = bytes.IndexByte(stream[i:], '\\')
		if nextSlashIdx < 0 || nextSlashIdx+i > nextQuotesIdx+1 {
			break
		}
	}
	raw = append(raw, stream[lastIdx:1+nextQuotesIdx]...)
	return raw, nextQuotesIdx + 1 + 1
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
		l := 4
		if l > len(raw) {
			l = len(raw)
		}
		panic(errors.New("incorrect format: " + string(raw[:l])))
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
