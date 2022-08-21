<!-- # golang 错误处理的一些优化想法 -->

## 0. 前言
1. 由于笔者水平有限，文章中难免出现各种错误，欢迎吐槽。
2. 由于篇幅所限，大部分代码细节并未讲的很清楚，如有疑问欢迎讨论。

##  1. 当前存在的问题

### 1.1 标准库的 error 信息量少
golang 自带的 errors.New() 和 fmt.Errorf() 都只能记录 string 信息，错误现场只能靠 log 提供，如果需要获取调用栈，需要额外的代码逻辑。即使后来增加了 "%w" 的支持也没有本质的改变。

令人欣喜的是，[pkg/errors](https://github.com/pkg/errors) 补齐了这个短板，其 [WithStack()](https://github.com/pkg/errors/blob/v0.9.1/errors.go#L143) 提供了完整的调用栈信息，且 [Wrap()](https://github.com/pkg/errors/blob/v0.9.1/errors.go#L181) 也提供了更详细的 error 路径追踪能力。

### 1.2 [pkg/errors](https://github.com/pkg/errors) 好用，但性能较差
我们简单给 [pkg/errors](https://github.com/pkg/errors) 做个基准测试：
```go
package errors

import (
    "errors"
    "testing"

    pkgerrs "github.com/pkg/errors"
)

func deepCall(depth int, f func()) {
    if depth <= 0 {
        f()
        return
    }
    deepCall(depth-1, f)
}
func BenchmarkPkg(b *testing.B) {
    b.Run("pkg/errors", func(b *testing.B) {
        b.ReportAllocs()
        var err error
        deepCall(10, func() {
            for i := 0; i < b.N; i++ {
                err = pkgerrs.New("ye error")
                GlobalE = fmt.Sprintf("%+v", err)
            }
            b.StopTimer()
        })
    })
    b.Run("errors-fmt", func(b *testing.B) {
        b.ReportAllocs()
        var err error
        deepCall(10, func() {
            for i := 0; i < b.N; i++ {
                err = errors.New("ye error")
                GlobalE = fmt.Sprintf("%+v", err)
            }
            b.StopTimer()
        })
    })
    b.Run("errors-Errors", func(b *testing.B) {
        b.ReportAllocs()
        var err error
        deepCall(10, func() {
            for i := 0; i < b.N; i++ {
                err = errors.New("ye error")
                GlobalE = err.Error()
            }
            b.StopTimer()
        })
    })
}
/*
结果如下：
cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
BenchmarkPkg
BenchmarkPkg/pkg/errors
BenchmarkPkg/pkg/errors-12    105876    10353 ns/op    2354 B/op    36 allocs/op
BenchmarkPkg/errors-fmt
BenchmarkPkg/errors-fmt-12    8025952    154.4 ns/op    24 B/op    2 allocs/op
BenchmarkPkg/errors-Errors
BenchmarkPkg/errors-Errors-12    42631442    26.23 ns/op    16 B/op    1 allocs/op
*/
```
由基准测试结果可知，仅 10 个调用层级 [pkg/errors](https://github.com/pkg/errors) 就需要损耗近 10us;
而 errors.New() 在同等条件下只损耗 150ns，如果用 Error() 输出甚至只要不到 30ns，差距巨大。
这也是很多 gopher 犹豫是否用 [pkg/errors](https://github.com/pkg/errors) 替换 errors.New() 和 fmt.Errorf() 的原因。


### 1.3 [pkg/errors](https://github.com/pkg/errors) 信息冗余较严重
 [WithStack()](https://github.com/pkg/errors/blob/v0.9.1/errors.go#L143) 的冗余信息应该不算多，
 不过 [Wrap()](https://github.com/pkg/errors/blob/v0.9.1/errors.go#L181) 接口还带上完整的调用栈就显得没那么必要了，因为大部分和 [WithStack()](https://github.com/pkg/errors/blob/v0.9.1/errors.go#L143) 其实是重复的，而且如果多调用几次 [Wrap()](https://github.com/pkg/errors/blob/v0.9.1/errors.go#L181) 很容易会造成日志超长。

### 1.4 panic(err) 方便，但有可能会造成程序退出
很多同学受不了 if err!=nil{ ... return } 的错误处理方式，转而使用更方便的 defer...panic(err) 方式。
这种方式虽然可以方便的处理错误，不过也会带来致命的问题----如果忘记 defer recover() 将会造成整个程序的退出。
也是因此，很多团队把 defer...panic(err) 方式列为了红线，严禁触碰。

## 2. 有什么改进想法？

可以看到 [pkg/errors](https://github.com/pkg/errors) 实际上是使用 [runtime.Callers](https://github.com/golang/go/blob/release-branch.go1.18/src/runtime/extern.go#L215) 获取完整调用栈的，而该函数实际性能比较差，简单基准测试如下：
```go
func BenchmarkStack(b *testing.B) {
    b.Run("runtime.Callers", func(b *testing.B) {
        deepCall(10, func() {
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                pcs := pool.Get().(*[DefaultDepth]uintptr)
                n := runtime.Callers(2, pcs[:DefaultDepth])
                var cs []string
                traces, more, f := runtime.CallersFrames(pcs[:n]), true, runtime.Frame{}
                for more {
                    f, more = traces.Next()
                    cs = append(cs, f.File+":"+strconv.Itoa(f.Line))
                }
                pool.Put(pcs)
            }
        })
    })
}
/*
cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
BenchmarkStack
BenchmarkStack/runtime.Callers
BenchmarkStack/runtime.Callers-12    252842    4947 ns/op    1820 B/op    24 allocs/op
*/
```
由测试结果可知，其获取 stack 的损耗接近 5us。

[笔者之前的文章](https://tech.bytedance.net/articles/7111885881002164238) 里也有涉及 [runtime.Callers](https://github.com/golang/go/blob/release-branch.go1.18/src/runtime/extern.go#L215) 的优化，该文章虽然仅优化了获取当前代码行的性能，不过其思路依然可以作为此次优化的参考。

### 2.1 通过缓存提升 pc 到代码行号的性能
```go
package errors

import (
    "runtime"
    "strconv"
    "sync"
    "unsafe"

    "github.com/lxt1045/errors"
)

const DefaultDepth = 32

var (
    cacheStack = errors.AtomicCache[[DefaultDepth]uintptr{}, []string]{}
    pool       = sync.Pool{
        New: func() any { return &[DefaultDepth]uintptr{} },
    }
)

func NewStack(skip int) (stack []string) {
    pcs := pool.Get().(*[DefaultDepth]uintptr)
    n := runtime.Callers(skip+2, pcs[:DefaultDepth-skip])
    for i := n; i < DefaultDepth; i++ {
        pcs[i] = 0
    }

    stack = cacheStack.Get(*pcs)
    if len(stack) == 0 {
        stack = parseSlow(pcs[:n])
        cacheStack.Set(*pcs, stack)
    }
    pool.Put(pcs)
    return
}
func parseSlow(pcs []uintptr) (cs []string) {
    traces, more, f := runtime.CallersFrames(pcs), true, runtime.Frame{}
    for more {
        f, more = traces.Next()
        cs = append(cs, f.File+":"+strconv.Itoa(f.Line))
    }
    return
}
```
做个简单的基准测试：
```go
func BenchmarkCacheStack(b *testing.B) {
    b.Run("NewStack", func(b *testing.B) {
        deepCall(10, func() {
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                NewStack(0)
            }
        })
    })
    b.Run("runtime.Callers", func(b *testing.B) {
        deepCall(10, func() {
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                pcs := pool.Get().(*[DefaultDepth]uintptr)
                n := runtime.Callers(2, pcs[:DefaultDepth])
                var cs []string
                traces, more, f := runtime.CallersFrames(pcs[:n]), true, runtime.Frame{}
                for more {
                    f, more = traces.Next()
                    cs = append(cs, f.File+":"+strconv.Itoa(f.Line))
                }
                pool.Put(pcs)
            }
        })
    })
}
/*
结果：
cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
BenchmarkCacheStack
BenchmarkCacheStack/NewStack
BenchmarkCacheStack/NewStack-12    1285654    983.3 ns/op    0 B/op    0 allocs/op
BenchmarkCacheStack/runtime.Callers
BenchmarkCacheStack/runtime.Callers-12    272685    4435 ns/op    1820 B/op    24 allocs/op
*/
```
从测试结果来看，缓存的效果还是很明显的，由 4000ns 多提升到 1000ns 左右。


### 2.2 从调用栈中获取 pc 列表
我们来先看一下 golang 的调用栈,下图出自[曹春晖老师的github文章](https://github.com/cch123/asmshare/blob/master/layout.md#%E6%9F%A5%E7%9C%8B-go-%E8%AF%AD%E8%A8%80%E7%9A%84%E5%87%BD%E6%95%B0%E8%B0%83%E7%94%A8%E8%A7%84%E7%BA%A6) :
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

我们看到调用栈中的两个关键信息----"parent return address" 和 "caller BP(caller frame pointer)"，分别表示当前函数的栈帧的返回 pc 和 栈基地址。

通过这两者的配合，我们就可以获得较完整的调用栈：
```s
// func buildStack(s []uintptr) int
TEXT ·buildStack(SB), NOSPLIT, $24-8
    NO_LOCAL_POINTERS
    MOVQ     cap+16(FP), DX     // s.cap
    MOVQ     p+0(FP), AX        // s.ptr
    MOVQ    $0, CX            // loop.i = 0
loop:
    CMPQ    CX, DX            // if i >= s.cap { return }
    JAE    return                // 无符号大于等于就跳转

    MOVQ    +8(BP), BX        // last pc -> BX
    MOVQ    BX, 0(AX)(CX*8)        // s[i] = BX
    
    ADDQ    $1, CX            // CX++ / i++

    MOVQ    +0(BP), BP         // last BP; 展开调用栈至上一层
    CMPQ    BP, $0             // if (BP) <= 0 { return }
    JA loop                    // 无符号大于就跳转

return:
    MOVQ    CX,n+24(FP)     // ret n
    RET

```
但是，我们知道，golang 为了提高函数调用的性能，会对符合条件的小函数进行内联。所以，我们通过上面的 buildStack 函数，实际上是无法获取完整的调用栈的。

我们可以在[ golang 的源码](https://github.com/golang/go/blob/release-branch.go1.18/src/runtime/traceback.go#L356) 中找到展开内联函数获取 pc 的代码：
```go
// If there is inlining info, record the inner frames.
if inldata := funcdata(f, _FUNCDATA_InlTree); inldata != nil {
    inltree := (*[1 << 20]inlinedCall)(inldata)
    for {
        ix := pcdatavalue(f, _PCDATA_InlTreeIndex, tracepc, &cache)
        if ix < 0 {
            break
        }
        if inltree[ix].funcID == funcID_wrapper && elideWrapperCalling(lastFuncID) {
            // ignore wrappers
        } else if skip > 0 {
            skip--
        } else if n < max {
            (*[1 << 20]uintptr)(unsafe.Pointer(pcbuf))[n] = pc
            n++
        }
        lastFuncID = inltree[ix].funcID
        // Back up to an instruction in the "caller".
        tracepc = frame.fn.entry() + uintptr(inltree[ix].parentPc)
        pc = tracepc + 1
    }
}
```
可以看到，它是在 _FUNCDATA_InlTree 中查找内联函数，并补充到 pc 列表中。我们也可以抄它，不过没必要，
显然 buildStack() 拿到的 pc 列表和 runtime.Callers() 拿到的 pc 列表肯定是一一对应的（如果有不同看法，欢迎来讨论），
所以我们可以利用缓存就能达到这个目的。

利用缓存提升性能的代码如下:
```go
func NewStack2(skip int) (stack []string) {
    pcs := pool.Get().(*[DefaultDepth]uintptr)
    n := buildStack(pcs[:])
    for i := n; i < DefaultDepth; i++ {
        pcs[i] = 0
    }

    stack = cacheStack.Get(*pcs)
    if len(stack) == 0 {
        pcs1 := make([]uintptr, DefaultDepth)
        npc1 := runtime.Callers(2, pcs1[:DefaultDepth])

        stack = parseSlow(pcs1[:npc1])
        cacheStack.Set(*pcs, stack)
    }
    if len(stack)<skip{
        stack=nil
    }else{
        stack = stack[skip:]
    }
    pool.Put(pcs)
    return
}

```
再做个简单的基准测试：
```go
func BenchmarkCacheStack(b *testing.B) {
    b.Run("NewStack", func(b *testing.B) {
        deepCall(10, func() {
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                NewStack(0)
            }
        })
    })
    b.Run("NewStack2", func(b *testing.B) {
        deepCall(10, func() {
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                NewStack2(0)
            }
        })
    })
    b.Run("runtime.Callers", func(b *testing.B) {
        deepCall(10, func() {
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                pcs := pool.Get().(*[DefaultDepth]uintptr)
                n := runtime.Callers(2, pcs[:DefaultDepth])
                var cs []string
                traces, more, f := runtime.CallersFrames(pcs[:n]), true, runtime.Frame{}
                for more {
                    f, more = traces.Next()
                    cs = append(cs, f.File+":"+strconv.Itoa(f.Line))
                }
                pool.Put(pcs)
            }
        })
    })
}
/*
结果：
cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
BenchmarkCacheStack
BenchmarkCacheStack/NewStack
BenchmarkCacheStack/NewStack-12    1287921    932.1 ns/op    0 B/op    0 allocs/op
BenchmarkCacheStack/NewStack2
BenchmarkCacheStack/NewStack2-12    23890822    49.06 ns/op    0 B/op    0 allocs/op
BenchmarkCacheStack/runtime.Callers
BenchmarkCacheStack/runtime.Callers-12    263892    4285 ns/op    1820 B/op    24 allocs/op
*/
```
基准测试的结果还是让人满意的，获取调用栈时间损耗从接近 4000ns 优化到了 50ns 左右了。

至此，提高获取调用栈的性能的目标我们已经达到了。

不过，这种优化方式也是有缺点的。
就是强依赖调用栈上的 BP 信息，而由 [go的下面这段代码](https://github.com/golang/go/blob/release-branch.go1.18/src/runtime/traceback.go#L265) 可知，golang 的 BP 信息只在 AMD64 和 ARM64 平台上才有。实际上在其他平台上，我们也可以通过 "debug/dwarf" 包读取 golang 的 DWARF 信息，从而解析出 pc 到栈帧的映射，不过如果这样做的话，性能可能和 runtime.Callers() 就没多大差别了。
此外，这种优化方式针对 CGO 需要做特殊处理
```go
// For architectures with frame pointers, if there's
// a frame, then there's a saved frame pointer here.
//
// NOTE: This code is not as general as it looks.
// On x86, the ABI is to save the frame pointer word at the
// top of the stack frame, so we have to back down over it.
// On arm64, the frame pointer should be at the bottom of
// the stack (with R29 (aka FP) = RSP), in which case we would
// not want to do the subtraction here. But we started out without
// any frame pointer, and when we wanted to add it, we didn't
// want to break all the assembly doing direct writes to 8(RSP)
// to set the first parameter to a called function.
// So we decided to write the FP link *below* the stack pointer
// (with R29 = RSP - 8 in Go functions).
// This is technically ABI-compatible but not standard.
// And it happens to end up mimicking the x86 layout.
// Other architectures may make different decisions.
if frame.varp > frame.sp && framepointer_enabled {
    frame.varp -= goarch.PtrSize
}

//  src/runtime/runtime2.go
// Must agree with internal/buildcfg.Experiment.FramePointer.
const framepointer_enabled = GOARCH == "amd64" || GOARCH == "arm64"
```

### 2.3 减少 stack 冗余信息，并修改日志结构
针对 [WithStack()](https://github.com/pkg/errors/blob/v0.9.1/errors.go#L143)，可以压缩一下 stack 信息。
 ```go
 var (
    rootDirs = []string{"src/", "/pkg/mod/"} // file 会从rootDir开始截断

    // skipPkgs里的pkg会被忽略
    skipPrefixFiles = []string{
        "github.com/cloudwego/kitex",
        "testing/benchmark.go",
        "testing/testing.go",
    }
)
 func parseSlow(pcs []uintptr) (cs []string) {
    traces, more, f := runtime.CallersFrames(pcs), true, runtime.Frame{}
    for more {
        f, more = traces.Next()
        c := toCaller(f)
        if skipFile(c.File) && len(cs) > 0 {
            break
        }
        cs = append(cs, c.String())
        if strings.HasSuffix(f.Function, "main.main") && len(cs) > 0 {
            break
        }
    }
    return
}
func skipFile(f string) bool {
    for _, skipPkg := range skipPrefixFiles {
        if strings.HasPrefix(f, skipPkg) {
            return true
        }
    }
    return false
}

func toCaller(f runtime.Frame) caller { // nolint:gocritic
    funcName, file, line := f.Function, f.File, f.Line

    i := strings.LastIndex(funcName, pathSeparator)
    if i > 0 {
        rootDir := funcName[:i]
        funcName = funcName[i+1:]
        i = strings.Index(file, rootDir)
        if i > 0 {
            file = file[i:]
        }
    }
    if i <= 0 {
        for _, rootDir := range rootDirs {
            i = strings.Index(file, rootDir)
            if i > 0 {
                i += len(rootDir)
                file = file[i:]
                break
            }
        }
    }

    return caller{
        File: file + ":" + strconv.Itoa(line),
        Func: funcName, 
    }
}
 ```
 对比一下 [pkg/errors](https://github.com/pkg/errors) 的打印信息，可以有个比较直观的感受：
 ```log
// 改进后的：
 88888, error msg;
    (github.com/lxt1045/errors/code_test.go:136) errors.Test_Text.func1.1,
    (github.com/lxt1045/errors/try_catch_test.go:115) errors.deepCall,
    (github.com/lxt1045/errors/try_catch_test.go:118) errors.deepCall,
    (github.com/lxt1045/errors/try_catch_test.go:118) errors.deepCall,
    (github.com/lxt1045/errors/try_catch_test.go:118) errors.deepCall,
    (github.com/lxt1045/errors/code_test.go:135) errors.Test_Text.func1;

// pkg/errors 的：
error msg
github.com/lxt1045/errors.Test_Text.func2.1
    /Users/bytedance/go/src/github.com/lxt1045/errors/code_test.go:142
github.com/lxt1045/errors.deepCall
    /Users/bytedance/go/src/github.com/lxt1045/errors/try_catch_test.go:115
github.com/lxt1045/errors.deepCall
    /Users/bytedance/go/src/github.com/lxt1045/errors/try_catch_test.go:118
github.com/lxt1045/errors.deepCall
    /Users/bytedance/go/src/github.com/lxt1045/errors/try_catch_test.go:118
github.com/lxt1045/errors.deepCall
    /Users/bytedance/go/src/github.com/lxt1045/errors/try_catch_test.go:118
github.com/lxt1045/errors.Test_Text.func2
    /Users/bytedance/go/src/github.com/lxt1045/errors/code_test.go:141
testing.tRunner
    /usr/local/go/src/testing/testing.go:1439
runtime.goexit
    /usr/local/go/src/runtime/asm_amd64.s:1571
 ```
针对 [Wrap()](https://github.com/pkg/errors/blob/v0.9.1/errors.go#L181)，可以只保留一行源码信息即可。

比如：
```go
func TestWrap(t *testing.T) {
    t.Run("NewCode", func(t *testing.T) {
        deepCall(3, func() {
            err := NewErr(1, "error1")
            err = Wrap(err, "error2")
            err = Wrap(err, "error3")
            t.Logf("err:%+v", err)
        })
    })
}
```
以上代码调用栈会格式化成这样：
```log
1, error1;
    (github.com/lxt1045/errors/warpper_test.go:25) errors.TestWrap.func2.1,
    (github.com/lxt1045/errors/try_catch_test.go:115) errors.deepCall,
    (github.com/lxt1045/errors/try_catch_test.go:118) errors.deepCall,
    (github.com/lxt1045/errors/try_catch_test.go:118) errors.deepCall,
    (github.com/lxt1045/errors/try_catch_test.go:118) errors.deepCall,
    (github.com/lxt1045/errors/warpper_test.go:24) errors.TestWrap.func2;
error2,
    (github.com/lxt1045/errors/warpper_test.go:26) errors.TestWrap.func2.1;
error3,
    (github.com/lxt1045/errors/warpper_test.go:27) errors.TestWrap.func2.1;
```
当然，这个日志结构可以自己定义，关键是要减少无效、重复的信息量。

### 2.4 改进 error -> string 的性能
我们先简单测试一下 pkg/errors 生成 string 的性能：
```go
func BenchmarkCacheStack(b *testing.B) {
    b.Run("pkg.Sprintf", func(b *testing.B) {
        deepCall(10, func() {
            err := pkgerrs.WithStack(errors.New("test"))
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                str = fmt.Sprintf("%+v", err)
            }
        })
    })
}
/*
//结果
cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
BenchmarkCacheStack
BenchmarkCacheStack/pkg.Sprintf
BenchmarkCacheStack/pkg.Sprintf-12    132031    8642 ns/op    2202 B/op    20 allocs/op
*/
```
基准测试结果表明，[pkg/errors](https://github.com/pkg/errors) 打印错误栈也造成了巨大的损耗，8us 多。
因此，还是得想办法改善一下。

笔者知道的改善序列化性能的途径只有两个，一个是提高序列化速度，一个是减少内存复制次数。显然在拼接 string 的时候，前者在无法再优化，只能在后者上想想办法。

于是笔者尝试了“引入 buffer”和“提前计算 string 长度”两种方式后，选择了后者。

其实现如下：
```go
type Code struct {
    msg  string //业务错误信息
    code int    //业务错误码

    cache *callers
}

func (e *Code) fmt() (cs fmtCode) {
    return fmtCode{code: strconv.Itoa(e.code), msg: e.msg, callers: e.cache}
}

type callers struct {
    stack []string
}

type fmtCode struct {
    code      string
    msg       string
    *callers
}

func (f *fmtCode) textSize() (l int) {
    l = len(", ") + len(f.code) + len(f.msg)
    if f.callers == nil || len(f.stack) == 0 {
        return
    }
    l += len(f.stack) * len(", \n    ")
    for _, str := range f.stack {
        l += len(str) + 3
    }
    return
}

func (f *fmtCode) text(buf *writeBuffer) {
    buf.WriteString(f.code)
    buf.WriteString(", ")
    buf.WriteString(f.msg)
    if f.callers != nil && len(f.stack) > 0 {
        buf.WriteString(";\n")
        for i, str := range f.stack {
            if i != 0 {
                buf.WriteString(", \n")
            }
            buf.WriteString("    ")
            buf.WriteString(str)
        }
        buf.WriteByte(';')
    }
    return
}

func BenchmarkCaseMarshal(b *testing.B) {
    b.Run("text", func(b *testing.B) {
    err := &Code{
        msg:  "msg",
        code: 1,
        cache: &callers{
            stack: []string{
                "1234567890qwertyuiopasdfghjklzxcvbnm",
                "1234567890qwertyuiopasdfghjklzxcvbnm",
                "1234567890qwertyuiopasdfghjklzxcvbnm",
                "1234567890qwertyuiopasdfghjklzxcvbnm",
                "1234567890qwertyuiopasdfghjklzxcvbnm",
                "1234567890qwertyuiopasdfghjklzxcvbnm",
                "1234567890qwertyuiopasdfghjklzxcvbnm",
                "1234567890qwertyuiopasdfghjklzxcvbnm",
                "1234567890qwertyuiopasdfghjklzxcvbnm",
                "1234567890qwertyuiopasdfghjklzxcvbnm",
            },
        },
    }
    b.Run("text", func(b *testing.B) {
        b.ReportAllocs()
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            buf := &writeBuffer{}
            f := err.fmt()
            buf.Grow(f.textSize())
            f.text(buf)
        }
        b.StopTimer()
    })
    b.Run("fmt", func(b *testing.B) {
        b.ReportAllocs()
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            fmt.Sprintf("code:%d, msg:%s, stack:%v", err.code, err.msg, err.cache.stack)
        }
        b.StopTimer()
    })
    b.Run("bytes.NewBuffer", func(b *testing.B) {
        b.ReportAllocs()
        b.ResetTimer()
        buf := bytes.NewBuffer(nil)
        for i := 0; i < b.N; i++ {
            buf.WriteString("code:")
            buf.WriteString(strconv.Itoa(err.code))
            buf.WriteString("msg:")
            buf.WriteString(err.msg)
            buf.WriteString("stack:")
            for _, str := range err.cache.stack {
                buf.WriteString(str)
            }
        }
        b.StopTimer()
    })
}
/*
测试结果：
cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
BenchmarkCaseMarshal
BenchmarkCaseMarshal/text
BenchmarkCaseMarshal/text-12    4225810    287.4 ns/op    480 B/op    1 allocs/op
BenchmarkCaseMarshal/fmt
BenchmarkCaseMarshal/fmt-12    814053    1494 ns/op    616 B/op    13 allocs/op
BenchmarkCaseMarshal/bytes.NewBuffer
BenchmarkCaseMarshal/bytes.NewBuffer-12    2127382    495.9 ns/op    765 B/op    0 allocs/op
*/
```
由结果来看，当前场景下，性能比 bytes.NewBuffer 方式快了不少，令人意外。


### 2.5 改进 panic(err) 

笔者的想法是，在 panic(err) 之前检查一下该 goroutine 是否已执行过 defer Catch()。
如果未执行过，则不执行 panic(err)。

代码如下：
```go
var (
    mRoutineLastDefer = map[int64]struct{}{}
    lockRoutineDefer  sync.RWMutex

    tryEscapeErr = func(error){
        time.Sleep(time.Second*60)
        runtime.Goexit()  // 退出当前 goroutine
    }
)

type guard struct {
    gid int64
    own bool
    noCopy
}

//go:noinline
func NewGuard() guard {
    gid := Getg()  // 获取 goroutine_id
    lockRoutineDefer.Lock()
    _, ok := mRoutineLastDefer[gid]
    if !ok {
        mRoutineLastDefer[gid] = struct{}{}
    }
    lockRoutineDefer.Unlock()

    return guard{
        gid: gid,
        own: !ok,
    }
}

func Catch(g guard, f func(err interface{}) bool) { //nolint:govet
    if g.own {
        lockRoutineDefer.Lock()
        delete(mRoutineLastDefer, g.gid)
        lockRoutineDefer.Unlock()
    }
    e := recover()
    if e == nil {
        return
    }

    if f != nil && f(e) {
        return
    }
    panic(e)
}

func Try(err *Code) {
    gid := Getg()
    lockRoutineDefer.Lock()
    _, ok := mRoutineLastDefer[gid]
    lockRoutineDefer.Unlock()
    if !ok {
        cs := toCallers([]uintptr{GetPC()[0]})
        e := fmt.Errorf(`should call "defer Catch(NewGuard(),func()bool)" before call "Try(err))"; file:%s`, cs[0].File)
        if err != nil {
            e = fmt.Errorf("%w; %+v", err, e)
        }
        tryEscapeErr(e)
        return
    }
    if err != nil {
        panic(err)
    }

    return
}

// test.go
func Test_Catch(t *testing.T) {
    defer Catch(NewGuard(), func() {
        e := recover()
        if e != nil {
            t.Log("cache:", e)
        }
    })

...

    Try(err)
}

```
原理很简单，就是在执行 defer func(){} 前，先执行 NewGuard()。在 NewGuard() 中给当前 goroutine 打上一个标签。在 Try(err) 函数里， panic(err) 前检查该 goroutine 是否已打标签。如果未打标签则打印错误信息，并退出当前 goroutine。这比因 panic 而导致整个程序退出造成的后果要轻的多。

##  3. 不成功的探索

### 3.1 长跳转提前返回

笔者曾经想实现这样一个接口：
```go
func Do() error {
    tag, err1 := NewTag() // 当 tag.Try(err) 时，跳转此处并返回 err1
    if err1 != nil {
        return
    }

    err := Do2()
    tag.Try(err) //如果 err!=nil 则跳转到 NewTag() 调用处
    ...
}
```
显然，这个接口参考了 go2.0 的错误处理方式：
```go
func Do() error {
    handle err {
        return err
    }
    data := check Dosth()
}
```

实现代码如下：
```s
// func NewTag2() (tag, error)
// func tryJump(retpc, parent_pc uintptr, err error) uintptr
TEXT ·tryJump(SB),NOSPLIT, $0-40
    NO_LOCAL_POINTERS

    MOVQ    (BP), R14         // 获取 parent pc (caller pc)
    MOVQ    +8(R14), R13    
    MOVQ    R13, ret+32(FP)  // 返回 parent pc
    
    CMPQ    pc+8(FP), R13   // parent pc 是否与传入参数相等？
    JE    checkerr
    RET
checkerr:
    CMPQ    err+24(FP), $0 // err == nil ?
    JHI    gototag
    RET
gototag:
    MOVQ    pc+0(FP), CX // retpc -> return addr
    MOVQ    CX, 8(BP)    // caller的返回地址替换为 传入的参数retpc
    MOVQ    CX, 16(BP)   // tag.pc = retpc
    MOVQ    pc+8(FP), CX  
    MOVQ    CX, 24(BP)  // tag.parent = parent_pc
    MOVQ    pc+16(FP), CX 
    MOVQ    CX, 32(BP)  // 设置caller的返回值err
    MOVQ    pc+24(FP), CX 
    MOVQ    CX, 40(BP)  // 设置caller的返回值err
    RET



// func NewTag() (tag, error)
TEXT ·NewTag(SB),NOSPLIT,$32-24
    NO_LOCAL_POINTERS
    MOVQ    $0, ret+0(FP)  // 返回值清零
    MOVQ    $0, ret+8(FP)
    MOVQ    $0, ret+16(FP)
    GO_RESULTS_INITIALIZED
    MOVQ    pc-8(FP), R13  // pc
    MOVQ    R13, ret+0(FP) // tag.pc = pc

    MOVQ    (BP), R14       // tag.parent = parent_pc
    MOVQ    +8(R14), R13
    MOVQ    R13, ret+8(FP)

    RET
```
```go

var tryTagErr func(error)

func NewTag() (tag, error)

func tryJump(pc, parent uintptr, err error) uintptr

type tag struct {
    pc     uintptr
    parent uintptr
}

//go:noinline
func (t tag) Try(err error) {
    //还是要加上检查，否则报错信息太难看
    // 但是检查时只要检查 更上一级的 PC 是否相等即可，不需要复杂的 map 存储了！！！
    parent := tryJump(t.pc, t.parent, err)
    if parent != t.parent {
        cs := toCallers([]uintptr{parent, t.parent, GetPC()[0]})
        e := fmt.Errorf("tag.Try() should be called in [%s] not in [%s]; file:%s",
            cs[1].Func, cs[0].Func, cs[2].File)
        if err != nil {
            e = fmt.Errorf("%w; %+v", err, e)
        }
        if tryTagErr != nil {
            tryTagErr(e)
            return
        }
        panic(e)
    }
}

```

这里的实现原理也很简单，就是在 NewTag() 函数里，把 NewTag() 函数的返回地址(pc)记录到 tag.pc 中。后续执行 tag.Try(err) 时，如果 err!=nil 就将其返回地址替换为 tag.pc，这样就跳转到了 NewTag() 的返回处继续执行。由于此时 err1!=nil，自然就进入 if 逻辑并执行 return 让函数提前返回。

但是，这样虽然能保证按预计逻辑执行，却有一个比较致命的问题，就是 debug 模式和 release 模式下，defer 函数的执行预期不一致。

比如下面的例子：
```go
func Do() error {
    defer func() {
        fmt.Printf("2")
    }()
    tag, err1 := NewTag() // 当 tag.Try(err) 时，跳转此处并返回 err1
    if err1 != nil {
        return
    }
    defer func() {
        fmt.Printf("1")  // release 模式下不执行
    }()
    err := errors.New("err")
    tag.Try(err)          // 在此处执行跳转
}
```
debug 模式下打印 "12"，relese 模式下只打印 "1"。
原因是 debug 模式下， golang 编译器没有对 defer 函数做优化，会在执行到 defer 时，将 defer 函数挂到 g._defer 链表上。
而 release 模式下，golang 编译器会将 defer 函数内联到 return（即汇编的 RET）命令前。
我们通过汇编将 tag.Try(err) 的返回地址替换为 NewTag() 的返回地址，却并没有将 NewTag() 之后的 defer 函数添加到 return 命令前，所以 NewTag() 之后的 defer 函数并能被执行。

不过，也有办法处理，就是比较”反人类“。
像这样：
```go
func Do() error {
gototag:
    defer func() {
        fmt.Printf("2")
    }()
    tag, err1 := NewTag() // 当 tag.Try(err) 时，跳转此处并返回 err1
    if err1 != nil {
        return
    }
    defer func() {
        fmt.Printf("1")
    }()
    err := errors.New("err")
    tag.Try(err)

    return
    goto gototag
}
```
上面代码的做法，就是把函数体用 goto tag 包裹起来，这样 golang 编译器就会取消 defer 的内联优化，把 defer 函数挂到 g._defer 链表上，从而使代码在 release 和 debug 下得到一致结果。
如果这样做，这个接口就太不友善了，所以这只是一个“失败的探索”

### 3.2 检查 g._defer 链表

同样的原因导致的“失败探索”，还有下面这个例子:
```s
// stack.asm
// func GetDefer() *_defer
TEXT ·GetDefer(SB), NOSPLIT, $0-8
    MOVQ (TLS), AX
    ADDQ ·g__defer_offset(SB),AX
    MOVQ (AX), BX
    MOVQ BX, ret+0(FP)
    RET

// func getgi() interface{}
TEXT ·getgi(SB), NOSPLIT, $32-16
    NO_LOCAL_POINTERS

    MOVQ $0, ret_type+0(FP)
    MOVQ $0, ret_data+8(FP)
    GO_RESULTS_INITIALIZED

    // get runtime.g
    MOVQ (TLS), AX

    // get runtime.g type
    MOVQ $type·runtime·g(SB), BX

    // return interface{}
    MOVQ BX, ret_type+0(FP)
    MOVQ AX, ret_data+8(FP)
    RET

```
```go
// stack.go
func GetDefer() *_defer
func getgi() interface{}

var g__defer_offset uintptr = func() uintptr {
    g := getgi()
    if f, ok := reflect.TypeOf(g).FieldByName("_defer"); ok {
        return f.Offset
    }
    panic("can not find g.goid field")
}()

type _defer struct {
    started bool
    heap    bool
    openDefer bool
    sp        uintptr // sp at time of defer
    pc        uintptr // pc at time of defer
    fn        func()  // can be nil for open-coded defers
    _panic    uintptr
    link      *_defer // next defer on G; can point to either heap or stack!

    fd   unsafe.Pointer
    varp uintptr    
    framepc uintptr
}

func (d *_defer) Next() *_defer {
    return d.link
}
func (d *_defer) PC() uintptr {
    return d.pc
}

```
```go
//main.go
func main() {
    GetDefer()
}
func GetDefer() {
    defer func() {
        log.Println("defer 1")
    }()
    d := stack.GetDefer()
    for ; d != nil; d = d.Next() {
        log.Printf("defer pc:%d", d.PC())
    }
    return
}

```
以上代码，本意是想通过 g._defer 链表，检查 panic(err) 前是否有执行过 defer Catch() （检查代码未写出）。
不过也是由于 golang 编译器对 defer 的优化，导致无法简单的获取到正确的 g._defer 链表。

## 4. 附录
根据以上优化想法，笔者写了一个 errors 库 [lxt1045/errors](https://github.com/lxt1045/errors)，当前还只是一个玩具，欢迎来吐槽。

先来看一下调用栈输出格式：
```go
package errors

import (
    "errors"
    "fmt"
    "runtime"
    "strconv"
    "testing"

    lxterrs "github.com/lxt1045/errors"
    pkgerrs "github.com/pkg/errors"
)

var str string

func TestErrors(t *testing.T) {
    t.Run("lxt1045/errors", func(t *testing.T) {
        deepCall(3, func() {
            err := lxterrs.NewErr(1, "test")
            str = fmt.Sprintf("%+v", err)
            t.Log(str)
        })
    })
    t.Run("lxt1045/errors", func(t *testing.T) {
        deepCall(3, func() {
            err := lxterrs.NewErr(1, "test")
            str = err.Error()
            t.Log(str)
        })
    })
    t.Run("pkg/errors", func(t *testing.T) {
        deepCall(3, func() {
            err := errors.New("test")
            err = pkgerrs.WithStack(err)
            str = fmt.Sprintf("%+v", err)
            t.Log(str)
        })
    })
}
/*
输出日志：
=== RUN   TestErrors
=== RUN   TestErrors/lxt1045/errors
    /Users/bytedance/go/src/github.com/lxt1045/blog/sample/errors/errors_test.go:114: 1, test;
            (github.com/lxt1045/blog/sample/errors/errors_test.go:112) errors.TestErrors.func1.1,
            (github.com/lxt1045/blog/sample/errors/errors_test.go:29) errors.deepCall,
            (github.com/lxt1045/blog/sample/errors/errors_test.go:32) errors.deepCall,
            (github.com/lxt1045/blog/sample/errors/errors_test.go:32) errors.deepCall,
            (github.com/lxt1045/blog/sample/errors/errors_test.go:32) errors.deepCall,
            (github.com/lxt1045/blog/sample/errors/errors_test.go:111) errors.TestErrors.func1;
=== RUN   TestErrors/lxt1045/errors#01
    /Users/bytedance/go/src/github.com/lxt1045/blog/sample/errors/errors_test.go:121: 1, test;
            (github.com/lxt1045/blog/sample/errors/errors_test.go:119) errors.TestErrors.func2.1,
            (github.com/lxt1045/blog/sample/errors/errors_test.go:29) errors.deepCall,
            (github.com/lxt1045/blog/sample/errors/errors_test.go:32) errors.deepCall,
            (github.com/lxt1045/blog/sample/errors/errors_test.go:32) errors.deepCall,
            (github.com/lxt1045/blog/sample/errors/errors_test.go:32) errors.deepCall,
            (github.com/lxt1045/blog/sample/errors/errors_test.go:118) errors.TestErrors.func2;
=== RUN   TestErrors/pkg/errors
    /Users/bytedance/go/src/github.com/lxt1045/blog/sample/errors/errors_test.go:129: test
        github.com/lxt1045/blog/sample/errors.TestErrors.func3.1
            /Users/bytedance/go/src/github.com/lxt1045/blog/sample/errors/errors_test.go:127
        github.com/lxt1045/blog/sample/errors.deepCall
            /Users/bytedance/go/src/github.com/lxt1045/blog/sample/errors/errors_test.go:29
        github.com/lxt1045/blog/sample/errors.deepCall
            /Users/bytedance/go/src/github.com/lxt1045/blog/sample/errors/errors_test.go:32
        github.com/lxt1045/blog/sample/errors.deepCall
            /Users/bytedance/go/src/github.com/lxt1045/blog/sample/errors/errors_test.go:32
        github.com/lxt1045/blog/sample/errors.deepCall
            /Users/bytedance/go/src/github.com/lxt1045/blog/sample/errors/errors_test.go:32
        github.com/lxt1045/blog/sample/errors.TestErrors.func3
            /Users/bytedance/go/src/github.com/lxt1045/blog/sample/errors/errors_test.go:125
        testing.tRunner
            /usr/local/go/src/testing/testing.go:1439
        runtime.goexit
            /usr/local/go/src/runtime/asm_amd64.s:1571
*/
```
可以看到，相对于 [pkg/errors](https://github.com/pkg/errors) 来说，日志格式还是有所改善的。

再做个简单的基准测试：
```go
var str string
func BenchmarkErrors(b *testing.B) {
    b.Run("lxt1045/errors-fmt", func(b *testing.B) {
        deepCall(10, func() {
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                err := lxterrs.NewErr(1, "test")
                str = fmt.Sprintf("%+v", err)
            }
        })
    })
    b.Run("lxt1045/errors-Errors", func(b *testing.B) {
        deepCall(10, func() {
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                err := lxterrs.NewErr(1, "test")
                str = err.Error()
            }
        })
    })
    b.Run("pkg/errors", func(b *testing.B) {
        deepCall(10, func() {
            err := errors.New("test")
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                err := pkgerrs.WithStack(err)
                str = fmt.Sprintf("%+v", err)
            }
        })
    })
}
/*
测试结果：
cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
BenchmarkErrors
BenchmarkErrors/lxt1045/errors-fmt
BenchmarkErrors/lxt1045/errors-fmt-12    1642436    712.3 ns/op    2338 B/op    3 allocs/op
BenchmarkErrors/lxt1045/errors-Errors
BenchmarkErrors/lxt1045/errors-Errors-12    2753948    452.8 ns/op    1184 B/op    2 allocs/op
BenchmarkErrors/pkg/errors
BenchmarkErrors/pkg/errors-12    113755    9030 ns/op    2515 B/op    24 allocs/op
*/
```
可以看到，就能能来说，相对于 [pkg/errors](https://github.com/pkg/errors)，有接近 10~20倍 的差距，提升还是很明显的。
