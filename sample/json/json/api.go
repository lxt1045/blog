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
	prv := reflectValueToValue(&vi)
	goType := prv.typ
	tag, err := LoadTagNode(vi, goType.Hash)
	if err != nil {
		return
	}

	store := PoolStore{
		tag:  tag,
		obj:  prv.ptr, // eface.Value,
		pool: tag.Builder.NewFromPool(),
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

	vi := reflect.Indirect(reflect.ValueOf(in))
	if !vi.CanSet() {
		err = fmt.Errorf("%T cannot set", in)
		return
	}
	prv := reflectValueToValue(&vi)
	goType := prv.typ
	tag, err := LoadTagNode(vi, goType.Hash)
	if err != nil {
		return
	}

	store := Store{
		tag: tag,
		obj: prv.ptr, // eface.Value,
	}

	pbs := bsPool.Get().(*[]byte)
	if cap(*pbs) < 128 {
		pbs = bsPool.New().(*[]byte)
	}
	bs = *pbs
	defer func() {
		*pbs = bs[len(bs):]
		bs = bs[:len(bs):len(bs)]
		bsPool.Put(pbs)
	}()

	bs = marshalObj(bs[:0], store)
	return
}
