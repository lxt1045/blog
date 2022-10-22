package json

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"unsafe"

	lxterrs "github.com/lxt1045/errors"
)

type Store struct {
	obj unsafe.Pointer
	tag *TagInfo
}
type PoolStore struct {
	obj  unsafe.Pointer
	pool unsafe.Pointer
	tag  *TagInfo
}

//TagInfo 拥有tag的struct的成员的解析结果
type TagInfo struct {
	Offset          uintptr //偏移量
	Type            reflect.Type
	TypeSize        int
	BaseType        reflect.Type
	BaseKind        reflect.Kind // 次成员可能是 **string,[]int 等这种复杂类型,这个 用来指示 "最里层" 的类型
	JSONType        Type         // 次成员可能是 **string,[]int 等这种复杂类型,这个 用来指示 "最里层" 的类型
	TagName         string       //
	StringTag       bool         // `json:"field,string"`: 此情形下,需要把struct的int转成json的string
	OmitemptyTag    bool         //  `json:"some_field,omitempty"`
	Anonymous       bool
	Children        map[string]*TagInfo
	ChildList       []*TagInfo // 遍历的顺序和速度
	MChildren       tagMap
	MChildrenEnable bool

	fSet setFunc
	fGet getFunc

	fUnm   unmFunc
	fM     mFunc
	hasPtr bool // TODO： 是否有指针，如果没有指针，则不需要用 reflect.New reflect.MakeSlice,可以更快速的分配空间
	Pool   *SlicePool
	fMake  func(l int) unsafe.Pointer // make slice/map
	MakeN  int
	SPool  sync.Pool

	Builder     *TypeBuilder
	PointerPool sync.Pool
}

func (t *TagInfo) buildChildMap() {
	if len(t.ChildList) == 0 {
		return
	}
	nodes := make([]mapNode, 0, len(t.ChildList))
	for _, child := range t.ChildList {
		if len(child.TagName) == 0 {
			continue
		}
		nodes = append(nodes, mapNode{
			K: []byte(child.TagName),
			V: child,
		})
	}
	if len(t.ChildList) <= 128 {
		t.MChildrenEnable = true
		mc := buildTagMap(nodes)
		t.MChildren = mc
	}
}

func (t *TagInfo) GetChildFromMap(key []byte) *TagInfo {
	return t.Children[string(key)]
}

func (t *TagInfo) AddChild(c *TagInfo) (err error) {
	if len(t.Children) == 0 {
		t.Children = make(map[string]*TagInfo)
	}
	if _, ok := t.Children[c.TagName]; ok {
		err = fmt.Errorf("error, tag[%s]类型配置出错,字段重复", c.TagName)
		return
	}
	t.ChildList = append(t.ChildList, c)
	t.Children[c.TagName] = c
	return
}

// []byte 是一种特殊的底层数据类型，需要 base64 编码
func isBytes(typ reflect.Type) bool {
	bsType := reflect.TypeOf(&[]byte{})
	return typ.PkgPath() == bsType.PkgPath() && typ.String() == bsType.String()
}

func (ti *TagInfo) setFuncs(typ reflect.Type) (err error) {
	ptrDeep, baseType := 0, typ
	for ; ; typ = typ.Elem() {
		if typ.Kind() == reflect.Ptr {
			ptrDeep++
			continue
		}
		baseType = typ
		break
	}

	// 先从最后一个基础类型开始处理
	switch baseType.Kind() {
	case reflect.Bool:
		ti.JSONType = Bool
		ti.fSet, ti.fGet = boolFuncs(ti.Builder, ti.TagName, ptrDeep)
		ti.fUnm, ti.fM = boolMFuncs()
	case reflect.Uint, reflect.Uint64, reflect.Uintptr:
		ti.JSONType = Number
		ti.fSet, ti.fGet = uint64Funcs(ti.Builder, ti.TagName, ptrDeep)
		ti.fUnm, ti.fM = numMFuncs()
	case reflect.Int, reflect.Int64:
		ti.JSONType = Number
		ti.fSet, ti.fGet = int64Funcs(ti.Builder, ti.TagName, ptrDeep)
		ti.fUnm, ti.fM = numMFuncs()
	case reflect.Uint32:
		ti.JSONType = Number
		ti.fSet, ti.fGet = uint32Funcs(ti.Builder, ti.TagName, ptrDeep)
		ti.fUnm, ti.fM = numMFuncs()
	case reflect.Int32:
		ti.JSONType = Number
		ti.fSet, ti.fGet = int32Funcs(ti.Builder, ti.TagName, ptrDeep)
		ti.fUnm, ti.fM = numMFuncs()
	case reflect.Uint16:
		ti.JSONType = Number
		ti.fSet, ti.fGet = uint16Funcs(ti.Builder, ti.TagName, ptrDeep)
		ti.fUnm, ti.fM = numMFuncs()
	case reflect.Int16:
		ti.JSONType = Number
		ti.fSet, ti.fGet = int16Funcs(ti.Builder, ti.TagName, ptrDeep)
		ti.fUnm, ti.fM = numMFuncs()
	case reflect.Uint8:
		ti.JSONType = Number
		ti.fSet, ti.fGet = uint8Funcs(ti.Builder, ti.TagName, ptrDeep)
		ti.fUnm, ti.fM = numMFuncs()
	case reflect.Int8:
		ti.JSONType = Number
		ti.fSet, ti.fGet = int8Funcs(ti.Builder, ti.TagName, ptrDeep)
		ti.fUnm, ti.fM = numMFuncs()
	case reflect.Float64:
		ti.JSONType = Number
		ti.fSet, ti.fGet = float64Funcs(ti.Builder, ti.TagName, ptrDeep)
		ti.fUnm, ti.fM = numMFuncs()
	case reflect.Float32:
		ti.JSONType = Number
		ti.fSet, ti.fGet = float32Funcs(ti.Builder, ti.TagName, ptrDeep)
		ti.fUnm, ti.fM = numMFuncs()
	case reflect.String:
		ti.JSONType = String
		ti.fSet, ti.fGet = stringFuncs(ti.Builder, ti.TagName, ptrDeep)
		ti.fUnm, ti.fM = stringMFuncs()
		// ti.fUnm, ti.fM = stringMFuncs_0(ti.Builder, ti.TagName, ptrDeep)
	case reflect.Slice: // &[]byte; Array
		ti.fUnm, ti.fM = sliceMFuncs()
		ti.MakeN = 4
		if isBytes(baseType) {
			ti.JSONType = Bytes
			ti.fSet, ti.fGet = bytesFuncs(ti.Builder, ti.TagName, ptrDeep)
		} else {
			ti.JSONType = Slice
			ti.BaseType = baseType
			sliceType := baseType.Elem()
			son := &TagInfo{
				TagName:  `"son"`,
				Type:     sliceType,
				TypeSize: int(sliceType.Size()),
				Builder:  ti.Builder,
			}
			err = son.setFuncs(sliceType)
			if err != nil {
				return lxterrs.Wrap(err, "Struct")
			}
			err = ti.AddChild(son)
			if err != nil {
				return lxterrs.Wrap(err, "Struct")
			}
			ti.fSet, ti.fGet = sliceFuncs(ti.Builder, ti.TagName, ptrDeep)

			if ptrDeep > 0 || son.hasPtr {
				ti.hasPtr = true
			}
			if !son.hasPtr {
			} else {
			}
			ti.Pool = NewSlicePool(ti.BaseType, son.Type)
			ti.SPool.New = func() any {
				v := reflect.MakeSlice(ti.BaseType, 0, 1024)
				p := reflectValueToPointer(&v)
				pH := (*SliceHeader)(p)
				pH.Cap = pH.Cap * int(son.Type.Size())
				return (*[]uint8)(p)
			}
		}
	case reflect.Struct:
		ti.fUnm, ti.fM = structMFuncs()
		ti.JSONType = Struct
		son, e := NewStructTagInfo(baseType, false, ti.Builder)
		if err = e; err != nil {
			return lxterrs.Wrap(err, "Struct")
		}
		// 匿名成员的处理; 这里只能处理费指针嵌入，指针嵌入逻辑在上一层
		if !ti.Anonymous {
			for _, c := range son.ChildList {
				err = ti.AddChild(c)
				if err != nil {
					return lxterrs.Wrap(err, "AddChild")
				}
			}
			ti.buildChildMap()
			ti.fSet, ti.fGet = structChildFuncs(ti.Builder, ti.TagName, ptrDeep, baseType)
		} else {
			for _, c := range son.ChildList {
				c.fSet, c.fGet = anonymousStructFuncs(
					ti.Builder, ti.TagName, ptrDeep, baseType, ti.Offset, c.fSet, c.fGet)
				err = ti.AddChild(c)
				if err != nil {
					return lxterrs.Wrap(err, "AddChild")
				}
			}
		}
	case reflect.Interface:
		ti.JSONType = Interface
		// Interface 需要根据实际类型创建
		ti.fSet, ti.fGet = iterfaceFuncs(ti.Builder, ti.TagName, ptrDeep)
		ti.fUnm, ti.fM = interfaceMFuncs()
	case reflect.Map:
		ti.JSONType = Map
		ti.fSet, ti.fGet = mapFuncs(ti.Builder, ti.TagName, ptrDeep)
		valueType := baseType.Elem()
		son := &TagInfo{
			TagName:  `"son"`,
			Type:     valueType,
			Builder:  ti.Builder,
			TypeSize: int(valueType.Size()),
		}
		err = ti.AddChild(son)
		if err != nil {
			return lxterrs.Wrap(err, "Struct")
		}
		err = son.setFuncs(valueType)
		if err != nil {
			return lxterrs.Wrap(err, "Struct")
		}
	default:
		return lxterrs.New("errors type:%s", baseType)
	}

	//一些共同的操作
	return
}

//NewStructTagInfo 解析struct的tag字段，并返回解析的结果
//只需要type, 不需要 interface 也可以? 不着急,分步来
func NewStructTagInfo(typIn reflect.Type, noBuildmap bool, builder *TypeBuilder) (ti *TagInfo, err error) {
	if typIn.Kind() != reflect.Struct {
		err = lxterrs.New("NewStructTagInfo only accepts structs; got %v", typIn.Kind())
		return
	}
	if builder == nil {
		builder = NewTypeBuilder()
	}
	ti = &TagInfo{
		TagName:  typIn.String(),
		BaseKind: typIn.Kind(), // 解析出最内层类型
		JSONType: Struct,
		Builder:  builder,
		TypeSize: int(typIn.Size()),
	}

	for i := 0; i < typIn.NumField(); i++ {
		field := typIn.Field(i)
		son := &TagInfo{
			Type:      field.Type,
			BaseType:  field.Type,
			TagName:   `"` + field.Name + `"`,
			Offset:    field.Offset,
			BaseKind:  field.Type.Kind(),
			Anonymous: field.Anonymous,
			Builder:   builder,
			TypeSize:  int(field.Type.Size()),
		}

		if !field.IsExported() {
			continue // 非导出成员不处理
		}

		tagv := field.Tag.Get("json")  //从tag列表中取出下标为i的tag //json:"field,string"
		tagv = strings.TrimSpace(tagv) //去除两头的空格
		if len(tagv) > 0 && tagv == "-" {
			continue //如果tag字段没有内容，则不处理
		}
		if len(tagv) == 0 {
			tagv = field.Name // 没有 tag 则以成员名为 tag
		}

		tvs := strings.Split(tagv, ",")
		for i := range tvs {
			tvs[i] = strings.TrimSpace(tvs[i])
		}
		son.TagName = `"` + tvs[0] + `"` // 此处加上 双引号 是为了方便使用 改进后的 hash map
		for i := 1; i < len(tvs); i++ {
			if strings.TrimSpace(tvs[i]) == "string" {
				son.StringTag = true
				continue
			}
			if strings.TrimSpace(tvs[i]) == "omitempty" {
				son.OmitemptyTag = true
				continue
			}
		}

		err = son.setFuncs(field.Type)
		if err != nil {
			err = lxterrs.Wrap(err, "son.setFuncs")
			return
		}
		if !son.Anonymous {
			err = ti.AddChild(son)
			if err != nil {
				return
			}
		} else {
			// 如果是匿名成员类型，需要将其子成员插入为父节点的子成员；
			// 此外，get set 函数也要做相应修改
			for _, c := range son.ChildList {
				err = ti.AddChild(c)
				if err != nil {
					return
				}
			}
		}
	}
	if !noBuildmap {
		ti.buildChildMap()
	}
	return
}
