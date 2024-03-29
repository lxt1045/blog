package hashmap

import (
	"unsafe"
)

//go:linkname memequal runtime.memequal
func memequal(a, b unsafe.Pointer, size uintptr) bool

//go:linkname reflect_mapassign_faststr reflect.mapassign_faststr
func reflect_mapassign_faststr(t *maptype, h *hmap, key string, elem unsafe.Pointer)

//go:linkname reflect_typedmemmove reflect.typedmemmove
func reflect_typedmemmove(typ *GoType, dst, src unsafe.Pointer)

//go:linkname unsafe_New reflect.unsafe_New
func unsafe_New(*GoType) unsafe.Pointer

//go:linkname unsafe_NewArray reflect.unsafe_NewArray
func unsafe_NewArray(typ *GoType, n int) unsafe.Pointer

//go:linkname reflect_ifaceE2I runtime.reflect_ifaceE2I
func reflect_ifaceE2I(inter *interfacetype, e GoEface, dst *GoIface)

//go:linkname roundupsize runtime.roundupsize
func roundupsize(size uintptr) uintptr

// //go:linkname mapassign runtime.makemap
// func makemap(t *GoType, h unsafe.Pointer, k unsafe.Pointer) unsafe.Pointer

//go:linkname bucketShift runtime.bucketShift
func bucketShift(b uint8) uintptr

//go:linkname overLoadFactor runtime.overLoadFactor
func overLoadFactor(count int, B uint8) bool

//go:linkname reflect_memclrNoHeapPointers reflect.memclrNoHeapPointers
func reflect_memclrNoHeapPointers(ptr unsafe.Pointer, n uintptr)

//go:linkname memclrHasPointers runtime.memclrHasPointers
func memclrHasPointers(ptr unsafe.Pointer, n uintptr)

//go:linkname makeBucketArray runtime.makeBucketArray
func makeBucketArray(t *maptype, b uint8, dirtyalloc unsafe.Pointer) (buckets unsafe.Pointer, nextOverflow *bmap)

/*
func makemap(t *GoType, hint int, h *hmap) *hmap {
	if h == nil {
		h = new(hmap)
	}
	h.hash0 = fastrand()

	// Find the size parameter B which will hold the requested # of elements.
	// For hint < 0 overLoadFactor returns false since hint < bucketCnt.
	B := uint8(0)
	for overLoadFactor(hint, B) {
		B++
	}
	h.B = B

	// allocate initial hash table
	// if B == 0, the buckets field is allocated lazily later (in mapassign)
	// If hint is large zeroing this memory could take a while.
	if h.B != 0 {
		var nextOverflow *bmap
		h.buckets, nextOverflow = makeBucketArray(t, h.B, nil)
		if nextOverflow != nil {
			h.extra = new(mapextra)
			h.extra.nextOverflow = nextOverflow
		}
	}

	return h
}//*/

// A header for a Go map.
type hmap struct {
	// Note: the format of the hmap is also encoded in cmd/compile/internal/reflectdata/reflect.go.
	// Make sure this stays in sync with the compiler's definition.
	count     int // # live cells == size of map.  Must be first (used by len() builtin)
	flags     uint8
	B         uint8  // log_2 of # of buckets (can hold up to loadFactor * 2^B items)
	noverflow uint16 // approximate number of overflow buckets; see incrnoverflow for details
	hash0     uint32 // hash seed

	buckets    unsafe.Pointer // array of 2^B Buckets. may be nil if count==0.
	oldbuckets unsafe.Pointer // previous bucket array of half the size, non-nil only when growing
	nevacuate  uintptr        // progress counter for evacuation (buckets less than this have been evacuated)

	extra *mapextra // optional fields
}

// A bucket for a Go map.
type bmap struct {
	tophash [8]uint8
}

const ptrSize = 4 << (^uintptr(0) >> 63)

func pointerOffset(p unsafe.Pointer, offset uintptr) (pOut unsafe.Pointer) {
	return unsafe.Pointer(uintptr(p) + uintptr(offset))
}
func (b *bmap) setoverflow(t *maptype, ovf *bmap) {
	*(**bmap)(pointerOffset(unsafe.Pointer(b), uintptr(t.bucketsize)-ptrSize)) = ovf
}

type maptype struct {
	typ    GoType
	key    *GoType
	elem   *GoType
	bucket *GoType // internal type representing a hash bucket
	// function for hashing keys (ptr to key, seed) -> hash
	hasher     func(unsafe.Pointer, uintptr) uintptr
	keysize    uint8  // size of key slot
	elemsize   uint8  // size of elem slot
	bucketsize uint16 // size of bucket
	flags      uint32
}

type mapextra struct {
	// If both key and elem do not contain pointers and are inline, then we mark bucket
	// type as containing no pointers. This avoids scanning such maps.
	// However, bmap.overflow is a pointer. In order to keep overflow buckets
	// alive, we store pointers to all overflow buckets in hmap.extra.overflow and hmap.extra.oldoverflow.
	// overflow and oldoverflow are only used if key and elem do not contain pointers.
	// overflow contains overflow buckets for hmap.buckets.
	// oldoverflow contains overflow buckets for hmap.oldbuckets.
	// The indirection allows to store a pointer to the slice in hiter.
	overflow    *[]*bmap
	oldoverflow *[]*bmap

	// nextOverflow holds a pointer to a free overflow bucket.
	nextOverflow *bmap
}
