package json

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	"github.com/lxt1045/errors"
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
	ti, err := tagParse(typ, "json")
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
	ti, err := tagParse(typ, "json")
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
	tagInfo  map[string]*TagInfo
	pkgCache cache[string, *tagNodeStr] //如果 name 相等，则从这个缓存中获取
}

type tagNode struct {
	pkgPath  uintptr
	tagInfo  map[string]*TagInfo
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

	fSet func(field reflect.StructField, pStruct unsafe.Pointer, pIn unsafe.Pointer)
	fGet func(field reflect.StructField, pStruct unsafe.Pointer, pOut unsafe.Pointer)
}

func (p *TagInfo) cacheKey() (k string) {
	return p.TagName
}
func (t *TagInfo) GetChild(key []byte) *TagInfo {
	return t.Children[string(key)]
}

func (t *TagInfo) Set(pStruct unsafe.Pointer, pIn unsafe.Pointer) {
	switch t.BaseKind {
	case reflect.String:
		// setFieldString(t.StructField, pStruct, pIn)
	case reflect.Int:
		setFieldInt(t.StructField, pStruct, pIn)
	case reflect.Bool:
		setFieldBool(t.StructField, pStruct, pIn)
	default:
		setField(t.StructField, pStruct, pIn)
	}
}
func (t *TagInfo) Get(pStruct unsafe.Pointer, pOut unsafe.Pointer) {
	switch t.BaseKind {
	case reflect.String:
		getField(t.StructField, pStruct, pOut)
	default:
		getField(t.StructField, pStruct, pOut)
	}
}
func (tag *TagInfo) Store(tis map[string]*TagInfo) {
	if _, ok := tis[tag.cacheKey()]; ok {
		err := fmt.Errorf("error, tag[%s]类型配置出错,字段重复", tag.TagName)
		panic(err)
	}
	tis[tag.cacheKey()] = tag
	return
}

func hasBaseElem(typ reflect.Type) bool {
	return typ.Kind() == reflect.Ptr || typ.Kind() == reflect.Map || typ.Kind() == reflect.Slice || typ.Kind() == reflect.Array
}
func baseElem(typ reflect.Type) reflect.Type {
	for hasBaseElem(typ) {
		typ = typ.Elem()
	}
	return typ
}

//tagParse 解析struct的tag字段，并返回解析的结果
//只需要type, 不需要 interface 也可以? 不着急,分步来
func tagParse(typIn reflect.Type, tagKey string) (tis map[string]*TagInfo, err error) {
	if typIn.Kind() != reflect.Struct {
		err = fmt.Errorf("IfaceToHBaseMutation() only accepts structs; got %vFrom", typIn.Kind())
		return
	}
	tis = make(map[string]*TagInfo)

	for i := 0; i < typIn.NumField(); i++ {
		field := typIn.Field(i)
		baseType := baseElem(field.Type)
		if field.Anonymous { //匿名类型
			if baseType.Kind() == reflect.Struct {
				children, e := tagParse(baseType, tagKey)
				if err = e; err != nil {
					return
				}
				for key, ti := range children {
					if field.Type.Kind() == reflect.Ptr {
						fSet, fGet := ti.fSet, ti.fSet
						ti.fSet = func(field reflect.StructField, pStruct, pIn unsafe.Pointer) {
							// TODO
							if fSet != nil {
								fSet(field, pStruct, pIn)
							}
						}
						ti.fGet = func(field reflect.StructField, pStruct, pOut unsafe.Pointer) {
							// TODO
							if fGet != nil {
								fGet(field, pStruct, pOut)
							}
						}
					} else {
						ti.Offset += field.Offset
					}
					tis[key] = ti
				}
			}
			continue
		}
		tagInfo := &TagInfo{
			StructField: field,
			TagName:     field.Name,
			Offset:      field.Offset,
			BaseKind:    baseType.Kind(), // 解析出最内层类型
		}

		tagv := field.Tag.Get(tagKey)  //从tag列表中取出下标为i的tag //json:"field,string"
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
		if baseType.Kind() == reflect.Struct {
			children, e := tagParse(baseType, tagKey)
			if err = e; err != nil {
				return
			}
			tagInfo.Children = children
		}
		tagInfo.Store(tis)
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
	switch bs[i] {
	case '{':
		parseObj(bs[i+1:], empty.word, n.tagInfo)
	case '[':
	default:
		panicIncorrectFormat(bs[i+1:])
		err = fmt.Errorf("json must start with '{' or '[', %s", ErrStream(bs[i:]))
		return
	}

	return
}
