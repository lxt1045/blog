线程本地存储


[Go语言goroutine调度器初始化 (12)](https://zhuanlan.zhihu.com/p/64672362)\

下面我们详细来详细看一下settls函数是如何实现线程私有全局变量的。

runtime/sys_linx_amd64.s : 606
```asm
    // set tls base to DI
    TEXT runtime·settls(SB),NOSPLIT,$32
    //......
    //DI寄存器中存放的是m.tls[0]的地址，m的tls成员是一个数组，读者如果忘记了可以回头看一下m结构体的定义
    //下面这一句代码把DI寄存器中的地址加8，为什么要+8呢，主要跟ELF可执行文件格式中的TLS实现的机制有关
    //执行下面这句指令之后DI寄存器中的存放的就是m.tls[1]的地址了
    ADDQ $8, DI    // ELF wants to use -8(FS)

    //下面通过arch_prctl系统调用设置FS段基址
    MOVQ DI, SI                 //SI存放arch_prctl系统调用的第二个参数
    MOVQ $0x1002, DI            // ARCH_SET_FS //arch_prctl的第一个参数
    MOVQ $SYS_arch_prctl, AX    //系统调用编号
    SYS CALL
    CMP QAX, $0xfffffffffffff001
    JLS 2(PC)
    MOVL $0xf1, 0xf1            // crash //系统调用失败直接crash
    RET
```
从代码可以看到，这里通过arch_prctl系统调用把m0.tls[1]的地址设置成了fs段的段基址。CPU中有个叫fs的段寄存器与之对应，而每个线程都有自己的一组CPU寄存器值，操作系统在把线程调离CPU运行时会帮我们把所有寄存器中的值保存在内存中，调度线程起来运行时又会从内存中把这些寄存器的值恢复到CPU，这样，在此之后，工作线程代码就可以通过fs寄存器来找到m.tls，读者可以参考上面初始化tls之后对tls功能验证的代码来理解这一过程。


runtime·setg 函数
/usr/local/go/src/runtime/asm_amd64.s
```asm
// func setg(gg *g)
// set g. for use by needm.
TEXT runtime·setg(SB), NOSPLIT, $0-8
	MOVQ	gg+0(FP), BX
	get_tls(CX)
	MOVQ	BX, g(CX)
	RET

```

/usr/local/go/src/runtime/asm_386.s
```asm
    // void setg(G*); set g. for use by needm.
    TEXT runtime·setg(SB), NOSPLIT, $0-4
        MOVL	gg+0(FP), BX
    #ifdef GOOS_windows
        CMPL	BX, $0
        JNE	settls
        MOVL	$0, 0x14(FS)
        RET
    settls:
        MOVL	g_m(BX), AX
        LEAL	m_tls(AX), AX
        MOVL	AX, 0x14(FS)
    #endif
        get_tls(CX)
        MOVL	BX, g(CX)
        RET

```
/usr/local/go/src/runtime/go_tls.h
```c
#ifdef GOARCH_arm
#define LR R14
#endif

#ifdef GOARCH_amd64
#define	get_tls(r)	MOVQ TLS, r
#define	g(r)	0(r)(TLS*1)
#endif

#ifdef GOARCH_386
#define	get_tls(r)	MOVL TLS, r
#define	g(r)	0(r)(TLS*1)
#endif
```


[go语言调度器源代码情景分析之二：CPU寄存器](https://www.cnblogs.com/abozhang/p/10766689.html)
```md
不同体系结构的CPU，其内部寄存器的数量、种类以及名称可能大不相同，这里我们只介绍目前使用最为广泛的AMD64这种体系结构的CPU，这种CPU共有20多个可以直接在汇编代码中使用的寄存器，其中有几个寄存器在操作系统代码中才会见到，而应用层代码一般只会用到如下分为三类的19个寄存器。

通用寄存器：rax, rbx, rcx, rdx, rsi, rdi, rbp, rsp, r8, r9, r10, r11, r12, r13, r14, r15寄存器。CPU对这16个通用寄存器的用途没有做特殊规定，程序员和编译器可以自定义其用途（下面会介绍，rsp/rbp寄存器其实是有特殊用途的）；

程序计数寄存器（PC寄存器，有时也叫IP寄存器）：rip寄存器。它用来存放下一条即将执行的指令的地址，这个寄存器决定了程序的执行流程；

段寄存器：fs和gs寄存器。一般用它来实现线程本地存储（TLS），比如AMD64 linux平台下go语言和pthread都使用fs寄存器来实现系统线程的TLS，在本章线程本地存储一节和第二章详细分析goroutine调度器的时候我们可以分别看到Linux平台下Pthread线程库和go是如何使用fs寄存器的。

上述这些寄存器除了fs和gs段寄存器是16位的，其它都是64位的，也就是8个字节，其中的16个通用寄存器还可以作为32/16/8位寄存器使用，只是使用时需要换一个名字，比如可以用eax这个名字来表示一个32位的寄存器，它使用的是rax寄存器的低32位。

```

[线程本地存储及实现原理](https://www.cnblogs.com/abozhang/p/10800332.html)

线程本地存储又叫线程局部存储，其英文为Thread Local Storage，简称TLS，看似一个很高大上的东西，其实就是线程私有的全局变量而已。

有过多线程编程的读者一定知道，普通的全局变量在多线程中是共享的，一个线程对其进行了修改，所有线程都可以看到这个修改，而线程私有的全局变量与普通全局变量不同，线程私有全局变量是线程的私有财产，每个线程都有自己的一份副本，某个线程对其所做的修改只会修改到自己的副本，并不会修改到其它线程的副本。

[浅谈FS段寄存器在用户层和内核层的使用](https://cloud.tencent.com/developer/article/1471370)\
[汇编之FS段寄存器](https://www.cnblogs.com/milantgh/p/3878771.html) \
greate: [FS寄存器 和 段寄存器线索](https://tokameine.top/2022/01/31/fs-register/)\
greate: [线程本地存储及实现原理](https://www.cnblogs.com/abozhang/p/10800332.html) \
greate： [Go语言goroutine调度器概述(11)](https://www.cnblogs.com/abozhang/p/10802319.html)\

如果只有一个工作线程，那么就只会有一个m结构体对象，问题就很简单，定义一个全局的m结构体变量就行了。可是我们有多个工作线程和多个m需要一一对应，怎么办呢？还记得第一章我们讨论过的线程本地存储吗？当时我们说过，线程本地存储其实就是线程私有的全局变量，这不正是我们所需要的吗？！只要每个工作线程拥有了各自私有的m结构体全局变量，我们就能在不同的工作线程中使用相同的全局变量名来访问不同的m结构体对象，这完美的解决我们的问题。

具体到goroutine调度器代码，每个工作线程在刚刚被创建出来进入调度循环之前就利用线程本地存储机制为该工作线程实现了一个指向m结构体实例对象的私有全局变量，这样在之后的代码中就使用该全局变量来访问自己的m结构体对象以及与m相关联的p和g对象。

有了上述数据结构以及工作线程与数据结构之间的映射机制，我们可以把前面的调度伪代码写得更丰满一点：


```go
// 程序启动时的初始化代码
......
    for i = 0; i < N; i++ { // 创建N个操作系统线程执行schedule函数
        create_os_thread(schedule) // 创建一个操作系统线程执行schedule函数
    }


    // 定义一个线程私有全局变量，注意它是一个指向m结构体对象的指针
    // ThreadLocal用来定义线程私有全局变量
    ThreadLocal self *m
    //schedule函数实现调度逻辑
    schedule() {
        // 创建和初始化m结构体对象，并赋值给私有全局变量self
        self = initm()   
        for { //调度循环
            if(self.p.runqueue is empty) {
                    // 根据某种算法从全局运行队列中找出一个需要运行的goroutine
                    g = find_a_runnable_goroutine_from_global_runqueue()
            } else {
                    // 根据某种算法从私有的局部运行队列中找出一个需要运行的goroutine
                    g = find_a_runnable_goroutine_from_local_runqueue()
            }
            run_g(g) // CPU运行该goroutine，直到需要调度其它goroutine才返回
            save_status_of_g(g) // 保存goroutine的状态，主要是寄存器的值
        }
    } 

```
仅仅从上面这个伪代码来看，我们完全不需要线程私有全局变量，只需在schedule函数中定义一个局部变量就行了。但真实的调度代码错综复杂，不光是这个schedule函数会需要访问m，其它很多地方还需要访问它，所以需要使用全局变量来方便其它地方对m的以及与m相关的g和p的访问。

 

在简单的介绍了Go语言调度器以及它所需要的数据结构之后，下面我们来看一下Go的调度代码中对上述的几个结构体的定义。

----------------
[](https://www.golang-tech-stack.com/post/3168)

主线程与m0绑定

设置好g0栈之后，我们跳过CPU型号检查以及cgo初始化相关的代码，直接从164行继续分析。

greate:
```
    //getg() 函数在源代码中没有对应的定义，由编译器插入类似下面两行代码
    //get_tls(CX) 
    //MOVQ g(CX), BX; BX存器里面现在放的是当前g结构体对象的地址
```
runtime/asm_amd64.s : 164
```asm
    //下面开始初始化tls(thread local storage,线程本地存储)
    LEAQ runtime·m0+m_tls(SB), DI //DI=&m0.tls，取m0的tls成员的地址到DI寄存器
    CALL runtime·settls(SB) //调用settls设置线程本地存储，settls函数的参数在DI寄存器中

    // store through it, to make sure it works
    //验证settls是否可以正常工作，如果有问题则abort退出程序
    get_tls(BX) //获取fs段基地址并放入BX寄存器，其实就是m0.tls[0]的地址，get_tls的代码由编译器生成
    MOVQ $0x123, g(BX) //把整型常量0x123拷贝到fs段基地址偏移-8的内存位置，也就是m0.tls[0] =0x123
    MOVQ runtime·m0+m_tls(SB), AX//AX=m0.tls[0]
    CMPQ AX, $0x123 //检查m0.tls[0]的值是否是通过线程本地存储存入的0x123来验证tls功能是否正常
    JEQ 2(PC)
    CALL runtime·abort(SB) //如果线程本地存储不能正常工作，退出程序
```
这段代码首先调用settls函数初始化主线程的线程本地存储(TLS)，目的是把m0与主线程关联在一起，至于为什么要把m和工作线程绑定在一起，我们已经在上一节介绍过了，这里就不再重复。设置了线程本地存储之后接下来的几条指令在于验证TLS功能是否正常，如果不正常则直接abort退出程序。

下面我们详细来详细看一下settls函数是如何实现线程私有全局变量的。

runtime/sys_linx_amd64.s : 606
```asm
// set tls base to DI
TEXT runtime·settls(SB),NOSPLIT,$32
    //......
    //DI寄存器中存放的是m.tls[0]的地址，m的tls成员是一个数组，读者如果忘记了可以回头看一下m结构体的定义
    //下面这一句代码把DI寄存器中的地址加8，为什么要+8呢，主要跟ELF可执行文件格式中的TLS实现的机制有关
    //执行下面这句指令之后DI寄存器中的存放的就是m.tls[1]的地址了
    ADDQ $8, DI// ELF wants to use -8(FS)

    //下面通过arch_prctl系统调用设置FS段基址
    MOVQ DI, SI//SI存放arch_prctl系统调用的第二个参数
    MOVQ $0x1002, DI// ARCH_SET_FS //arch_prctl的第一个参数
    MOVQ $SYS_arch_prctl, AX//系统调用编号
    SYSCALL
    CMPQ AX, $0xfffffffffffff001
    JLS 2(PC)
    MOVL $0xf1, 0xf1 // crash //系统调用失败直接crash
    RET
```
从代码可以看到，这里通过arch_prctl系统调用把m0.tls[1]的地址设置成了fs段的段基址。CPU中有个叫fs的段寄存器与之对应，而每个线程都有自己的一组CPU寄存器值，操作系统在把线程调离CPU运行时会帮我们把所有寄存器中的值保存在内存中，调度线程起来运行时又会从内存中把这些寄存器的值恢复到CPU，这样，在此之后，工作线程代码就可以通过fs寄存器来找到m.tls，读者可以参考上面初始化tls之后对tls功能验证的代码来理解这一过程。

下面继续分析rt0_go，

runtime/asm_amd64.s : 174
```asm
    ok:
    // set the per-goroutine and per-mach "registers"
    get_tls(BX) //获取fs段基址到BX寄存器
    LEAQ runtime·g0(SB), CX//CX=g0的地址
    MOVQ CX, g(BX) //把g0的地址保存在线程本地存储里面，也就是m0.tls[0]=&g0
    LEAQ runtime·m0(SB), AX//AX=m0的地址

    //把m0和g0关联起来m0->g0 =g0，g0->m =m0
    // save m->g0 =g0
    MOVQ CX, m_g0(AX) //m0.g0 =g0
    // save m0 to g0->m 
    MOVQ AX, g_m(CX) //g0.m =m0
```
上面的代码首先把g0的地址放入主线程的线程本地存储中，然后通过
```go
m0.g0 = &g0
g0.m = &m0
```
把m0和g0绑定在一起，这样，之后在主线程中通过get_tls可以获取到g0，通过g0的m成员又可以找到m0，于是这里就实现了m0和g0与主线程之间的关联。从这里还可以看到，保存在主线程本地存储中的值是g0的地址，也就是说工作线程的私有全局变量其实是一个指向g的指针而不是指向m的指针，目前这个指针指向g0，表示代码正运行在g0栈。

-------------
几个关键数据结构
[golang调度器笔记](http://xlim.cn/post/read/golang/golang%E8%B0%83%E5%BA%A6%E5%99%A8%E7%AC%94%E8%AE%B0/)
```go
type gobuf struct {
    // The offsets of sp, pc, and g are known to (hard-coded in) libmach.
    //
    // ctxt is unusual with respect to GC: it may be a
    // heap-allocated funcval, so GC needs to track it, but it
    // needs to be set and cleared from assembly, where it's
    // difficult to have write barriers. However, ctxt is really a
    // saved, live register, and we only ever exchange it between
    // the real register and the gobuf. Hence, we treat it as a
    // root during stack scanning, which means assembly that saves
    // and restores it doesn't need write barriers. It's still
    // typed as a pointer so that any other writes from Go get
    // write barriers.
    sp   uintptr  // 保存CPU的rsp寄存器的值
    pc   uintptr  // 保存CPU的rip寄存器的值
    g    guintptr // 记录当前这个gobuf对象属于哪个goroutine
    ctxt unsafe.Pointer
 
    // 保存系统调用的返回值，因为从系统调用返回之后如果p被其它工作线程抢占，
    // 则这个goroutine会被放入全局运行队列被其它工作线程调度，其它线程需要知道系统调用的返回值。
    ret  sys.Uintreg  
    lr   uintptr
 
    // 保存CPU的rbp寄存器的值
    bp   uintptr // for GOEXPERIMENT=framepointer
}
// 前文所说的g结构体，它代表了一个goroutine
type g struct {
    // Stack parameters.
    // stack describes the actual stack memory: [stack.lo, stack.hi).
    // stackguard0 is the stack pointer compared in the Go stack growth prologue.
    // It is stack.lo+StackGuard normally, but can be StackPreempt to trigger a preemption.
    // stackguard1 is the stack pointer compared in the C stack growth prologue.
    // It is stack.lo+StackGuard on g0 and gsignal stacks.
    // It is ~0 on other goroutine stacks, to trigger a call to morestackc (and crash).
 
    // 记录该goroutine使用的栈
    stack       stack   // offset known to runtime/cgo
    // 下面两个成员用于栈溢出检查，实现栈的自动伸缩，抢占调度也会用到stackguard0
    stackguard0 uintptr // offset known to liblink
    stackguard1 uintptr // offset known to liblink

    ......
 
    // 此goroutine正在被哪个工作线程执行
    m              *m      // current m; offset known to arm liblink
    // 保存调度信息，主要是几个寄存器的值
    sched          gobuf
 
    ......
    // schedlink字段指向全局运行队列中的下一个g，
    //所有位于全局运行队列中的g形成一个链表
    schedlink      guintptr

    ......
    // 抢占调度标志，如果需要抢占调度，设置preempt为true
    preempt        bool       // preemption signal, duplicates stackguard0 = stackpreempt

   ......
}
type m struct {
    // g0主要用来记录工作线程使用的栈信息，在执行调度代码时需要使用这个栈
    // 执行用户goroutine代码时，使用用户goroutine自己的栈，调度时会发生栈的切换
    g0      *g     // goroutine with scheduling stack

    // 通过TLS实现m结构体对象与工作线程之间的绑定
    tls           [6]uintptr   // thread-local storage (for x86 extern register)
    mstartfn      func()
    // 指向工作线程正在运行的goroutine的g结构体对象
    curg          *g       // current running goroutine
 
    // 记录与当前工作线程绑定的p结构体对象
    p             puintptr // attached p for executing go code (nil if not executing go code)
    nextp         puintptr
    oldp          puintptr // the p that was attached before executing a syscall
   
    // spinning状态：表示当前工作线程正在试图从其它工作线程的本地运行队列偷取goroutine
    spinning      bool // m is out of work and is actively looking for work
    blocked       bool // m is blocked on a note
   
    // 没有goroutine需要运行时，工作线程睡眠在这个park成员上，
    // 其它线程通过这个park唤醒该工作线程
    park          note
    // 记录所有工作线程的一个链表
    alllink       *m // on allm
    schedlink     muintptr

    // Linux平台thread的值就是操作系统线程ID
    thread        uintptr // thread handle
    freelink      *m      // on sched.freem

    ......
}
type p struct {
    lock mutex

    status       uint32 // one of pidle/prunning/...
    link            puintptr
    schedtick   uint32     // incremented on every scheduler call
    syscalltick  uint32     // incremented on every system call
    sysmontick  sysmontick // last tick observed by sysmon
    m                muintptr   // back-link to associated m (nil if idle)

    ......

    // Queue of runnable goroutines. Accessed without lock.
    //本地goroutine运行队列
    runqhead uint32  // 队列头
    runqtail uint32     // 队列尾
    runq     [256]guintptr  //使用数组实现的循环队列
    // runnext, if non-nil, is a runnable G that was ready'd by
    // the current G and should be run next instead of what's in
    // runq if there's time remaining in the running G's time
    // slice. It will inherit the time left in the current time
    // slice. If a set of goroutines is locked in a
    // communicate-and-wait pattern, this schedules that set as a
    // unit and eliminates the (potentially large) scheduling
    // latency that otherwise arises from adding the ready'd
    // goroutines to the end of the run queue.
    runnext guintptr

    // Available G's (status == Gdead)
    gFree struct {
        gList
        n int32
    }

    ......
}

```


---------------
[Go语言内幕（5）：运行时启动过程](https://studygolang.com/articles/7211)\
[Golang内部构件，第5部分：运行时引导程序](https://segmentfault.com/a/1190000039745029)

运行下面的命令：
```
objdump -f 6.out
```
你可以看到包含开始地址的输出信息：
```
6.out:     file format elf64-x86-64
architecture: i386:x86-64, flags 0x00000112:
EXEC_P, HAS_SYMS, D_PAGED
start address 0x000000000042f160
```
接下来，我们要反汇编可执行程序，再找到在开始位置处到底是什么函数：
```
objdump -d 6.out > disassemble.txt
```
现在，我们可以打开 disassemble.txt 文件并搜索 “42f160”，可以得到如下结果：
```
000000000042f160 <_rt0_amd64_linux>:
  42f160:	48 8d 74 24 08       		lea    0x8(%rsp),%rsi
  42f165:	48 8b 3c 24          		mov    (%rsp),%rdi
  42f169:	48 8d 05 10 00 00 00 	lea    0x10(%rip),%rax        # 42f180 

  42f170:	ff e0               		 	jmpq   *%rax
  ```
很好，我们找到它了。在我的这台电脑上（与 OS 以及机器的架构有关）入口点的函数为 _rt0_amd64_linux。

TLS 内部实现\
如果你仔细阅读过前面的代码，很容易就会发现只有几行是真正起作用的代码：
```asm
LEAQ	runtime·tls0(SB), DI
CALL	runtime·settls(SB)
````
所有其它的代码都是在你的系统不支持 TLS 时跳过 TLS 设置或者检测 TLS 是否正常工作的代码。这两行代码将 runtime.tlso 变量的地址存储到 DI 寄存器中，然后调用 runtime.settls 函数。这个函数的代码如下：
```asm
// set tls base to DI
TEXT runtime·settls(SB),NOSPLIT,$32
	ADDQ	$8, DI	// ELF wants to use -8(FS)

	MOVQ	DI, SI
	MOVQ	$0x1002, DI	// ARCH_SET_FS
	MOVQ	$158, AX	// arch_prctl
	SYSCALL
	CMPQ	AX, $0xfffffffffffff001
	JLS	2(PC)
	MOVL	$0xf1, 0xf1  // crash
	RET
```
从注释可以看出，这个函数执行了 arch_prctl 系统调用，并将 ARCH_SET_FS 作为参数传入。我们也可以看到，系统调用使用 FS 寄存器存储基址。在这个例子中，我们将 TLS 指向 runtime.tls[0] 变量。

还记得 main 开始时的汇编指令吗？
```asm
0x0000 00000 (test.go:3)	MOVQ	(TLS),CX
```
在前面我已经解释了这条指令将 runtime.g 结构体实例的地址加载到 CX 寄存器中。这个结构体描述了当前 goroutine，且存储到 TLS 中。现在我们明白了这条指令是如何被汇编成机器指令的了。打开之前是创建的 disasembly.txt 文件，搜索 main.main 函数，你会看到其中第一条指令为：
```asm
400c00:       64 48 8b 0c 25 f0 ff    mov    %fs:0xfffffffffffffff0,%rcx
```
这条指令中的冒号（%fs:0xfffffffffffffff0）表示段寻址（更多内容请参考[这里](https://thestarman.pcministry.com/asm/debug/Segments.html)）。

[](https://taoshu.in/go/monkey/monkey-2.html)

[Monkey Patching in Go](https://bou.ke/blog/monkey-patching-in-go/)\
[翻译](https://berryjam.github.io/2018/12/golang%E6%9B%BF%E6%8D%A2%E8%BF%90%E8%A1%8C%E6%97%B6%E5%87%BD%E6%95%B0%E4%BD%93%E5%8F%8A%E5%85%B6%E5%8E%9F%E7%90%86/)

