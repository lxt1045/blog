package json

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unsafe"

	"github.com/lxt1045/errors"
	lxterrs "github.com/lxt1045/errors"
)

var (
	cacheStructTagInfoP   = newCache[uintptr, *tagNode]()
	cacheStructTagInfoStr = newCache[string, *tagNodeStr]()
)

// 获取 string 的起始地址
func strToUintptr(p string) uintptr {
	return *(*uintptr)(unsafe.Pointer(&p))
}
func LoadTagNode(typ reflect.Type) (n *tagNode, err error) {
	pname := strToUintptr(typ.Name())
	ppkg := strToUintptr(typ.PkgPath())
	n, ok := cacheStructTagInfoP.Get(pname)
	if ok {
		if n.pkgPath == ppkg {
			return
		}
		if n, ok := n.pkgCache.Get(ppkg); ok {
			return n, nil
		}
	}
	ti, err := tagParse(typ)
	if err != nil {
		return nil, err
	}
	n = &tagNode{
		pkgPath:  ppkg,
		tagInfo:  ti,
		pkgCache: newCache[uintptr, *tagNode](),
	}
	if !ok {
		cacheStructTagInfoP.Set(pname, n)
	} else {
		n.pkgCache.Set(ppkg, n)
	}
	return
}
func LoadTagNodeStr(typ reflect.Type) (n *tagNodeStr) {
	pname := typ.Name()
	ppkg := typ.PkgPath()
	n, ok := cacheStructTagInfoStr.Get(pname)
	if ok {
		if n.pkgPath == ppkg {
			return
		}
		if n, ok := n.pkgCache.Get(ppkg); ok {
			return n
		}
	}
	ti, err := tagParse(typ)
	if err != nil {
		panic(err)
	}
	n = &tagNodeStr{
		pkgPath:  ppkg,
		tagInfo:  ti,
		pkgCache: newCache[string, *tagNodeStr](),
	}
	if !ok {
		cacheStructTagInfoStr.Set(pname, n)
	} else {
		n.pkgCache.Set(ppkg, n)
	}
	return
}

type tagNodeStr struct {
	pkgPath  string
	tagInfo  *TagInfo
	pkgCache cache[string, *tagNodeStr] //如果 name 相等，则从这个缓存中获取
}

type tagNode struct {
	pkgPath  uintptr
	tagInfo  *TagInfo
	pkgCache cache[uintptr, *tagNode] //如果 name 相等，则从这个缓存中获取
}

/*
JSON的基本数据类型：

数值： 十进制数，不能有前导0，可以为负数，可以有小数部分。还可以用e或者E表示指数部分。不能包含非数，如NaN。
      不区分整数与浮点数。JavaScript用双精度浮点数表示所有数值。
字符串：以双引号""括起来的零个或多个Unicode码位。支持反斜杠开始的转义字符序列。
布尔值：表示为true或者false。
数组：有序的零个或者多个值。每个值可以为任意类型。序列表使用方括号[，]括起来。元素之间用逗号,分割。形如：[value, value]
对象：若干无序的“键-值对”(key-value pairs)，其中键只能是字符串[1]。建议但不强制要求对象中的键是独一无二的。
     对象以花括号{开始，并以}结束。键-值对之间使用逗号分隔。键与值之间用冒号:分割。
空值：值写为null

token(6种标点符号、字符串、数值、3种字面量)之间可以存在有限的空白符并被忽略。四个特定字符被认为是空白符：空格符、
水平制表符、回车符、换行符。空白符不能出现在token内部(但空格符可以出现在字符串内部)。JSON标准不允许有字节序掩码，
不提供注释的句法。 一个有效的JSON文档的根节点必须是一个对象或一个数组。

JSON交换时必须编码为UTF-8。[2]转义序列可以为：“\\”、“\"”、“\/”、“\b”、“\f”、“\n”、“\r”、“\t”，或Unicode16
进制转义字符序列(\u后面跟随4位16进制数字)。对于不在基本多文种平面上的码位，必须用UTF-16代理对(surrogate pair)
表示，例如对于Emoji字符——喜极而泣的表情(U+1F602 😂 face with tears of joy)在JSON中应表示为：

------------
在 Go 中并不是所有的类型都能进行序列化：
	JSON object key 只支持 string
	Channel、complex、function 等 type 无法进行序列化
	数据中如果存在循环引用，则不能进行序列化，因为序列化时会进行递归
	Pointer 序列化之后是其指向的值或者是 nil
	只有 struct 中支持导出的 field 才能被 JSON package 序列化，即首字母大写的 field。
反序列化:
	`json:"field,string"`
	`json:"some_field,omitempty"`
	`json:"-"`
默认的 JSON 只支持以下几种 Go 类型：
	bool for JSON booleans
	float64 for JSON numbers
	string for JSON strings
	nil for JSON null
反序列化对 slice、map、pointer 的处理:
如果我们序列化之前不知道其数据格式，我们可以使用 interface{} 来存储我们的 decode 之后的数据：
	var f interface{}
	err := json.Unmarshal(b, &f)
	key 是 string，value 是存储在 interface{} 内的。想要获得 f 中的数据，我们首先需要进行 type assertion，
然后通过 range 迭代获得 f 中所有的 key ：
		m := f.(map[string]interface{})
		for k, v := range m {
			switch vv := v.(type) {
			case string:
				fmt.Println(k, "is string", vv)
			case float64:
				fmt.Println(k, "is float64", vv)
			case []interface{}:
				fmt.Println(k, "is an array:")
				for i, u := range vv {
					fmt.Println(i, u)
				}
			default:
				fmt.Println(k, "is of a type I don't know how to handle")
			}
		}
Stream JSON:
	除了 marshal 和 unmarshal 函数，Go 还提供了 Decoder 和 Encoder 对 stream JSON 进行处理，常见 request
中的 Body、文件等

嵌入式 struct 的序列化:
	Go 支持对 nested struct 进行序列化和反序列化:
自定义序列化函数:
	Go JSON package 中定了两个 Interface Marshaler 和 Unmarshaler ，实现这两个 Interface 可以让你定义的
type 支持序列化操作。
*/

//TagInfo 拥有tag的struct的成员的解析结果
type TagInfo struct {
	// Value       reflect.Value
	StructField reflect.StructField //`json:"-"`

	Offset       uintptr      //偏移量
	BaseKind     reflect.Kind // 次成员可能是 **string,[]int 等这种复杂类型,这个 用来指示 "最里层" 的类型
	TagName      string       //
	StringTag    bool         // `json:"field,string"`: 此情形下,需要把struct的int转成json的string
	OmitemptyTag bool         //  `json:"some_field,omitempty"`
	Children     map[string]*TagInfo
	ChildList    []*TagInfo // 遍历的顺序和速度

	fSet func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer, err error)
	fGet func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte)
}

func (p *TagInfo) cacheKey() (k string) {
	return p.TagName
}
func (t *TagInfo) GetChild(key []byte) *TagInfo {
	return t.Children[string(key)]
}
func (t *TagInfo) AddChild(c *TagInfo) {
	if len(t.Children) == 0 {
		t.Children = make(map[string]*TagInfo)
	}
	if _, ok := t.Children[c.TagName]; ok {
		err := fmt.Errorf("error, tag[%s]类型配置出错,字段重复", c.TagName)
		panic(err)
	}
	t.ChildList = append(t.ChildList, c)
	t.Children[c.TagName] = c
	return
}

// []byte 是一种特殊的底层数据类型，需要 base64 编码
func isBytes(typ reflect.Type) bool {
	bsType := reflect.TypeOf(&[]byte{})
	return typ.PkgPath() == bsType.PkgPath() && typ.Name() == bsType.Name()
}
func (ti *TagInfo) setFuncs() (err error) {
	ptrDeep, baseType := 0, ti.StructField.Type
	for typ := ti.StructField.Type; ; typ = typ.Elem() {
		if typ.Kind() == reflect.Ptr {
			ptrDeep++
			continue
		}
		baseType = typ
		break
	}

	// 先从最后一个基础类型开始处理
	var fSet func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer, err error)
	var fGet func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte)

	// 先从最后一个基础类型开始处理
	switch baseType.Kind() {
	case reflect.Bool:
		fSet = func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer, err error) {
			pBase = pObj
			if raw[0] == 't' {
				*(*bool)(pObj) = true
			} else {
				*(*bool)(pObj) = false
			}
			return
		}
		fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
			pBase = pObj
			if *(*bool)(pObj) {
				out = append(in, []byte("false")...)
			} else {
				out = append(in, []byte("true")...)
			}
			return
		}
		if ptrDeep > 0 {
			if ptrDeep > 0 {
				fSet, fGet = getBaseTypeFuncs[bool](ptrDeep, fSet, fGet)
			}
		}
	case reflect.Uint, reflect.Uint64, reflect.Uintptr:
		fSet = func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer, err error) {
			pBase = pObj
			num, err := strconv.ParseUint(bytesString(raw), 10, 64)
			if err != nil {
				err = lxterrs.Wrap(err, ErrStream(raw))
				return
			}
			*(*uint64)(pObj) = num
			return
		}
		fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
			pBase = pObj
			num := *(*uint64)(pObj)
			str := strconv.FormatUint(num, 10)
			out = append(in, str...)
			return
		}
		if ptrDeep > 0 {
			fSet, fGet = getBaseTypeFuncs[uint64](ptrDeep, fSet, fGet)
		}
	case reflect.Int, reflect.Int64:
		fSet = func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer, err error) {
			pBase = pObj
			num, err := strconv.ParseInt(bytesString(raw), 10, 64)
			if err != nil {
				err = lxterrs.Wrap(err, ErrStream(raw))
				return
			}
			*(*int64)(pObj) = num
			return
		}
		fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
			pBase = pObj
			num := *(*int64)(pObj)
			str := strconv.FormatInt(num, 10)
			out = append(in, str...)
			return
		}
		if ptrDeep > 0 {
			fSet, fGet = getBaseTypeFuncs[int64](ptrDeep, fSet, fGet)
		}
	case reflect.Uint32:
		fSet = func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer, err error) {
			pBase = pObj
			num, err := strconv.ParseUint(bytesString(raw), 10, 32)
			if err != nil {
				err = lxterrs.Wrap(err, ErrStream(raw))
				return
			}
			*(*uint32)(pObj) = uint32(num)
			return
		}
		fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
			pBase = pObj
			num := *(*uint32)(pObj)
			str := strconv.FormatUint(uint64(num), 10)
			out = append(in, str...)
			return
		}
		if ptrDeep > 0 {
			fSet, fGet = getBaseTypeFuncs[uint32](ptrDeep, fSet, fGet)
		}
	case reflect.Int32:
		fSet = func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer, err error) {
			num, err := strconv.ParseInt(bytesString(raw), 10, 32)
			if err != nil {
				err = lxterrs.Wrap(err, ErrStream(raw))
				return
			}
			*(*int32)(pObj) = int32(num)
			return
		}
		fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
			pBase = pObj
			num := *(*int32)(pObj)
			str := strconv.FormatInt(int64(num), 10)
			out = append(in, str...)
			return
		}
		if ptrDeep > 0 {
			fSet, fGet = getBaseTypeFuncs[int32](ptrDeep, fSet, fGet)
		}
	case reflect.Uint16:
		fSet = func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer, err error) {
			pBase = pObj
			num, err := strconv.ParseUint(bytesString(raw), 10, 32)
			if err != nil {
				err = lxterrs.Wrap(err, ErrStream(raw))
				return
			}
			*(*uint16)(pObj) = uint16(num)
			return
		}
		fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
			pBase = pObj
			num := *(*uint16)(pObj)
			str := strconv.FormatUint(uint64(num), 10)
			out = append(in, str...)
			return
		}
		if ptrDeep > 0 {
			fSet, fGet = getBaseTypeFuncs[uint16](ptrDeep, fSet, fGet)
		}
	case reflect.Int16:
		fSet = func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer, err error) {
			pBase = pObj
			num, err := strconv.ParseInt(bytesString(raw), 10, 32)
			if err != nil {
				err = lxterrs.Wrap(err, ErrStream(raw))
				return
			}
			*(*int16)(pObj) = int16(num)
			return
		}
		fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
			pBase = pObj
			num := *(*int16)(pObj)
			str := strconv.FormatInt(int64(num), 10)
			out = append(in, str...)
			return
		}
		if ptrDeep > 0 {
			fSet, fGet = getBaseTypeFuncs[int16](ptrDeep, fSet, fGet)
		}
	case reflect.Uint8:
		fSet = func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer, err error) {
			pBase = pObj
			num, err := strconv.ParseUint(bytesString(raw), 10, 32)
			if err != nil {
				err = lxterrs.Wrap(err, ErrStream(raw))
				return
			}
			*(*uint8)(pObj) = uint8(num)
			return
		}
		fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
			pBase = pObj
			num := *(*uint8)(pObj)
			str := strconv.FormatUint(uint64(num), 10)
			out = append(in, str...)
			return
		}
		if ptrDeep > 0 {
			fSet, fGet = getBaseTypeFuncs[uint8](ptrDeep, fSet, fGet)
		}
	case reflect.Int8:
		fSet = func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer, err error) {
			pBase = pObj
			num, err := strconv.ParseInt(bytesString(raw), 10, 32)
			if err != nil {
				err = lxterrs.Wrap(err, ErrStream(raw))
				return
			}
			*(*int8)(pObj) = int8(num)
			return
		}
		fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
			pBase = pObj
			num := *(*int8)(pObj)
			str := strconv.FormatInt(int64(num), 10)
			out = append(in, str...)
			return
		}
		if ptrDeep > 0 {
			fSet, fGet = getBaseTypeFuncs[int8](ptrDeep, fSet, fGet)
		}
	case reflect.Float64:
		fSet = func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer, err error) {
			pBase = pObj
			f, err := strconv.ParseFloat(bytesString(raw), 64)
			if err != nil {
				err = lxterrs.Wrap(err, ErrStream(raw))
				return
			}
			*(*float64)(pObj) = f
			return
		}
		fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
			pBase = pObj
			num := *(*float64)(pObj)
			out = strconv.AppendFloat(in, float64(num), 'f', -1, 64)
			return
		}
		if ptrDeep > 0 {
			fSet, fGet = getBaseTypeFuncs[float64](ptrDeep, fSet, fGet)
		}
	case reflect.Float32:
		fSet = func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer, err error) {
			pBase = pObj
			f, err := strconv.ParseFloat(bytesString(raw), 64)
			if err != nil {
				err = lxterrs.Wrap(err, ErrStream(raw))
				return
			}
			*(*float64)(pObj) = f
			return
		}
		fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
			pBase = pObj
			num := *(*float64)(pObj)
			out = strconv.AppendFloat(in, float64(num), 'f', -1, 64)
			return
		}
		if ptrDeep > 0 {
			fSet, fGet = getBaseTypeFuncs[float32](ptrDeep, fSet, fGet)
		}
	case reflect.Complex64:
		fSet = func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer, err error) {
			pBase = pObj
			return
		}
		fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
			pBase, out = pObj, in
			return
		}
	case reflect.String:
		fSet = func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer, err error) {
			pBase = pObj
			*(*string)(pObj) = *(*string)(unsafe.Pointer(&raw))
			return
		}
		fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
			pBase = pObj
			str := *(*string)(pObj)
			out = append(in, str...)
			return
		}
		if ptrDeep > 0 {
			fSet, fGet = getBaseTypeFuncs[string](ptrDeep, fSet, fGet)
		}
	case reflect.Slice: // &[]byte
		if isBytes(baseType) {
			// []byte 是一种特殊的底层数据类型，需要 base64 编码
			fSet = func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer, err error) {
				pBase = pObj
				pbs := (*[]byte)(pObj)
				*pbs = make([]byte, len(raw)*2)
				n, err := base64.StdEncoding.Decode(*pbs, raw)
				if err != nil {
					err = lxterrs.Wrap(err, ErrStream(raw))
					return
				}
				*pbs = (*pbs)[:n]
				return
			}
			fGet = func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
				pBase = pObj
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
			if ptrDeep > 0 {
				fSet, fGet = getBaseTypeFuncs[[]byte](ptrDeep, fSet, fGet)
			}
		}

	case reflect.Struct:
		son, e := tagParse(baseType)
		if err = e; err != nil {
			return lxterrs.Wrap(err, "Struct")
		}
		// 匿名成员的处理; 这里只能处理费指针嵌入，指针嵌入逻辑在上一层
		if !ti.StructField.Anonymous {
			ti.AddChild(son)
			if ptrDeep > 0 {
				//
				fSet = func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer, err error) {
					p := *(*unsafe.Pointer)(pObj)
					if p == nil {
						v := reflect.New(baseType)
						p = reflectValueToPointer(&v)
						*(*unsafe.Pointer)(pObj) = p
					}
					return unsafe.Pointer(p), nil
				}
				fGet := func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
					p := *(*unsafe.Pointer)(pObj)
					return p, in
				}

				for i := 0; i < ptrDeep; i++ {
					fSet1 := func(pObj unsafe.Pointer, bs []byte) (pBase unsafe.Pointer, err error) {
						var p unsafe.Pointer
						*(**unsafe.Pointer)(pObj) = &p
						return fSet(unsafe.Pointer(&p), bs)
					}
					fGet1 := func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
						p := *(*unsafe.Pointer)(pObj)
						return fGet(p, in)
					}
					fSet, fGet = fSet1, fGet1
				}
			}
		} else {
			if ptrDeep <= 0 {
				for _, c := range son.Children {
					fSet := func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer, err error) {
						pBase = pObj
						pSon := pointerOffset(pObj, ti.Offset)
						return c.fSet(pSon, raw)
					}
					fGet := func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
						pBase = pObj
						pSon := pointerOffset(pObj, ti.Offset)
						return c.fGet(pSon, in)
					}
					c.fSet, c.fGet = fSet, fGet
					ti.AddChild(c)
				}
			} else {
				// 指针匿名嵌入数据结构
				for _, c := range son.Children {
					fSet := func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer, err error) {
						p := *(*unsafe.Pointer)(pObj)
						if p == nil {
							v := reflect.New(baseType)
							p = reflectValueToPointer(&v)
							*(*unsafe.Pointer)(pObj) = p
						}
						return c.fSet(p, raw)
					}
					fGet := func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
						p := *(*unsafe.Pointer)(pObj)
						if p != nil {
							return c.fGet(p, in)
						}
						return nil, in
					}

					for i := 0; i < ptrDeep; i++ {
						fSet1 := func(pObj unsafe.Pointer, bs []byte) (pBase unsafe.Pointer, err error) {
							var p unsafe.Pointer
							*(**unsafe.Pointer)(pObj) = &p
							return fSet(unsafe.Pointer(&p), bs)
						}
						fGet1 := func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
							p := *(*unsafe.Pointer)(pObj)
							return fGet(p, in)
						}
						fSet, fGet = fSet1, fGet1
					}
					c.fSet, c.fGet = fSet, fGet
					ti.AddChild(c)
				}
			}
		}

	case reflect.Interface:
		// Interface 需要根据实际类型创建
	case reflect.Map:
		// Map 要怎么处理？
		if ptrDeep <= 0 {
			fSet = func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer, err error) {
				p := (*map[string]interface{})(pObj)
				*p = make(map[string]interface{})
				return pObj, nil
			}
		} else {
			// 指针匿名嵌入数据结构
			fSet = func(pObj unsafe.Pointer, raw []byte) (pBase unsafe.Pointer, err error) {
				p := (**map[string]interface{})(pObj)
				m := make(map[string]interface{})
				*p = &m
				return unsafe.Pointer(&m), nil
			}
			fGet := func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
				p := *(*unsafe.Pointer)(pObj)
				return p, in
			}

			for i := 0; i < ptrDeep; i++ {
				fSet1 := func(pObj unsafe.Pointer, bs []byte) (pBase unsafe.Pointer, err error) {
					var p unsafe.Pointer
					*(**unsafe.Pointer)(pObj) = &p
					return fSet(unsafe.Pointer(&p), bs)
				}
				fGet1 := func(pObj unsafe.Pointer, in []byte) (pBase unsafe.Pointer, out []byte) {
					p := *(*unsafe.Pointer)(pObj)
					return fGet(p, in)
				}
				fSet, fGet = fSet1, fGet1
			}
		}
	default:
		// Array
		// Interface
		// Map
		// Ptr
		// Slice
		// String,[]byte
		// Struct
		// UnsafePointer
	}
	ti.fSet, ti.fGet = fSet, fGet

	//一些共同的操作
	return
}

func hasBaseElem(typ reflect.Type) bool {
	return typ.Kind() == reflect.Ptr || typ.Kind() == reflect.Map || typ.Kind() == reflect.Slice || typ.Kind() == reflect.Array
}
func baseElem(typ reflect.Type) (typ2 reflect.Type) {
	typ2 = typ
	for hasBaseElem(typ) {
		typ = typ.Elem()
	}
	return
}

//tagParse 解析struct的tag字段，并返回解析的结果
//只需要type, 不需要 interface 也可以? 不着急,分步来
func tagParse(typIn reflect.Type) (ret *TagInfo, err error) {
	if typIn.Kind() != reflect.Struct {
		err = fmt.Errorf("IfaceToHBaseMutation() only accepts structs; got %vFrom", typIn.Kind())
		return
	}
	ret = &TagInfo{
		TagName:  typIn.Name(),
		BaseKind: typIn.Kind(), // 解析出最内层类型
	}

	for i := 0; i < typIn.NumField(); i++ {
		field := typIn.Field(i)
		tagInfo := &TagInfo{
			StructField: field,
			TagName:     field.Name,
			Offset:      field.Offset,
		}

		tagv := field.Tag.Get("json")  //从tag列表中取出下标为i的tag //json:"field,string"
		tagv = strings.TrimSpace(tagv) //去除两头的空格
		if len(tagv) <= 0 || tagv == "-" {
			continue //如果tag字段没有内容，则不处理
		}

		tvs := strings.Split(tagv, ",")
		for i := range tvs {
			tvs[i] = strings.TrimSpace(tvs[i])
		}
		tagInfo.TagName = tvs[0]
		for i := range tvs[1:] {
			if strings.TrimSpace(tvs[i]) == "string" {
				tagInfo.StringTag = true
				continue
			}
			if strings.TrimSpace(tvs[i]) == "omitempty" {
				tagInfo.OmitemptyTag = true
				continue
			}
		}

		err = tagInfo.setFuncs()
		if err != nil {
			err = lxterrs.Wrap(err, "tagInfo.setFuncs")
			return
		}
		if !tagInfo.StructField.Anonymous {
			ret.AddChild(tagInfo)
		}
	}
	return
}

//var allType = make(map[string]map[string]TagInfo, 64)

// emptyInterface is the header for an interface{} value.
type emptyInterface struct {
	typ  uintptr
	word unsafe.Pointer
}

const PANIC = true

func tryPanic(e any) {
	if PANIC {
		panic(e)
	}
}

//Unmarshal 转成struct
func Unmarshal(bs []byte, in interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New("%+v", e)
			return
		}
	}()
	i := trimSpace(bs)

	if _, ok := in.(*map[string]interface{}); ok {
		if bs[i] != '{' {
			err = fmt.Errorf("json must start with '{' or '[', %s", ErrStream(bs[i:]))
			return
		}
		out := make(map[string]interface{})
		parseObjToMap(bs[i+1:], out)
		return
	}
	if _, ok := in.(*[]interface{}); ok {
		if bs[i] != '[' {
			err = fmt.Errorf("json must start with '{' or '[', %s", ErrStream(bs[i:]))
			return
		}
		out := make([]interface{}, 0, 32)
		parseObjToSlice(bs[i+1:], out)
		return
	}

	vi := reflect.ValueOf(in)
	vi = reflect.Indirect(vi)
	if !vi.CanSet() {
		err = fmt.Errorf("%T cannot set", in)
		tryPanic(err)
		return
	}
	typ := vi.Type()
	for typ.Kind() == reflect.Ptr || typ.Kind() == reflect.Slice {
		vi.Set(reflect.New(vi.Type().Elem()))
		vi = vi.Elem()
		typ = typ.Elem()
	}
	n, err := LoadTagNode(typ)
	if err != nil {
		tryPanic(err)
		return
	}

	empty := (*emptyInterface)(unsafe.Pointer(&in))
	err = parseRoot(bs[i:], empty.word, n.tagInfo)
	return
}

type Value struct {
	typ  uintptr
	ptr  unsafe.Pointer
	flag uintptr
}

func reflectValueToPointer(v *reflect.Value) unsafe.Pointer {
	return (*Value)(unsafe.Pointer(v)).ptr
}

func bytesString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
