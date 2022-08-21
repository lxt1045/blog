
# 如何优雅的获取 gid 和 pid
<!-- 
/usr/local/go/src/runtime/traceback.go:314
panic 下找 deferfunc 的 pc


https://github.com/rpccloud/goid    有各个平台的 goid 的 asm 实现，虽然获取 goid 偏移量的方式不太合适

关于编译期绑定，或者编译期预定义宏，可以利用 .h c的头文件，参考：
/usr/local/go/src/runtime/go_tls.h

https://chai2010.cn/advanced-go-programming-book/ch3-asm/ch3-08-goroutine-id.html
曹春晖和柴树杉老师的书本上的例子，很详细了！

https://github.com/timandy/gohack
通过 type 的方式获取偏移量，和 曹春晖和柴树杉老师的书本上的例子一样，不过拿的是全局变量 g0：
也会是比较全面的各平台汇编：
原版可能是： https://github.com/go-eden/routine
```s
TEXT ·getgp(SB), NOSPLIT, $0-8
    get_tls(CX)
    MOVQ    g(CX), AX
    MOVQ    AX, ret+0(FP)
    RET

TEXT ·getg0(SB), NOSPLIT, $0-16
    NO_LOCAL_POINTERS
    MOVQ    $0, ret_type+0(FP)
    MOVQ    $0, ret_data+8(FP)
    GO_RESULTS_INITIALIZED
    //get runtime.g type
    MOVQ    $type·runtime·g(SB), AX
    //get runtime·g0 variable // src/runtime/proc.go:115   定义的全局变量
    MOVQ    $runtime·g0(SB), BX
    //return interface{}
    MOVQ    AX, ret_type+0(FP)
    MOVQ    BX, ret_data+8(FP)
    RET
```
```go
// getgt returns the type of runtime.g.
//go:nosplit
func getgt() reflect.Type {
	return reflect.TypeOf( getg0() )
}
// offset returns the offset of the specified field.
func offset(t reflect.Type, f string) uintptr {
	field, found := t.FieldByName(f)
	if found {
		return field.Offset
	}
	panic(fmt.Sprintf("No such field '%v' of struct '%v.%v'.", f, t.PkgPath(), t.Name()))
}
func init() {
	gt := getgt()
	offsetGoid = offset(gt, "goid")
	offsetPaniconfault = offset(gt, "paniconfault")
	offsetGopc = offset(gt, "gopc")
	offsetLabels = offset(gt, "labels")
}
其中下面这句可以借鉴：
	field, found := t.FieldByName(f)
	if found {
		return field.Offset
	}
```

https://github.com/kortschak/goroutine/blob/master/gid.go#L40
string to type, 可以借鉴！！！
```go
// ID returns the runtime ID of the calling goroutine.
func ID() int64 {
	return *(*int64)(add(getg(), goidoff))
}

func getg() unsafe.Pointer {
	return *(*unsafe.Pointer)(add(getm(), curgoff))
}

//go:linkname add runtime.add
//go:nosplit
func add(p unsafe.Pointer, x uintptr) unsafe.Pointer

//go:linkname getm runtime.getm
func getm() unsafe.Pointer

var (
	curgoff = offset("*runtime.m", "curg")
	goidoff = offset("*runtime.g", "goid")
)

// offset returns the offset into typ for the given field.
func offset(typ, field string) uintptr {
	rt := toType(typesByString(typ)[0])
	f, _ := rt.Elem().FieldByName(field)
	return f.Offset
}

//go:linkname typesByString reflect.typesByString
func typesByString(s string) []unsafe.Pointer

//go:linkname toType reflect.toType
func toType(t unsafe.Pointer) reflect.Type
```

关于 framepointer_enabled ：
	/usr/local/go/src/runtime/stack.go:528
	/usr/local/go/src/runtime/runtime2.go:1134
	/usr/local/go/src/runtime/traceback.go:265 
	https://github.com/cch123/golang-notes/blob/master/assembly.md

关于 ARCH：
/usr/local/go/src/cmd/asm/internal/arch/arch.go:51   寄存器映射
386
amd64
arm
arm64
mips
mipsle
mips64
mips64le
ppc64
ppc64le
riscv64
s390x
wasm

/usr/local/go/src/cmd/dist/build.go:61      ：支持的 CPU 和 OS
```go

// The known architectures.
var okgoarch = []string{
	"386",
	"amd64",
	"arm",
	"arm64",
	"mips",
	"mipsle",
	"mips64",
	"mips64le",
	"ppc64",
	"ppc64le",
	"riscv64",
	"s390x",
	"sparc64",
	"wasm",
}

// The known operating systems.
var okgoos = []string{
	"darwin",
	"dragonfly",
	"illumos",
	"ios",
	"js",
	"linux",
	"android",
	"solaris",
	"freebsd",
	"nacl", // keep;
	"netbsd",
	"openbsd",
	"plan9",
	"windows",
	"aix",
}
```


/usr/local/go/src/cmd/cgo/main.go:172   指针位宽



-->

## 1. 获取goid方案
1. [petermattis/goid](https://github.com/petermattis/goid)
相信大家都知道这个方案，这个方案最关键的是[这几行汇编代码](https://github.com/petermattis/goid/blob/master/goid_go1.5_amd64.s#L26)：
```s
// func Get() int64
TEXT ·Get(SB),NOSPLIT,$0-8
	MOVQ (TLS), R14
	MOVQ g_goid(R14), R13
	MOVQ R13, ret+0(FP)
	RET

```
不过这个方案有个比加大的兼容问题，就是强依赖 g_goid ，它是 [runtime.g](https://github.com/golang/go/blob/go1.18/src/runtime/runtime2.go#L405) 结构体里 "goid" 成员的偏移量。随着 golang 的升级，g_goid 有可能就要跟着作兼容性升级。否则就无法得到正确的 goid。

2. [曹春晖老师也写了一个库](https://github.com/cch123/goroutineid)，其中[关键汇编代码](https://github.com/cch123/goroutineid/blob/master/goid.s)如下：
```s
// func GetGoID() int64
TEXT ·GetGoID(SB), NOSPLIT, $0-8
	get_tls(CX)
	MOVQ g(CX), AX
	MOVQ ·offset(SB), BX
	LEAQ 0(AX)(BX*1), DX
	MOVQ (DX), AX
	MOVQ AX, ret+0(FP)
	RET

```
其缺点和 [petermattis/goid](https://github.com/petermattis/goid) 类似，需要手动维护 go 编译器版本和  g.goid 的品一两的映射关系。

3. 柴树杉老师和曹春晖老师联合出版的树 [《Go语言高级编程》](https://chai2010.cn/advanced-go-programming-book/ch3-asm/ch3-08-goroutine-id.html) 里也讲了一个例子。

该例子讲的很清楚，大家可以去围观一下，这里就不展开仅列一下关键代码：
```s
// func getg() unsafe.Pointer
TEXT ·getg(SB), NOSPLIT, $0-8
    MOVQ (TLS), AX
    MOVQ AX, ret+0(FP)
    RET

// func GetGroutine() interface{}
TEXT ·GetGroutine(SB), NOSPLIT, $32-16
    NO_LOCAL_POINTERS

    MOVQ $0, ret_type+0(FP)
    MOVQ $0, ret_data+8(FP)
    GO_RESULTS_INITIALIZED

    // get runtime.g
    MOVQ (TLS), AX

    // get runtime.g type
    MOVQ $type·runtime·g(SB), BX

    // convert (*g) to interface{}
    MOVQ AX, 8(SP)
    MOVQ BX, 0(SP)
    CALL runtime·convT2E(SB)
    MOVQ 16(SP), AX
    MOVQ 24(SP), BX

    // return interface{}
    MOVQ AX, ret_type+0(FP)
    MOVQ BX, ret_data+8(FP)
    RET

```
```go

var g_goid_offset uintptr = func() uintptr {
    g := GetGroutine()
    if f, ok := reflect.TypeOf(g).FieldByName("goid"); ok {
        return f.Offset
    }
    panic("can not find g.goid field")
}()

func GetGroutineId() int64 {
    g := getg()
    p := (*int64)(unsafe.Pointer(uintptr(g) + g_goid_offset))
    return *p
}

```
这里利用了 reflect 包的 FieldByName 方法，获取 goid 在 runtime·g 的偏移量，算是一种令人眼前一亮的实现方式。因为这笔前两种方案更具向前兼容性，只要 goid 的名字不变，就不用升级。

不过针对新版本 golang 还有个小问题，go1.18 runtime·convT2E 已经不在提供了，所以这里需要再改一下。如果了解 interface{} 的struct的话，就知道可以很简单的吧汇编代码改成这样即可。
```s
// func GetGroutine() interface{}
TEXT ·GetGroutine(SB), NOSPLIT, $32-16
    NO_LOCAL_POINTERS

    MOVQ $0, ret_type+0(FP)
    MOVQ $0, ret_data+8(FP)
    GO_RESULTS_INITIALIZED

    // get runtime.g
    MOVQ (TLS), AX

    // get runtime.g type
    MOVQ $type·runtime·g(SB), BX

    // convert (*g) to interface{}
    // MOVQ AX, 8(SP)
    // MOVQ BX, 0(SP)
    // CALL runtime·convT2E(SB)
    // MOVQ 16(SP), AX
    // MOVQ 24(SP), BX

    // return interface{}
    MOVQ BX, ret_type+0(FP)
    MOVQ AX, ret_data+8(FP)
```

4. 因为上面第 3 种实现方式已经很完美了，我们这里仅对它进行简单修改，把计算偏移量放到汇编里。
```s
// func GetGroutineId() unsafe.Pointer
TEXT ·GetGroutineId(SB), NOSPLIT, $0-8
    MOVQ (TLS), AX
	ADDQ ·g_goid_offset(SB),AX
    MOVQ (AX), BX
    MOVQ BX, ret+0(FP)
    RET
```
这样，可以提高些许性能。


## 2. 获取pid方案

1. [choleraehyq/pid](https://github.com/choleraehyq/pid)
这个方案是 fork 了获取 goid 的开源方案 [petermattis/goid](https://github.com/petermattis/goid) ，然后用同样的方法获取了 pid ，其关键代码如下：
```s
// func getPid() int64
TEXT ·getPid(SB),NOSPLIT,$0-8
	MOVQ (TLS), R14
	MOVQ g_m(R14), R13
	MOVQ m_p(R13), R14
	MOVL p_id(R14), R13
	MOVQ R13, ret+0(FP)
	RET

```
我们字节内部的 [gopkg/pid](https://code.byted.org/gopkg/pid) 就是 fork 了 [choleraehyq/pid](https://github.com/choleraehyq/pid) ，并对 go1.18 版本做了兼容性升级。在字节内部使用还是很广泛的。

2. [sync.Pool](https://github.com/golang/go/blob/dev.boringcrypto.go1.18/src/sync/pool.go#L196) 里的 pid 获取方式

其关键代码如下：
```go 
func runtime_procPin() int
func runtime_procUnpin()
```
并没有实现，这是为什么呢？

原来，这个被动链接函数，发起链接的地方在 [runtime/proc.go](https://github.com/golang/go/blob/dev.boringcrypto.go1.18/src/runtime/proc.go#L6046) 里
```go

//go:nosplit
func procPin() int {
	_g_ := getg()
	mp := _g_.m

	mp.locks++
	return int(mp.p.ptr().id)
}

//go:nosplit
func procUnpin() {
	_g_ := getg()
	_g_.m.locks--
}

//go:linkname sync_runtime_procPin sync.runtime_procPin
//go:nosplit
func sync_runtime_procPin() int {
	return procPin()
}

//go:linkname sync_runtime_procUnpin sync.runtime_procUnpin
//go:nosplit
func sync_runtime_procUnpin() {
	procUnpin()
}
```

3. 

附录：


关于 pid ，这里还涉及一个知识点：使用 pid 的时候需要给 g.m 上锁！！！

那这个锁是用来干什么的呢？

我们可以参考这 golang 源码的两个注释：
1. [do not reschedule between printlock++ and lock(&debuglock).](https://github.com/golang/go/blob/dev.boringcrypto.go1.18/src/runtime/print.go#L68)
2. [Double-check trace.enabled now that we've done m.locks++ and acquired bufLock. This protects from races between traceEvent and StartTrace/StopTrace.](https://github.com/golang/go/blob/dev.boringcrypto.go1.18/src/runtime/trace.go#L523)
3. [disable preemption because it can be holding p in a local var](https://github.com/golang/go/blob/dev.boringcrypto.go1.18/src/runtime/proc.go#L850)
4. [The caller owns _p_, but we may borrow (i.e., acquirep) it. We must disable preemption to ensure it is not stolen, which would make the caller lose ownership.](https://github.com/golang/go/blob/dev.boringcrypto.go1.18/src/runtime/proc.go#L1708)
5. [Callers passing a non-nil P must already be in non-preemptible context, otherwise such preemption could occur on function entry to startm. Callers passing a nil P may be preemptible, so we must disable preemption before acquiring a P from pidleget below.](https://github.com/golang/go/blob/dev.boringcrypto.go1.18/src/runtime/proc.go#L2262)
6. [disable preemption because it can be holding p in a local var](https://github.com/golang/go/blob/dev.boringcrypto.go1.18/src/runtime/proc.go#L4077)
7. [这里说goroutine 调度的时候需要 getg().m.locks == 0 ](https://github.com/golang/go/blob/dev.boringcrypto.go1.18/src/runtime/sema.go#L195)


 [Increment locks to ensure that the goroutine is not preempted in the middle of sweep thus leaving the span in an inconsistent state for next GC](https://github.com/golang/go/blob/dev.boringcrypto.go1.18/src/runtime/mgcsweep.go#L335) 

由已上注释可知，如果不给 g.m.locks++ 的话，在使用 pid 的过程中，有可能正在使用的 pid 已经不是与当前 goroutine 绑定的 pid 了。因为，这期间有可能 goroutine 已经被调度到其他 P 上执行了！！！


