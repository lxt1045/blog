


<!-- # 0、Go语言的汇编能做什么

 0.1、 算法加速，golang编译器为了编译速度，放弃了很多指令优化，远比不上C/C++的gcc/clang生成的优化程度高，甚至和 Java 比都有差距，毕竟时间沉淀在那里。因此某些场景下某些优化逻辑、CPU指令可以让我们的算法运行速度更快，如 SIMD 指令。

 0.2、 摆脱golang编译器的一些约束，如通过汇编[调用其他package的私有函数](https://sitano.github.io/2016/04/28/golang-private/)。

 0.3、 进行一些hack的事，如通过汇编适配其他语言的ABI来直接调用其他语言的函数:[petermattis/fastcgo](https://github.com/petermattis/fastcgo)。
  -->


参考文档：
<!-- 柴树杉和曹春晖的书 -->
[《Go语言高级编程》第三章](https://github.com/chai2010/advanced-go-programming-book/tree/master/ch3-asm)

[Go 语言汇编的官方文档](https://go.dev/doc/asm)

<!-- 柴树杉的两篇文章 [plan9 assembly 完全解析](https://github.com/cch123/golang-notes/blob/master/assembly.md) 和 [汇编分享](https://github.com/cch123/asmshare/blob/master/layout.md) -->

[GoFunctionsInAssembly](https://lrita.github.io/images/posts/go/GoFunctionsInAssembly.pdf)

[A Quick Guide to Go's Assembler](https://go.dev/doc/asm)

<!-- 欧长坤老师的书 -->
[《Go语言原本》1.4 Plan 9 汇编语言](https://golang.design/under-the-hood/zh-cn/part1basic/ch01basic/asm/)


<!-- # 前言
我们知道 Go 语言一些核心成员是 Plan 9 的遗老遗少，而且属于比较高傲的的学院派，这导致 Go 语言的汇编采用了令人抓狂的 Plan 9 风格。 -->

### 广告时间

笔者写了两个性能提升库，欢迎大家来吐槽。
1. [lxt1045/errors](https://github.com/lxt1045/errors): 有两个模块\
    1.1 errors：功能和 [pkg/errors](https://github.com/pkg/errors) 类似，性能比后者高一个数量级以上。\
    1.2 errors/loglus：功能和 [sirupsen/logrus](https://github.com/sirupsen/logrus) 一样，目标是 100% 兼容。 利用了 errors 的高性能获取代码行号的接口，性能比后者高 35% 以上。
2. [json](https://github.com/lxt1045/json): 以纯 Go 语言实现，在性能上追平了 SIMD 实现的 [sonic](https://github.com/bytedance/sonic)，甚至略胜一筹。

# 1、 Go 汇编基础知识
## 1.1、通用寄存器
<!-- [go语言调度器源代码情景分析之二：CPU寄存器](https://www.cnblogs.com/abozhang/p/10766689.html) -->

不同体系结构的CPU，其内部寄存器的数量、种类以及名称可能大不相同，这里我们只介绍 AMD64 的寄存器。
AMD64 有20多个可以直接在汇编代码中使用的寄存器，其中有几个寄存器在操作系统代码中才会见到，而应用层代码一般只会用到如下三类寄存器。

| 寄存器类型 | 用途 | 寄存器 |位宽|
| :---- | :---- | :---- |:----|
|通用寄存器|没有做特殊规定，程序员和编译器可以自定义其用途|RAX, RBX, RCX, RDX, RSI, RDI, RBP, RSP, R8, R9, R10, R11, R12, R13, R14, R15| 64bit |
|程序计数寄存器|IP 寄存器（Go 中叫 PC 寄存器）用来存放下一条即将执行的指令的地址，这个寄存器决定了程序的执行流程| RIP | 64bit |
|段寄存器|CS、DS、ES和SS在 AMD64 中已不用；FS 和 GS 一般用它来实现线程本地存储（TLS），比如AMD64 linux平台下 Go 语言使用 FS 寄存器来实现系统线程的TLS。| FS, GS, CS, DS, ES, SS | 16bit |

上述这些寄存器除了 FS和 GS 段寄存器是16位的，其它都是64位的，也就是8个字节，其中的16个通用寄存器还可以作为32/16/8位寄存器使用，只是使用时需要换一个名字，比如可以用 EAX 这个名字来表示一个32位的寄存器，它使用的是 RAX 寄存器的低32位。

AMD64的通用通用寄存器的名字在 plan9 中的对应关系:

|AMD64|RAX|RBX|RCX|RDX|RDI|RSI|RBP|RSP|R8|R9|R10|R11|R12|R13|R14|RIP|
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
|Plan9|AX|BX|CX|DX|DI|SI|BP|SP|R8|R9|R10|R11|R12|R13|R14|PC|

Go语言中寄存器一般用途：
| 助记符 | 名字 | 用途 |
| :---- | :---- | :---- |
|AX|累加寄存器(AccumulatorRegister)| 用于存放数据，包括算术、操作数、结果和临时存放地址等|
|BX|基址寄存器(BaseRegister)|用于存放访问存储器时的地址|
|CX|计数寄存器(CountRegister)|用于保存计算值，用作计数器|
|DX|数据寄存器(DataRegister)|用于数据传递，在寄存器间接寻址中的I/O指令中存放I/O端口的地址|
|BP|堆栈基指针(BasePointer)|保存在进入函数前的栈顶基址，SP一起指示了当前函数栈帧的范围|
|SI|源变址寄存器(SourceIndex)|用于存放源操作数的偏移地址|
|DI|目的寄存器(DestinationIndex)|用于存放目的操作数的偏移地址|
<!-- 
|SP|堆栈顶指针(StackPointer)|如果是symbol+offset(SP)的形式表示go汇编的伪寄存器；如果是offset(SP)的形式表示硬件寄存器|
|SB|静态基指针(StaticBasePointer)|go汇编的伪寄存器。foo(SB)用于表示变量在内存中的地址，foo+4(SB)表示foo起始地址往后偏移四字节。一般用来声明函数或全局变量|
|FP|栈帧指针(FramePointer)|go汇编的伪寄存器。引用函数的输入参数，形式是symbol+offset(FP)，例如arg0+0(FP)| 
-->

## 1.2、伪寄存器
伪寄存器是plan9伪汇编中的一个助记符, 也是 Plan9 比较反人类的语法之一。\
常见伪寄存器如下表所示：
| 助记符 | 说明 |  |
| :---- | :---- |  :---- | 
|SB(Static base pointer)|global symbols|
|PC(Program counter)|jumps and branches|
|SP(Stack pointer)|top of stack|
|FP(Frame pointer)|arguments and locals|

SB：指向全局符号表。相对于寄存器，SB 更像是一个声明标识，用于标识全局变量、函数等。通过 symbol(SB) 方式使用，symbol<>(SB)表示symbol只在当前文件可见，跟C中的static效果类似。此外可以在引用上加偏移量，如 symbol+4(SB) 表示 symbol+4bytes 的地址。

PC：程序计数器（Program Counter），指向下一条要执行的指令的地址，在AMD64对应 rip 寄存器。个人觉得，把他归为伪寄存器有点令人费解，可能是因为每个平台对应的物理寄存器名字不一样。

SP： SP 寄存器比较特殊，既可以当做物理寄存器也可以当做伪寄存器使用，不过这两种用法的使用语法不同。\
其中，伪寄存器使用语法是 symbol+offset(SP)，此场景下 SP 指向局部变量的起始位置(高地址处)；x-8(SP) 表示函数的第一个本地变量；\
物理SP(硬件SP)的使用语法则是 +offset(SP)，此场景下 SP 指向真实栈顶地址（栈帧最低地址处）。

FP：用于标识函数参数、返回值。被调用者（callee）的FP实际上是调用者（caller）的栈顶，即callee.SP(物理SP) == caller.FP；x+0(FP)表示第一个请求参数(参数返回值从右到左入栈)。
<!-- 非NOSPLIT：伪FP=伪SP+16；NOSPLIT（framepointer_enable?有无BP栈指针？）：伪FP=伪SP+8； -->

实际上，生成真正可执行代码时，伪 SP、FP 会由物理 SP 寄存器加上偏移量替换。所以执行过程中修改物理SP，会引起伪SP、FP同步变化，比如执行 SUBQ $16, SP 指令后，伪SP和伪FP都会 -16。
而且，反汇编二进制而生成的汇编代码中，只有物理 SP 寄存器。 即 go tool objdump/go tool compile -S 输出的汇编代码中，没有伪 SP 和 伪FP 寄存器，只有物理 SP 寄存器。

另外还有1个比较特殊的伪寄存器：\
TLS：存储当前 goroutine 的 g 结构体的指针。实际上，X86 和 AMD64 下的 TLS 是通过段寄存器 FS 实现的线程本地存储基地址，而当前 g 的指针是线程本地存储的第一个变量。

操作系统对内存的一般划分如下图所示：
```
  高地址 +------------------+
        |                  |
        |     内核空间      |
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
        |     全局数据      |
        |------------------|
        |                  |
        |     静态代码      |
        |                  |
        |------------------|
        |     系统保留      |
  低地址 |------------------|    
```

这里提个疑问，我们知道协程分为有栈协程和无栈协程，go语言是有栈协程。那你知道普通gorutine的调用栈是在哪个内存区吗？

## 1.3、函数调用栈帧
我们先熟悉几个名词。

caller：函数调用者。\
callee：函数被调用者。比如函数main中调用sum函数，那么main就是caller，而sum函数就是callee。\
栈帧：stack frame，即执行中的函数所持有的、独立连续的栈区段。一般用来保存函数参数、返回值、局部变量、返回PC值等信息。golang 的 ABI 规定，由 caller 管理函数参数和返回值。

下图是 golang 的调用栈，源于[曹春晖老师的github文章](https://github.com/cch123/asmshare/blob/master/layout.md#%E6%9F%A5%E7%9C%8B-go-%E8%AF%AD%E8%A8%80%E7%9A%84%E5%87%BD%E6%95%B0%E8%B0%83%E7%94%A8%E8%A7%84%E7%BA%A6) ，做了简单修改:
```
                                                                                                                              
                                       caller                                                                                 
                                 +------------------+                                                                         
                                 |                  |                                                                         
       +---------------------->  +------------------+                                                                        
       |                         |                  |                                                                         
       |                         | caller parent BP |                                                                         
       |           BP(pseudo SP) +------------------+                                                                         
       |                         |                  |                                                                         
       |                         |   Local Var0     |                                                                         
       |                         +------------------+                                                                         
       |                         |                  |                                                                         
       |                         |   .......        |                                                                         
       |                         +------------------+                                                                         
       |                         |                  |                                                                         
       |                         |   Local VarN     |                                                                         
                                 +------------------+                                                                         
 caller stack frame              |                  |                                                                         
                                 |   callee arg2    |                                                                         
       |                         +------------------+                                                                         
       |                         |                  |                                                                         
       |                         |   callee arg1    |                                                                         
       |                         +------------------+                                                                         
       |                         |                  |                                                                         
       |                         |   callee arg0    |                                                                         
       |   SP(Real Register) ->  +------------------+--------------------------+   FP(virtual register)                       
       |                         |                  |                          |                                              
       |                         |   return addr    |  parent return address   |                                              
       +---------------------->  +------------------+--------------------------+    <-----------------------+         
                                                    |  caller BP               |                            |         
                                                    |  (caller frame pointer)  |                            |         
                                     BP(pseudo SP)  +--------------------------+                            |         
                                                    |                          |                            |         
                                                    |     Local Var0           |                            |         
                                                    +--------------------------+                            |         
                                                    |                          |                                      
                                                    |     Local Var1           |                                      
                                                    +--------------------------+                    callee stack frame
                                                    |                          |                                      
                                                    |       .....              |                                      
                                                    +--------------------------+                            |         
                                                    |                          |                            |         
                                                    |     Local VarN           |                            |         
     High                         SP(Real Register) +--------------------------+                            |         
      ^                                             |                          |                            |         
      |                                             |                          |                            |         
      |                                             |                          |                            |         
      |                                             |                          |                            |         
      |                                             |                          |                            |         
      |                                             +--------------------------+    <-----------------------+         
     Low                                                                                                                      
                                                              callee
```

需要指出的是，上图中的 CALLER BP 是在编译期由编译器在符合条件时自动插入。所以手写汇编时，计算 framesize 时不应包括CALLER BP的空间。\
是否插入CALLER BP的主要判断依据如下：
```go
// Must agree with internal/buildcfg.FramePointerEnabled.
const framepointer_enabled = GOARCH == "amd64" || GOARCH == "arm64"
```

以下是 Go语言函数栈展开逻辑的[一段代码](https://github.com/golang/go/blob/969bea8d59daa6bdd478b71f6e99d8b8f625a140/src/runtime/traceback.go#L281)，它侧面验证了 BP 插入的条件：\
1. 函数的栈帧大小大于0; 
2. 常量 [framepointer_enabled](https://github.com/golang/go/blob/e4435cb8448514d2413f9d9aa3ee40738d26fd67/src/runtime/runtime2.go#L1190) 值为 true。
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

## 1.4、golang常用汇编指令
参考文档：\
[Go支持的X86指令](https://github.com/golang/arch/blob/v0.2.0/x86/x86.csv)\
[Go支持的ARM64指令](https://github.com/golang/arch/blob/v0.2.0/arm64/arm64asm/inst.json)\
[Go支持的ARM指令](https://github.com/golang/arch/blob/v0.2.0/arm/arm.csv)

常用指令：
|指令类型|示例|描述|
| :---- | :---- | :---- |
|数据传输指令|MOVB、MOVW、MOVD、MOVQ|移动寄存器/内存中的数据|
|算术运算指令|ADDQ、SUBD、IMULQ、IDIVQ、CMPQ||
|跳转指令|JMP、JNZ、JLS|执行代码跳转，伴随着 PC 寄存器的变化|
|栈操作指令| PUSHQ、POPQ|Go 语言中几乎不使用栈操作指令，而通过操作 SP 寄存器缩/扩栈|
|地址管理指令|LEAQ、LEAD|取地址操作|
|函数调用指令|CALL、RET|函数调用指令，一般伴随着 SP 和 PC 的变化|
|SIMD指令|MOVOU、PCMPEQB、PMOVMSKB|Go 汇编已经可以直接使用大部分 SIMD 指令|

例如
```asm
MOVB $1, DI     // 1 byte； 将 DI 的第一个 Byte 的值设置为 1
MOVW $0x10, BX  // 2bytes
MOVD $1, DX     // 4 bytes
MOVQ $-10, AX   // 8 bytes
SUBQ $0x18, SP  //对SP做减法，扩栈
ADDQ $0x18, SP  //对SP做加法，缩栈
ADDQ AX, BX     // BX += AX
SUBQ AX, BX     // BX -= AX
IMULQ AX, BX    // BX *= AX
JMP addr        // 跳转到地址，地址可为代码中的地址，不过实际上手写一般不会出现
JMP label       // 跳转到标签，可以跳转到同一函数内的标签位置
JMP 2(PC)       // 向前转2行
JMP -2(PC)      // 向后跳转2行
JNZ target      // 如果zero flag被set过，则跳转

```
<!-- 助记符	指令种类	用途	示例
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
可以看到，表中的PUSHQ和POPQ被去掉了，这是因为在go汇编中，对栈的操作并不是出栈入栈，而是通过对SP进行运算来实现的。 -->

常用标志位：

|助记符|名字|描述|
| :---- | :---- | :---- |
|OF|溢出|0为无溢出 1为溢出|
|CF|进位|0为最高位无进位或错位 1为有|
|PF|奇偶|0表示数据最低8位中1的个数为奇数，1则表示1的个数为偶数|
|AF|辅助进位||
|ZF|零|0表示结果不为0 1表示结果为0|
|SF|符号|0表示最高位为0 1表示最高位为1|

## 1.5 全局变量
参考文档：\
[《Go语言高级编程》的章节3.3 常量和全局变量](https://github.com/chai2010/advanced-go-programming-book/blob/master/ch3-asm/ch3-03-const-and-var.md)

### 1.5.1 使用语法
使用 GLOBL 关键字声明全局变量，用 DATA 定义指定内存的值：
```asm
// DATA 汇编指令指定对应内存中的值; width 必须是 1、2、4、8 几个宽度之一
DATA    symbol+offset(SB)/width, value // symbol+offset 偏移量，width 宽度, value 初始值

// GLOBL 指令声明一个变量对应的符号，以及变量对应的内存大小
GLOBL symbol(SB), flag, width  // 名为 symbol, 内存宽度为 width, flag可省略
```
例子：
```asm
DATA age+0x00(SB)/4, $18    // age = 18
GLOBL age(SB), RODATA, $4   // 声明全局变量 age，占用 4Byte 内存空间

DATA pi+0(SB)/8, $3.1415926
GLOBL pi(SB), RODATA, $8

DATA bio<>+0(SB)/8, $"hello wo"     // <> 表示只在当前文件生效
DATA bio<>+8(SB)/8, $"old !!!!"     // bio = "hello world !!!!"
GLOBL bio<>(SB), RODATA, $16     
```
其中 flag 的字面量定义在 [textflag.h](https://github.com/golang/go/blob/go1.18.10/src/runtime/textflag.h)，需要在汇编文件中 #include "textflag.h"，其类型有有如下几个：
| flag | value | 说明 |
| :-- | :-- |  :-- | 
|NOPROF|1|  (TEXT项使用) 不优化NOPROF标记的函数。这个标志已废弃。<br>(For TEXT items.) Don't profile the marked function. This flag is deprecated.</br> |
|DUPOK|2|  在二进制文件中允许一个符号的多个实例。链接器会选择其中之一。<br>It is legal to have multiple instances of this symbol in a single binary. The linker will choose one of the duplicates to use.</br> |
|NOSPLIT|4|  (TEXT项使用) 不插入检测栈分裂（扩张）的前导指令代码（减少开销，一般用于叶子节点函数（函数内部不调用其他函数））。程序的栈帧中，如果调用其他函数会增加栈帧的大小，必须在栈顶留出可用空间。用来保护程序，例如堆栈拆分代码本身。<br>(For TEXT items.) Don't insert the preamble to check if the stack must be split. The frame for the routine, plus anything it calls, must fit in the spare space at the top of the stack segment. Used to protect routines such as the stack splitting code itself. </br>|
|RODATA|8| (DATA和GLOBAL项使用) 将这个数据放在只读的块中。<br>(For DATA and GLOBL items.) Put this data in a read-only section.</br> |
|NOPTR|16|（用于DATA和GLOBL项目）这个数据不包含指针所以就不需要垃圾收集器来扫描。<br>(For DATA and GLOBL items.) This data contains no pointers and therefore does not need to be scanned by the garbage collector.</br> |
|WRAPPER|32|(TEXT项使用)这是包装函数<br>(For TEXT items.) This is a wrapper function and should not count as disabling recover. </br>|
|NEEDCTXT|64|（TEXT项使用）此函数是一个闭包，因此它将使用其传入的上下文寄存器。 <br> (For TEXT items.) This function is a closure so it uses its incoming context register. </br>|
|TLSBSS|256| （用于DATA和GLOBL项目）将此数据放入线程本地存储中。<br>Allocate a word of thread local storage and store the offset from the thread local base to the thread local storage in this variable. </br>|
|NOFRAME|512| （TEXT项使用）不要插入指令为此函数分配栈帧。仅在声明帧大小为0的函数上有效。（函数必须是叶子节点函数，且以0标记堆栈函数，没有保存帧指针（或link寄存器架构上的返回地址））TODO(mwhudson)：目前仅针对 ppc64x 实现。<br>Do not insert instructions to allocate a stack frame for this function. Only valid on functions that declare a frame size of 0. TODO(mwhudson): only implemented for ppc64x at present. </br>|
|REFLECTMETHOD|1024| 函数可以调用 reflect.Type.Method 或 reflect.Type.MethodByName。<br>Function can call reflect.Type.Method or reflect.Type.MethodByName. </br>|
|TOPFRAME|2048| （TEXT项使用）函数是调用堆栈的顶部。栈回溯应在此功能处停止。<br>Function is the outermost frame of the call stack. Call stack unwinders should stop at this function. </br>|
|ABIWRAPPER|4096| 函数是一个 ABI 包装器。<br>Function is an ABI wrapper. </br>|


其中 NOSPLIT 需要特别注意，它表示该函数运行不会导致栈分裂，用户也可以使用 //go:nosplit 强制给go函数指定NOSPLIT属性。\
例如：
```go
//go:nosplit
func someFunc() {
}
```
汇编中直接给函数标记 NOSPLIT 即可：
```asm
// 表示someFunc函数执行时最多需要 24 字节本地变量和 8 字节参数空间
TEXT ·someFunc(SB), NOSPLIT, $24-8  
    RET
```
链接器认为标记为 NOSPLIT 的函数，最多需要使用StackLimit字节空间，所以不需要插入栈分裂(溢出)检查，函数调用损耗更小。不过，使用该标志的时候要特别小心，万一发生意外容易导致栈溢出错误，溢出时会在执行期报 nosplit stack overflow 错。\
[go/src/runtime/HACKING.md](https://github.com/golang/go/blob/master/src/runtime/HACKING.md#nosplit-functions) 中有如下说明：
```md
nosplit functions
Most functions start with a prologue that inspects the stack pointer and the current G's stack bound and calls morestack if the stack needs to grow.

Functions can be marked //go:nosplit (or NOSPLIT in assembly) to indicate that they should not get this prologue. This has several uses:

 - Functions that must run on the user stack, but must not call into stack growth, for example because this would cause a deadlock, or because they have untyped words on the stack.
 - Functions that must not be preempted on entry.
 - Functions that may run without a valid G. For example, functions that run in early runtime start-up, or that may be entered from C code such as cgo callbacks or the signal handler.

Splittable functions ensure there's some amount of space on the stack for nosplit functions to run in and the linker checks that any static chain of nosplit function calls cannot exceed this bound.

Any function with a //go:nosplit annotation should explain why it is nosplit in its documentation comment.
```

另外，当函数处于调用链的叶子节点，且栈帧小于 StackSmall（128）字节时，则自动标记为NOSPLIT。 此逻辑的代码如下：\
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

### 1.5.2 Go 语言中的常用用法

在汇编代码中使用 go 变量：
```asm
#include "textflag.h"

TEXT ·get(SB), NOSPLIT, $0-8
    MOVQ ·a(SB), AX             // 把 go 代码定义的全局变量读到 AX 中
    MOVQ AX, ret+0(FP)          // 把 AX 的值写入返回值位置
    RET
```
```go
package main

var a = 999
func get() int

func main() {
    println(get())
}
```
go 代码中使用汇编定义的变量：
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
```go
var Name,Helloworld string
func doSth() {
	fmt.Printf("Name: %s\n", Name)               // 读取汇编中初始化的变量 Name
	fmt.Printf("Helloworld: %s\n", Helloworld)   // 读取汇编中初始化的变量 Helloworld
}
// 输出： 
// Name: gopher
// Helloworld: Hello World!

```

## 1.6 函数调用
<!-- 
以下来之欧长坤的《go 语言原本》
```md
运行时协调
为保证垃圾回收正确运行，在大多数栈帧中，运行时必须知道所有全局数据的指针。 Go 编译器会将这部分信息耦合到 Go 源码文件中，但汇编程序必须进行显式定义。

被标记为 NOPTR 标志的数据符号会视为不包含指向运行时分配数据的指针。 带有 R0DATA 标志的数据符号在只读存储器中分配，因此被隐式标记为 NOPTR。 总大小小于指针的数据符号也被视为隐式标记 NOPTR。 在一份汇编源文件中是无法定义包含指针的符号的，因此这种符号必须定义在 Go 源文件中。 一个良好的经验法则是 R0DATA 在 Go 中定义所有非符号而不是在汇编中定义。

每个函数还需要注释，在其参数，结果和本地堆栈框架中给出实时指针的位置。 对于没有指针结果且没有本地堆栈帧或没有函数调用的汇编函数， 唯一的要求是在同一个包中的 Go 源文件中为函数定义 Go 原型。 汇编函数的名称不能包含包名称组件 （例如，syscall 包中的函数 Syscall 应使用名称 ·Syscall 而不是 syscall·Syscall 其 TEXT 指令中的等效名称）。 对于更复杂的情况，需要显式注释。 这些注释使用标准 #include 文件中定义的伪指令 funcdata.h。

如果函数没有参数且没有结果（返回值），则可以省略指针信息。这是由一个参数大小 $n-0 注释指示 TEXT 对指令。 否则，指针信息必须由 Go 源文件中的函数的 Go 原型提供，即使对于未直接从 Go 调用的汇编函数也是如此。 （原型也将 go vet 检查参数引用。）在函数的开头，假定参数被初始化但结果假定未初始化。 如果结果将在调用指令期间保存实时指针，则该函数应首先将结果归零， 然后执行伪指令 GO_RESULTS_INITIALIZED。 此指令记录结果现在已初始化，应在堆栈移动和垃圾回收期间进行扫描。 通常更容易安排汇编函数不返回指针或不包含调用指令; 标准库中没有汇编函数使用 GO_RESULTS_INITIALIZED。

如果函数没有本地堆栈帧，则可以省略指针信息。这由 TEXT 指令上的本地帧大小 $0-n 注释表示。如果函数不包含调用指令，也可以省略指针信息。否则，本地堆栈帧不能包含指针，并且汇编必须通过执行伪指令 TEXTNO_LOCAL_POINTERS 来确认这一事实。因为通过移动堆栈来实现堆栈大小调整，所以堆栈指针可能在任何函数调用期间发生变化：甚至指向堆栈数据的指针也不能保存在局部变量中。

汇编函数应始终给出 Go 原型，既可以提供参数和结果的指针信息，也可以 go vet 检查用于访问它们的偏移量是否正确。
``` -->


### 1.6.1 使用语法

Go 语言汇编中，函数声明格式如下：
```asm
告诉汇编器该数据放到TEXT区
  ^             静态基地址指针(告诉汇编器这是基于静态地址的数据)
  |                     ^    
  |                     |   标签   函数入参+返回值占用空间大小
  |                     |    ^      ^
  |                     |    |      |
TEXT pkgname·funcname<ABIInternal>(SB),TAG,$16-24
     ^       ^       ^           ^
     |       |       |           |
函数所属包名  函数名  表示ABI类型   函数栈帧大小(本地变量占用空间大小)           
```
|格式|描述|
| :---- | :---- |
|TEXT|定义函数标识|
|pkgname|函数包名，可以不写，也可以""替代|
|·|在程序链接后会转换为.|
|\<ABIInternal\>|表示ABI类型为ABIInternal，不过汇编 ABIInternal 只能在 runtime 中使用。ABI0 不用写|
|(SB)|让 SB 认识这个函数，即生成全局符号，有用绝对地址|
|TAG|标签，表示函数某些特殊功能，多个标签可以通过\|连接。定义和全局变量的 tag 一样|
|$16-24|16表示函数栈帧大小，24表示入参和返回大小|

一些说明：
1. 栈帧大小包括局部变量和可能需要的额外调用函数的参数空间的总大小，但不不包含调用其他函数时的ret address的大小。\
2. 汇编文件中，函数名以 '·' 开头或连接 pkgname 是固定格式。
3. go 函数采用的是 caller-save 模式，被调用者的参数、返回值、栈位置都由调用者维护。
<!-- 若不指定NOSPLIT，arguments size必须指定。 -->

go 语言编译成汇编:
```sh
go tool compile -S xxx.go
go build -gcflags -S xxx.go
```
<!-- go build -gcflags "-N -l" -ldflags=-compressdwarf=false -o main.out main.go -->
从二进制反编译为汇编：
```sh
go tool objdump -s "main.main" main.out > main.S
```


### 1.6.2 使用例子

Go 函数调用汇编函数：
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

汇编调用 go语言函数：
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
```asm
#include "textflag.h"

// func output(a,b int) int
TEXT ·output(SB), NOSPLIT, $24-24
    MOVQ a+0(FP), DX    // arg a
    MOVQ DX, 0(SP)      // arg x
    MOVQ b+8(FP), CX    // arg b
    MOVQ CX, 8(SP)      // arg y
    CALL ·add(SB)       // 在调用 add 之前，已经把参数都通过物理寄存器 SP 搬到了函数的栈顶
    MOVQ 16(SP), AX     // add 函数会把返回值放在这个位置
    MOVQ AX, ret+16(FP) // return result
    RET
```


### 1.6.1 汇编函数中用到的一些特殊命令（伪指令）

GO_RESULTS_INITIALIZED：
如果Go 汇编函数返回值含指针，则该指针信息必须由 Go 源文件中的函数的 Go 原型提供，即使对于未直接从 Go 调用的汇编函数也是如此。 如果返回值将在调用指令期间保存实时指针，则该函数中应首先将结果归零， 然后执行伪指令 GO_RESULTS_INITIALIZED。 表明该堆栈位置应该执行进行GC扫描，避免其指向的内存地址呗 GC 意外回收。

NO_LOCAL_POINTERS: 就是字面意思，表示函数没有指针类型的局部变量。

<!-- [《Go语言高级编程》3.6.3 PCDATA 和 FUNCDATA](https://chai2010.cn/advanced-go-programming-book/ch3-asm/ch3-06-func-again.html#363-pcdata-%E5%92%8C-funcdata) -->

PCDATA: Go 语言生成的汇编，利用此伪指令表明汇编所在的原始 Go 源码的位置(file&line&func)，用于生成 PC 表格。runtime.FuncForPC 函数就是通过 PC 表格得到结果的。一般由编译器自动插入，手动维护并不现实。

FUNCDATA: 和 PCDATA 的格式类似，用于生成 FUNC 表格。FUNC 表格用于记录函数的参数、局部变量的指针信息，GC 依据它来跟踪栈中指针指向内存的生命周期，同时栈扩缩容的时候也是依据它来确认是否需要调整栈指针的值（如果指向的地址在需要扩缩容的栈中，则需要同步修改）。

<!-- [debug/gosym 官方文档](https://pkg.go.dev/debug/gosym)\
[debug/gosym: 腾讯开发者社区翻译](https://cloud.tencent.com/developer/section/1141111)\
[腾讯的张杰的《Go语言调试器开发》”pkg debug/gosym 应用“](https://www.hitzhangjie.pro/debugger101.io/7-headto-sym-debugger/6-gopkg-debug/2-gosym.html)\
[Go 1.2 Runtime Symbol Information](https://docs.google.com/document/d/1lyPIbmsYbXnpNj57a261hgOYVpNRcgydurVQIyZOz_o/pub)\
[issue-debug/gosym: report different line number than cmd/addr2line](https://github.com/golang/go/issues/56869)\ 

```md
debug/dwarf   //dwarf包实现了可执行文件的调试信息
debug/elf     //elf包实现了对ELF对象文件的访问接口
debug/gosym   //gosym实现了对gc编译器生成的go二进制文件中嵌入的go符号和行号表的访问接
debug/macho   //macho包实现了Mach-O对象文件的访问接口
debug/pe    //pe包实现了对PE文件的访问接口
debug/plan9obj  //plan9obj包实现了 Plan 9 a.out对象文件的访问接口
```
-->

<!-- [《Go语言高级编程》3.6.4 方法函数](https://chai2010.cn/advanced-go-programming-book/ch3-asm/ch3-06-func-again.html#364-%E6%96%B9%E6%B3%95%E5%87%BD%E6%95%B0) -->



## 1.7 条件编译
Go 语言仅支持有限的条件编译规则:
  1. 根据文件名编译。
  2. 根据 build 注释编译。
 
根据文件名编译类似 *_test.go，通过添加平台后缀区分，比如: asm_386.s、asm_amd64.s、asm_arm.s、asm_arm64.s、asm_mips64x.s、asm_linux_amd64.s、asm_bsd_arm.s 等.

根据 build 注释编译，就是在源码中加入区分平台和编译器版本的注释。比如：
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

gofmt 可以兼容处理两种形式，处理原则是：如果一个源码文件只有// +build形式的指示符，gofmt会将与其等价的//go:build 行加入。否则，如果一个源文件中同时存在这两种形式的指示符行，那么//+build行的信息将被//go:build行的信息所覆盖。



# 2、 go语言ABI
参考文档：\
[Go internal ABI specification](https://go.googlesource.com/go/+/refs/heads/dev.regabi/src/cmd/compile/internal-abi.md)\
[Proposal: Create an undefined internal calling convention](https://go.googlesource.com/proposal/+/master/design/27539-internal-abi.md)
<!-- [翻译](http://hushi55.github.io/2021/04/19/Go-internal-ABI-specification)\ -->
<!-- [internal/abi:文档，abi0 寄存器传参规则](https://pkg.go.dev/internal/abi)\ -->
<!-- [issue-register ABI: allow selecting ABI when Go code refers to assembly routine #44065](https://github.com/golang/go/issues/44065)\ -->
<!-- [Proposal: Create an undefined internal calling convention](https://github.com/golang/proposal/blob/master/design/27539-internal-abi.md#compatibility) : ABIInternal 封装成 ABI0 的桥接代码\ -->
<!-- [翻译？](https://elkeid.bytedance.com/Chinese/RASP/golang.html#golang-internal-abi) -->


名词解释：\
ABI: application binary interface, 应用程序二进制接口，规定了程序在机器层面的操作规范和调用规约。\
调用规约: calling convention, 所谓“调用规约”是调用方和被调用方对于函数调用的一个明确的约定，包括：函数参数与返回值的传递方式、传递顺序。只有双方都遵守同样的约定，函数才能被正确地调用和执行。如果不遵守这个约定，函数将无法正确执行。

Go从1.17.1版本开始支持多ABI：\
    1. 为了兼容性各平台保持通用性，保留历史版本 ABI，并更名为 ABI0。\
    2. 为了更好的性能，增加新版本 ABI 取名 ABIInternal。\
ABI0 遵循平台通用的函数调用约定，实现简单，不用担心底层cpu架构寄存器的差异；ABIInternal 可以指定特定的函数调用规范，可以针对特定性能瓶颈进行优化，在多个Go版本之间可以迭代，灵活性强，支持寄存器传参提升性能。\
Go 汇编为了兼容已存在的汇编代码，保持使用旧的 ABI0。

Go为什么在有了 ABI0 之后，还要引入 ABIInternal？ 当然是为了性能！据官方测试，寄存器传参可以带来 5% 的性能提升。

我们看一个例子：
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
运行上面代码会报错：
main.Print: relocation target fmt.Println not defined for ABI0 (but is defined for ABIInternal)

原因是，fmt·Println 函数默认使用的 ABI 标准是 ABIInternal，而 Go 语言手写的汇编使用的 ABI 格式是 ABI0，二者标准不一样不能直接调用。不过 Go 语言可以通过 //go:linkname 的方式为 ABIInternal 生成 ABI0 包装。
```go
package main
import ( 
  "fmt"
)
//go:linkname Println fmt.Println
func Println(a ...any) (n int, err error)

func Print(delta interface{})

func main() { 
  Print("hello")
}
```
```asm
#include "textflag.h"
TEXT ·Print(SB), NOSPLIT, $48-16   
  LEAQ strp+0(FP),AX
  MOVQ AX, 0(SP)        // []interface{} slice 的 pointer
  MOVQ $1, BX    
  MOVQ BX, 8(SP)        // slice 的 len
  MOVQ BX, 16(SP)       // slice 的 cap
  CALL fmt·Println(SB)  // //go:linkname 为 fmt.Println 生成一个 ABI0 包装后，汇编可以直接调用
  RET

```
简单说明：\
函数 fmt.Println 是一个变参函数，变参(a ...any)实际上是 (a []any)的语法糖。参数中，slice 占24Byte，int 占 8Byte，error 是 interface 类型，占 16Byte，加起来是 48 Byte。所以，调用 此函数时，caller 需要再栈上准备 24Byte 空间。而 Print 的入参刚好是一个 interface{} 类型，和 any 一致，所以只要把 Print 函数的入参的地址赋给 a 的指针，并把 a 的 len 和 cap 设置为1，就可以 调用fmt·Println函数了。如以上代码所示。



# 3、内存管理和 GC 对汇编的影响

## 3.1 调用栈扩缩容对汇编的影响

为了减少对内存的占用，goroutine 启动时 runtime 只给它分配了很少的栈内存。所有函数（标记 go:nosplit 的除外）的序言部分（启动指令）会插入分段检查，当发现栈溢出（栈空间不足）时，就会调用 runtime.morestack，执行栈拓展逻辑:
1. 旧版本的 Go 编译器采用了分段栈机制实现栈拓展，当一个 goroutine 的执行栈溢出时，就增加一个栈内存作为调用栈的补充，新旧栈彼此没有连续。 这种设计的缺陷很容易破坏缓存的局部性原理，从而降低程序的运行时性能。 
2. Go 1.3 版本开始，引入了连续栈（拷贝栈）机制，并把 goroutine 的初始栈大小由 8KB 降低到了 2KB。当一个执行栈发生溢出时，新建一个两倍于原栈大小的新栈，并将原栈整个拷贝到新栈上，保证整个栈是连续的。

栈的拷贝有些副作用：
1. 如果栈上存在指向当前被拷贝栈的指针，当栈拷贝执行完成后，这个指针还是指向原栈，需要更新。
2. goroutine 的 g 结构体上的 gobuf 成员也还是指向旧的栈，也需要更新。

<!-- 此外，不难发现，只有栈上分配的指针才能指向栈上的地址，否则这个指针指向的对象会重新在堆中进行分配（逃逸）。 -->

除了正在拷贝的栈中可能存在指向自己的的指针外，还有没有其他存活中的内存有指向即将失效的栈空间的指针呢？\
答案在 [go 逃逸分析源码](https://github.com/golang/go/blob/go1.18.10/src/cmd/compile/internal/escape/escape.go#L22) 中：
```go
// Escape analysis.
//
// Here we analyze functions to determine which Go variables
// (including implicit allocations such as calls to "new" or "make",
// composite literals, etc.) can be allocated on the stack. The two
// key invariants we have to ensure are: (1) pointers to stack objects
// cannot be stored in the heap, and (2) pointers to a stack object
// cannot outlive that object (e.g., because the declaring function
// returned and destroyed the object's stack frame, or its space is
// reused across loop iterations for logically distinct variables).
//
```
其中 “(1) pointers to stack objects cannot be stored in the heap” 表明指向栈对象的指针不能存储在堆中。

拷贝栈理论上没有上限，但是一般都设置了上限。
当新的栈大小超过了 maxstacksize 就会抛出”stack overflow“的异常。maxstacksize是在runtime.main中设置的。64位 系统下栈的最大值1GB、32位系统是250MB。\
参考代码：
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

由拷贝栈的原理可知，拷贝栈对 Go 汇编是透明的。

## 3.2 GC 对汇编的影响
由于 GC 会动态回收没有被引用的堆内存，而 goroutine 的调用栈在堆空间，所以如果调用栈中存了动态申请的内存，就需要告诉 GC 栈中含内存。上文中说到的伪指令 FUNCDATA、GO_RESULTS_INITIALIZED、NO_LOCAL_POINTERS 就是干这个事的。\
由于 FUNCDATA 伪指令几乎只能由编译器维护，所以在手写的汇编函数本地内存栈中保存指向动态内存的指针几乎是一种奢望。

# 4、 函数内联和汇编
参考文档：
<!-- [详解Go内联优化](https://segmentfault.com/a/1190000039146279) \ -->
[Go: Inlining Strategy & Limitation](https://medium.com/a-journey-with-go/go-inlining-strategy-limitation-6b6d7fc3b1be)
<!-- [译-Go语言inline内联的策略与限制](https://www.pengrl.com/p/20028/)\ -->
<!-- 
[Go: Inlining Strategy & Limitation](https://medium.com/a-journey-with-go/go-inlining-strategy-limitation-6b6d7fc3b1be)\
[[译] Go语言inline内联的策略与限制](https://www.pengrl.com/p/20028/)\
[Go中的内联优化](https://dashen.tech/2021/05/22/Go%E4%B8%AD%E7%9A%84%E5%86%85%E8%81%94%E4%BC%98%E5%8C%96/)\
[Go 内联](https://blog.crazytaxii.com/posts/go_inline_strategy_and_limitation/)\
[通过实例理解Go内联优化](https://tonybai.com/2022/10/17/understand-go-inlining-optimisations-by-example/) : greate -->

## 4.1 查看内联情况
可以通过执行以下命令，输出被内联的函数：
```sh
go build -gcflags="-m" main.go
# 输出结果：
# ./op.go:3:6: can inline add
# ./op.go:7:6: can inline sub
# ./main.go:16:11: inlining call to sub
# ./main.go:14:11: inlining call to add
# ./main.go:7:12: inlining call to fmt.Printf
```
或者使用参数 -gflags="-m -m" 运行，查看编译器的详细优化策略：
```sh
go build -gcflags="-m -m" main.go
```
输出很详细：
```sh
# command-line-arguments
./main.go:10:6: cannot inline main: function too complex: cost 106 exceeds budget 80
./main.go:20:12: inlining call to fmt.Printf
./main.go:23:6: can inline toEface with cost 0 as: func() {  }
./main.go:18:2: shlx escapes to heap:
./main.go:18:2:   flow: i = &{storage for shlx}:
./main.go:18:2:     from shlx (spill) at ./main.go:18:2
./main.go:18:2:     from i = shlx (assign) at ./main.go:18:4
./main.go:18:2:   flow: {storage for ... argument} = i:
./main.go:18:2:     from ... argument (slice-literal-element) at ./main.go:20:12
./main.go:18:2:   flow: fmt.a = &{storage for ... argument}:
./main.go:18:2:     from ... argument (spill) at ./main.go:20:12
./main.go:18:2:     from fmt.format, fmt.a := "%+v", ... argument (assign-pair) at ./main.go:20:12
./main.go:18:2:   flow: {heap} = *fmt.a:
./main.go:18:2:     from fmt.Fprintf(io.Writer(os.Stdout), fmt.format, fmt.a...) (call parameter) at ./main.go:20:12
./main.go:17:2: x escapes to heap:
./main.go:17:2:   flow: i = &{storage for x}:
./main.go:17:2:     from x (spill) at ./main.go:17:2
./main.go:17:2:     from i = x (assign) at ./main.go:17:4
./main.go:17:2:   flow: {storage for ... argument} = i:
./main.go:17:2:     from ... argument (slice-literal-element) at ./main.go:20:12
./main.go:17:2:   flow: fmt.a = &{storage for ... argument}:
./main.go:17:2:     from ... argument (spill) at ./main.go:20:12
./main.go:17:2:     from fmt.format, fmt.a := "%+v", ... argument (assign-pair) at ./main.go:20:12
./main.go:17:2:   flow: {heap} = *fmt.a:
./main.go:17:2:     from fmt.Fprintf(io.Writer(os.Stdout), fmt.format, fmt.a...) (call parameter) at ./main.go:20:12
./main.go:17:2: x escapes to heap
./main.go:18:2: shlx escapes to heap
./main.go:20:12: ... argument does not escape
```

Go编译器默认将进行内联优化，可以通过 -gcflags="-l" 选项全局禁用内联，与一个-l禁用内联相反，如果传递两个或两个以上的-l则会打开内联，并启用更激进的内联策略。\
例如以下代码：
```go
// 3.1: var closure = NewClosure()
func main() {
	// 3.2: var closure func() int
	var closure = NewClosure()
	closure()
	// 3.3: closure = NewClosure()
	closure()
}

func NewClosure() func() int {
	i := 0
	return func() int {
		i++
		return i
	}
}
```
命令 go build -gcflags="-m" main.go 和 go build -gcflags="-m -l -l" main.go 都是输出：
```sh
./main.go:19:6: can inline NewClosure
./main.go:21:9: can inline NewClosure.func1
./main.go:13:26: inlining call to NewClosure
./main.go:21:9: can inline main.func1
./main.go:14:9: inlining call to main.func1
./main.go:16:9: inlining call to main.func1
./main.go:13:26: func literal does not escape
./main.go:20:2: moved to heap: i
./main.go:21:9: func literal escapes to heap
```
命令 go build -gcflags="-m" main.go 输出：
```sh
./main.go:20:2: moved to heap: i
./main.go:21:9: func literal escapes to heap
```

## 4.2 内联前后性能对比
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

func BenchmarkInline(b *testing.B) {
	x, y := 1, 2
	b.Run("BenchmarkNoInline", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			maxNoinline(x, y)
		}
	})
	b.Run("BenchmarkInline", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			maxInline(x, y)
		}
	})
}
```
在程序代码中，想要禁止编译器内联优化很简单，在函数定义前一行添加//go:noinline即可。以下是性能对比结果:
```sh
BenchmarkInline/BenchmarkNoInline-12  886137398  1.248  ns/op  0 B/op  0 allocs/op
BenchmarkInline/BenchmarkInline-12   1000000000  0.2506 ns/op  0 B/op  0 allocs/op
```
因为函数体内部的执行逻辑非常简单，此时内联与否的性能差异主要体现在函数调用的固定开销上。显而易见，该差异是非常大的。

## 4.3 内联条件

Go 语言代码函数内联的策略每个编译器版本都有细微差别，比如新版已支持含 for 和 闭包 的函数内联。1.18 版本的部分无法内联的规则如下:
1. 函数标注 "go:noinline" 注释。
2. 函数标注 "go:norace" 注释，且使用 "-gcflags=-d checkptr" 参数编译。
3. 函数标注 "go:cgo_unsafe_args" 注释。
4. 函数标注 "go:uintptrescapes" 注释。
5. 函数只有声明而没有函数体：比如函数实体在汇编文件 xxx.s 中。
6. 超过小代码量边界的函数：内联的小代码量边界是 80 个节点（抽象语法树AST的节点）。
7. 函数中含某些关键字的函数：比如 select、defer、go、recover 等。
8. 一些特殊的内部函数：比如 runtime.getcallerpc、runtime.getcallersp （这俩太特殊了）。
9. 函数内部使用 type 关键字重定义了类型：比如 "type Int int" 或 "type Int = int"。
10. 作为尾递归调用时。

此外，还有一些编译器觉得内联成本很低，所以必然内联的函数：\
1. "runtime" package 下的 "heapBits.nextArena" 和 "builtin" package 下的 "append"。
2. "encoding/binary" package 下的：
 "littleEndian.Uint64", "littleEndian.Uint32", "littleEndian.Uint16","bigEndian.Uint64", "bigEndian.Uint32", "bigEndian.Uint16","littleEndian.PutUint64", "littleEndian.PutUint32", "littleEndian.PutUint16","bigEndian.PutUint64", "bigEndian.PutUint32", "bigEndian.PutUint16", "append"。

由规则 5 可知，Go 语言汇编是无法内联的。

此外，关于闭包内联是一个比较复杂的话题，据笔者测试，1.18 有如上规则：
1. 满足条件的闭包可以内联。
2. 闭包占用函数的 15 个 AST 节点。
3. 变量保存的闭包，如果是局部变量且没有重新赋值过，则可以被内联。

关于闭包内联的第 3 条规则，有如下例子：
```go
// 3.1: var closure = NewClosure()
func main() {
	// 3.2: var closure func() int
	var closure = NewClosure()
	closure()
	// 3.3: closure = NewClosure()
	closure()
}

func NewClosure() func() int {
	i := 0
	return func() int {
		i++
		return i
	}
}
```
执行 go build -gcflags="-m" ./ 输出如下
```sh
./main.go:19:6: can inline NewClosure
./main.go:21:9: can inline NewClosure.func1
./main.go:13:26: inlining call to NewClosure
./main.go:21:9: can inline main.func1
./main.go:14:9: inlining call to main.func1
./main.go:16:9: inlining call to main.func1
./main.go:13:26: func literal does not escape
./main.go:20:2: moved to heap: i
./main.go:21:9: func literal escapes to heap
```
表明闭包 closure 可以内联。\
如果把 3.1 或 3.2 或 3.3 的注释打开，则将会输出：
```sh
./main.go:19:6: can inline NewClosure
./main.go:21:9: can inline NewClosure.func1
./main.go:13:22: inlining call to NewClosure
./main.go:13:22: func literal does not escape
./main.go:20:2: moved to heap: i
./main.go:21:9: func literal escapes to heap
```
表明闭包 closure 无法内联。

此外，如果想禁用闭包内联，可以使用 -gcflags="-d=inlfuncswithclosures=0" 或-gcflags="-d inlfuncswithclosures=0" 参数编译。
```sh
go build -gcflags="-d=inlfuncswithclosures=0" main.go
go build -gcflags="-d inlfuncswithclosures=0" main.go
```

如果想了解 go 1.18 的内联检查逻辑，可以看这个源码： [inline.CanInline](https://github.com/golang/go/blob/go1.18.10/src/cmd/compile/internal/inline/inl.go#L76) 和 [(*inline.hairyVisitor).doNode](https://github.com/golang/go/blob/go1.18.10/src/cmd/compile/internal/inline/inl.go#L250)。
其调用顺序是： inline.CanInline --> inline.hairyVisitor.tooHairy --> inline.hairyVisitor.doNode。
```go
// CanInline determines whether fn is inlineable.
// If so, CanInline saves copies of fn.Body and fn.Dcl in fn.Inl.
// fn and fn.Body will already have been typechecked.
func CanInline(fn *ir.Func) {
...
	// If marked "go:noinline", don't inline
	if fn.Pragma&ir.Noinline != 0 {
		reason = "marked go:noinline"
		return
	}

	// If marked "go:norace" and -race compilation, don't inline.
	if base.Flag.Race && fn.Pragma&ir.Norace != 0 {
		reason = "marked go:norace with -race compilation"
		return
	}

	// If marked "go:nocheckptr" and -d checkptr compilation, don't inline.
	if base.Debug.Checkptr != 0 && fn.Pragma&ir.NoCheckPtr != 0 {
		reason = "marked go:nocheckptr"
		return
	}

	// If marked "go:cgo_unsafe_args", don't inline, since the
	// function makes assumptions about its argument frame layout.
	if fn.Pragma&ir.CgoUnsafeArgs != 0 {
		reason = "marked go:cgo_unsafe_args"
		return
	}

	// If marked as "go:uintptrescapes", don't inline, since the
	// escape information is lost during inlining.
	if fn.Pragma&ir.UintptrEscapes != 0 {
		reason = "marked as having an escaping uintptr argument"
		return
	}

	// The nowritebarrierrec checker currently works at function
	// granularity, so inlining yeswritebarrierrec functions can
	// confuse it (#22342). As a workaround, disallow inlining
	// them for now.
	if fn.Pragma&ir.Yeswritebarrierrec != 0 {
		reason = "marked go:yeswritebarrierrec"
		return
	}

	// If fn has no body (is defined outside of Go), cannot inline it.
	if len(fn.Body) == 0 {
		reason = "no function body"
		return
	}
...
    visitor := hairyVisitor{
		budget:        inlineMaxBudget, // inlineMaxBudget == 80 
		extraCallCost: cc,
	}
    if visitor.tooHairy(fn) {
		reason = visitor.reason
		return
	}
...
}

func (v *hairyVisitor) tooHairy(fn *ir.Func) bool {
	v.do = v.doNode // cache closure
	if ir.DoChildren(fn, v.do) {
		return true
	}
...
}

func (v *hairyVisitor) doNode(n ir.Node) bool {
...
	case ir.OSELECT,
		ir.OGO,
		ir.ODEFER,
		ir.ODCLTYPE, // can't print yet
		ir.OTAILCALL:
		v.reason = "unhandled op " + n.Op().String()
		return true
...
}
```
<!-- ```go
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
``` -->


<!-- ## 4.3 内联表
函数内联就是直接将调用函数的地方替换为函数的实现。这样就会将函数调用的过程抹掉，会导致调用栈信息不完整，所以需要 Go 在内部维持了一份内联函数的映射关系。
首先它会生成一个内联树，我们可以通过-gcflags="-d pctab=pctoinline"参数查看。
Go在生成的代码中映射了内联函数。并且，也映射了行号和源码文件可以通过-d pctab=pctoline 参数查看行号，可以通过-gcflags="-d pctab=pctofile"查看源文件。

由以上信息，我们可以得到一张 PC-函数/文件/行号映射表. -->



# 5、 有哪些有意思的使用场景

## 5.1、 获取 goid
goid 即 goroutine id，最常用三方库应该就是 [petermattis/goid](https://github.com/petermattis/goid)， 里通过汇编获取 goid 的代码关键逻辑如下：

[runtime_go1.9.go](https://github.com/petermattis/goid/blob/master/runtime_go1.9.go) 代码：
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
[goid_go1.5_amd64.go](https://github.com/petermattis/goid/blob/master/goid_go1.5_amd64.go) 代码：
```go
//go:build (amd64 || amd64p32) && gc && go1.5
// +build amd64 amd64p32
// +build gc
// +build go1.5

package goid

func Get() int64
```
[goid_go1.5_amd64.s](https://github.com/petermattis/goid/blob/master/goid_go1.5_amd64.s) 代码：
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
不过这样获取 goid 有一个局限性，就是如果当前处于 g0 调用栈（系统调用或CGO函数中）时，拿到的不是当前 g 的 goid，而是 是 g0 的 goid。在这种情况下 g.m.curg.goid 才是当前 g 的 goid。
参考[go/src/runtime/HACKING.md](https://github.com/golang/go/blob/go1.18.10/src/runtime/HACKING.md#getg-and-getgmcurg)
```md
getg() and getg().m.curg

To get the current user g, use getg().m.curg.

getg() alone returns the current g, but when executing on the system or signal stacks, this will return the current M's "g0" or "gsignal", respectively. This is usually not what you want.

To determine if you're running on the user stack or the system stack, use getg() == getg().m.curg.
```

除了 goid，pid也可以用汇编获取：\
[choleraehyq/pid](https://github.com/choleraehyq/pid)
是一个 fork 
[petermattis/goid](https://github.com/petermattis/goid) 
的仓库，里面增加了获取 pid 的实现，实现代码如下：

[p_m_go1.19.go](https://github.com/choleraehyq/pid/blob/master/p_m_go1.19.go) 代码：
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
[pid_go1.5.go](https://github.com/choleraehyq/pid/blob/master/pid_go1.5.go) 代码：
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
[pid_go1.5_amd64.s](https://github.com/choleraehyq/pid/blob/master/pid_go1.5_amd64.s) 代码：
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
不过，通过这种方式获取的 pid 也有一个局限性：在持有 pid 之后的时间里，可能当前 goroutine 已经被调度到其他 P 上了，也就是在使用 pid 的时候当前 pid 已经改变了。\
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

	mp.locks++     // 锁定 P 的调度
	return int(mp.p.ptr().id)
}

//go:nosplit
func procUnpin() {
	_g_ := getg()
	_g_.m.locks--
}
```
通过 procPin 函数锁定 P 的调度后再使用 pid，然后通过 procUnpin 释放 P。不过这里也需要谨慎使用，使用不当会对性能产生严重影响。

以上获取 goid 的方式以前是还有一个比较大的确定，就是如果 Go 编译器修改了 g 的结构体，就需要重新适配。\

[《Go语言高级编程》第三章第8节](https://github.com/chai2010/advanced-go-programming-book/blob/master/ch3-asm/ch3-08-goroutine-id.md) 的实现可以避免这个问题。
其原理是，通过汇编构建一个 g 类型的 interface{}，然后通过反射获取 goid 成员的偏移量。
根据原理，可以如下实现：
```go
func Getg() int64
func getgi() interface{}
var g_goid_offset uintptr = func() uintptr {
	g := getgi()
	if f, ok := reflect.TypeOf(g).FieldByName("goid"); ok {
		return f.Offset
	}
	panic("can not find g.goid field")
}()
```

```asm
TEXT ·Getg(SB), NOSPLIT, $0-8
    MOVQ (TLS), AX
	ADDQ ·g_goid_offset(SB),AX
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
    // MOVQ (TLS), AX
    MOVQ $0, AX

    // get runtime.g type
    MOVQ $type·runtime·g(SB), BX

    // MOVQ BX, ·runtime_g_type(SB)

    // return interface{}
    MOVQ BX, ret_type+0(FP)
    MOVQ AX, ret_data+8(FP)
    RET

```
实际上还可以继续简化：
```go
var runtime_g_type uint64  // go 源码中声明

var gGoidOffset uintptr = func() uintptr { //nolint
	var iface interface{}
	type eface struct {
		_type uint64
		data  unsafe.Pointer
	}
    // 结构 iface 后，修改他的类型为 g
	(*eface)(unsafe.Pointer(&iface))._type = runtime_g_type
	if f, ok := reflect.TypeOf(iface).FieldByName("goid"); ok {
		return f.Offset
	}
	panic("can not find g.goid field")
}()
```
```asm
GLOBL ·runtime_g_type(SB),NOPTR,$8  
DATA ·runtime_g_type+0(SB)/8,$type·runtime·g(SB)  // 汇编中初始化。汇编中可以访问 package 的私有变量
```

## 5.2、Monkey Patch
Go 语言实现猴子打点的 package 不一定需要使用汇编，比如 [bouk/monkey](https://github.com/bouk/monkey) 和 [go-kiss/monkey](https://github.com/go-kiss/monkey)。
不过字节开源的 [monkey](https://github.com/bytedance/mockey) 和 内部的 [mockito](https://code.byted.org/luoshiqi/mockito) 都使用了汇编。他们有一个同源的依赖库，分别在
[mockey/internal/monkey](https://github.com/bytedance/mockey/tree/v1.1.1/internal/monkey) 目录和 
[mockito/monkey](https://code.byted.org/luoshiqi/mockito/tree/master/monkey) 目录下。

<!-- 该库来源不明，不过 README.md 文档有如下说明
```md
This library fork form chenzhuoyu/monkey
enhancement:

allocPage/freePage  used go.runtime
support windows
support arm64 darwin (and refactored)
``` -->

其 Patch() 的调用路径如下：\
Build() -> Patch() -> PatchValue() -> WriteWithSTW() -> Write() -> do_replace_code()\
其中 do_replace_code() 是汇编实现的，作用是使用 mprotect 系统调用来修改内存权限。原因是：可执行代码区是只读的，需要修改为可读写后才能修改，修改为可执行的才能执行（有想用 Go 写病毒的，可以参考一下）。

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
而 Write 函数的实现在
[github.com/bytedance/mockey/internal/monkey/mem/write_linux.go](https://github.com/bytedance/mockey/blob/v1.1.1/internal/monkey/mem/write_linux.go)，其代码如下：
```go
package mem

import (
	"syscall"

	"github.com/bytedance/mockey/internal/monkey/common"
)

func Write(target uintptr, data []byte) error {
	do_replace_code(target, common.PtrOf(data), uint64(len(data)), syscall.
    , syscall.PROT_READ|syscall.PROT_WRITE, syscall.PROT_READ|syscall.PROT_EXEC)
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
do_replace_code 函数的汇编实现在
[github.com/bytedance/mockey/internal/monkey/mem/write_linux_amd64.s](https://github.com/bytedance/mockey/blob/v1.1.1/internal/monkey/mem/write_darwin_amd64.s)，代码如下：
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

## 5.2、 优化获取行号性能
[golang文件行号探索](https://tech.bytedance.net/articles/7111885881002164238) 中有详细说明，代码如下：
```go
//stack_amd64.go

type Line uintptr

func NewLine() Line

var rcuCache unsafe.Pointer = func() unsafe.Pointer {
    m := make(map[Line]string)
    return unsafe.Pointer(&m)
}()

func (l Line) LineNO() (line string) {
    mPCs := *(*map[Line]string)(atomic.LoadPointer(&rcuCache))
    line， ok := mPCs[l]
    if !ok {
        file， n := runtime.FuncForPC(uintptr(l)).FileLine(uintptr(l))
        line = file + ":" + strconv.Itoa(n)
        mPCs2 := make(map[Line]string， len(mPCs)+10)
        mPCs2[l] = line
        for {
            p := atomic.LoadPointer(&rcuCache)
            mPCs = *(*map[Line]string)(p)
            for k， v := range mPCs {
                mPCs2[k] = v
            }
            swapped := atomic.CompareAndSwapPointer(&rcuCache p， unsafe.Pointer(&mPCs2))
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
	SUBQ	 $1, AX             // 注意，这里要 -1
    MOVQ     AX， ret+0(FP)
    RET

```
该代码除了使用汇编获取行号外，还是用了无锁的 RCU(Read-copy update) 算法提升并发查询速度。\
还有一点要注意的，retpc-8(FP) 是函数返回地址，也就是调用指令 CALL 的下一行指令， 所以需要 -1 才能得到CALL指令的 pc，[参考](https://github.com/golang/go/blob/dev.boringcrypto.go1.18/src/runtime/traceback.go#L354)。

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
    SUBQ	 $1, BX 
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
Go 语言源码里的字符串比较函数，实际上使用了 SIMD 指令加速，由汇编实现。源码在：
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
loop:
	CMPQ	R8, $16
	JBE	_0through16
	MOVOU	(SI), X0
	MOVOU	(DI), X1
	PCMPEQB X0, X1
	PMOVMSKB X1, AX
	XORQ	$0xffff, AX	// convert EQ to NE
	JNE	diff16	// branch if at least one byte is not equal
	ADDQ	$16, SI
	ADDQ	$16, DI
	SUBQ	$16, R8
	JMP	loop
    ···
```
这里 MOVOU、PCMPEQB、PMOVMSKB 等就是 SIMD 指令。\
如果想详细了解 SIMD 指令可以看一下 Intel 的官方文档 [Intel® Intrinsics Guide](https://www.intel.com/content/www/us/en/docs/intrinsics-guide/index.html)。\
笔者收集了相对较完整的 SSE 和 SSE2 指令集说明（ [simd_sse 指令集.md](https://github.com/lxt1045/blog/blob/main/sample/asm/simd_sse%20%E6%8C%87%E4%BB%A4%E9%9B%86.md)、[simd_sse2指令集.md](https://github.com/lxt1045/blog/blob/main/sample/asm/simd_sse2%E6%8C%87%E4%BB%A4%E9%9B%86.md)），感兴趣的同学欢迎围观。\
另，据笔者的尝试，SSE 和 SSE2 指令是可以直接在 Go 语言会便利使用的。有想法的同学可以自己验证一下其他 SIMD 指令。

## 5.5、 字符串搜索
我们常用的字符串搜索函数 strings.Index，也使用了汇编实现的 SIMD 指令加速。代码在：
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
IndexByteString 函数声明在
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
汇编实现在
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
笔者验证了一下 IndexByte 和自定义通过 for 循环实现建的差别：
```go
func BenchmarkIndexByte(b *testing.B) {
	b.Run("IndexByte", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			str := testdata.TwitterJsonOut
			n := 0
			k := 0
			for {
				j := strings.IndexByte(str[k:], ']')
				if j < 0 {
					break
				}
				n++
				k += j + 1
			}
			_ = n
		}
		b.SetBytes(int64(b.N))
		b.StopTimer()
	})
	b.Run("for", func(b *testing.B) {
		b.ReportAllocs()
		str := testdata.TwitterJsonOut
		for i := 0; i < b.N; i++ {
			n := 0
			for i := 0; i < len(str); i++ {
				if str[i] == ']' {
					n++
				}
			}
			_ = n
		}
		b.SetBytes(int64(b.N))
		b.StopTimer()
	})
}
```
结果如下：
```
BenchmarkIndexByte/IndexByte
BenchmarkIndexByte/IndexByte-12    3072980    387.5 ns/op    7929621.19 MB/s    0 B/op	       0 allocs/op
BenchmarkIndexByte/for
BenchmarkIndexByte/for-12          516663    2417 ns/op      213777.66 MB/s    0 B/op	       0 allocs/op
```
由结果可知，SIMD 的加速性能还是挺好的。不过，实际上如果 strings.IndexByte() 字符串很短 或 所查找的字符在字符串中大量存在 的话，性能甚至会比 for 循环慢。这个可以自行验证一下。

## 5.6、 自定义SIMD优化
如果感兴趣，可以照着 Go 编译器里的汇编抄，慢慢尝试。\
github 上也有许多项目可以抄，比如：
[minio/sha256-simd](https://github.com/minio/sha256-simd)

## 5.7、 随意跳转
[关于 golang 错误处理的一些优化想法](https://tech.bytedance.net/articles/7120632282095812644) 中有详细说明。
测试代码如下：
```go
func TestTagTry0(t *testing.T) {
	defer func() {
		fmt.Printf("1 -> ")
	}()

	tag, err1 := NewTag() // 当 tag.Try(err) 时，跳转此处并返回 err1
	fmt.Printf("2 -> ")
	if err1 != nil {
		fmt.Printf("3 -> ")
		return
	}

	defer func() {
		fmt.Printf("4 -> ") // 由于的缺陷：这里 debug 下 defer 不内联，会执行；release 下 defer 内联，不会执行
	}()

	fmt.Printf("5 -> ")
    err2 := errors.New("err2")
	tag.Try(err2)  // 这里 err2!=nil，则会跳转到 tag 创建处的下一行指令执行，即 fmt.Printf("2 -> ")

	fmt.Printf("6 -> ")
	return
}
```
测试结果：
```sh
# release 下 defer 内联，不会输出 4
2 -> 5 -> 2 -> 3 -> 1 ->
# debug 下 defer 不内联，会输出 4
2 -> 5 -> 2 -> 3 -> 4 -> 1
```


## 5.8 调用其他package的私有函数
通过过摆脱golang编译器的一些约束，调用其他package的私有函数。如这篇文章：[How to call private functions (bind to hidden symbols) in GoLang](https://sitano.github.io/2016/04/28/golang-private/)。

上面 goid 的例子的最后，也讲了通过汇编使用 package 私有的类型，即 DATA ·runtime_g_type+0(SB)/8,$type·runtime·g(SB) ，这里不在重复。

## 5.8 提高 CGO 调用的性能
我们知道，CGO和系统调用时，Go 语言需要把 goroutine 的调用栈切换回 g0 调用栈，并使用 g0 调用，整个过程性能损耗比较大。实际上，我们可以通过汇编适配 C 语言的ABI来直接调用 C 语言的函数，参考这个库: [petermattis/fastcgo](https://github.com/petermattis/fastcgo)。不过，这么做也有很大的局限性，比如导致栈溢出、因goroutine 无法被抢占而影响 GC 性能等。

<!-- 
rust 才可以参考: 
[rustgo](https://words.filippo.io/rustgo/) \
[FiloSottile/ed25519-dalek-rustgo](https://github.com/FiloSottile/ed25519-dalek-rustgo)
-->