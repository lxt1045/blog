

go语言的汇编可以为我们做什么呢？

# 0、汇编有何用
 Go 的汇编有什么用呢？能用在什么场景？

 0.1、 算法加速，golang编译器为了编译速度，放弃了很多指令优化，远比不上C/C++的gcc/clang生成的优化程度高，甚至和 Java 比都有差距，毕竟时间沉淀在那里。因此某些场景下某些优化逻辑、CPU指令可以让我们的算法运行速度更快，如 SIMD 指令。

 0.2、 摆脱golang编译器的一些约束，如通过汇编[调用其他package的私有函数](https://sitano.github.io/2016/04/28/golang-private/)。

 0.3、 进行一些hack的事，如通过汇编适配其他语言的ABI来直接调用其他语言的函数:[petermattis/fastcgo](https://github.com/petermattis/fastcgo)。
 

# 1、 go如何使用汇编
参考文档：

柴树杉和曹春晖的书[《Go语言高级变成》第三章](ssssshttps://github.com/chai2010/advanced-go-programming-book/tree/master/ch3-asm)

[Go 语言汇编的官方文档](https://go.dev/doc/asm)

柴树杉的两篇文章 [plan9 assembly 完全解析](https://github.com/cch123/golang-notes/blob/master/assembly.md) 和 [汇编分享](https://github.com/cch123/asmshare/blob/master/layout.md)

[GoFunctionsInAssembly](https://lrita.github.io/images/posts/go/GoFunctionsInAssembly.pdf)

[A Quick Guide to Go's Assembler](https://go.dev/doc/asm)

欧长坤老师的书[《Go语言原本》1.4 Plan 9 汇编语言](https://golang.design/under-the-hood/zh-cn/part1basic/ch01basic/asm/)



我们知道 Go 语言一些核心成员是 Plan 9 的遗老遗少，而且属于比较高傲的的学院派，这导致 Go 语言的汇编采用了令人抓狂的 Plan 9 风格。

## 1.1、通用寄存器
[go语言调度器源代码情景分析之二：CPU寄存器](https://www.cnblogs.com/abozhang/p/10766689.html)

不同体系结构的CPU，其内部寄存器的数量、种类以及名称可能大不相同，这里我们只介绍 AMD64 的寄存器。AMD64 有20多个可以直接在汇编代码中使用的寄存器，其中有几个寄存器在操作系统代码中才会见到，而应用层代码一般只会用到如下分为三类的19个寄存器。

| 寄存器类型 | 用途 | 寄存器 |位宽|
| :---- | :---- | :---- |:----|
|通用寄存器|有做特殊规定，程序员和编译器可以自定义其用途（下面会介绍，rsp/rbp寄存器其实是有特殊用途的）|rax, rbx, rcx, rdx, rsi, rdi, rbp, rsp, r8, r9, r10, r11, r12, r13, r14, r15|64bit|
|程序计数寄存器|IP 寄存器（Go 中叫 PC 寄存器）用来存放下一条即将执行的指令的地址，这个寄存器决定了程序的执行流程|rip|64bit|
|段寄存器|cs、ds、es和ss在 AMD64 中已不用；fs 和 gs 一般用它来实现线程本地存储（TLS），比如AMD64 linux平台下go语言使用fs寄存器来实现系统线程的TLS。|fs, gs, cs, ds, es, ss|16bit|

上述这些寄存器除了fs和gs段寄存器是16位的，其它都是64位的，也就是8个字节，其中的16个通用寄存器还可以作为32/16/8位寄存器使用，只是使用时需要换一个名字，比如可以用eax这个名字来表示一个32位的寄存器，它使用的是rax寄存器的低32位。

AMD64的通用通用寄存器的名字在 plan9 中的对应关系:

|AMD64|rax|rbx|rcx|rdx|rdi|rsi|rbp|rsp|r8|r9|r10|r11|r12|r13|r14|rip|
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
|Plan9|AX|BX|CX|DX|DI|SI|BP|SP|R8|R9|R10|R11|R12|R13|R14|PC|

Go语言使用寄存器常用规则：
| 助记符 | 名字 | 用途 |
| :---- | :---- | :---- |
|AX| 累加寄存器(AccumulatorRegister) | 用于存放数据，包括算术、操作数、结果和临时存放地址 |
|BX|基址寄存器(BaseRegister)|用于存放访问存储器时的地址|
|CX|计数寄存器(CountRegister)|用于保存计算值，用作计数器|
|DX|数据寄存器(DataRegister)|用于数据传递，在寄存器间接寻址中的I/O指令中存放I/O端口的地址|
|BP|堆栈基指针(BasePointer)|保存在进入函数前的栈顶基址|
|SI|源变址寄存器(SourceIndex)|用于存放源操作数的偏移地址|
|DI|目的寄存器(DestinationIndex)|用于存放目的操作数的偏移地址|
<!-- 
|SP|堆栈顶指针(StackPointer)|如果是symbol+offset(SP)的形式表示go汇编的伪寄存器；如果是offset(SP)的形式表示硬件寄存器|
|SB|静态基指针(StaticBasePointer)|go汇编的伪寄存器。foo(SB)用于表示变量在内存中的地址，foo+4(SB)表示foo起始地址往后偏移四字节。一般用来声明函数或全局变量|
|FP|栈帧指针(FramePointer)|go汇编的伪寄存器。引用函数的输入参数，形式是symbol+offset(FP)，例如arg0+0(FP)| 
-->

## 1.2、伪寄存器

| 助记符 | 说明 |
| :---- | :---- | 
|FP(Frame pointer)|arguments and locals|
|PC(Program counter)|jumps and branches|
|SB(Static base pointer)|global symbols|
|SP(Stack pointer)|top of stack|

伪寄存器是plan9伪汇编中的一个助记符, Plan9 比较反人类的语法之一。\
SB：伪寄存器，可以理解为原始内存，foo(SB)的意思是用foo来代表内存中的一个地址。foo(SB)可以用来定义全局的function和数据，foo<>(SB)表示foo只在当前文件可见，跟C中的static效果类似。此外可以在引用上加偏移量，如foo+4(SB)表示foo+4bytes的地址。\
PC：PC寄存器，在x86下对应IP寄存器，amd64上则是RIP。

伪寄存器SP和硬件SP不是一回事。\
比如：\
a+0(SP) 表示伪寄存器SP；a是一个标记符(symbol)\
 +8(SP) 则表示硬件寄存器SP 

在编译和反汇编的结果中，只有真实的SP寄存器。 \
即 go tool objdump/go tool compile -S输出的代码中，是没有伪SP和FP寄存器的，上面说的区分伪SP 和硬件SP寄存器的方法对此没法使用的。


SP：伪栈指针，通过symbol+offset(SP)使用，指向局部变量的起始位置(高地址处)；x-8(SP) 表示第一个本地变量；\
物理SP、硬件SP：真实栈顶地址（栈帧最低地址处），通过+offset(SP)使用，修改物理SP，会引起伪SP、FP同步变化，比如：
    SUBQ $16, SP // 伪SP/FP都会-16

FP：伪寄存器，用来标识函数参数、返回值，被调用者（callee）的FP实际上是调用者（caller）的栈帧；x+0(FP)表示第一个请求参数(参数返回值从又到左入栈)；
关系：callee.SP(物理SP)==caller.FP；\
非NOSPLIT：伪FP=伪SP+16；NOSPLIT（framepointer_enable?有无BP栈指针？）：伪FP=伪SP+8；

另外还有两个比较特殊的伪寄存器：|
BP：栈帧的起始位置(最高地址处，SP表示最低地址处)；\
TLS：存储当前 goroutine 的 g 结构地址。\
FP和SP是根据当前函数栈空间计算出来的一个相对于物理寄存器SP的一个偏移量坐标。当在一个函数中，如果用户手动修改了物理寄存器SP的偏移，则伪寄存器FP和SP也随之发生对应的偏移。

```
  高地址 +------------------+
        |                  |
        |       内核        |
        |                  |
        --------------------
        |                  |
        |       栈         |
        |                  |
        --------------------
        |                  |
        |     .......      |
        |                  |
        --------------------
        |                  |
        |       堆         |
        |                  |
        --------------------
        |      全局数据     |
        |------------------|
        |                  |
        |       代码        |
        |                  |
        |------------------|
        |     系统保留      |
  低地址 |------------------|    

```

## 1.3、函数调用栈帧
我们先了解几个名词。

caller：函数调用者。\
callee：函数被调用者。 \
比如函数main中调用sum函数，那么main就是caller，而sum函数就是callee。

栈帧：\
栈帧即stack frame，即未完成函数所持有的，独立连续的栈区域，用来保存其局部变量，返回地址等信息。

函数栈：\
当前函数作为caller，其本身拥有的栈帧以及其所有callee的栈帧，可以称为该函数的函数栈。一般情况下函数栈大小是固定的，如果 超出栈空间，就会栈溢出异常。比如递归求斐波拉契，这时候可以使用尾调用来优化。用火焰图分析性能时候，火焰越高，说明栈越深。

我们知道协程分为有栈协程和无栈协程，go语言是有栈协程。那你知道普通gorutine的调用栈在哪个内存区吗？

下图是 golang 的调用栈，出自[曹春晖老师的github文章](https://github.com/cch123/asmshare/blob/master/layout.md#%E6%9F%A5%E7%9C%8B-go-%E8%AF%AD%E8%A8%80%E7%9A%84%E5%87%BD%E6%95%B0%E8%B0%83%E7%94%A8%E8%A7%84%E7%BA%A6) :
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
       |   SP(Real Register) ->  ----------------------------------------------+   FP(virtual register)                       
       |                         |                  |                          |                                              
       |                         |   return addr    |  parent return address   |                                              
       +---------------------->  +------------------+---------------------------    <-----------------------+         
                                                    |  caller BP               |                            |         
                                                    |  (caller frame pointer)  |                            |         
                                     BP(pseudo SP)  ----------------------------                            |         
                                                    |                          |                            |         
                                                    |     Local Var0           |                            |         
                                                    ----------------------------                            |         
                                                    |                          |                                      
                                                    |     Local Var1           |                                      
                                                    ----------------------------                    callee stack frame
                                                    |                          |                                      
                                                    |       .....              |                                      
                                                    ----------------------------                            |         
                                                    |                          |                            |         
                                                    |     Local VarN           |                            |         
     High                         SP(Real Register) ----------------------------                            |         
      ^                                             |                          |                            |         
      |                                             |                          |                            |         
      |                                             |                          |                            |         
      |                                             |                          |                            |         
      |                                             |                          |                            |         
      |                                             +--------------------------+    <-----------------------+         
     Low                                                                                                                      
                                                              callee
```
上图指示出了 FP 和 SP 的指向。

需要指出的是:\
CALLER BP:在编译期由编译器在符合条件时自动插入。所以手写汇编时，计算framesize时不包括CALLER BP的空间;
是否插入CALLER BP的主要判断依据是：函数的栈帧大小大于0; 常量 [framepointer_enabled](https://github.com/golang/go/blob/e4435cb8448514d2413f9d9aa3ee40738d26fd67/src/runtime/runtime2.go#L1190) 值为 true。
```go
// Must agree with internal/buildcfg.FramePointerEnabled.
const framepointer_enabled = GOARCH == "amd64" || GOARCH == "arm64"
```

[参考 go 源码](https://github.com/golang/go/blob/969bea8d59daa6bdd478b71f6e99d8b8f625a140/src/runtime/traceback.go#L281)
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
```
TLS伪寄存器
该寄存器存储当前goroutine g结构地址



## 1.3、常用指令
以下来之欧长坤的《go 语言原本》
```md
运行时协调
为保证垃圾回收正确运行，在大多数栈帧中，运行时必须知道所有全局数据的指针。 Go 编译器会将这部分信息耦合到 Go 源码文件中，但汇编程序必须进行显式定义。

被标记为 NOPTR 标志的数据符号会视为不包含指向运行时分配数据的指针。 带有 R0DATA 标志的数据符号在只读存储器中分配，因此被隐式标记为 NOPTR。 总大小小于指针的数据符号也被视为隐式标记 NOPTR。 在一份汇编源文件中是无法定义包含指针的符号的，因此这种符号必须定义在 Go 源文件中。 一个良好的经验法则是 R0DATA 在 Go 中定义所有非符号而不是在汇编中定义。

每个函数还需要注释，在其参数，结果和本地堆栈框架中给出实时指针的位置。 对于没有指针结果且没有本地堆栈帧或没有函数调用的汇编函数， 唯一的要求是在同一个包中的 Go 源文件中为函数定义 Go 原型。 汇编函数的名称不能包含包名称组件 （例如，syscall 包中的函数 Syscall 应使用名称 ·Syscall 而不是 syscall·Syscall 其 TEXT 指令中的等效名称）。 对于更复杂的情况，需要显式注释。 这些注释使用标准 #include 文件中定义的伪指令 funcdata.h。

如果函数没有参数且没有结果（返回值），则可以省略指针信息。这是由一个参数大小 $n-0 注释指示 TEXT 对指令。 否则，指针信息必须由 Go 源文件中的函数的 Go 原型提供，即使对于未直接从 Go 调用的汇编函数也是如此。 （原型也将 go vet 检查参数引用。）在函数的开头，假定参数被初始化但结果假定未初始化。 如果结果将在调用指令期间保存实时指针，则该函数应首先将结果归零， 然后执行伪指令 GO_RESULTS_INITIALIZED。 此指令记录结果现在已初始化，应在堆栈移动和垃圾回收期间进行扫描。 通常更容易安排汇编函数不返回指针或不包含调用指令; 标准库中没有汇编函数使用 GO_RESULTS_INITIALIZED。

如果函数没有本地堆栈帧，则可以省略指针信息。这由 TEXT 指令上的本地帧大小 $0-n 注释表示。如果函数不包含调用指令，也可以省略指针信息。否则，本地堆栈帧不能包含指针，并且汇编必须通过执行伪指令 TEXTNO_LOCAL_POINTERS 来确认这一事实。因为通过移动堆栈来实现堆栈大小调整，所以堆栈指针可能在任何函数调用期间发生变化：甚至指向堆栈数据的指针也不能保存在局部变量中。

汇编函数应始终给出 Go 原型，既可以提供参数和结果的指针信息，也可以 go vet 检查用于访问它们的偏移量是否正确。
```


[golang-汇编](https://blog.djaigo.com/golang/golang-hui-bian.html)
text
 静态基地址(static-base) 指针
                        |    标签
                        |    |      add函数入参+返回值总大小
                        |    |      |
TEXT pkgname·funcname<>(SB),TAG,$16-24
     |       |       |           |
函数所属包名  函数名    |           add函数栈帧大小
                     表示堆栈是否符合GOABI

TEXT，定义函数标识
pkgname，函数包名，可以不写，也可以""替代
·，在程序链接后会转换为.
<>，表示堆栈结构是否符合GOABI，也可以写作<ABIInternal>
(SB)，让SB认识这个函数，即生成全局符号，有用绝对地址
TAG，标签，表示函数某些特殊功能，多个标签可以通过|连接，常用标签
NOSPLIT，向编译器表明，不应该插入stack-split的用来检查栈需要扩张的前导指令，减少开销，一般用于叶子节点函数（函数内部不调用其他函数）
NOFRAME，不分配函数堆栈，函数必须是叶子节点函数，且以0标记堆栈函数，没有保存帧指针（或link寄存器架构上的返回地址）
$16-24，16表示函数栈帧大小，24表示入参和返回大小
伪SP和FP的相对位置是会变的，所以不应该尝试用伪SP寄存器去找那些用FP+offset来引用的值，例如函数的 入参和返回值。

X86 和 AMD64 常用指令：（相对常用的 AT&T 格式有些差异，不过大同小异。）
[Go支持的X86指令](https://github.com/golang/arch/blob/v0.2.0/x86/x86.csv)
[Go支持的ARM64指令](https://github.com/golang/arch/blob/v0.2.0/arm64/arm64asm/inst.json)
[Go支持的ARM指令](https://github.com/golang/arch/blob/v0.2.0/arm/arm.csv)
SIMD： 

源码文件规则：

  1、类似 *_test.go，通过添加平台后缀区分，比如: asm_386.s、asm_amd64.s、asm_arm.s、asm_arm64.s、asm_mips64x.s、asm_linux_amd64.s、asm_bsd_arm.s 等

  2、通过注释区分平台和编译器版本
```asm
//go:build (darwin || freebsd || netbsd || openbsd) && gc
// +build darwin freebsd netbsd openbsd
// +build gc
```

Go 1.17之前，我们可以通过在源码文件头部放置+build构建约束指示符来实现构建约束，但这种形式十分易错，并且它并不支持&&和||这样的直观的逻辑操作符，而是用逗号、空格替代，下面是原+build形式构建约束指示符的用法及含义：
| build tags | 含义 |
| :---- | :---- |
| // +build tag1 tag2 | tag1 OR tag2 |
| // +build tag1,tag2 | tag1 AND tag2 |
| // +build !tag1 | NOT tag1 |
| // +build tag1 <br> // +build tag2 <br /> | tag1 AND tag2 |
| // +build tag1,tag2 tag3,!tag4 | (tag1 AND tag2) OR (tag3 AND (NOT tag4)) |

Go 1.17 引入了
[//go:build形式的构建约束指示符](https://go.googlesource.com/proposal/+/master/design/draft-gobuild.md)，支持&&和||逻辑操作符，如下代码所示：
```go
//go:build linux && (386 || amd64 || arm || arm64 || mips64 || mips64le || ppc64 || ppc64le)
//go:build linux && (mips64 || mips64le)
//go:build linux && (ppc64 || ppc64le)
//go:build linux && !386 && !arm
```
考虑到兼容性，Go命令可以识别这两种形式的构建约束指示符，但推荐Go 1.17之后都用新引入的这种形式。

gofmt可以兼容处理两种形式，处理原则是：如果一个源码文件只有// +build形式的指示符，gofmt会将与其等价的//go:build 行加入。否则，如果一个源文件中同时存在这两种形式的指示符行，那么//+build行的信息将被//go:build行的信息所覆盖。


全局变量：
[柴树杉和曹春晖的树《Go语言高级编程》的章节3.3 常量和全局变量](https://github.com/chai2010/advanced-go-programming-book/blob/master/ch3-asm/ch3-03-const-and-var.md)
```asm
// GLOBL 指令声明一个变量对应的符号，以及变量对应的内存大小
GLOBL symbol(SB), flag, width  // 名为 symbol, 内存宽度为 width, flag可省略

// DATA 汇编指令指定对应内存中的数据; width 必须是 1、2、4、8 几个宽度之一
DATA    symbol+offset(SB)/width, value // symbol+offset 偏移量，width 宽度, value 初始值
```
例子：
```asm
DATA age+0x00(SB)/4, $18  // forever 18
GLOBL age(SB), RODATA, $4

DATA pi+0(SB)/8, $3.1415926
GLOBL pi(SB), RODATA, $8

DATA bio<>+0(SB)/8, $"hello wo"
DATA bio<>+8(SB)/8, $"old !!!!"  // <> 表示只在当前文件生效
GLOBL bio<>(SB), RODATA, $16     // bio = "hello world !!!!"
```
flag 的类型有有如下几个：\
当使用这些 flag 的字面量时，需要在汇编文件中 #include "textflag.h"。
| flag | value | 说明 |
| :-- | :-- |  :-- | 
|NOPROF|1| (For TEXT items.) Don't profile the marked function. This flag is deprecated. |
|DUPOK|2| It is legal to have multiple instances of this symbol in a single binary. The linker will choose one of the duplicates to use. |
|NOSPLIT|4| (For TEXT items.) Don't insert the preamble to check if the stack must be split. The frame for the routine, plus anything it calls, must fit in the spare space at the top of the stack segment. Used to protect routines such as the stack splitting code itself. |
|RODATA|8| (For DATA and GLOBL items.) Put this data in a read-only section. |
|NOPTR|16| (For DATA and GLOBL items.) This data contains no pointers and therefore does not need to be scanned by the garbage collector. |
|WRAPPER|32| (For TEXT items.) This is a wrapper function and should not count as disabling recover. |
|NEEDCTXT|64| (For TEXT items.) This function is a closure so it uses its incoming context register. |
|TLSBSS|256| Allocate a word of thread local storage and store the offset from the thread local base to the thread local storage in this variable. |
|NOFRAME|512| Do not insert instructions to allocate a stack frame for this function. Only valid on functions that declare a frame size of 0. TODO(mwhudson): only implemented for ppc64x at present. |
|REFLECTMETHOD|1024| Function can call reflect.Type.Method or reflect.Type.MethodByName. |
|TOPFRAME|2048| Function is the outermost frame of the call stack. Call stack unwinders should stop at this function. |
|ABIWRAPPER|4096| Function is an ABI wrapper. |

<!-- 
| flag | value | 说明 |
| :-- | :-- |  :-- | 
|NOPROF|1| (TEXT项使用) 不优化NOPROF标记的函数。这个标志已废弃。 |
|DUPOK|2| 在二进制文件中允许一个符号的多个实例。链接器会选择其中之一。 |
|NOSPLIT|4| (TEXT项使用) 不插入预先检测是否将栈空间分裂的代码。程序的栈帧中，如果调用任何其他代码都会增加栈帧的大小，必须在栈顶留出可用空间。用来保护处理栈空间分裂的代码本身。 |
|RODATA|8| (DATA和GLOBAL项使用) 将这个数据放在只读的块中。 |
|NOPTR|16| （用于DATA和GLOBL项目）这个数据不包含指针所以就不需要垃圾收集器来扫描。 |
|WRAPPER|32| （对于TEXT项。）这是包装函数，不应算作禁用recover。|
|NEEDCTXT|64| （对于TEXT项。）此函数是一个闭包，因此它将使用其传入的上下文寄存器。|
|LOCAL|128| 此符号位于动态共享库的本地。|
|TLSBSS|256| （用于DATA和GLOBL项目。）将此数据放入线程本地存储中。|
|NOFRAME|512| （对于TEXT项。）即使这不是叶函数，也不要插入指令来分配堆栈帧并保存/恢复返回地址。仅在声明帧大小为0的函数上有效。|
|TOPFRAME|64| （对于TEXT项。）函数是调用堆栈的顶部。回溯应在此功能处停止。| 
-->
其中 NOSPLIT 表示该函数运行不会导致栈分裂，用户也可以使用 //go:nosplit 强制给函数指定NOSPLIT属性。\
例如：
```asm
// 表示函数执行时最多需要 24 字节本地变量和 8 字节参数空间
TEXT ·buildStack(SB), NOSPLIT, $24-8  
    RET
```
```go
//go:nosplit
func someFunction() {
}
```
标记为NOSPLIT的函数，链接器知道该函数最多需要使用StackLimit字节空间，所以不需要栈分裂(溢出)检查，提高性能。不过，使用该标志的时候要特别小心，万一发生意外，容易导致栈溢出错误。\
但如果函数实际真的溢出了，则会在编译期就报错nosplit stack overflow。\
[go/src/runtime/HACKING.md](https://github.com/golang/go/blob/go1.18.10/src/runtime/HACKING.md#nosplit-functions)

另外，当函数处于调用链的叶子节点，且栈帧小于StackSmall（128）字节时，则自动标记为NOSPLIT。 代码如下：\
 [src/cmd/internal/obj/x86/obj6.go](/usr/local/go/src/cmd/internal/obj/x86/obj6.go:620)
```go
    //const StackSmall  = 128
	if ctxt.Arch.Family == sys.AMD64 && autoffset < objabi.StackSmall && !p.From.Sym.NoSplit() {
		leaf := true
	LeafSearch:
		for q := p; q != nil; q = q.Link {
			switch q.As {
			case obj.ACALL:
				// Treat common runtime calls that take no arguments
				// the same as duffcopy and duffzero.
				if !isZeroArgRuntimeCall(q.To.Sym) {
					leaf = false
					break LeafSearch
				}
				fallthrough
			case obj.ADUFFCOPY, obj.ADUFFZERO:
				if autoffset >= objabi.StackSmall-8 {
					leaf = false
					break LeafSearch
				}
			}
		}

		if leaf {
			p.From.Sym.Set(obj.AttrNoSplit, true)
		}
	}

```
一些示例:

在汇编代码中使用 go 变量：
```go
package main

var a = 999
func get() int

func main() {
    println(get())
}
```
```asm
#include "textflag.h"

TEXT ·get(SB), NOSPLIT, $0-8
    MOVQ ·a(SB), AX     // 把 go 代码定义的全局变量读到 AX 中
    MOVQ AX, ret+0(FP)
    RET
```
go 代码中使用汇编定义的变量：
```go
var Name,Helloworld string
func doSth() {
	fmt.Printf("Name:%s\n", Name)               // 读取汇编中初始化的变量 Name
	fmt.Printf("Helloworld:%s\n", Helloworld)   // 读取汇编中初始化的变量 Helloworld
}
// 输出： 
// Name:gopher
```
```asm
// string 定义形式 1： 在 String 结构体后多分配一个 [n]byte 数组存放静态字符串
DATA ·Name+0(SB)/8,$·Name+16(SB)    // StringHeader.Data
DATA ·Name+8(SB)/8,$6               // StringHeader.Len
DATA ·Name+16(SB)/8,$"gopher"       // [6]byte{'g','o','p','h','e','r'}
GLOBL ·Name(SB),NOPTR,$24           // struct{Data uintptr, Len int, str [6]byte}

// string 定义形式 2：独立分配一个仅当前文件可见的 [n]byte 数组存放静态字符串
DATA str<>+0(SB)/8,$"Hello Wo"      // str[0:8]={'H','e','l','l','o',' ','W','o'}
DATA str<>+8(SB)/8,$"rld!"          // str[9:12]={'r','l','d','!''}
GLOBL str<>(SB),NOPTR,$16           // 定义全局数组 var str<> [16]byte
DATA ·Helloworld+0(SB)/8,$str<>(SB) // StringHeader.Data = &str<>
DATA ·Helloworld+8(SB)/8,$12        // StringHeader.Len = 12
GLOBL ·Helloworld(SB),NOPTR,$16     // struct{Data uintptr, Len int}

```

函数：

1、go 文件中声明
```go
package fun

//go:noinline
func Swap(a, b int) (int, int)
```
```asm
#include "textflag.h"

// func Swap(a,b int) (int,int)
告诉汇编器该数据放到TEXT区
 ^    告诉汇编器这是基于静态地址的数据(static base)
 |             ^      本地变量占用空间大小
 |             |          ^  参数+返回值占用空间大小
 |             |          |   ^
 |             |          |   |
TEXT fun·Swap(SB),NOSPLIT,$0-32
    MOVQ a+0(FP), AX  // FP(Frame pointer)栈帧指针 这里指向栈帧最低位
    MOVQ b+8(FP), BX
    MOVQ BX ,ret0+16(FP)
    MOVQ AX ,ret1+24(FP)
    RET
```
stack frame size 栈帧大小(局部变量+可能需要的额外调用函数的参数空间的总大小，但不不包含调用其他函数时的ret address的大小)

arguments size 参数及返回值大小

若不指定NOSPLIT，arguments size必须指定。
  
2、s 文件中实现，函数名已 '·'开头


例子：
```go
// add.go
package main
import "fmt"
func add(x, y int64) int64
func main() {
	fmt.Println(add(2, 3))
}
```
```asm
// add_amd64.s
// add(x,y) -> x+y
TEXT ·add(SB),NOSPLIT,$0
	MOVQ x+0(FP), BX
	MOVQ y+8(FP), BP
	ADDQ BP, BX
	MOVQ BX, ret+16(FP)
	RET
```

go 语言编译成汇编
```sh
go tool compile -S xxx.go
```
从二进制反编译为汇编：
```sh
go tool objdump 

go build -gcflags "-N -l" -ldflags=-compressdwarf=false -o main.out main.go
go tool objdump -s "main.main" main.out > main.S
# or
go tool compile -S main.go
# or
go build -gcflags -S main.go

//

// 编译
go build -gcflags="-S"
go tool compile -S hello.go
go tool compile -l -N -S hello.go // 禁止内联 禁止优化
// 反编译
go tool objdump <binary>
```

汇编调用 go语言函数：
```asm
#include "textflag.h"

// func output(a,b int) int
TEXT ·output(SB), NOSPLIT, $24-24
    MOVQ a+0(FP), DX // arg a
    MOVQ DX, 0(SP) // arg x
    MOVQ b+8(FP), CX // arg b
    MOVQ CX, 8(SP) // arg y
    CALL ·add(SB) // 在调用 add 之前，已经把参数都通过物理寄存器 SP 搬到了函数的栈顶
    MOVQ 16(SP), AX // add 函数会把返回值放在这个位置
    MOVQ AX, ret+16(FP) // return result
    RET
```
```go
package main
import "fmt"

func add(x, y int) int {
    return x + y
}

func output(a, b int) int

func main() {
    s := output(10, 13)
    fmt.Println(s)
}
```
有一些需要注意的细节，比如调用过程中的栈拓展，nosplit 等

FUNCDATA 和 PCDATA 指令包含了由垃圾回收器使用的信息，他们由编译器引入

[《Go语言高级编程》3.6.3 PCDATA 和 FUNCDATA](https://chai2010.cn/advanced-go-programming-book/ch3-asm/ch3-06-func-again.html#363-pcdata-%E5%92%8C-funcdata)
```md
3.6.3 PCDATA 和 FUNCDATA
Go 语言中有个 runtime.Caller 函数可以获取当前函数的调用者列表。我们可以非常容易在运行时定位每个函数的调用位置，以及函数的调用链。因此在 panic 异常或用 log 输出信息时，可以精确定位代码的位置。

比如以下代码可以打印程序的启动流程：


func main() {
    for skip := 0; ; skip++ {
        pc, file, line, ok := runtime.Caller(skip)
        if !ok {
            break
        }

        p := runtime.FuncForPC(pc)
        fnfile, fnline := p.FileLine(0)

        fmt.Printf("skip = %d, pc = 0x%08X\n", skip, pc)
        fmt.Printf("func: file = %s, line = L%03d, name = %s, entry = 0x%08X\n", fnfile, fnline, p.Name(), p.Entry())
        fmt.Printf("call: file = %s, line = L%03d\n", file, line)
    }
}
其中 runtime.Caller 先获取当时的 PC 寄存器值，以及文件和行号。然后根据 PC 寄存器表示的指令位置，通过 runtime.FuncForPC 函数获取函数的基本信息。Go 语言是如何实现这种特性的呢？

Go 语言作为一门静态编译型语言，在执行时每个函数的地址都是固定的，函数的每条指令也是固定的。如果针对每个函数和函数的每个指令生成一个地址表格（也叫 PC 表格），那么在运行时我们就可以根据 PC 寄存器的值轻松查询到指令当时对应的函数和位置信息。而 Go 语言也是采用类似的策略，只不过地址表格经过裁剪，舍弃了不必要的信息。因为要在运行时获取任意一个地址的位置，必然是要有一个函数调用，因此我们只需要为函数的开始和结束位置，以及每个函数调用位置生成地址表格就可以了。同时地址是有大小顺序的，在排序后可以通过只记录增量来减少数据的大小；在查询时可以通过二分法加快查找的速度。

在汇编中有个 PCDATA 用于生成 PC 表格，PCDATA 的指令用法为：PCDATA tableid, tableoffset。PCDATA 有个两个参数，第一个参数为表格的类型，第二个是表格的地址。在目前的实现中，有 PCDATA_StackMapIndex 和 PCDATA_InlTreeIndex 两种表格类型。两种表格的数据是类似的，应该包含了代码所在的文件路径、行号和函数的信息，只不过 PCDATA_InlTreeIndex 用于内联函数的表格。

此外对于汇编函数中返回值包含指针的类型，在返回值指针被初始化之后需要执行一个 GO_RESULTS_INITIALIZED 指令：


#define GO_RESULTS_INITIALIZED	PCDATA $PCDATA_StackMapIndex, $1
GO_RESULTS_INITIALIZED 记录的也是 PC 表格的信息，表示 PC 指针越过某个地址之后返回值才完成被初始化的状态。

Go 语言二进制文件中除了有 PC 表格，还有 FUNC 表格用于记录函数的参数、局部变量的指针信息。FUNCDATA 指令和 PCDATA 的格式类似：FUNCDATA tableid, tableoffset，第一个参数为表格的类型，第二个是表格的地址。目前的实现中定义了三种 FUNC 表格类型：FUNCDATA_ArgsPointerMaps 表示函数参数的指针信息表，FUNCDATA_LocalsPointerMaps 表示局部指针信息表，FUNCDATA_InlTree 表示被内联展开的指针信息表。通过 FUNC 表格，Go 语言的垃圾回收器可以跟踪全部指针的生命周期，同时根据指针指向的地址是否在被移动的栈范围来确定是否要进行指针移动。

在前面递归函数的例子中，我们遇到一个 NO_LOCAL_POINTERS 宏。它的定义如下：


#define FUNCDATA_ArgsPointerMaps 0 /* garbage collector blocks */
#define FUNCDATA_LocalsPointerMaps 1
#define FUNCDATA_InlTree 2

#define NO_LOCAL_POINTERS FUNCDATA $FUNCDATA_LocalsPointerMaps, runtime·no_pointers_stackmap(SB)
因此 NO_LOCAL_POINTERS 宏表示的是 FUNCDATA_LocalsPointerMaps 对应的局部指针表格，而 runtime·no_pointers_stackmap 是一个空的指针表格，也就是表示函数没有指针类型的局部变量。

PCDATA 和 FUNCDATA 的数据一般是由编译器自动生成的，手工编写并不现实。如果函数已经有 Go 语言声明，那么编译器可以自动输出参数和返回值的指针表格。同时所有的函数调用一般是对应 CALL 指令，编译器也是可以辅助生成 PCDATA 表格的。编译器唯一无法自动生成是函数局部变量的表格，因此我们一般要在汇编函数的局部变量中谨慎使用指针类型。

对于 PCDATA 和 FUNCDATA 细节感兴趣的同学可以尝试从 debug/gosym 包入手，参考包的实现和测试代码。
```
[debug/gosym 官方文档](https://pkg.go.dev/debug/gosym)
[debug/gosym: 腾讯开发者社区翻译](https://cloud.tencent.com/developer/section/1141111)
[腾讯的张杰的《Go语言调试器开发》”pkg debug/gosym 应用“](https://www.hitzhangjie.pro/debugger101.io/7-headto-sym-debugger/6-gopkg-debug/2-gosym.html)
[Go 1.2 Runtime Symbol Information](https://docs.google.com/document/d/1lyPIbmsYbXnpNj57a261hgOYVpNRcgydurVQIyZOz_o/pub)
[issue-debug/gosym: report different line number than cmd/addr2line](https://github.com/golang/go/issues/56869)
```md
debug/dwarf   //dwarf包实现了可执行文件的调试信息
debug/elf     //elf包实现了对ELF对象文件的访问接口
debug/gosym   //gosym实现了对gc编译器生成的go二进制文件中嵌入的go符号和行号表的访问接
debug/macho   //macho包实现了Mach-O对象文件的访问接口
debug/pe    //pe包实现了对PE文件的访问接口
debug/plan9obj  //plan9obj包实现了 Plan 9 a.out对象文件的访问接口
```


[《Go语言高级编程》3.6.4 方法函数](https://chai2010.cn/advanced-go-programming-book/ch3-asm/ch3-06-func-again.html#364-%E6%96%B9%E6%B3%95%E5%87%BD%E6%95%B0)
```md
3.6.4 方法函数
Go 语言中方法函数和全局函数非常相似，比如有以下的方法：


package main

type MyInt int

func (v MyInt) Twice() int {
    return int(v)*2
}

func MyInt_Twice(v MyInt) int {
    return int(v)*2
}
其中 MyInt 类型的 Twice 方法和 MyInt_Twice 函数的类型是完全一样的，只不过 Twice 在目标文件中被修饰为 main.MyInt.Twice 名称。我们可以用汇编实现该方法函数：


// func (v MyInt) Twice() int
TEXT ·MyInt·Twice(SB), NOSPLIT, $0-16
    MOVQ a+0(FP), AX   // v
    ADDQ AX, AX        // AX *= 2
    MOVQ AX, ret+8(FP) // return v
    RET
不过这只是接收非指针类型的方法函数。现在增加一个接收参数是指针类型的 Ptr 方法，函数返回传入的指针：


func (p *MyInt) Ptr() *MyInt {
    return p
}
在目标文件中，Ptr 方法名被修饰为 main.(*MyInt).Ptr，也就是对应汇编中的 ·(*MyInt)·Ptr。不过在 Go 汇编语言中，星号和小括弧都无法用作函数名字，也就是无法用汇编直接实现接收参数是指针类型的方法。

在最终的目标文件中的标识符名字中还有很多 Go 汇编语言不支持的特殊符号（比如 type.string."hello" 中的双引号），这导致了无法通过手写的汇编代码实现全部的特性。或许是 Go 语言官方故意限制了汇编语言的特性。
```

汇编直接调用内建函数会报错：
```go
package main
import _ "fmt"
func Print(delta string)
func main() {
   Print("hello")
}
```
```asm
#include "textflag.h"
TEXT ·Print(SB), NOSPLIT, $8
    CALL fmt·Println(SB)
    RET
```
运行上面代码会报错：main.Print: relocation target fmt.Println not defined for ABI0 (but is defined for ABIInternal)

```go
package main
import ( 
  "fmt"
)
func Print(str string)
func main() { 
  Print("hello")
}
func Println(str string) { 
  fmt.Println(str)
}
```
```asm
#include "textflag.h"
TEXT ·Print(SB), NOSPLIT, $16-16    
  MOVQ strp+0(FP), AX    
  MOVQ AX, 0(SP)  // 第一个参数：数据的开始指针    
  MOVQ size+8(FP), BX    
  MOVQ BX, 8(SP)  // 第二个参数：string的大小。改成 MOVQ $100, 8(SP) 试试，会发现打印了其他地方的数据    
  CALL ·Println(SB)    
  RET// 这里一定要有换行，否则编译报错

```


plan9汇编操作数方向 与intel汇编方向相反
```asm
//plan9 汇编
MOVQ $123, AX
//intel汇编
mov rax, 123
```
栈操作
plan9中栈操作并没有push pop，而是采用sub和add SP
```asm
SUBQ $0x18, SP //对SP做减法 为函数分配函数栈帧
ADDQ $0x18, SP //对SP做加法 清楚函数栈帧
```
数据copy
```asm
MOVB $1, DI // 1 byte
MOVW $0x10, BX // 2bytes
MOVD $1, DX // 4 bytes
MOVQ $-10, AX // 8 bytes
```
计算指令
```asm
ADDQ AX, BX // BX += AX
SUBQ AX, BX // BX -= AX
IMULQ AX, BX // BX *= AX
```
跳转
```asm
//无条件跳转
JMP addr // 跳转到地址，地址可为代码中的地址 不过实际上手写不会出现这种东西
JMP label // 跳转到标签 可以跳转到同一函数内的标签位置
JMP 2(PC) // 以当前置顶为基础，向前/后跳转x行
JMP -2(PC) //同上
//有条件跳转
JNZ target // 如果zero flag被set过，则跳转
```

AT&T 汇编语法
AT＆T汇编语法是类Unix的系统上的标准汇编语法，比如gcc、gdb中默认都是使用AT&T汇编语法。AT&T汇编的指令格式如下：

instruction src dst
其中instruction是指令助记符，也叫操作码，比如mov就是一个指令助记符，src是源操作数，dst是目的操作。

当引用寄存器时候，应在寄存器名称加前缀%，对于常数，则应加前缀 $。

指令分类
数据传输指令
汇编指令	逻辑表达式	含义
mov $0x05, %ax	R[ax] = 0x05	将数值5存储到寄存器ax中
mov %ax, -4(%bp)	mem[R[bp] -4] = R[ax]	将ax寄存器中存储的数据存储到
bp寄存器存的地址减去4之后的内存地址中，
mov -4(%bp), %ax	R[ax] = mem[R[bp] -4]	bp寄存器存储的地址减去4值，
然后改地址对应的内存存储的信息存储到ax寄存器中
mov $0x10, (%sp)	mem[R[sp]] = 0x10	将16存储到sp寄存器存储的地址对应的内存
push $0x03	mem[R[sp]] = 0x03
R[sp] = R[sp] - 4	将数值03入栈，然后sp寄存器存储的地址减去4
pop	R[sp] = R[sp] + 4	将当前sp寄存器指向的地址的变量出栈，
并将sp寄存器存储的地址加4
call func1	---	调用函数func1
ret	---	函数返回，将返回值存储到寄存器中或caller栈中，
并将return address弹出到ip寄存器中
当使用mov指令传递数据时，数据的大小由mov指令的后缀决定。

movb $123, %eax // 1 byte
movw $123, %eax // 2 byte
movl $123, %eax // 4 byte
movq $123, %eax // 8 byte
算术运算指令
指令	含义
subl $0x05, %eax	R[eax] = R[eax] - 0x05
subl %eax, -4(%ebp)	mem[R[ebp] -4] = mem[R[ebp] -4] - R[eax]
subl -4(%ebp), %eax	R[eax] = R[eax] - mem[R[ebp] -4]
跳转指令
指令	含义
cmpl %eax %ebx	计算 R[eax] - R[ebx], 然后设置flags寄存器
jmp location	无条件跳转到location
je location	如果flags寄存器设置了相等标志，则跳转到location
jg, jge, jl, gle, jnz, ... location	如果flags寄存器设置了>, >=, <, <=, != 0等标志，则跳转到location
栈与地址管理指令
指令	含义	等同操作
pushl %eax	将R[eax]入栈	subl $4, %esp;
movl %eax, (%esp)
popl %eax	将栈顶数据弹出，然后存储到R[eax]	movl (%esp), %eax
addl $4, %esp
leave	Restore the callers stack pointer	movl %ebp, %esp
pop %ebp
lea 8(%esp), %esi	将R[esp]存放的地址加8，然后存储到R[esi]	R[esi] = R[esp] + 8
lea 是load effective address的缩写，用于将一个内存地址直接赋给目的操作数。

函数调用指令
指令	含义
call label	调用函数，并将返回地址入栈
ret	从栈中弹出返回地址，并跳转至该返回地址
leave	恢复调用者者栈指针
提示
以上指令分类并不规范和完整，比如`call`,`ret`都可以算作无条件跳转指令，这里面是按照功能放在函数调用这一分类了。

操作指令
用于指导汇编如何进行。以下指令后缀<mark>Q</mark>说明是64位上的汇编指令。

助记符	指令种类	用途	示例
MOVQ	传送	数据传送	MOVQ 48, AX表示把48传送AX中
LEAQ	传送	地址传送	LEAQ AX, BX表示把AX有效地址传送到BX中
PUSHQ	传送	栈压入	PUSHQ AX表示先修改栈顶指针，将AX内容送入新的栈顶位置在go汇编中使用SUBQ代替
POPQ	传送	栈弹出	POPQ AX表示先弹出栈顶的数据，然后修改栈顶指针在go汇编中使用ADDQ代替
ADDQ	运算	相加并赋值	ADDQ BX, AX表示BX和AX的值相加并赋值给AX
SUBQ	运算	相减并赋值	略，同上
IMULQ	运算	无符号乘法	略，同上
IDIVQ	运算	无符号除法	IDIVQ CX除数是CX，被除数是AX，结果存储到AX中
CMPQ	运算	对两数相减，比较大小	CMPQ SI CX表示比较SI和CX的大小。与SUBQ类似，只是不返回相减的结果
CALL	转移	调用函数	CALL runtime.printnl(SB)表示通过<mark>println</mark>函数的内存地址发起调用
JMP	转移	无条件转移指令	JMP 389无条件转至0x0185地址处(十进制389转换成十六进制0x0185)
JLS	转移	条件转移指令	JLS 389上一行的比较结果，左边小于右边则执行跳到0x0185地址处(十进制389转换成十六进制0x0185)
可以看到，表中的PUSHQ和POPQ被去掉了，这是因为在go汇编中，对栈的操作并不是出栈入栈，而是通过对SP进行运算来实现的。

标志位
助记符	名字	用途
OF	溢出	0为无溢出 1为溢出
CF	进位	0为最高位无进位或错位 1为有
PF	奇偶	0表示数据最低8位中1的个数为奇数，1则表示1的个数为偶数
AF	辅助进位	
ZF	零	0表示结果不为0 1表示结果为0
SF	符号	0表示最高位为0 1表示最高位为1

# 2、 go语言ABI
[Go internal ABI specification](https://go.googlesource.com/go/+/refs/heads/dev.regabi/src/cmd/compile/internal-abi.md)
[翻译](http://hushi55.github.io/2021/04/19/Go-internal-ABI-specification)
[internal/abi:文档，获取 abi0 寄存器传参](https://pkg.go.dev/internal/abi)
[issue-register ABI: allow selecting ABI when Go code refers to assembly routine #44065](https://github.com/golang/go/issues/44065)
[Proposal: Create an undefined internal calling convention](https://github.com/golang/proposal/blob/master/design/27539-internal-abi.md#compatibility) : ABIInternal 封装成 ABI0 的桥接代码
[Proposal: Create an undefined internal calling convention](https://go.googlesource.com/proposal/+/master/design/27539-internal-abi.md)
[翻译？](https://elkeid.bytedance.com/Chinese/RASP/golang.html#golang-internal-abi)
[Go internal ABI specification](https://go.googlesource.com/go/+/refs/heads/dev.regabi/src/cmd/compile/internal-abi.md)


为何高版本Go要改用寄存器传参？
至于为什么Go1.17.1函数调用的参数传递开始基于寄存器进行传递，原因无外乎。

第一，CPU访问寄存器比访问栈要快的多。函数调用通过寄存器传参比栈传参，性能要高5%。

第二，早期Go版本为了降低实现的复杂度，统一使用栈传递参数和返回值，不惜牺牲函数调用的性能。

第三，Go从1.17.1版本，开始支持多ABI（application binary interface 应用程序二进制接口，规定了程序在机器层面的操作规范，主要包括调用规约calling convention），主要是两个ABI：一个是老版本Go采用的平台通用ABI0，一个是Go独特的ABIInternal，前者遵循平台通用的函数调用约定，实现简单，不用担心底层cpu架构寄存器的差异；后者可以指定特定的函数调用规范，可以针对特定性能瓶颈进行优化，在多个Go版本之间可以迭代，灵活性强，支持寄存器传参提升性能。

所谓“调用规约(calling convention)”是调用方和被调用方对于函数调用的一个明确的约定，包括：函数参数与返回值的传递方式、传递顺序。只有双方都遵守同样的约定，函数才能被正确地调用和执行。如果不遵守这个约定，函数将无法正确执行。

性能差异
go 函数采用的是 caller-save 模式，被调用者的参数、返回值、栈位置都由调用者维护。


# 3、 go语言内存管理和垃圾回收对汇编的影响
拷贝栈
早年的 Go 运行时使用分段栈的机制，即当一个 Goroutine 的执行栈溢出时， 栈的扩张操作是在另一个栈上进行的，这两个栈彼此没有连续。 这种设计的缺陷很容易破坏缓存的局部性原理，从而降低程序的运行时性能。 因此现在 Go 运行时开始使用连续栈机制，当一个执行栈发生溢出时， 新建一个两倍于原栈大小的新栈，再将原栈整个拷贝到新栈上。 从而整个栈总是连续的。栈的拷贝并非想象中的那样简单，因为一个栈上可能保留指向被拷贝栈的指针， 从而当栈发生拷贝后，这个指针可能还指向原栈，从而造成错误。 此外，Goroutine 上原本的 gobuf 也需要被更新，这也是使用连续栈的难点之一
总而言之，没有被 go:nosplit 标记的函数的序言部分会插入分段检查，从而在发生栈溢出的情况下， 触发 runtime.morestack 调用，如果函数不需要 ctxt，则会调用 runtime.morestack_noctxt 从而抛弃 ctxt 再调用 morestack：
前面我们已经提到了，栈拷贝的其中一个难点就是 Go 中栈上的变量会包含自己的地址， 当我们拷贝了一个指向原栈的指针时，拷贝后的指针会变为无效指针。 不难发现，只有栈上分配的指针才能指向栈上的地址，否则这个指针指向的对象会重新在堆中进行分配（逃逸）。

当执行栈扩容时，会在内存空间中分配更大的栈内存空间，然后将旧栈中的所有内容复制到新栈中，并修改指向旧栈对应变量的指针重新指向新栈，最后销毁并回收旧栈的内存空间，从而实现栈的动态扩容。

从 go 逃逸分析的一个条件，”堆上的指针指向的内存将会逃逸“ 反推可知，只有站上的指针才会指向栈上的地址。

拷贝栈理论上没有上限，但是是加上设置了上限。
当新的栈大小超过了maxstacksize就会抛出”stack overflow“的异常。maxstacksize是在runtime.main中设置的。64位 系统下栈的最大值1GB、32位系统是250MB
```go
if newsize > maxstacksize || newsize > maxstackceiling {
    if maxstacksize < maxstackceiling {
        print("runtime: goroutine stack exceeds ", maxstacksize, "-byte limit\n")
    } else {
        print("runtime: goroutine stack exceeds ", maxstackceiling, "-byte limit\n")
    }
    print("runtime: sp=", hex(sp), " stack=[", hex(gp.stack.lo), ", ", hex(gp.stack.hi), "]\n")
    throw("stack overflow")
}
```

# 4、 内联函数
[详解Go内联优化](https://segmentfault.com/a/1190000039146279)
[译-Go语言inline内联的策略与限制](https://www.pengrl.com/p/20028/)
[原文-Go: Inlining Strategy & Limitation](https://medium.com/a-journey-with-go/go-inlining-strategy-limitation-6b6d7fc3b1be)
## 4.1 性能对比
首先，看一下函数内联与非内联的性能差异。
内联可以避免函数调用过程中的一些开销：创建栈帧，读写寄存器。不过，对函数体进行拷贝也会增大二进制文件的大小。
据 Go 官方宣传，内联大概会有 5~6% 的性能提升。
```go
//go:noinline
func maxNoinline(a, b int) int {
    if a < b {
        return b
    }
    return a
}

func maxInline(a, b int) int {
    if a < b {
        return b
    }
    return a
}

func BenchmarkNoInline(b *testing.B) {
    x, y := 1, 2
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        maxNoinline(x, y)
    }
}

func BenchmarkInline(b *testing.B) {
    x, y := 1, 2
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        maxInline(x, y)
    }
}
```
在程序代码中，想要禁止编译器内联优化很简单，在函数定义前一行添加//go:noinline即可。以下是性能对比结果

BenchmarkNoInline-8     824031799                1.47 ns/op
BenchmarkInline-8       1000000000               0.255 ns/op
因为函数体内部的执行逻辑非常简单，此时内联与否的性能差异主要体现在函数调用的固定开销上。显而易见，该差异是非常大的。

内联条件
执行命令查看编译器的优化策略
```sh
go build -gcflags="-m -m" xxx.go
```
或者使用 参数-gflags="-m"运行，可显示被内联的函数
```sh
./op.go:3:6: can inline add
./op.go:7:6: can inline sub
./main.go:16:11: inlining call to sub
./main.go:14:11: inlining call to add
./main.go:7:12: inlining call to fmt.Printf
```
1. 函数中含某些关键字： 闭包调用，select，for，defer，go
```go
//src/cmd/compile/internal/gc/inl.go
   case OCLOSURE,
        OCALLPART,
        ORANGE,
        OFOR,
        OFORUNTIL,
        OSELECT,
        OTYPESW,
        OGO,
        ODEFER,
        ODCLTYPE, // can't print yet
        OBREAK,
        ORETJMP:
        v.reason = "unhandled op " + n.Op.String()
        return true
```
2. 小代码量
超过80个节点（抽象语法树AST的节点）的代码量就不再内联。

## 4.1 内联表
内联会将函数调用的过程抹掉，会导致调用栈信息不完整。
Go在内部维持了一份内联函数的映射关系。
首先它会生成一个内联树，我们可以通过-gcflags="-d pctab=pctoinline"参数查看。
Go在生成的代码中映射了内联函数。并且，也映射了行号和源码文件可以通过-d pctab=pctoline参数查看行号，可以通过-gcflags="-d pctab=pctofile"查看源文件。

由以上信息，我们可以得到一张 PC-函数/文件/行号映射表.

```
此外，PCDATA 和 FUNCDATA 信息在二进制中也会存在。

## 4.2 内联控制
Go程序编译时，默认将进行内联优化。我们可通过-gcflags="-l"选项全局禁用内联，与一个-l禁用内联相反，如果传递两个或两个以上的-l则会打开内联，并启用更激进的内联策略。如果不想全局范围内禁止优化，则可以在函数定义时添加 //go:noinline 编译指令来阻止编译器内联函数。




# 5、 有哪些有意思的使用场景

## 5.1、 goroutine id
[petermattis/goid](https://github.com/petermattis/goid) 里通过汇编获取 goid 的代码关键逻辑如下：

runtime_go1.9.go 代码：
```go
//go:build gc && go1.9
// +build gc,go1.9

package goid

type stack struct {
	lo uintptr
	hi uintptr
}

type gobuf struct {
	sp   uintptr
	pc   uintptr
	g    uintptr
	ctxt uintptr
	ret  uintptr
	lr   uintptr
	bp   uintptr
}

type g struct {
	stack       stack
	stackguard0 uintptr
	stackguard1 uintptr

	_panic       uintptr
	_defer       uintptr
	m            uintptr
	sched        gobuf
	syscallsp    uintptr
	syscallpc    uintptr
	stktopsp     uintptr
	param        uintptr
	atomicstatus uint32
	stackLock    uint32
	goid         int64 // Here it is!
}
```
goid_go1.5_amd64.go代码：
```go
//go:build (amd64 || amd64p32) && gc && go1.5
// +build amd64 amd64p32
// +build gc
// +build go1.5

package goid

func Get() int64
```
goid_go1.5_amd64.s 代码：
```asm
//go:build (amd64 || amd64p32) && gc && go1.5
// +build amd64 amd64p32
// +build gc
// +build go1.5

#include "go_asm.h"
#include "textflag.h"

// func Get() int64
TEXT ·Get(SB),NOSPLIT,$0-8
	MOVQ (TLS), R14
	MOVQ g_goid(R14), R13
	MOVQ R13, ret+0(FP)
	RET

```
这样或 goid 也有一个局限性，就是如果当前处于系统调用或处于 CGO 函数中，拿到的 goid 是 g0 的 id。
在这种情况下 g.m.curg.goid 才是真正的 goroutine id。
参考[go/src/runtime/HACKING.md](https://github.com/golang/go/blob/go1.18.10/src/runtime/HACKING.md#getg-and-getgmcurg)
```md
getg() and getg().m.curg

To get the current user g, use getg().m.curg.

getg() alone returns the current g, but when executing on the system or signal stacks, this will return the current M's "g0" or "gsignal", respectively. This is usually not what you want.

To determine if you're running on the user stack or the system stack, use getg() == getg().m.curg.
```

pid也可以用汇编获取：
[choleraehyq/pid](https://github.com/choleraehyq/pid)
是一个 fork 
[petermattis/goid](https://github.com/petermattis/goid) 
的仓库，里面获取 pid 的实现：
p_m_go1.19.go 代码：
```go

//go:build gc && go1.19 && !go1.21
// +build gc,go1.19,!go1.21

package goid

type p struct {
	id int32 // Here is pid
}

type m struct {
	g0      uintptr // goroutine with scheduling stack
	morebuf gobuf   // gobuf arg to morestack
	divmod  uint32  // div/mod denominator for arm - known to liblink
	_       uint32

	// Fields not known to debuggers.
	procid     uint64       // for debuggers, but offset not hard-coded
	gsignal    uintptr      // signal-handling g
	goSigStack gsignalStack // Go-allocated signal handling stack
	sigmask    sigset       // storage for saved signal mask
	tls        [6]uintptr   // thread-local storage (for x86 extern register)
	mstartfn   func()
	curg       uintptr // current running goroutine
	caughtsig  uintptr // goroutine running during fatal signal
	p          *p      // attached p for executing go code (nil if not executing go code)
}
```
pid_go1.5.go 代码：
```go

//go:build (amd64 || amd64p32 || arm64) && !windows && gc && go1.5
// +build amd64 amd64p32 arm64
// +build !windows
// +build gc
// +build go1.5

package goid

//go:nosplit
func getPid() uintptr

//go:nosplit
func GetPid() int {
	return int(getPid())
}
```
pid_go1.5_amd64.s 代码：
```asm

// +build amd64 amd64p32
// +build gc,go1.5

#include "go_asm.h"
#include "textflag.h"

// func getPid() int64
TEXT ·getPid(SB),NOSPLIT,$0-8
	MOVQ (TLS), R14
	MOVQ g_m(R14), R13
	MOVQ m_p(R13), R14
	MOVL p_id(R14), R13
	MOVQ R13, ret+0(FP)
	RET

```
不过在伺候持有 pid 的过程中，可能当前 goroutine 已经被调度到其他 P 上了。
如果想要持有在持有 pid 的过程中持续帮当当前 P，可以使用一下方式：
```go
import "unsafe"

var _ = unsafe.Sizeof(0)

//go:linkname procPin runtime.procPin
//go:nosplit
func procPin() int

//go:linkname procUnpin runtime.procUnpin
//go:nosplit
func procUnpin()
```
runtime.procPin 和 runtime.procUnpin的实现代码在 [src/runtime/proc.go](https://github.com/golang/go/blob/go1.18.10/src/runtime/proc.go#L6054):
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
```
通过 mp.locks++ 锁定 P 的调度，不过也需要谨慎使用，会影响性能。


## 5.1、mockito和monkey
之所以把这两个放一起讲，是因为他们有一个同源的依赖库，分别在
[github.com/bytedance/mockey/internal/monkey](https://github.com/bytedance/mockey/tree/v1.1.1/internal/monkey) 目录和 
[code.byted.org/luoshiqi/mockito/monkey](https://code.byted.org/luoshiqi/mockito/tree/master/monkey) 目录下。
该库来源不明，不过 README.md 文档有如下说明
```md
This library fork form chenzhuoyu/monkey
enhancement:

allocPage/freePage  used go.runtime
support windows
support arm64 darwin (and refactored)
```

github.com/bytedance/mockey/internal/monkey/mem/write_linux.go 代码：
```go
package mem

import (
	"syscall"

	"github.com/bytedance/mockey/internal/monkey/common"
)

func Write(target uintptr, data []byte) error {
	do_replace_code(target, common.PtrOf(data), uint64(len(data)), syscall.SYS_MPROTECT, syscall.PROT_READ|syscall.PROT_WRITE, syscall.PROT_READ|syscall.PROT_EXEC)
	return nil
}

func do_replace_code(
	_ uintptr, // void   *addr
	_ uintptr, // void   *data
	_ uint64, // size_t  size
	_ uint64, // int     mprotect
	_ uint64, // int     prot_rw
	_ uint64, // int     prot_rx
)
```
[github.com/bytedance/mockey/internal/monkey/mem/write_linux_amd64.s](https://github.com/bytedance/mockey/blob/v1.1.1/internal/monkey/mem/write_darwin_amd64.s) 代码：
```asm

#include "textflag.h"

#define NOP8 BYTE $0x90; BYTE $0x90; BYTE $0x90; BYTE $0x90; BYTE $0x90; BYTE $0x90; BYTE $0x90; BYTE $0x90;
#define NOP64 NOP8; NOP8; NOP8; NOP8; NOP8; NOP8; NOP8; NOP8;
#define NOP512 NOP64; NOP64; NOP64; NOP64; NOP64; NOP64; NOP64; NOP64;
#define NOP4096 NOP512; NOP512; NOP512; NOP512; NOP512; NOP512; NOP512; NOP512;

#define addr        arg + 0x00(FP)
#define data        arg + 0x08(FP)
#define size        arg + 0x10(FP)
#define mprotect    arg + 0x18(FP)
#define prot_rw     arg + 0x20(FP)
#define prot_rx     arg + 0x28(FP)

#define CMOVNEQ_AX_CX   \
    BYTE $0x48          \
    BYTE $0x0f          \
    BYTE $0x45          \
    BYTE $0xc8

TEXT ·do_replace_code(SB), NOSPLIT, $0x30 - 0
    JMP START
    NOP4096
START:
    MOVQ    addr, DI
    MOVQ    size, SI
    MOVQ    DI, AX
    ANDQ    $0x0fff, AX
    ANDQ    $~0x0fff, DI
    ADDQ    AX, SI
    MOVQ    SI, CX
    ANDQ    $0x0fff, CX
    MOVQ    $0x1000, AX
    SUBQ    CX, AX
    TESTQ   CX, CX
    CMOVNEQ_AX_CX
    ADDQ    CX, SI
    MOVQ    DI, R8
    MOVQ    SI, R9
    MOVQ    mprotect , AX
    MOVQ    prot_rw  , DX
    SYSCALL
    MOVQ    addr, DI
    MOVQ    data, SI
    MOVQ    size, CX
    REP
    MOVSB
    MOVQ    R8, DI
    MOVQ    R9, SI
    MOVQ    mprotect , AX
    MOVQ    prot_rx  , DX
    SYSCALL
    JMP     RETURN
    NOP4096
RETURN:
    RET

```
其 Patch() 的调用路径如下：

```go
func (builder *MockBuilder) Build() *Mocker {
	mocker := Mocker{target: reflect.ValueOf(builder.target), builder: builder}
	mocker.buildHook(builder)
	mocker.Patch()
	return &mocker
}
func (mocker *Mocker) Patch() *Mocker {
	mocker.lock.Lock()
	defer mocker.lock.Unlock()
	if mocker.isPatched {
		return mocker
	}
	mocker.patch = monkey.PatchValue(mocker.target, mocker.hook, reflect.ValueOf(mocker.proxy), mocker.builder.unsafe)
	mocker.isPatched = true
	addToGlobal(mocker)

	mocker.outerCaller = tool.OuterCaller()
	return mocker
}

// PatchValue replace the target function with a hook function, and stores the target function in the proxy function
// for future restore. Target and hook are values of function. Proxy is a value of proxy function pointer.
func PatchValue(target, hook, proxy reflect.Value, unsafe bool) *Patch {
	tool.Assert(hook.Kind() == reflect.Func, "'%s' is not a function", hook.Kind())
	tool.Assert(proxy.Kind() == reflect.Ptr, "'%v' is not a function pointer", proxy.Kind())
	tool.Assert(hook.Type() == target.Type(), "'%v' and '%s' mismatch", hook.Type(), target.Type())
	tool.Assert(proxy.Elem().Type() == target.Type(), "'*%v' and '%s' mismatch", proxy.Elem().Type(), target.Type())

	targetAddr := target.Pointer()
	// The first few bytes of the target function code
	const bufSize = 64
	targetCodeBuf := common.BytesOf(targetAddr, bufSize)
	// construct the branch instruction, i.e. jump to the hook function
	hookCode := inst.BranchInto(common.PtrAt(hook))
	// search the cutting point of the target code, i.e. the minimum length of full instructions that is longer than the hookCode
	cuttingIdx := inst.Disassemble(targetCodeBuf, len(hookCode), !unsafe)

	// construct the proxy code
	proxyCode := common.AllocatePage()
	// save the original code before the cutting point
	copy(proxyCode, targetCodeBuf[:cuttingIdx])
	// construct the branch instruction, i.e. jump to the cutting point
	copy(proxyCode[cuttingIdx:], inst.BranchTo(targetAddr+uintptr(cuttingIdx)))
	// inject the proxy code to the proxy function
	fn.InjectInto(proxy, proxyCode)

	tool.DebugPrintf("PatchValue: hook code len(%v), cuttingIdx(%v)\n", len(hookCode), cuttingIdx)

	// replace target function codes before the cutting point
	mem.WriteWithSTW(targetAddr, hookCode)

	return &Patch{base: targetAddr, code: proxyCode, size: cuttingIdx}
}

// WriteWithSTW copies data bytes to the target address and replaces the original bytes, during which it will stop the
// world (only the current goroutine's P is running).
func WriteWithSTW(target uintptr, data []byte) {
	common.StopTheWorld()
	defer common.StartTheWorld()
	err := Write(target, data)
	tool.Assert(err == nil, err)
}
```

## 5.2、 优化获取行号性能
[golang文件行号探索](https://tech.bytedance.net/articles/7111885881002164238) 中有详细说明，代码如下：
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
```asm
# stack_amd64.s
TEXT    ·NewLine(SB)， NOSPLIT， $0-8
    MOVQ     retpc-8(FP)， AX
    MOVQ     AX， ret+0(FP)
    RET

```

## 5.3、 优化获取调用栈性能
[关于 golang 错误处理的一些优化想法](https://tech.bytedance.net/articles/7120632282095812644) 中有详细说明。
stack_amd64.go 代码：
```go
//go:build amd64
// +build amd64

package errors

import (
	_ "unsafe"
)

func buildStack(s []uintptr) int
```
stack_amd64.s 代码：
```asm
//go:build amd64 || amd64p32 || arm64
// +build amd64 amd64p32 arm64

#include "go_asm.h"
#include "textflag.h"
#include "funcdata.h"

// func buildStack(s []uintptr) int
TEXT ·buildStack(SB), NOSPLIT, $24-8
	NO_LOCAL_POINTERS
	MOVQ 	cap+16(FP), DX 	// s.cap
	MOVQ 	p+0(FP), AX		// s.ptr
	MOVQ	$0, CX			// loop.i
loop:
	MOVQ	+8(BP), BX		// last pc -> BX
	MOVQ	BX, 0(AX)(CX*8)		// s[i] = BX
	
	ADDQ	$1, CX			// CX++ / i++
	CMPQ	CX, DX			// if s.len >= s.cap { return }
	JAE	return				// 无符号大于等于就跳转

	MOVQ	+0(BP), BP 		// last BP; 展开调用栈至上一层
	CMPQ	BP, $0 			// if (BP) <= 0 { return }
	JA loop					// 无符号大于就跳转

return:
	MOVQ	CX,n+24(FP) 	// ret n
	RET

```

## 5.4、 字符串比较
[src/cmd/compile/internal/typecheck/builtin/runtime.go](https://github.com/golang/go/blob/go1.18.10/src/cmd/compile/internal/typecheck/builtin/runtime.go#L74)
```go
func cmpstring(string, string) int
```
[src/internal/bytealg/compare_amd64.s](https://github.com/golang/go/blob/go1.18.10/src/internal/bytealg/compare_amd64.s)
```asm
    TEXT ·Compare<ABIInternal>(SB),NOSPLIT,$0-56
        // AX = a_base (want in SI)
        // BX = a_len  (want in BX)
        // CX = a_cap  (unused)
        // DI = b_base (want in DI)
        // SI = b_len  (want in DX)
        // R8 = b_cap  (unused)
        MOVQ	SI, DX
        MOVQ	AX, SI
        JMP	cmpbody<>(SB)

    TEXT runtime·cmpstring<ABIInternal>(SB),NOSPLIT,$0-40
        // AX = a_base (want in SI)
        // BX = a_len  (want in BX)
        // CX = b_base (want in DI)
        // DI = b_len  (want in DX)
        MOVQ	AX, SI
        MOVQ	DI, DX
        MOVQ	CX, DI
        JMP	cmpbody<>(SB)

    // input:
    //   SI = a
    //   DI = b
    //   BX = alen
    //   DX = blen
    // output:
    //   AX = output (-1/0/1)
    TEXT cmpbody<>(SB),NOSPLIT,$0-0
        CMPQ	SI, DI
...
```

## 5.5、 搜索字符串里的字符
Go 语言中，我们常用的 string 搜索函数是 strings.Index，该函数也调用了 汇编实现的函数：
[src/strings/strings.go](https://github.com/golang/go/blob/go1.18.10/src/strings/strings.go#L112)
```go
// Index returns the index of the first instance of substr in s, or -1 if substr is not present in s.
func Index(s, substr string) int {
	n := len(substr)
	switch {
	case n == 0:
		return 0
	case n == 1:
		return IndexByte(s, substr[0])
	case n == len(s):
		if substr == s {
			return 0
		}
		return -1
	case n > len(s):
		return -1
	case n <= bytealg.MaxLen:
		// Use brute force when s and substr both are small
		if len(s) <= bytealg.MaxBruteForce {
			return bytealg.IndexString(s, substr)
    ...
}
// IndexByte returns the index of the first instance of c in s, or -1 if c is not present in s.
func IndexByte(s string, c byte) int {
	return bytealg.IndexByteString(s, c)
}
```
[src/internal/bytealg/indexbyte_native.go](https://github.com/golang/go/blob/go1.18.10/src/internal/bytealg/indexbyte_native.go)
```go
//go:build 386 || amd64 || s390x || arm || arm64 || ppc64 || ppc64le || mips || mipsle || mips64 || mips64le || riscv64 || wasm

package bytealg

//go:noescape
func IndexByte(b []byte, c byte) int

//go:noescape
func IndexByteString(s string, c byte) int
```
[src/internal/bytealg/index_native.go](https://github.com/golang/go/blob/go1.18.10/src/internal/bytealg/index_native.go)
```go
//go:build amd64 || arm64 || s390x || ppc64le || ppc64

package bytealg

//go:noescape

// Index returns the index of the first instance of b in a, or -1 if b is not present in a.
// Requires 2 <= len(b) <= MaxLen.
func Index(a, b []byte) int

//go:noescape

// IndexString returns the index of the first instance of b in a, or -1 if b is not present in a.
// Requires 2 <= len(b) <= MaxLen.
func IndexString(a, b string) int
```
[src/internal/bytealg/indexbyte_amd64.s](https://github.com/golang/go/blob/go1.18.10/src/internal/bytealg/indexbyte_amd64.s)
```asm
    #include "go_asm.h"
    #include "textflag.h"

    TEXT	·IndexByte(SB), NOSPLIT, $0-40
        MOVQ b_base+0(FP), SI
        MOVQ b_len+8(FP), BX
        MOVB c+24(FP), AL
        LEAQ ret+32(FP), R8
        JMP  indexbytebody<>(SB)

    TEXT	·IndexByteString(SB), NOSPLIT, $0-32
        MOVQ s_base+0(FP), SI
        MOVQ s_len+8(FP), BX
        MOVB c+16(FP), AL
        LEAQ ret+24(FP), R8
        JMP  indexbytebody<>(SB)

    // input:
    //   SI: data
    //   BX: data len
    //   AL: byte sought
    //   R8: address to put result
    TEXT	indexbytebody<>(SB), NOSPLIT, $0
        // Shuffle X0 around so that each byte contains
        // the character we're looking for.
        MOVD AX, X0
        PUNPCKLBW X0, X0
        PUNPCKLBW X0, X0
        PSHUFL $0, X0, X0

...
```
[src/internal/bytealg/index_amd64.s](https://github.com/golang/go/blob/go1.18.10/src/internal/bytealg/index_amd64.s)
```asm
    #include "go_asm.h"
    #include "textflag.h"

    TEXT ·Index(SB),NOSPLIT,$0-56
        MOVQ a_base+0(FP), DI
        MOVQ a_len+8(FP), DX
        MOVQ b_base+24(FP), R8
        MOVQ b_len+32(FP), AX
        MOVQ DI, R10
        LEAQ ret+48(FP), R11
        JMP  indexbody<>(SB)

    TEXT ·IndexString(SB),NOSPLIT,$0-40
        MOVQ a_base+0(FP), DI
        MOVQ a_len+8(FP), DX
        MOVQ b_base+16(FP), R8
        MOVQ b_len+24(FP), AX
        MOVQ DI, R10
        LEAQ ret+32(FP), R11
        JMP  indexbody<>(SB)

    // AX: length of string, that we are searching for
    // DX: length of string, in which we are searching
    // DI: pointer to string, in which we are searching
    // R8: pointer to string, that we are searching for
    // R11: address, where to put return value
    // Note: We want len in DX and AX, because PCMPESTRI implicitly consumes them
    TEXT indexbody<>(SB),NOSPLIT,$0
...
```

## 5.6、 自定义SIMD优化
[minio/sha256-simd](https://github.com/minio/sha256-simd)

## 5.7、 乱七八招的跳转
[关于 golang 错误处理的一些优化想法](https://tech.bytedance.net/articles/7120632282095812644) 中有详细说明。
stack_amd64.go 代码：


