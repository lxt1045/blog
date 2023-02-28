package hashmap

import (
	"reflect"
	"sync"
	"unsafe"
)

//go:linkname strhash runtime.strhash
func strhash(p unsafe.Pointer, h uintptr) uintptr

//go:noescape
func Hash(bs, cs []byte) int

type N struct {
	pick   [1024]byte
	mask   [1024]byte
	shiftL [128]int64
}

var Pick08 = [...]byte{
	0x00, 0x0F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
}

//go:noescape
func Hashx(bs []byte, cs []N) int

//go:noescape
func IndexByte(bs []byte, c byte) int

//go:noescape
func IndexBytes(bs []byte, cs []byte) int

//go:noescape
func IndexBytes1(bs []byte, cs []byte) int

//go:noescape
func IndexBytes2(bs []byte, cs []byte) int

func Test1(x, y int) (a, b int)
func Test2(a int, xs []byte) (n int)

var SpaceBytes = [8][16]byte{
	fillBytes16('\t'),
	fillBytes16('\n'),
	fillBytes16('\v'),
	fillBytes16('\f'),
	fillBytes16('\r'),
	fillBytes16(' '),
	fillBytes16(0x85),
	fillBytes16(0xA0),
}

func fillBytes16(b byte) (bs [16]byte) {
	for i := 0; i < 16; i++ {
		bs[i] = b
	}
	return
}

// asm 中读入 X0 寄存器
var SpaceQ = [8]byte{0x85, 0xA0, '\t', '\n', '\v', '\f', '\r', ' '}

// 在 asm 中实现
func InSpaceQ(b byte) bool

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

	SPool  sync.Pool // TODO：slice pool 和 store.pool 放在一起吧，通过 id 来获取获取 pool，并把剩余的”垃圾“放回 sync.Pool 中共下次复用
	SPoolN int32

	slicePool sync.Pool // &dynamicPool{} 的 pool，用于批量非配 slice
	// idxStackDynamic uintptr   // 在 store.pool 的 index 文字

	sPooloffset  int32 // slice pool 在 PoolStore的偏移量； TODO
	psPooloffset int32 // pointer slice pool  在 PoolStore的偏移量
	bsMarshalLen int32 // 缓存上次 生成的 bs 的大小，如果 cache 小于这个值，则丢弃
	bsHaftCount  int32 // 记录上次低于 bsMarshalLen/2 的次数
}
