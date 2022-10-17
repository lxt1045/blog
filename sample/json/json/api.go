package json

import (
	"fmt"
	"reflect"

	lxterrs "github.com/lxt1045/errors"
)

//Unmarshal 转成struct
func Unmarshal(bs []byte, in interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			if err1, ok := e.(*lxterrs.Code); ok {
				err = err1
			} else {
				err = lxterrs.New("%+v", e)
			}
			return
		}
	}()
	i := trimSpace(bs)

	if mIn, ok := in.(*map[string]interface{}); ok {
		if bs[i] != '{' {
			err = fmt.Errorf("json must start with '{' or '[', %s", ErrStream(bs[i:]))
			return
		}
		m, _, _ := parseMapInterface(-1, bs[i+1:])
		*mIn = m
		return nil
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

	vi := reflect.Indirect(reflect.ValueOf(in))
	if !vi.CanSet() {
		err = fmt.Errorf("%T cannot set", in)
		return
	}
	typ := vi.Type()

	prv := reflectValueToValue(&vi)
	goType := prv.typ
	tag, ok := cacheStructTagInfo.Get(goType.Hash)
	if !ok {
		tag, err = LoadTagNodeSlow(typ, goType.Hash)
		if err != nil {
			return
		}
	}
	store := PoolStore{
		tag:  tag,
		obj:  prv.ptr, // eface.Value,
		pool: tag.NewPool(),
	}
	err = parseRoot(bs[i:], store)
	return
}

//Marshal []byte
func Marshal(in interface{}) (bs []byte, err error) {
	defer func() {
		if e := recover(); e != nil {
			if err1, ok := e.(*lxterrs.Code); ok {
				err = err1
			} else {
				err = lxterrs.New("%+v", e)
			}
			return
		}
	}()

	if mIn, ok := in.(*map[string]interface{}); ok {
		_ = mIn
		return
	}
	if _, ok := in.(*[]interface{}); ok {

		return
	}

	vi := reflect.ValueOf(in)
	vi = reflect.Indirect(vi)
	if !vi.CanSet() {
		err = fmt.Errorf("%T cannot set", in)
		return
	}
	typ := vi.Type()

	prv := reflectValueToValue(&vi)
	goType := prv.typ
	tag, ok := cacheStructTagInfo.Get(goType.Hash)
	if !ok {
		tag, err = LoadTagNodeSlow(typ, goType.Hash)
		if err != nil {
			return
		}
	}
	store := Store{
		tag: tag,
		obj: prv.ptr, // eface.Value,
	}
	bs, err = marshalRoot(store)
	return
}
