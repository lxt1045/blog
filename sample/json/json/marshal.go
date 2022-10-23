package json

import (
	"encoding"
	stdjson "encoding/json"
	"reflect"
	"strconv"
	"sync"

	lxterrs "github.com/lxt1045/errors"
)

/*
    fget 函数 list ，打入slice中，然后依次遍历slice，避免了遍历struct 的多个for循环！！ greate！
    golang不支持尾递归，可能嵌套调用性能没有fget list 方式好
*/

var bsPool = sync.Pool{New: func() any {
	s := make([]byte, 0, 1<<20)
	return (*[]byte)(&s)
}}

// 排列一个 fGet list，优化掉多个 for 循环
func marshalStruct(in []byte, store Store) (out []byte) {
	out = append(in, '{')
	for _, tag := range store.tag.ChildList {
		out = append(out, tag.TagName...)
		out = append(out, ':')

		pObj := pointerOffset(store.obj, tag.Offset)
		if tag.fGet != nil {
			out = tag.fGet(Store{obj: pObj, tag: tag}, out)
		} else {
			out = marshalT(out, Store{
				obj: pObj,
				tag: tag,
			})
			// out= marshalValue(out,)
		}
		out = append(out, ',')
	}
	out[len(out)-1] = '}'
	return
}

//marshalT 序列化明确的类型
func marshalT(in []byte, store Store) (out []byte) {
	out = in

	switch store.tag.BaseKind {
	case reflect.Map:
		// TODO: reflect.ValueOf
		m := *(*map[string]interface{})(store.obj)
		out = marshalMapInterface(out, m)
	case reflect.Interface:
		iface := *(*interface{})(store.obj)
		out = marshalInterface(out, iface)
	case reflect.Struct:
		out = marshalStruct(out, Store{
			obj: store.obj,
			tag: store.tag,
		})
	// case reflect.Array:
	case reflect.Slice:
		pHeader := (*SliceHeader)(store.obj)
		son := store.tag.ChildList[0]
		out = marshalSlice(out, Store{obj: pHeader.Data, tag: son}, pHeader.Len)
	}
	return
}
func marshalSlice(bs []byte, store Store, l int) (out []byte) {
	out = append(bs, '[')
	if l <= 0 {
		out = append(out, ']')
		return
	}
	son := store.tag
	size := son.TypeSize
	for i := 0; i < l; i++ {
		pSon := pointerOffset(store.obj, uintptr(i*size))
		if son.fGet != nil {
			out = son.fGet(Store{obj: pSon, tag: son}, out)
		} else {
			out = marshalT(out, Store{
				obj: pSon,
				tag: son,
			})
		}
		out = append(out, ',')
	}
	out[len(out)-1] = ']'
	return
}
func marshalInterface(bs []byte, iface interface{}) (out []byte) {
	if iface == nil {
		out = append(bs, "null"...)
		return
	}
	value := reflect.ValueOf(iface)
	out = marshalValue(bs, value)
	return
}

func marshalMapInterface(bs []byte, m map[string]interface{}) (out []byte) {
	out = bs
	out = append(out, '{')
	for k, v := range m {
		out = append(out, '"')
		out = append(out, k...)
		out = append(out, '"')
		out = append(out, ':')
		out = marshalInterface(out, v)
		out = append(out, ',')
	}
	out[len(out)-1] = '}'
	return
}

func marshalValue(bs []byte, value reflect.Value) (out []byte) {
	out = bs

	// 针对指针
	for value.Kind() == reflect.Ptr {
		if value.IsNil() {
			out = append(out, "null"...)
			return
		}
		value = value.Elem()
	}

	switch value.Kind() {
	case reflect.Interface:
		if value.IsNil() {
			out = append(out, "null"...)
			return
		}
		// UnpackEface(&value)
		// 从 iterface{} 里取出原始类型
		value := reflect.ValueOf(value.Interface())
		out = marshalValue(bs, value)
		return
	case reflect.Map:
		if value.IsNil() {
			out = append(out, "null"...)
			return
		}
		iter := value.MapRange()
		if iter == nil {
			return
		}
		out = append(out, '{')
		l := len(out)
		for iter.Next() {
			out = marshalKey(out, iter.Key())
			out = append(out, ':')
			out = marshalValue(out, iter.Value())
			out = append(out, ',')
		}
		if l < len(out) {
			out = out[:len(out)-1]
		}
		out = append(out, '}')
		return
	case reflect.Slice:
		if value.IsNil() {
			out = append(out, "null"...)
			return
		}
		for i := 0; i < value.Len(); i++ {
			v := value.Index(i)
			out = marshalValue(out, v)
		}
		return
	case reflect.Struct:
		prv := reflectValueToValue(&value)
		goType := prv.typ

		tag, err := LoadTagNode(value, goType.Hash)
		if err != nil {
			return
		}

		store := Store{
			tag: tag,
			obj: prv.ptr, // eface.Value,
		}
		out = marshalStruct(out, store)
		return
	case reflect.Bool:
		if value.Bool() {
			out = append(out, "true"...)
		} else {
			out = append(out, "false"...)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		out = strconv.AppendUint(out, value.Uint(), 10)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		out = strconv.AppendInt(out, value.Int(), 10)
	case reflect.Float64:
		out = strconv.AppendFloat(out, value.Float(), 'f', -1, 64)
	case reflect.Float32:
		out = strconv.AppendFloat(out, value.Float(), 'f', -1, 32)
	case reflect.String:
		if value.Type() == jsonNumberType {
			numStr := value.String()
			// TODO: 检查 numStr 的有效性？
			out = append(out, numStr...)
			return
		}
		out = append(out, '"')
		out = append(out, value.String()...)
		out = append(out, '"')
		return
	default:
		out = append(out, "null"...)
	}

	return
}

var jsonNumberType = reflect.TypeOf(stdjson.Number(""))

func marshalKey(in []byte, k reflect.Value) (out []byte) {
	out = in
	if k.Kind() == reflect.String {
		// key = k.String()
		out = append(out, '"')
		out = append(out, k.String()...)
		out = append(out, '"')
		return
	}
	if tm, ok := k.Interface().(encoding.TextMarshaler); ok {
		if k.Kind() == reflect.Pointer && k.IsNil() {
			return
		}
		bs, err := tm.MarshalText()
		if err != nil {
			err = lxterrs.Wrap(err, "MarshalText() got error")
			panic(err)
		}
		// key = string(bs)
		out = append(out, '"')
		out = append(out, bs...)
		out = append(out, '"')
		return
	}
	switch k.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// key = strconv.FormatInt(k.Int(), 10)
		out = append(out, '"')
		out = strconv.AppendInt(out, k.Int(), 10)
		out = append(out, '"')
		return
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		// key = strconv.FormatUint(k.Uint(), 10)
		out = append(out, '"')
		out = strconv.AppendUint(out, k.Uint(), 10)
		out = append(out, '"')
		return
	}
	err := lxterrs.New("unexpected map key type")
	panic(err)
}
