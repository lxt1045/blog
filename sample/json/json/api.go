package json

import (
	"fmt"
	"reflect"
	"sync/atomic"

	lxterrs "github.com/lxt1045/errors"
)

/*
来一个更狠的、可以用于炫技的，自己 mask pointer 的 bitmap，
这样就可以不需要建那么多的 pool，只要一个 nopointer 一个 pointer 的 pool 就可以了
*/

//Unmarshal 转成struct
func Unmarshal(bsIn []byte, in interface{}) (err error) {
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
	bs := bytesString(bsIn)
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

	pool := tag.stack.Get().(*dynamicPool)
	pool.structPool = tag.Builder.NewFromPool()
	store := PoolStore{
		tag:  tag,
		obj:  prv.ptr, // eface.Value,
		pool: pool,
	}
	err = parseRoot(bs[i:], store)
	pool.structPool = nil
	tag.stack.Put(pool)
	return
}

//UnmarshalString Unmarshal string
func UnmarshalString(bs string, in interface{}) (err error) {
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
	pool := tag.stack.Get().(*dynamicPool)
	pool.structPool = tag.Builder.NewFromPool()
	store := PoolStore{
		tag:  tag,
		obj:  prv.ptr, // eface.Value,
		pool: pool,
	}
	err = parseRoot(bs[i:], store)
	pool.structPool = nil
	tag.stack.Put(pool)
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

	pbs := bsPool.Get().(*[]byte)
	bs = *pbs
	var lLeft int32 = 1024
	defer func() {
		if cap(bs)-len(bs) >= int(lLeft) {
			*pbs = bs[len(bs):]
			bs = bs[:len(bs):len(bs)]
			bsPool.Put(pbs)
		}
	}()

	if mIn, ok := in.(*map[string]interface{}); ok {
		bs = marshalMapInterface(bs[:0], *mIn)
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

	bs = marshalStruct(store, bs[:0])

	l := int32(len(bs))
	lLeft = atomic.LoadInt32(&tag.bsMarshalLen)
	if lLeft > l*2 {
		bsHaftCount := atomic.AddInt32(&tag.bsHaftCount, -1)
		if bsHaftCount < 1000 {
			atomic.StoreInt32(&tag.bsMarshalLen, l)
			lLeft = l
		}
	} else if lLeft < l {
		atomic.StoreInt32(&tag.bsMarshalLen, l)
		lLeft = l
	} else {
		atomic.AddInt32(&tag.bsHaftCount, 1)
	}
	return
}
