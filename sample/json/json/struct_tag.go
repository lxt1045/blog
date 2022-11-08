package json

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"unsafe"

	lxterrs "github.com/lxt1045/errors"
)

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

	fUnm unmFunc
	fM   mFunc

	SPool  sync.Pool // TODO：slice pool 和 store.pool 放在一起吧，通过 id 来获取获取 pool，并把剩余的”垃圾“放回 sync.Pool 中共下次复用
	SPoolN int32

	stack        sync.Pool
	stackBuilder *TypeBuilder // dynamicPool 的 builder 用于创建 dynamicPool 对象
	idxCacheST   int32        // 在 store.pool 的 index 文字

	sPooloffset  int32 // slice pool 在 PoolStore的偏移量； TODO
	psPooloffset int32 // pointer slice pool  在 PoolStore的偏移量
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
	t.stack.New = func() any {
		return &dynamicPool{}
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

type ancestor struct {
	hash uint32
	tag  *TagInfo
}

func (ti *TagInfo) setFuncs(typ reflect.Type, anonymous bool, ancestors []ancestor) (err error) {
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
		ti.fUnm, ti.fM = boolMFuncs2(pidx)
	case reflect.Uint, reflect.Uint64, reflect.Uintptr:
		ti.fUnm, ti.fM = uint64MFuncs(pidx)
	case reflect.Int, reflect.Int64:
		ti.fUnm, ti.fM = int64MFuncs(pidx)
	case reflect.Uint32:
		ti.fUnm, ti.fM = uint32MFuncs(pidx)
	case reflect.Int32:
		ti.fUnm, ti.fM = int32MFuncs(pidx)
	case reflect.Uint16:
		ti.fUnm, ti.fM = uint16MFuncs(pidx)
	case reflect.Int16:
		ti.fUnm, ti.fM = int16MFuncs(pidx)
	case reflect.Uint8:
		ti.fUnm, ti.fM = uint8MFuncs(pidx)
	case reflect.Int8:
		ti.fUnm, ti.fM = int8MFuncs(pidx)
	case reflect.Float64:
		ti.fUnm, ti.fM = float64MFuncs(pidx)
	case reflect.Float32:
		ti.fUnm, ti.fM = float32MFuncs(pidx)
	case reflect.String:
		ti.fUnm, ti.fM = stringMFuncs2(pidx)
	case reflect.Slice: // &[]byte; Array
		if isBytes(baseType) {
			ti.fUnm, ti.fM = bytesMFuncs(pidx)
		} else {
			ti.BaseType = baseType
			sliceType := baseType.Elem()
			if isStrings(baseType) {
				ti.fUnm, ti.fM = sliceStringsMFuncs()
			} else if ptrDeep > 0 {
				ti.fUnm, ti.fM = sliceNoscanMFuncs(pidx)
			} else if UnpackType(sliceType).Hash == UnpackType(reflect.TypeOf(int(0))).Hash && ptrDeep == 0 {
				// ti.fUnm, ti.fM = sliceIntsMFuncs(pidx)
				ti.fUnm, ti.fM = sliceNoscanMFuncs(pidx)
			} else if UnpackType(sliceType).PtrData == 0 && ptrDeep == 0 {
				ti.fUnm, ti.fM = sliceNoscanMFuncs(pidx)
			} else {
				ti.fUnm, ti.fM = sliceMFuncs(pidx)
			}
			son := &TagInfo{
				TagName:  `"son"`,
				TypeSize: int(sliceType.Size()),
				Builder:  ti.Builder,
			}

			err = son.setFuncs(sliceType, false /*anonymous*/, ancestors)
			if err != nil {
				return lxterrs.Wrap(err, "Struct")
			}
			err = ti.AddChild(son)
			if err != nil {
				return lxterrs.Wrap(err, "Struct")
			}

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
		ti.fUnm, ti.fM = structMFuncs2(pidx)

		son, e := NewStructTagInfo(baseType, ti.stackBuilder, ti.Builder, ancestors)
		if err = e; err != nil {
			return lxterrs.Wrap(err, "Struct")
		}
		if son == nil {
			// TODO: fUnm中需要重建 store.pool，并从
			// tag, err := LoadTagNode(vi, goType.Hash) 获取 tag，需要延后处理？
			ti.fUnm, ti.fM = structMFuncs2(pidx)
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
		} else {
			for _, c := range son.ChildList {
				if ptrDeep == 0 {
					c.Offset += ti.Offset
				} else {
					fUnm, fM := c.fUnm, c.fM
					c.fM = func(store Store, in []byte) (out []byte) {
						store.obj = *(*unsafe.Pointer)(store.obj)
						if store.obj != nil {
							return fM(store, in)
						}
						out = append(in, "null"...)
						return
					}
					c.fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
						store.obj = *(*unsafe.Pointer)(store.obj)
						if store.obj == nil {
							store.obj = store.Idx(*pidx)
						}
						return fUnm(idxSlash, store, stream)
					}
				}
				err = ti.AddChild(c)
				if err != nil {
					return lxterrs.Wrap(err, "AddChild")
				}
			}
		}
	case reflect.Interface:
		// Interface 需要根据实际类型创建
		ti.fUnm, ti.fM = interfaceMFuncs(pidx)

	case reflect.Map:
		ti.fUnm, ti.fM = mapMFuncs(pidx)
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
		err = son.setFuncs(valueType, false /*anonymous*/, ancestors)
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
		fUnm, fM := ti.fUnm, ti.fM
		ti.fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			store.obj = store.Idx(*idxP)
			return fUnm(idxSlash, store, stream)
		}
		ti.fM = func(store Store, in []byte) (out []byte) {
			store.obj = *(*unsafe.Pointer)(store.obj)
			return fM(store, in)
		}
	}

	return
}

//NewStructTagInfo 解析struct的tag字段，并返回解析的结果
/*
   每个 struct 都搞一个 pointerCacheType，在使用的时候直接获取； 再搞一个 slicePool 在是 slice 时使用；
   二者不会同一时刻出现，是不是可以合并为同一个值？

*/
func NewStructTagInfo(typIn reflect.Type, stackBuilder, builder *TypeBuilder, ancestors []ancestor) (ti *TagInfo, err error) {
	if typIn.Kind() != reflect.Struct {
		err = lxterrs.New("NewStructTagInfo only accepts structs; got %v", typIn.Kind())
		return
	}
	if builder == nil {
		builder = NewTypeBuilder()
	}
	if stackBuilder == nil {
		stackBuilder = NewTypeBuilder()
	}

	ti = &TagInfo{
		TagName:  typIn.String(),
		BaseKind: typIn.Kind(), // 解析出最内层类型
		Builder:  builder,
		TypeSize: int(typIn.Size()),
	}

	goType := UnpackType(typIn)
	for _, a := range ancestors {
		if a.hash == goType.Hash {
			return // 避免嵌套循环
		}
	}
	ancestors = append(ancestors, ancestor{goType.Hash, ti})
	if builder == nil {
		builder = NewTypeBuilder()
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

		err = son.setFuncs(field.Type, field.Anonymous, ancestors)
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
	return
}
