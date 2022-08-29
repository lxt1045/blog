package json

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"unicode/utf16"
	"unicode/utf8"
	"unsafe"

	lxterrs "github.com/lxt1045/errors"
)

//解析 []: curType 因为可能 StructField 可能包含了baseElem,所以需要这个参数
func setSliceField(key string, pObj unsafe.Pointer, stream []byte, typ Type, tag *TagInfo, curType reflect.Type) (i int) {
	if len(stream) < 2 || stream[0] != '[' {
		// 此 struct 结束语法分析
	}
	for i < len(stream) {
		switch tag.StructField.Type.Kind() {
		case reflect.Struct:
			pField := unsafe.Pointer(uintptr(pObj) + uintptr(tag.StructField.Offset))
			size := parseNextUnit(stream, pField, tag.Children)
			i += size
		case reflect.Slice:
			curType = curType.Elem()
			size := setSliceField(key, pObj, stream, typ, tag, curType)
			i += size
		}
		if stream[i] == ']' {
			// 此 slice 结束语法分析
		}
	}
	return
}

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
	tryPanic(str)
	return str
}
func panicIncorrectType(typ Type, tag *TagInfo) {
	t := strconv.Itoa(int(typ))
	panic(errors.New("incorrect type: " + t + " to " + tag.BaseKind.String()))
}

func setObjField(pObj unsafe.Pointer, tag *TagInfo, raw []byte) (i int) {
	if tag.BaseKind != reflect.Struct {
		panicIncorrectType(False, tag)
	}
	pField := unsafe.Pointer(uintptr(pObj) + uintptr(tag.StructField.Offset))
	size := parseNextUnit(raw, pField, tag.Children)
	i += size
	return
}

var spaceTable = [256]bool{'\t': true, '\n': true, '\v': true, '\f': true, '\r': true, ' ': true, 0x85: true, 0xA0: true}

//inline
func trimSpace(stream []byte) (i int) {
	// for i = 0; i < len(stream) && spaceTable[stream[i]]; i++ {}
	for i = 0; spaceTable[stream[i]]; i++ {
	}
	// for i = 0; InSpaceQ(stream[i]); i++ {
	// }
	return
}

// 解析：冒号 和 逗号 等单字符
func parseByte(stream []byte, b byte) (i int) {
	var n byte = 0
	for ; ; i++ {
		if stream[i] == b {
			n++
			i++
		}
		if !spaceTable[stream[i]] {
			break
		}
	}
	if n != 1 {
		panicIncorrectFormat(stream[i:])
	}
	return
}

func parseObjToMap(stream []byte, m map[string]interface{}) (i int) {
	return 0
}
func parseObjToSlice(stream []byte, s []interface{}) (i int) {
	return 0
}

//解析 obj: {}, 或 []
func parseNextUnit(stream []byte, pObj unsafe.Pointer, tis map[string]*TagInfo) (i int) {
	var key []byte
	i++
	for i < len(stream) {
		// i += trimSpace(stream[i:])
		switch stream[i] {
		default:
			if (stream[i] >= '0' && stream[i] <= '9') || stream[i] == '-' {
				if len(key) <= 0 {
					panicIncorrectFormat(stream[i:])
				}
				raw, size := parseNum(stream[i:])
				i += size
				if tag := tis[string(key)]; tag != nil {
					setNumberField(pObj, tag, raw, Number)
				}
				key = nil
			} else {
				log.Printf("error byte:%x", stream[i])
				panicIncorrectFormat(stream[i:])
			}
		case '}', ']':
			i++
			return // 此 struct 结束语法分析
		case '{': // obj
			if len(key) <= 0 {
				panicIncorrectFormat(stream[i:])
			}
			// i++
			if tag := tis[string(key)]; tag != nil {
				i += setObjField(pObj, tag, stream[i:])
			} else {
				i += parseNextUnit(stream[i:], nil, tag.Children)
			}
			key = nil
		case '[': // obj
			if len(key) <= 0 {
				panicIncorrectFormat(stream[i:])
			}
			i++
			if tag := tis[string(key)]; tag != nil {
				i += setObjField(pObj, tag, stream[i:])
			} else {
				i += parseNextUnit(stream[i:], nil, tag.Children)
			}
			key = nil
		case 'n':
			if len(key) <= 0 {
				panicIncorrectFormat(stream[i:])
			}
			if stream[i+1] != 'u' || stream[i+2] != 'l' || stream[i+3] != 'l' {
				panicIncorrectFormat(stream[i:])
			}
			i += 4
			key = nil
		case 't':
			if len(key) <= 0 {
				panicIncorrectFormat(stream[i:])
			}
			if stream[i+1] != 'r' || stream[i+2] != 'u' || stream[i+3] != 'e' {
				panicIncorrectFormat(stream[i:])
			}
			i += 4
			if tag := tis[string(key)]; tag != nil {
				setBoolField(pObj, tag, true)
			}
			key = nil
		case 'f':
			if len(key) <= 0 {
				panicIncorrectFormat(stream[i:])
			}
			if stream[i+1] != 'a' || stream[i+2] != 'l' || stream[i+3] != 's' || stream[i+4] != 'e' {
				panicIncorrectFormat(stream[i:])
			}
			i += 5
			if tag := tis[string(key)]; tag != nil {
				setBoolField(pObj, tag, false)
			}
			key = nil
		case '"':
			if len(key) <= 0 {
				// i += trimSpace(stream[i:])
				size := 0
				key, size = parseStr(stream[i:]) //先解析key 再解析value
				i += size
				i += parseByte(stream[i:], ':')
				continue
			} else {
				raw, size := parseStr(stream[i:])
				i += size
				if tag := tis[string(key)]; tag != nil {
					setStringField(pObj, tag, raw)
				}
				key = nil
			}
		}
		i += trimSpace(stream[i:])
		if stream[i] == ',' {
			i++
			continue
		}
	}
	return
}

func parseKey(stream []byte) (key []byte, i int, err error) {
	if stream[i] != '"' {
		lxterrs.New("errors key:%s", ErrStream(stream))
		return
	}
	key, size := parseStr(stream[i+1:]) //先解析key 再解析value
	i += size
	return
}

// 解析 {}
func parseObj(stream []byte, pObj unsafe.Pointer, tag *TagInfo) (i int, err error) {
	for i < len(stream) && stream[i] != '}' {
		key, n, err := parseKey(stream[i+1:])
		if err != nil {
			lxterrs.Wrap(err, "errors key:%s", ErrStream(stream))
			return i, err
		}
		tag := tag.Children[string(key)]
		key = nil

		n = parseByte(stream[i+1:], ':')
		i += n
		n, err = parseValue(stream[i+1:], pObj, tag)
		if err != nil {
			lxterrs.Wrap(err, "errors key:%s", ErrStream(stream))
			return i, err
		}

		n = parseByte(stream[i+1:], ',')
		i += n
	}
	if stream[i] == '}' {
		return
	}
	panicIncorrectFormat(stream[i:])
	return
}

func parseSlice(stream []byte, pObj unsafe.Pointer, tag *TagInfo) (i int, err error) {
	size := int(tag.StructField.Type.Size())
	raw := make([]byte, 0, 8*size) //先分配 8 个
	s := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(&raw[0])),
		Len:  0,
		Cap:  8,
	}
	for i < len(stream) && stream[i] != ']' {
		if s.Len >= s.Cap {
			raw = append(raw, make([]byte, s.Len*2*size)...)
			s.Data = uintptr(unsafe.Pointer(&raw[0]))
			s.Cap = s.Len * 2
		}
		p := unsafe.Pointer(&raw[s.Len*size])
		n, err := parseValue(stream[i+1:], p, tag)
		if err != nil {
			lxterrs.Wrap(err, "errors key:%s", ErrStream(stream))
			return i, err
		}
		s.Len++

		n = parseByte(stream[i+1:], ',')
		i += n
	}
	if stream[i] == ']' {
		return
	}
	panicIncorrectFormat(stream[i:])
	return
}

// key 后面的单元: Num, str, bool, slice, obj, null
func parseValue(stream []byte, pObj unsafe.Pointer, tag *TagInfo) (i int, err error) {
	// i = trimSpace(stream)
	n := 0
	switch stream[i] {
	default: // num
		if (stream[i] >= '0' && stream[i] <= '9') || stream[i] == '-' {
			raw, size := parseNum(stream[i:])
			i += size
			if tag != nil {
				if tag.fSet != nil {
					// TODO:
					// tag.fSet()
				} else {
					setNumberField(pObj, tag, raw, Number)
				}
			}
		} else {
			log.Printf("error byte:%x", stream[i])
			panicIncorrectFormat(stream[i:])
		}
	case '{': // obj
		if tag != nil && (tag.BaseKind == reflect.Map || tag.BaseKind == reflect.Struct) {
			n, err = parseObj(stream[i:], pObj, tag)

			i += n
		} else {
			n, err = parseObj(stream[i:], pObj, nil)

			i += n
		}
	case '[': // slice
		// TODO
		n = trimSpace(stream[i+1:])
		i += n
		parseSlice(stream[i:], pObj, tag)
	case 'n':
		if stream[i+1] != 'u' || stream[i+2] != 'l' || stream[i+3] != 'l' {
			err = lxterrs.New("should be \"null\", not [%s]", ErrStream(stream))
			return
		}
		i += 4
	case 't':
		if stream[i+1] != 'r' || stream[i+2] != 'u' || stream[i+3] != 'e' {
			err = lxterrs.New("should be \"true\", not [%s]", ErrStream(stream))
			return
		}
		i += 4
		if tag != nil {
			err = setBoolField(pObj, tag, true)
			if err != nil {
				return
			}
		}
	case 'f':
		if stream[i+1] != 'a' || stream[i+2] != 'l' || stream[i+3] != 's' || stream[i+4] != 'e' {
			err = lxterrs.New("should be \"true\", not [%s]", ErrStream(stream))
			return
		}
		i += 5
		if tag != nil {
			err = setBoolField(pObj, tag, false)
			if err != nil {
				return
			}
		}
	case '"':
		raw, n := parseStr(stream[i:])
		i += n
		if tag != nil {
			err = setStringField(pObj, tag, raw)
			if err != nil {
				return
			}
		}
	}

	return
}

//解析 obj: {}, 或 []
func parseRoot(stream []byte, pObj unsafe.Pointer, tag *TagInfo) (err error) {
	i := 1
	n := trimSpace(stream[i:])
	i += n
	if stream[i] == '{' {
		for i < len(stream) {
			n, err := parseObj(stream[i:], pObj, tag)
			if err != nil {
				err = lxterrs.Wrap(err, "parseObj")
				return err
			}
			i += n
			i += trimSpace(stream[i:])
		}
		if stream[i] == '}' {
			return
		}
		panicIncorrectFormat(stream[i:])
	}
	if stream[i] == '[' {
		i += parseObjToSlice(stream[i:], nil)
		i += trimSpace(stream[i:])
		if stream[i] == ']' {
			return
		}
		panicIncorrectFormat(stream[i:])
	}
	panicIncorrectFormat(stream[i:])
	err = fmt.Errorf("json must start with '{' or '[', %s", ErrStream(bs[i:]))
	return
}

func parseNum(stream []byte) (raw []byte, i int) {
	for ; i < len(stream); i++ {
		c := stream[i]
		if spaceTable[c] || c == ']' || c == '}' || c == ',' {
			raw, i = stream[:i], i+1
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

	nextSlashIdx := -1
	// nextSlashIdx = bytes.IndexByte(stream[i:], '\\')
	nextQuotesIdx := bytes.IndexByte(stream, '"')
	if nextQuotesIdx < 0 {
		panic(string(stream[i:]))
	}
	if nextSlashIdx > 0 && nextSlashIdx < nextQuotesIdx {
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

	raw = stream[1 : 1+nextQuotesIdx]
	return raw, nextQuotesIdx + 1
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
