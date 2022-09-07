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

func panicIncorrectFormat(stream []byte) {
	if len(stream[:]) > 128 {
		stream = stream[:128]
	}
	panic(errors.New("incorrect format: " + string(stream)))
}
func ErrStream(stream []byte) string {
	if len(stream[:]) > 128 {
		stream = stream[:128]
	}
	str := string(stream)
	return str
}
func panicIncorrectType(typ Type, tag *TagInfo) {
	t := strconv.Itoa(int(typ))
	panic(errors.New("incorrect type: " + t + " to " + tag.BaseKind.String()))
}

var spaceTable = [256]bool{'\t': true, '\n': true, '\v': true, '\f': true, '\r': true, ' ': true, 0x85: true, 0xA0: true}

//inline
func trimSpace(stream []byte) (i int) {
	// for i = 0; i < len(stream) && spaceTable[stream[i]]; i++ {}
	for ; spaceTable[stream[i]]; i++ {
	}
	// for i = 0; InSpaceQ(stream[i]); i++ { }
	return
}

// 解析：冒号 和 逗号 等单字符
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

func parseObjToMap(stream []byte, m map[string]interface{}) (i int) {
	return 0
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
			nextSlashIdx := -1 // nextSlashIdx = bytes.IndexByte(stream[1:], '\\')

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
		if son != nil && son.fUnm != nil {
			n = son.fUnm(pObj, stream[i:], son)
		} else {
			n = parseValue(stream[i:], pObj, son)
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
	m = *mapCache.Get()
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

// 解析 {}
func parseMap(stream []byte, pObj unsafe.Pointer, tag *TagInfo) (i int) {
	m := *(*map[string]interface{})(pObj)
	for i < len(stream) && stream[i] != '}' {
		i += trimSpace(stream[i:])
		key, n := parseStr(stream[i:])
		i += n
		key = key[1 : len(key)-1]
		n, nB := parseByte(stream[i:], ':')
		if nB != 1 {
			panic(lxterrs.New(ErrStream(stream[i:])))
		}
		i += n
		var value interface{}
		pValue := unsafe.Pointer(&value)
		son := tag.ChildList[0]
		n = parseValue(stream[i:], pValue, son)
		i += n
		m[string(key)] = value
		n, nB = parseByte(stream[i:], ',')
		if nB != 1 {
			panic(lxterrs.New(ErrStream(stream[i:])))
		}
		i += n
	}
	if stream[i] == '}' {
		i++
		return
	}
	err := lxterrs.New("errors json string:%s", ErrStream(stream))
	panic(err)
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
			n = parseValue(stream[i:], p, son)
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
	n := 0
	switch stream[0] {
	default: // num
		var f float64
		f, n = float64UnmFuncs(stream[i:])
		i = n
		*p = f
		return
	case '{': // obj
		i++
		var m map[string]interface{}
		m, n = parseMapInterface(stream[i:])
		i += n
		*p = m
		return
	case '[': // slice
		// TODO
		i++
		var s []interface{}
		s, n = parseSliceInterface(stream[i:])
		i += n
		*p = s
		return
	case 'n':
		if stream[i+1] != 'u' || stream[i+2] != 'l' || stream[i+3] != 'l' {
			err := lxterrs.New("should be \"null\", not [%s]", ErrStream(stream))
			panic(err)
		}
		i += 4
		return
	case 't':
		if stream[i+1] != 'r' || stream[i+2] != 'u' || stream[i+3] != 'e' {
			err := lxterrs.New("should be \"true\", not [%s]", ErrStream(stream))
			panic(err)
		}
		i += 4
		*p = true
		return
	case 'f':
		if stream[i+1] != 'a' || stream[i+2] != 'l' || stream[i+3] != 's' || stream[i+4] != 'e' {
			err := lxterrs.New("should be \"false\", not [%s]", ErrStream(stream))
			panic(err)
		}
		i += 5
		*p = false
		return
	case '"':
		raw, n := parseStr(stream[i:])
		raw = raw[1 : len(raw)-1]
		i += n
		pStr := strCache.Get()
		*pStr = bytesString(raw)
		*p = pStr
		// *p = bytesString(raw)
		return
	}
	return
}

// key 后面的单元: Num, str, bool, slice, obj, null
func parseInterface1(stream []byte) (iface interface{}, i int) {
	// i = trimSpace(stream)
	n := 0
	switch stream[0] {
	default: // num
		f, n := float64UnmFuncs(stream[i:])
		i = n
		iface = f
		return
	case '{': // obj
		i++
		var m map[string]interface{}
		m, n = parseMapInterface(stream[i:])
		i += n
		iface = m
		return
	case '[': // slice
		// TODO
		i++
		var s []interface{}
		s, n = parseSliceInterface(stream[i:])
		i += n
		iface = s
		return
	case 'n':
		if stream[i+1] != 'u' || stream[i+2] != 'l' || stream[i+3] != 'l' {
			err := lxterrs.New("should be \"null\", not [%s]", ErrStream(stream))
			panic(err)
		}
		i += 4
		return
	case 't':
		if stream[i+1] != 'r' || stream[i+2] != 'u' || stream[i+3] != 'e' {
			err := lxterrs.New("should be \"true\", not [%s]", ErrStream(stream))
			panic(err)
		}
		i += 4
		iface = true
		return
	case 'f':
		if stream[i+1] != 'a' || stream[i+2] != 'l' || stream[i+3] != 's' || stream[i+4] != 'e' {
			err := lxterrs.New("should be \"false\", not [%s]", ErrStream(stream))
			panic(err)
		}
		i += 5
		iface = false
		return
	case '"':
		raw, n := parseStr(stream[i:])
		raw = raw[1 : len(raw)-1]
		i += n
		iface = bytesString(raw)
		return
	}
	return
}

// key 后面的单元: Num, str, bool, slice, obj, null
func parseValue(stream []byte, pObj unsafe.Pointer, tag *TagInfo) (i int) {
	// i = trimSpace(stream)
	n := 0
	switch stream[i] {
	default: // num
		if (stream[i] >= '0' && stream[i] <= '9') || stream[i] == '-' {
			raw, size := parseNum(stream[i:])
			i += size
			if tag != nil {
				if tag.JSONType != Number {
					err := lxterrs.New("error type:%s", ErrStream(stream[i:]))
					panic(err)
				}
				tag.fSet(pointerOffset(pObj, tag.Offset), raw)
			}
		} else {
			err := lxterrs.New("error type:%s", ErrStream(stream[i:]))
			panic(err)
		}
	case '{': // obj
		i++
		tagObj := tag
		if tagObj != nil && tagObj.JSONType != Map && tagObj.JSONType != Struct {
			err := lxterrs.New("error type:%s", ErrStream(stream[i:]))
			panic(err)
		}
		if tagObj == nil {
			pObj = nil // TODO
		} else {
			pObj = pointerOffset(pObj, tagObj.Offset)
			if tagObj.fSet != nil {
				pObj = tag.fSet(pObj, stream[i:])
			}
		}
		if tagObj != nil && tagObj.JSONType == Map {
			n = parseMap(stream[i:], pObj, tagObj)
		} else {
			n = parseObj(stream[i:], pObj, tagObj)
		}
		i += n
	case '[': // slice
		// TODO
		i++
		n = parseSlice(stream[i:], pObj, tag)
		i += n
	case 'n':
		if stream[i+1] != 'u' || stream[i+2] != 'l' || stream[i+3] != 'l' {
			err := lxterrs.New("should be \"null\", not [%s]", ErrStream(stream))
			panic(err)
		}
		i += 4
	case 't':
		if stream[i+1] != 'r' || stream[i+2] != 'u' || stream[i+3] != 'e' {
			err := lxterrs.New("should be \"true\", not [%s]", ErrStream(stream))
			panic(err)
		}
		tag.fSet(pointerOffset(pObj, tag.Offset), stream[i:i+4])
		i += 4
	case 'f':
		if stream[i+1] != 'a' || stream[i+2] != 'l' || stream[i+3] != 's' || stream[i+4] != 'e' {
			err := lxterrs.New("should be \"false\", not [%s]", ErrStream(stream))
			panic(err)
		}
		tag.fSet(pointerOffset(pObj, tag.Offset), stream[i:i+5])
		i += 5
	case '"':
		raw, n := parseStr(stream[i:])
		raw = raw[1 : len(raw)-1]
		i += n
		if tag != nil {
			tag.fSet(pointerOffset(pObj, tag.Offset), raw)
		}
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
		i += trimSpace(stream[i:])
		i += parseObjToSlice(stream[i:], nil)
		i += trimSpace(stream[i:])
		if stream[i] == ']' {
			return
		}
		return
	}
	return
}

func parseNum(stream []byte) (raw []byte, i int) {
	for ; i < len(stream); i++ {
		c := stream[i]
		if spaceTable[c] || c == ']' || c == '}' || c == ',' {
			raw = stream[:i]
			return
		}
	}
	raw = stream
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

const deBruijn64ctz = 0x0218a392cd3d5dbf

var deBruijnIdx64ctz = [64]byte{
	0, 1, 2, 7, 3, 13, 8, 19,
	4, 25, 14, 28, 9, 34, 20, 40,
	5, 17, 26, 38, 15, 46, 29, 48,
	10, 31, 35, 54, 21, 50, 41, 57,
	63, 6, 12, 18, 24, 27, 33, 39,
	16, 37, 45, 47, 30, 53, 49, 56,
	62, 11, 23, 32, 36, 44, 52, 55,
	61, 22, 43, 51, 60, 42, 59, 58,
}

// Ctz64 counts trailing (low-order) zeroes,
// and if all are zero, then 64.
func Ctz64(x uint64) int {
	x &= -x                       // isolate low-order bit
	y := x * deBruijn64ctz >> 58  // extract part of deBruijn sequence
	i := int(deBruijnIdx64ctz[y]) // convert to bit index
	z := int((x - 1) >> 57 & 64)  // adjustment if zero
	return i + z
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

var SpaceQ = [8]byte{0x85, 0xA0, '\t', '\n', '\v', '\f', '\r', ' '}

var TrueB = true

var FalseB = false

func InSpaceQ(b byte) bool
