package json

import (
	"fmt"
	"log"
	"reflect"
	"unsafe"
)

func init() {
	log.SetFlags(log.Flags() | log.Lmicroseconds | log.Lshortfile) //log.Llongfile
}

// 构造器
type TypeBuilder struct {
	// 用于存储属性字段
	fields      []reflect.StructField
	Type        reflect.Type
	lazyOffsets []*uintptr

	goType *GoType
}

func NewTypeBuilder() *TypeBuilder {
	return &TypeBuilder{}
}

// 根据预先添加的字段构建出结构体
func (b *TypeBuilder) New() unsafe.Pointer {
	v := reflect.New(b.Type)
	p := reflectValueToPointer(&v)
	return p
}

// 根据预先添加的字段构建出结构体
func (b *TypeBuilder) NewSlice() unsafe.Pointer {
	p := unsafe_NewArray(b.goType, 1024)
	return p
}

func (b *TypeBuilder) Interface() interface{} {
	v := reflect.New(b.Type)
	return v.Interface()
}
func (b *TypeBuilder) PInterface() (unsafe.Pointer, interface{}) {
	v := reflect.New(b.Type)
	return reflectValueToPointer(&v), v.Interface()
}

// 根据预先添加的字段构建出结构体
func (b *TypeBuilder) Build() reflect.Type {
	typ := b.Type
	if b.Type == nil {
		typ = reflect.StructOf(b.fields)
		b.Type = typ
		b.goType = UnpackType(typ)
	}
	for i := 0; i < typ.NumField(); i++ {
		if len(b.lazyOffsets) > i && b.lazyOffsets[i] != nil {
			*b.lazyOffsets[i] = typ.Field(i).Offset
		}
	}
	return typ
}

/*
针对 slice，要添加一个 [4]type 的空间作为预分配的资源
*/
func (b *TypeBuilder) AppendTagField(name string, typ reflect.Type, lazyOffset *uintptr) *TypeBuilder {
	name = "X_" + name[1:len(name)-1]
	b.fields = append(b.fields, reflect.StructField{Name: name, Type: typ})
	b.lazyOffsets = append(b.lazyOffsets, lazyOffset)
	if typ.Kind() != reflect.Slice {
		return b
	}

	arrayType := reflect.ArrayOf(4, typ.Elem())
	b.fields = append(b.fields, reflect.StructField{Name: "Array_" + name, Type: arrayType})
	b.lazyOffsets = append(b.lazyOffsets, nil)
	return b
}

func (b *TypeBuilder) AppendField(name string, typ reflect.Type, lazyOffset *uintptr) *TypeBuilder {
	b.fields = append(b.fields, reflect.StructField{Name: name, Type: typ})
	b.lazyOffsets = append(b.lazyOffsets, lazyOffset)
	if typ.Kind() != reflect.Slice {
		return b
	}

	arrayType := reflect.ArrayOf(4, typ.Elem())
	b.fields = append(b.fields, reflect.StructField{Name: "Array_" + name, Type: arrayType})
	b.lazyOffsets = append(b.lazyOffsets, nil)
	return b
}

func (b *TypeBuilder) AppendPointer(name string, lazyOffset *uintptr) *TypeBuilder {
	var p unsafe.Pointer
	return b.AppendField(name, reflect.TypeOf(p), lazyOffset)
}

func (b *TypeBuilder) AppendIntSlice(name string, lazyOffset *uintptr) *TypeBuilder {
	var s []int
	return b.AppendField(name, reflect.TypeOf(s), lazyOffset)
}

func (b *TypeBuilder) AppendString(name string, lazyOffset *uintptr) *TypeBuilder {
	return b.AppendField(name, reflect.TypeOf(""), lazyOffset)
}

func (b *TypeBuilder) AppendBool(name string, lazyOffset *uintptr) *TypeBuilder {
	return b.AppendField(name, reflect.TypeOf(true), lazyOffset)
}

func (b *TypeBuilder) AppendInt64(name string, lazyOffset *uintptr) *TypeBuilder {
	return b.AppendField(name, reflect.TypeOf(int64(0)), lazyOffset)
}

func (b *TypeBuilder) AppendFloat64(name string, lazyOffset *uintptr) *TypeBuilder {
	return b.AppendField(name, reflect.TypeOf(float64(1.2)), lazyOffset)
}

// 添加字段
func (b *TypeBuilder) AddField(field string, typ reflect.Type) *TypeBuilder {
	b.fields = append(b.fields, reflect.StructField{Name: field, Type: typ})
	return b
}

func (b *TypeBuilder) AddString(name string) *TypeBuilder {
	return b.AddField(name, reflect.TypeOf(""))
}

func (b *TypeBuilder) AddBool(name string) *TypeBuilder {
	return b.AddField(name, reflect.TypeOf(true))
}

func (b *TypeBuilder) AddInt64(name string) *TypeBuilder {
	return b.AddField(name, reflect.TypeOf(int64(0)))
}

func (b *TypeBuilder) AddFloat64(name string) *TypeBuilder {
	return b.AddField(name, reflect.TypeOf(float64(1.2)))
}

func main() {
	b := NewTypeBuilder().
		AddString("Name").
		AddInt64("Age")

	p := b.New()
	i := b.Interface()
	pp := reflect.ValueOf(i).Elem().Addr().Interface()
	fmt.Printf("typ:%T, value:%+v, ponter1:%d,ponter1:%v\n", p, i, p, pp)
}