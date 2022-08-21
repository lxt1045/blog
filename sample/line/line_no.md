
我们打印log的时候，一般都伴随着行号的输出。
像C这种支持预定义宏(比如：\_\_FILE\_\_、\_\_LINE\_\_)的语言，可以在编译期计算代码行号，几乎没有运行期损耗，算是一种比较完美的实现方式。而对golang来说就没那么好的接口可以使用了。

以下是golang官方提供的获取行号的方法。

### sample 1. 
```go
func LineByRuntime() string {
    _， file， n， ok := runtime.Caller(depth)
    if !ok {
        return ""
    }
    return file + ":" + strconv.Itoa(n)
}
```
基准测试结果如下：

(其中 CPU是 Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz，下文不再说明.)
```
LineByRuntime-12      1585899      728.4 ns/op      272 B/op      4 allocs/op
```
由此观之，性能损失确实不小。

如果想要优化他的性能，我们该怎么做呢？

首先想到的肯定是本地缓存，每一个需要获取行号的地方都缓存一下：

### sample 2:
```go
//lib.go
type Line struct {
    sync.Once
    lineNO string
}

func (l *Line) Load() string {
    l.Do(func() {
        if l.lineNO != "" {
            return
        }
        _， file， line， ok := runtime.Caller(depth)
        if !ok {
            return
        }
        l.lineNO = file + ":" + strconv.Itoa(line)
    })
    return l.lineNO
}

//main.go
var lineNO Line
func main(){
    line := lineNO.Load()
    ...
}
```
简单做一下基准测试，结果比较惊喜：
```
LineNO.Load-12      537195462      2.199 ns/op      0 B/op      0 allocs/op
```
就性能来说，无可挑剔。不过在使用接口的时候还要提供一个全局变量，就显得不太友好了。

相对于上面 sample 2 采用的分散式的缓存，我们还可以改成集中式缓存，如下所示：

### sample 3：
```go
// lib.go
var intIDCache *[]*string = func() *[]*string {
    s := make([]*string， 8*1024)
    return &s
}()

// var escapesLine *string

func GetLineNO3(id int) string {
	cache := intIDCache
	if len(*cache) <= id {
		s := make([]*string， id*2)
		copy(s， *cache)
		// intIDCache = &s
		atomic.StoreUintptr((*uintptr)(unsafe.Pointer(&intIDCache))， (uintptr)(unsafe.Pointer(&s)))
		cache = &s
	}
	p := (*cache)[id]
	if p != nil && *p != "" {
		return *p
	}
	_， file， l， ok := runtime.Caller(1)
	if ok {
		line := file + ":" + strconv.Itoa(l)
		atomic.StoreUintptr((*uintptr)(unsafe.Pointer(&(*cache)[id]))， (uintptr)(unsafe.Pointer(&line)))
		(*cache)[id] = &line // 这句让编译器确保line逃逸，因为line未逃逸的话，会导致引用未初始化内存. (uintptr)(unsafe.Pointer(&)比较隐蔽的坑
		return line
	}
	return ""
}


// main.go
const (
    LogNO1 int = iota
    LogNO2

    LogNOMax
)
var _ = GetLineNO3(LogNOMax) //预分配

func main()  {
    line := GetLineNO3(LogNO2)
    ...
}
```
基准测试结果如下：
```
GetLineNO3-12      510159147      2.349 ns/op      0 B/op      0 allocs/op
```
相对 sample 2 而言，性能差不多，不过接口稍微友好了一点点。

换个思路，可能采用方便记忆的string类型id用起来会更舒服。不过因为只能用map缓存，性能可能会稍差些。

### sample 4：
```go
var (
    stringIDCache  = map[string]string{}
    lockMapIDCache sync.RWMutex
)

func GetLineNO4(id string) (line string) {
    lockMapIDCache.RLock()
    line， ok := stringIDCache[id]
    lockMapIDCache.RUnlock()
    if !ok {
        _， file， n， ok := runtime.Caller(1)
        if !ok {
            return
        }
        line = file + ":" + strconv.Itoa(n)
        lockMapIDCache.Lock()
        stringIDCache[id] = line
        lockMapIDCache.Unlock()
    }
    return
}

//main.go
var lineNO Line
func main(){
    line := GetLineNO4("main.001")
    ...
}
```
基准测试结果如下：
```
GetLineNO4-12      68144361      14.74 ns/op      0 B/op      0 allocs/op
```

很多同学可能已经想到了，这里如果用原子操作替换读写锁的话，性能应该还有优化空间。

### sample 5：
```go
var (
    stringIDCacheAtomic unsafe.Pointer = func() unsafe.Pointer {
        m := make(map[string]string)
        return unsafe.Pointer(&m)
    }()
)

func GetLineNO5(id string) (line string) {
    mPCs := *(*map[string]string)(atomic.LoadPointer(&stringIDCacheAtomic))
    line， ok := mPCs[id]
    if !ok {
        _， file， l， ok := runtime.Caller(1)
        if !ok {
            return
        }
        line = file + ":" + strconv.Itoa(l)
        mPCs2 := make(map[string]string， len(mPCs)+10)
        mPCs2[id] = line
        for {
            p := atomic.LoadPointer(&stringIDCacheAtomic)
            mPCs = *(*map[string]string)(p)
            for k， v := range mPCs {
                mPCs2[k] = v
            }
            swapped := atomic.CompareAndSwapPointer(&stringIDCacheAtomic， p， unsafe.Pointer(&mPCs2))
            if swapped {
                break
            }
        }
    }
    return
}
```
从基准测试结果来看，RWMutex转成atomic对性能的提升还是不错的：
```
GetLineNO5-12      135471256      8.912 ns/op      0 B/op      0 allocs/op
```
以上的尝试可以说都是卓有成效的，不过如果我们还要追求更加友好的接口，还是可以折腾一下的。


我们在reflect包里发现了[一个神奇方法](https://github.com/golang/go/blob/dev.boringcrypto.go1.18/src/reflect/value.go#L1957)：
```go
//
// If v's Kind is Func， the returned pointer is an underlying
// code pointer， but not necessarily enough to identify a
// single function uniquely. The only guarantee is that the
// result is zero if and only if v is a nil func Value.
//
//...
func (v Value) Pointer() uintptr {
    ...
```
也就是说，只要我们传一个函数或闭包进去，就能拿到底层代码的指针(实际上就是代码的pc)，这不就是我们想要的嘛! 接下来我们试一下。

### sample 6:
```go
//lib.go
var (
    mCache3 unsafe.Pointer = func() unsafe.Pointer {
        m := make(map[uintptr]string)
        return unsafe.Pointer(&m)
    }()
)

func GetLineByFunc(f func()) (line string) {
    pc := reflect.ValueOf(f).Pointer()
    mPCs := *(*map[uintptr]string)(atomic.LoadPointer(&mCache3))
    line， ok := mPCs[pc]
    if !ok {
        // 这里由于有 pc 所以就用性能更好的方式了. 但由于是一次性执行的操作，
        // 所以相对使用runtime.Caller来说，对整体性能的提升几乎可以忽略
        file， l := runtime.FuncForPC(pc).FileLine(pc)
        line = file + ":" + strconv.Itoa(l)
        mPCs2 := make(map[uintptr]string， len(mPCs)+10)
        mPCs2[pc] = line
        for {
            p := atomic.LoadPointer(&mCache3)
            mPCs = *(*map[uintptr]string)(p)
            for k， v := range mPCs {
                mPCs2[k] = v
            }
            swapped := atomic.CompareAndSwapPointer(&mCache3， p， unsafe.Pointer(&mPCs2))
            if swapped {
                break
            }
        }
    }
    return
}
// main.go
func main(){
    pc:=GetLineByFunc(func(){})
}
```
性能如下：
```
GetLineByFunc-12      100000000      10.68 ns/op      0 B/op      0 allocs/op
```
如果是妥协能力比较强的同学，到了这一步，基本上已经可以接受了。

不过我们追求极致的同学，肯定还是希望接口能够像 sampel 1 一样好用。

我们回到 sampel 1，也给它加上缓存。

### sample 7:
```go
var (
    mapRuntimeCache unsafe.Pointer = func() unsafe.Pointer {
        m := make(map[uintptr]string， 1024)
        return unsafe.Pointer(&m)
    }()
)

func GetLineByRuntimeCache() (line string) {
    var pcs [1]uintptr
    runtime.Callers(1， pcs[:])
    pc := pcs[0]
    mPCs := *(*map[uintptr]string)(atomic.LoadPointer(&mapRuntimeCache))
    line， ok := mPCs[pc]
    if !ok {
        file， l := runtime.FuncForPC(pc).FileLine(pc)
        line = file + ":" + strconv.Itoa(l)
        mPCs2 := make(map[uintptr]string， len(mPCs)+10)
        mPCs2[pc] = line
        for {
            p := atomic.LoadPointer(&mapRuntimeCache)
            mPCs = *(*map[uintptr]string)(p)
            for k， v := range mPCs {
                mPCs2[k] = v
            }
            swapped := atomic.CompareAndSwapPointer(&mapRuntimeCache， p， unsafe.Pointer(&mPCs2))
            if swapped {
                break
            }
        }
    }
    return
}
```
果然性能还是提升了不少，不过相较于其他方案差距还是有点大：
```
GetLineByRuntimeCache-12      8376738      141.3 ns/op      0 B/op      0 allocs/op
```
到这里，如果还要继续优化的话，只能拿runtime.Callers开刀了，该函数的功能是获取当前调用栈的pc列表，不过我们只需要上一级调用栈的pc值。

除了runtime.Callers，我们还可以在哪里可以拿到pc值呢? 当然是调用栈里！

我们来先看一下 golang 的调用栈，下图出自[曹春晖老师的github文章](https://github.com/cch123/asmshare/blob/master/layout.md#%E6%9F%A5%E7%9C%8B-go-%E8%AF%AD%E8%A8%80%E7%9A%84%E5%87%BD%E6%95%B0%E8%B0%83%E7%94%A8%E8%A7%84%E7%BA%A6) :
```
                                                                                                                              
                                       caller                                                                                 
                                 +------------------+                                                                         
                                 |                  |                                                                         
       +---------------------->  --------------------                                                                         
       |                         |                  |                                                                         
       |                         | caller parent BP |                                                                         
       |           BP(pseudo SP) --------------------                                                                         
       |                         |                  |                                                                         
       |                         |   Local Var0     |                                                                         
       |                         --------------------                                                                         
       |                         |                  |                                                                         
       |                         |   .......        |                                                                         
       |                         --------------------                                                                         
       |                         |                  |                                                                         
       |                         |   Local VarN     |                                                                         
                                 --------------------                                                                         
 caller stack frame              |                  |                                                                         
                                 |   callee arg2    |                                                                         
       |                         |------------------|                                                                         
       |                         |                  |                                                                         
       |                         |   callee arg1    |                                                                         
       |                         |------------------|                                                                         
       |                         |                  |                                                                         
       |                         |   callee arg0    |                                                                         
       |                         ----------------------------------------------+   FP(virtual register)                       
       |                         |                  |                          |                                              
       |                         |   return addr    |  parent return address   |                                              
       +---------------------->  +------------------+---------------------------    <-------------------------------+         
                                                    |  caller BP               |                                    |         
                                                    |  (caller frame pointer)  |                                    |         
                                     BP(pseudo SP)  ----------------------------                                    |         
                                                    |                          |                                    |         
                                                    |     Local Var0           |                                    |         
                                                    ----------------------------                                    |         
                                                    |                          |                                              
                                                    |     Local Var1           |                                              
                                                    ----------------------------                            callee stack frame
                                                    |                          |                                              
                                                    |       .....              |                                              
                                                    ----------------------------                                    |         
                                                    |                          |                                    |         
                                                    |     Local VarN           |                                    |         
                                  SP(Real Register) ----------------------------                                    |         
                                                    |                          |                                    |         
                                                    |                          |                                    |         
                                                    |                          |                                    |         
                                                    |                          |                                    |         
                                                    |                          |                                    |         
                                                    +--------------------------+    <-------------------------------+         
                                                                                                                              
                                                              callee
```

我们看到，golang的汇编语言(plan9)的伪寄存器FB的指向函数返回地址(指令CALL的下一个指令地址)的上方，我们完全可以把这个地址拿出来，就可以得到调用函数位置的PC了(这里需不需要减1，可以[参考go源码的这个注释](https://github.com/golang/go/blob/dev.boringcrypto.go1.18/src/runtime/traceback.go#L339))。

### sample 8:
```s
#  stack_amd64.s
TEXT ·GetPC(SB)，NOSPLIT，$0-8
    MOVQ    retpc-8(FP)， AX // MOVQ (SP)， R13
    MOVQ    AX， ret+0(FP)
    RET

```
```go
//stack_amd64.go
func GetPC() uintptr


var (
	mapPCByAsm unsafe.Pointer = func() unsafe.Pointer {
		m := make(map[uintptr]string， 1024)
		return unsafe.Pointer(&m)
	}()
)

func GetPCByAsm(pc uintptr) (line string) {
	mPCs := *(*map[uintptr]string)(atomic.LoadPointer(&mapPCByAsm))
	line， ok := mPCs[pc]
	if !ok {
		file， l := runtime.FuncForPC(pc).FileLine(pc)
		line = file + ":" + strconv.Itoa(l)
		mPCs2 := make(map[uintptr]string， len(mPCs)+10)
		mPCs2[pc] = line
		for {
			p := atomic.LoadPointer(&mapPCByAsm)
			mPCs = *(*map[uintptr]string)(p)
			for k， v := range mPCs {
				mPCs2[k] = v
			}
			swapped := atomic.CompareAndSwapPointer(&mapPCByAsm， p， unsafe.Pointer(&mPCs2))
			if swapped {
				break
			}
		}
	}
	return
}

//main.go
func main(){
    line := GetPCByAsm(GetPC())
    ...
 }
```
毫无意外，性能相当不错：
```
GetLineByAsm-12      150710400      7.938 ns/op      0 B/op      0 allocs/op
```

不过，这个接口只能这样使用：
```go
    line := GetLineByFunc3(GetPC())
```
和 sample 6 的接口没有本质的区别:
```go
    line :=GetLineByFunc(func(){})
```
所以还是需要继续改改。

### sample 9:
```s
# stack_amd64.s
TEXT    ·NewLine(SB)， NOSPLIT， $0-8
    MOVQ     retpc-8(FP)， AX
    MOVQ     AX， ret+0(FP)
    RET

```
```go
//stack_amd64.go

type Line uintptr

func NewLine() Line

var (
    mapByAsm unsafe.Pointer = func() unsafe.Pointer {
        m := make(map[Line]string)
        return unsafe.Pointer(&m)
    }()
)

func (l Line) LineNO() (line string) {
    mPCs := *(*map[Line]string)(atomic.LoadPointer(&mapByAsm))
    line， ok := mPCs[l]
    if !ok {
        file， n := runtime.FuncForPC(uintptr(l)).FileLine(uintptr(l))
        line = file + ":" + strconv.Itoa(n)
        mPCs2 := make(map[Line]string， len(mPCs)+10)
        mPCs2[l] = line
        for {
            p := atomic.LoadPointer(&mapByAsm)
            mPCs = *(*map[Line]string)(p)
            for k， v := range mPCs {
                mPCs2[k] = v
            }
            swapped := atomic.CompareAndSwapPointer(&mapByAsm， p， unsafe.Pointer(&mPCs2))
            if swapped {
                break
            }
        }
    }
    return
}
```
测试结果表明，性能和改之前一样：
```
stack.NewLine().LineNO()-12      149483650      8.165 ns/op      0 B/op      0 allocs/op
```
不过接口却更友好了：
```go
    line := NewLine().LineNO()
```

至此，我们现在就得到了比较完美的接口，已经完全摆脱了不友善的接口形式。
相较于辅助变量的方式，用 6ns 换一个相对友善的接口，应该是值得的。



附录：

本文所有代码都在[github的这个仓库](https://github.com/lxt1045/blog/tree/main/sample/line)下