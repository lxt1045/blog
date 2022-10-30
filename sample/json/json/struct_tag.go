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

func (ps PoolStore) Idx(idx uintptr) (p unsafe.Pointer) {
	p = pointerOffset(ps.pool, idx)
	*(*unsafe.Pointer)(ps.obj) = p
	return
}

//TagInfo 拥有tag的struct的成员的解析结果
type TagInfo struct {
	TagName      string       //
	BaseType     reflect.Type //
	BaseKind     reflect.Kind // 次成员可能是 **string,[]int 等这种复杂类型,这个 用来指示 "最里层" 的类型
	Offset       uintptr      //偏移量
	TypeSize     int          //
	StringTag    bool         // `json:"field,string"`: 此情形下,需要把struct的int转成json的string
	OmitemptyTag bool         //  `json:"some_field,omitempty"`

	/*
		MChildrenEnable: true 时表示使用 MChildren
		Children： son 超过 128 时tagMap解析很慢，用 map 替代
		ChildList： 遍历 map 性能较差，加个 list
	*/
	MChildrenEnable bool
	Children        map[string]*TagInfo
	ChildList       []*TagInfo // 遍历的顺序和速度
	MChildren       tagMap
	Builder         *TypeBuilder

	fSet setFunc
	fGet getFunc
	fUnm unmFunc
	fM   mFunc

	SPool        sync.Pool
	SPoolN       int32
	bsMarshalLen int32 // 缓存上次 生成的 bs 的大小，如果 cache 小于这个值，则丢弃
	bsHaftCount  int32 // 记录上次低于 bsMarshalLen/2 的次数
}

const SPoolN = 1024 // * 1024

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
			K: child.TagName,
			V: child,
		})
	}
	if len(t.ChildList) <= 64 {
		t.MChildrenEnable = true
		mc := buildTagMap(nodes)
		t.MChildren = mc
		t.Children = nil // 减少 gc 扫描指针
	}
}

func (t *TagInfo) GetChildFromMap(key string) *TagInfo {
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
	return UnpackType(bsType).Hash == UnpackType(typ).Hash
}
func isStrings(typ reflect.Type) bool {
	bsType := reflect.TypeOf([]string{})
	return UnpackType(bsType).Hash == UnpackType(typ).Hash
}

func (ti *TagInfo) setFuncs(typ reflect.Type, anonymous bool) (err error) {
	ptrDeep, baseType := 0, typ
	var pidx *uintptr
	for ; ; typ = typ.Elem() {
		if typ.Kind() == reflect.Ptr {
			ptrDeep++
			continue
		}
		baseType = typ
		break
	}
	if ptrDeep > 0 {
		pidx = &[]uintptr{0}[0]
		ti.Builder.AppendTagField(ti.TagName, baseType, pidx)
	}

	// 先从最后一个基础类型开始处理
	switch baseType.Kind() {
	case reflect.Bool:
		ti.fSet, ti.fGet = boolFuncs(pidx)
		ti.fUnm, ti.fM = boolMFuncs()
	case reflect.Uint, reflect.Uint64, reflect.Uintptr:
		ti.fSet, ti.fGet = uint64Funcs(pidx)
		ti.fUnm, ti.fM = numMFuncs()
	case reflect.Int, reflect.Int64:
		ti.fSet, ti.fGet = int64Funcs(pidx)
		ti.fUnm, ti.fM = numMFuncs()
	case reflect.Uint32:
		ti.fSet, ti.fGet = uint32Funcs(pidx)
		ti.fUnm, ti.fM = numMFuncs()
	case reflect.Int32:
		ti.fSet, ti.fGet = int32Funcs(pidx)
		ti.fUnm, ti.fM = numMFuncs()
	case reflect.Uint16:
		ti.fSet, ti.fGet = uint16Funcs(pidx)
		ti.fUnm, ti.fM = numMFuncs()
	case reflect.Int16:
		ti.fSet, ti.fGet = int16Funcs(pidx)
		ti.fUnm, ti.fM = numMFuncs()
	case reflect.Uint8:
		ti.fSet, ti.fGet = uint8Funcs(pidx)
		ti.fUnm, ti.fM = numMFuncs()
	case reflect.Int8:
		ti.fSet, ti.fGet = int8Funcs(pidx)
		ti.fUnm, ti.fM = numMFuncs()
	case reflect.Float64:
		ti.fSet, ti.fGet = float64Funcs(pidx)
		ti.fUnm, ti.fM = numMFuncs()
	case reflect.Float32:
		ti.fSet, ti.fGet = float32Funcs(pidx)
		ti.fUnm, ti.fM = numMFuncs()
	case reflect.String:
		ti.fSet, ti.fGet = stringFuncs(pidx)
		ti.fUnm, ti.fM = stringMFuncs()
	case reflect.Slice: // &[]byte; Array
		ti.fUnm, ti.fM = sliceMFuncs()
		if isBytes(baseType) {
			ti.fSet, ti.fGet = bytesFuncs(pidx)
		} else {
			if isStrings(baseType) {
				ti.fUnm, ti.fM = sliceStringsMFuncs()
			}
			ti.BaseType = baseType
			sliceType := baseType.Elem()
			son := &TagInfo{
				TagName:  `"son"`,
				TypeSize: int(sliceType.Size()),
				Builder:  ti.Builder,
			}
			err = son.setFuncs(sliceType, false /*anonymous*/)
			if err != nil {
				return lxterrs.Wrap(err, "Struct")
			}
			err = ti.AddChild(son)
			if err != nil {
				return lxterrs.Wrap(err, "Struct")
			}
			ti.fSet, ti.fGet = sliceFuncs(pidx)

			// ch := func() (ch chan *[]byte) {
			// 	ch = make(chan *[]uint8, 4)
			// 	go func() {
			// 		for {
			// 			v := reflect.MakeSlice(ti.BaseType, 0, SPoolN)
			// 			p := reflectValueToPointer(&v)
			// 			pH := (*SliceHeader)(p)
			// 			pH.Cap = pH.Cap * int(sliceType.Size())
			// 			ch <- (*[]uint8)(p)
			// 		}
			// 	}()
			// 	return
			// }()
			ti.SPoolN = (1 << 20) / int32(ti.BaseType.Size())
			ti.SPool.New = func() any {
				// return <-ch
				v := reflect.MakeSlice(ti.BaseType, 0, int(ti.SPoolN)) // SPoolN)
				p := reflectValueToPointer(&v)
				pH := (*SliceHeader)(p)
				pH.Cap = pH.Cap * int(sliceType.Size())
				return (*[]uint8)(p)
			}
		}
	case reflect.Struct:
		ti.fUnm, ti.fM = structMFuncs()
		son, e := NewStructTagInfo(baseType, false, ti.Builder)
		if err = e; err != nil {
			return lxterrs.Wrap(err, "Struct")
		}
		// 匿名成员的处理; 这里只能处理费指针嵌入，指针嵌入逻辑在上一层
		if !anonymous {
			for _, c := range son.ChildList {
				err = ti.AddChild(c)
				if err != nil {
					return lxterrs.Wrap(err, "AddChild")
				}
			}
			ti.buildChildMap()
			ti.fSet, ti.fGet = structChildFuncs(pidx)
		} else {
			for _, c := range son.ChildList {
				c.fSet, c.fGet = anonymousStructFuncs(pidx, ti.Offset, c.fSet, c.fGet)
				err = ti.AddChild(c)
				if err != nil {
					return lxterrs.Wrap(err, "AddChild")
				}
			}
		}
	case reflect.Interface:
		// Interface 需要根据实际类型创建
		ti.fSet, ti.fGet = iterfaceFuncs(pidx)
		ti.fUnm, ti.fM = interfaceMFuncs()
	case reflect.Map:
		ti.fSet, ti.fGet = mapFuncs(pidx)
		ti.fUnm, ti.fM = mapMFuncs()
		valueType := baseType.Elem()
		son := &TagInfo{
			TagName:  `"son"`,
			Builder:  ti.Builder,
			TypeSize: int(valueType.Size()), // TODO
		}
		err = ti.AddChild(son)
		if err != nil {
			return lxterrs.Wrap(err, "Struct")
		}
		err = son.setFuncs(valueType, false /*anonymous*/)
		if err != nil {
			return lxterrs.Wrap(err, "Struct")
		}
	default:
		return lxterrs.New("errors type:%s", baseType)
	}

	// 处理一下指针
	for i := 1; i < ptrDeep; i++ {
		var idxP *uintptr = &[]uintptr{0}[0]
		ti.Builder.AppendPointer(fmt.Sprintf("%s_%d", ti.TagName, i), idxP)
		fSet, fGet := ti.fSet, ti.fGet
		ti.fSet = func(store PoolStore, bs string) (pBase unsafe.Pointer) {
			p := pointerOffset(store.pool, *idxP)
			*(**unsafe.Pointer)(store.obj) = (*unsafe.Pointer)(p)
			store.obj = p
			return fSet(store, bs)
		}
		ti.fGet = func(store Store, in []byte) (out []byte) {
			store.obj = *(*unsafe.Pointer)(store.obj)
			return fGet(store, in)
		}
	}
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
		Builder:  builder,
		TypeSize: int(typIn.Size()),
	}

	for i := 0; i < typIn.NumField(); i++ {
		field := typIn.Field(i)
		son := &TagInfo{
			BaseType: field.Type,
			TagName:  `"` + field.Name + `"`,
			Offset:   field.Offset,
			BaseKind: field.Type.Kind(),
			Builder:  builder,
			TypeSize: int(field.Type.Size()),
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

		err = son.setFuncs(field.Type, field.Anonymous)
		if err != nil {
			err = lxterrs.Wrap(err, "son.setFuncs")
			return
		}
		if !field.Anonymous {
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
