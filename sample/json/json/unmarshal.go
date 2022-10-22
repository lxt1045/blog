package json

import (
	"bytes"
	"errors"
	"math"
	"reflect"
	"strconv"
	"sync"
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
func parseByte0(stream []byte, b byte) (i, n int) {
loop:
	if stream[i] == b {
		n++
		i++
		goto loop
	}
	if !spaceTable[stream[i]] {
		return
	}
	i++
	goto loop
}

func parseObjToSlice(stream []byte, s []interface{}) (i int) {
	return 0
}

func bsToStr(bs []byte) string {
	return *(*string)(unsafe.Pointer(&bs))
}

// 解析 {}
// func parseObj(sts status, stream []byte, store PoolStore,  tag *TagInfo) (i int) {
func parseObj(idxSlash int, stream []byte, store PoolStore) (i, iSlash int) {
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
		// son := store.tag.GetChild(key)
		// son := GetTagChild(store.tag, key)
		var son *TagInfo
		if store.tag.MChildrenEnable {
			son = TagMapGetV(&store.tag.MChildren, key)

			// idx := hash2(key, store.tag.MChildren.idxNTable, store.tag.MChildren.idxN)
			// n := store.tag.MChildren.S[idx]
			// if bytes.Equal(key, n.K) {
			// 	son = n.V
			// }
		} else {
			son = store.tag.GetChildFromMap(key)
		}

		if son != nil {
			storeSon := PoolStore{
				tag:  son,
				pool: store.pool,
				obj:  store.obj,
			}
			n, iSlash = son.fUnm(iSlash-i, storeSon, stream[i:])
			iSlash += i
		} else {
			//TODO 通过 IndexByte 的方式快速跳过； 在下一层处理，这里 设为 nil
			// 如果是 其他： 找 ','
			// 如果是obj: 1. 找 ’}‘; 2. 找'{'； 3. 如果 2 比 1 小则循环 1 2
			// 如果是 slice : 1. 找 ’]‘; 2. 找'['； 3. 如果 2 比 1 小则循环 1 2
			// var iface interface{}
			// n, iSlash = parseInterface(iSlash-i, stream[i:], &iface)
			// iSlash += i
			n = parseEmpty(stream[i:])
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

type pair struct {
	k []byte
	v interface{}
}

var pairPool = sync.Pool{
	New: func() any {
		s := make([]pair, 0, 64)
		return &s
	},
}

// TODO: ti 中存下 map 的 len，方便下次 make，通过 key 来存？
func parseMapInterface(idxSlash int, stream []byte) (m map[string]interface{}, i, iSlash int) {
	iSlash = idxSlash
	n, nB := 0, 0
	key := []byte{}
	// m = *mapCache.Get() //map 和 interface 一起获取，合并两次损耗为一次
	// m = make(map[string]interface{})
	// value := interfaceCache.Get()
	// var v interface{}
	// value := &v
	// pairs := make([]pair, 0, 8)
	ppairs := pairPool.Get().(*[]pair)
	pairs := *ppairs
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
		pairs = append(pairs, pair{
			k: key,
		})
		n, iSlash = parseInterface(iSlash-i, stream[i:], &pairs[len(pairs)-1].v)
		iSlash += i
		i += n
		// m[string(key)] = *value
		n, nB = parseByte(stream[i:], ',')
		i += n
		if nB != 1 {
			if nB == 0 && '}' == stream[i] {
				i++

				// map
				// m = make(map[string]interface{}, len(pairs))
				m = makeMapEface(len(pairs))

				for i := range pairs {
					m[*(*string)(unsafe.Pointer(&pairs[i].k))] = pairs[i].v
				}
				*ppairs = pairs[:0]
				pairPool.Put(ppairs)
				return
			}
			panic(lxterrs.New(ErrStream(stream[i:])))
		}
	}
}

var poolSliceInterface = sync.Pool{New: func() any {
	return make([]interface{}, 1024)
}}

func parseSliceInterface(idxSlash int, stream []byte) (s []interface{}, i, iSlash int) {
	iSlash = idxSlash
	i = trimSpace(stream[i:])
	var value interface{}
	s = poolSliceInterface.Get().([]interface{})
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
				if cap(s)-len(s) > 4 {
					sLeft := s[len(s):]
					poolSliceInterface.Put(sLeft)
					s = s[:len(s):len(s)]
				}
				return
			}
			panic(lxterrs.New(ErrStream(stream[i:])))
		}
	}
}
func parseSlice(idxSlash int, stream []byte, store PoolStore) (i, iSlash int) {
	iSlash = idxSlash
	i = trimSpace(stream[i:])
	store.obj = pointerOffset(store.obj, store.tag.Offset)
	pBase := store.tag.fSet(store, stream[i:])
	if stream[i] == ']' {
		i++
		pHeader := (*SliceHeader)(pBase)
		pHeader.Data = store.obj
		return
	}
	son := store.tag.ChildList[0]
	// size := int(son.Type.Size())
	size := son.TypeSize
	uint8s := store.tag.SPool.Get().(*[]uint8)
	// pHPool := (*SliceHeader)(unsafe.Pointer(uint8s))
	pHeader := (*SliceHeader)(pBase)
	bases := (*[]uint8)(pBase)
	for n, nB := 0, 0; ; {
		if len(*uint8s)+size > cap(*uint8s) {
			l := cap(*uint8s) / size
			c := l * 2
			if c < 1024 {
				c = 1024
			}
			v := reflect.MakeSlice(store.tag.BaseType, 0, c)
			p := reflectValueToPointer(&v)
			pH := (*SliceHeader)(p)
			pH.Cap = pH.Cap * size
			news := (*[]uint8)(p)

			copy(*news, *uint8s)
			// _ = append((*(*[]uint8)(p))[:0], *(*[]uint8)(unsafe.Pointer(pHeader))...)
			*uint8s = *news
		}
		l := len(*uint8s)
		*uint8s = (*uint8s)[:l+size]

		p := unsafe.Pointer(&(*uint8s)[l])
		if son != nil && son.fUnm != nil {
			n, iSlash = son.fUnm(iSlash-i, PoolStore{
				obj:  p,
				tag:  son,
				pool: store.pool,
			}, stream[i:])
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
				break
			}
			panic(lxterrs.New(ErrStream(stream[i:])))
		}
	}

	*bases = (*uint8s)[:len(*uint8s):len(*uint8s)]
	if cap(*uint8s)-len(*uint8s) > 4*size {
		*uint8s = (*uint8s)[len(*uint8s):]
		store.tag.SPool.Put(uint8s)
	}
	// pH.Data = uintptr(pointerOffset(unsafe.Pointer(pHeader.Data), uintptr(pHeader.Len)))
	// pH.Cap = pHeader.Cap - pHeader.Len
	pHeader.Len = pHeader.Len / size
	pHeader.Cap = pHeader.Cap / size

	return
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
		// *p = s
		ps := islicePool.Get()
		*ps = s
		pEface := (*GoEface)(unsafe.Pointer(p))
		pEface.Type = isliceEface.Type
		pEface.Value = unsafe.Pointer(ps)
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
		// *p = bytesString(raw)
		// return

		// pstr := strPool.Get() //
		pstr := GetStr(strPool) //
		// bytesCopyToString(raw, pstr)
		*pstr = *(*string)(unsafe.Pointer(&raw))
		pEface := (*GoEface)(unsafe.Pointer(p))
		pEface.Type = strEface.Type
		pEface.Value = unsafe.Pointer(pstr)

		// // TODO 探索无GC 分配方式
		// strs := stringPool.Get().(*[]string)
		// str := &(*strs)[0]
		// if len(*strs) >= 2 {
		// 	*strs = (*strs)[1:]
		// 	stringPool.Put(strs)
		// }
		// bytesCopyToString(raw, str)
		// pEface := (*GoEface)(unsafe.Pointer(p))
		// pEface.Type = strEface.Type
		// pEface.Value = unsafe.Pointer(str)

		return
	}
	return
}

var strPool = NewBatch[string]()
var islicePool = NewBatch[[]interface{}]()

var strEface = UnpackEface("")
var isliceEface = UnpackEface([]interface{}{})

var stringPool = sync.Pool{
	New: func() any {
		s := make([]string, 512, 512) //8*1024/16
		return &s
	},
}

func parseEmptyObjSlice(stream []byte, bLeft, bRight byte) (i int) {
	indexQuote := func(stream []byte, i int) int {
		for {
			iDQuote := bytes.IndexByte(stream[i:], '"')
			if iDQuote < 0 {
				return math.MaxInt32
			}
			i += iDQuote // 指向 '"'
			if stream[i-1] != '\\' {
				return i
			}
			j := i - 2
			for ; stream[j] == '\\'; j-- {
			}
			if (i-j)%2 == 0 {
				i++
				continue
			}
			return i
		}
	}
	i++
	nBrace := 0                                   // " 和 {
	iBraceL := bytes.IndexByte(stream[i:], bLeft) //通过 ’“‘ 的 idx 来确定'{' '}' 是否在字符串中
	iBraceR := bytes.IndexByte(stream[i:], bRight)
	if iBraceL < 0 {
		iBraceL = math.MaxInt32 // 保证 +i 后不会溢出
	}
	if iBraceR < 0 {
		iBraceR = math.MaxInt32
	}
	iBraceL, iBraceR = iBraceL+i, iBraceR+i

	iDQuoteL := indexQuote(stream, i)
	iDQuoteR := indexQuote(stream, iDQuoteL+1)

	for {
		// 1. 以 iBraceR 为边界
		if iBraceR < iBraceL {
			if iDQuoteR < iBraceR {
				// ']'在右区间
				iDQuoteL = indexQuote(stream, iDQuoteR+1)
				iDQuoteR = indexQuote(stream, iDQuoteL+1)
				continue
			} else if iBraceR < iDQuoteL {
				// ']'在左区间
				if nBrace == 0 {
					i = iBraceR + 1
					return
				}
				nBrace--
				iBraceR++
				iBraceRNew := bytes.IndexByte(stream[iBraceR:], bRight)
				if iBraceRNew < 0 {
					iBraceRNew = math.MaxInt32
				}
				iBraceR += iBraceRNew
				continue
			} else {
				// ']'在中间区间
				iBraceR = bytes.IndexByte(stream[iDQuoteR:], bRight)
				if iBraceR < 0 {
					iBraceR = math.MaxInt32
				}
				iBraceR += iDQuoteR
				continue
			}
		} else {
			// iBraceL < iBraceR
			// 2. 以 iBraceR 为边界

			if iDQuoteR < iBraceL {
				// ']'在右区间
				iDQuoteL = indexQuote(stream, iDQuoteR+1)
				iDQuoteR = indexQuote(stream, iDQuoteL+1)
				continue
			} else if iBraceL < iDQuoteL {
				// ']'在左区间
				nBrace++
				iBraceL++
				iBraceLNew := bytes.IndexByte(stream[iBraceL:], bLeft) //通过 ’“‘ 的 idx 来确定'{' '}' 是否在字符串中
				if iBraceLNew < 0 {
					iBraceLNew = math.MaxInt32 // 保证 +i 后不会溢出
				}
				iBraceL += iBraceLNew
				continue
			} else {
				// ']'在中间区间
				iBraceL = bytes.IndexByte(stream[iDQuoteR:], bLeft)
				if iBraceL < 0 {
					iBraceL = math.MaxInt32
				}
				iBraceL += iDQuoteR
				continue
			}
		}
	}
	return
}

// key 后面的单元: Num, str, bool, slice, obj, null
func parseEmpty(stream []byte) (i int) {
	//TODO 通过 IndexByte 的方式快速跳过； 在下一层处理，这里 设为 nil
	// 如果是 其他： 找 ','
	// 如果是obj: 1. 找 ’}‘; 2. 找'{'； 3. 如果 2 比 1 小则循环 1 2
	// 如果是 slice : 1. 找 ’]‘; 2. 找'['； 3. 如果 2 比 1 小则循环 1 2

	switch stream[0] {
	default: // num
		for ; i < len(stream); i++ {
			c := stream[i]
			if c == ']' || c == '}' || c == ',' {
				break
			}
		}
		return
	case '{': // obj
		n := parseEmptyObjSlice(stream[i:], '{', '}')
		i += n
		return
	case '[': // slice
		n := parseEmptyObjSlice(stream[i:], '[', ']')
		i += n
		return
	case ']', '}':
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
		return
	case 'f':
		if stream[i+1] != 'a' || stream[i+2] != 'l' || stream[i+3] != 's' || stream[i+4] != 'e' {
			err := lxterrs.New("should be \"false\", not [%s]", ErrStream(stream))
			panic(err)
		}
		i = 5
		return
	case '"':
		i++
		for {
			iDQuote := bytes.IndexByte(stream[i:], '"')
			i += iDQuote // 指向 '"'
			if stream[i-1] == '\\' {
				j := i - 2
				for ; stream[j] == '\\'; j-- {
				}
				if (i-j)%2 == 0 {
					i++
					continue
				}
			}
			i++
			return
		}
	}
	return
}

type status struct {
	idxUnescape int
}

//解析 obj: {}, 或 []
func parseRoot(stream []byte, store PoolStore) (err error) {
	idxSlash := bytes.IndexByte(stream[1:], '\\')
	if idxSlash < 0 {
		idxSlash = math.MaxInt
	}
	if stream[0] == '{' {
		parseObj(idxSlash, stream[1:], store)
		return
	}
	if stream[0] == '[' {
		parseSlice(idxSlash, stream[1:], store)
		return
	}
	return
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

func parseUnescapeStr(stream []byte, nextQuotesIdx, nextSlashIdxIn int) (raw []byte, i, nextSlashIdx int) {
	nextSlashIdx = nextSlashIdxIn
	if nextSlashIdx < 0 {
		nextSlashIdx = bytes.IndexByte(stream[1:], '\\')
		if nextSlashIdx < 0 {
			nextSlashIdx = math.MaxInt
			i += nextQuotesIdx
			i++
			return
		} else {
			nextSlashIdx++
		}

		// 处理 '\"'
		for {
			i += nextQuotesIdx // 指向 '"'
			if stream[i-1] == '\\' {
				j := i - 2
				for ; stream[j] == '\\'; j-- {
				}
				if (i-j)%2 == 0 {
					i++
					nextQuotesIdx = bytes.IndexByte(stream[i:], '"')
					continue
				}
			}
			i++
			return
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
			raw = make([]byte, 0, nextQuotesIdx)
			raw = append(raw[:0], stream[1:i]...) //新建 []byte 避免修改员 stream
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
