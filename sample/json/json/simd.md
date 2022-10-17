1. 来看看movups指令，这条指令名称上分为四个部分：

mov，表示数据移动，操作双方可以是内存也可以是寄存器。
u，表示 unaligned，内存未对齐。如果是a，表示 aligned，内存已对齐。
p，表示 packed，打包数据，会对128位所有数据执行操作。如果是s，则表示 scalar，标量数据，仅对128位内第一个数执行操作。

```log
    Packed                        Scalar
    A3    A2    A1    A0        A3    A2    A1    A0
    +     +     +     +         +     +     +     +
    B3    B2    B1    B0        B3    B2    B1    B0
    |     |     |     |         |     |     |     |
    v     v     v     v         v     v     v     v
  A3+B3 A2+B2 A1+B1 A0+B0       A3    A2    A1  A0+B0
```
s，表示 single precision floating point，将数据视为32位单精度浮点数，一组4个。如果是d，表示 double precision floating point，将数据视为64位双精度浮点，一组两个。

2. 从内存中向寄存器加载数据时，必须区分数据的对齐与否。
SSE指令要求数据按16字节对齐，未对齐数据必须使用movups，已对齐数据可以任意使用movups或者movaps。对齐的数据需要按照下面这样进行声明：
```c++
    // C++ 11 alignas
    alignas(16) float a[4] = { 1,2,3,4 };
```
对非对齐的数据使用movaps，会导致程序崩溃。理论上movups相比movaps性能会差一些，但在较新的CPU上性能差异已经基本可以忽视。(cacheline)

3. 下面是SSE指令的一般格式，由三部分组成，第一部分是表示指令的作用，比如加法add等，第二部分是s或者p分别表示scalar或packed，第三部分为s，表示单精度浮点数（single precision floating point data）。


```log
add/set/load...     s/p          s
     |               |           |   
     v               v           v 
  指令作用       scalar/packer  单精度浮点数
                execution
                  mode
```

4. SSE定址/寻址方式：
SSE 指令和一般的x86 指令很类似，基本上包括两种定址方式：寄存器-寄存器方式(reg-reg)和寄存器-内存方式(reg-mem)：
```asm
addps xmm0, xmm1 ; reg-reg
addps xmm0, [ebx] ; reg-mem
```

5. intrinsics的SSE指令

要使用SSE指令，可以使用intrinsics来简化编程，前面已经介绍过intrinsics的基础了，这里也不会展开。

SSE指令的intrinsics函数名称一般为：_m_operation[u/r...]_ss/ps，和上面的SSE指令的命名类似，只是增加了_m_前缀，另外，表示指令作用的操作后面可能会有一个可选的修饰符，表示一些特殊的作用，比如从内存加载，可能是反过来的顺序加载（不知道汇编指令有没有对应的修饰符，理论上应该没有，这个修饰符只是给编译器用于进行一些转换用的，具体待查）。

SSE指令中的intrinsics函数的数据类型为：__m128，正好对应 了上面提到的SSE新的数据类型，当然，这种数据类型只是一种抽象表示，实际是要转换为基本的数据类型的。

6. 字宽： moveb、movew、moved、moveq
b：byte，字节，8位
w：word，双字节，字，16位
d：doubleword，双字，32位
q：quadwords，四字，64位

packed byte: 以 byte 为单位，一共 16 个 byte
packed word: 以 word 为单位，一共 8 个 word
packed doubleword: 以 doubleword 为单位，一共 4 个 doubleword
packed quadword: 以 quadword 为单位，一共 2 个 quadword

6.1 普通紫菱不以 p 开头： 功能+位数
6.2 以p开头的指令： p+功能+其他(p/s)+位数

7. 操作码汇总表中的指令列
mm — An MMX register. The 64-bit MMX registers are: MM0 through MM7.
xmm — An XMM register. The 128-bit XMM registers are: XMM0 through XMM7; XMM8 through XMM15 are
available using REX.R in 64-bit mode.
ymm — A YMM register. The 256-bit YMM registers are: YMM0 through YMM7; YMM8 through YMM15 are
available in 64-bit mode.

7. Intel Intrinsics 文档
https://www.intel.com/content/www/us/en/docs/intrinsics-guide/index.html


8. 
punpcklbw
punpcklbw XMM,XMM/m128
把源存储器与目的寄存器低64位按字节交错排列,内存变量必须对齐内存16字节.
高64位 | 低64位
目的寄存器:         a0|a1| a2| a3| a4|a5| a6|a7| a8|a9| aA|aB| aC|aD| aE| aF
源寄存器:           b0|b1| b2| b3| b4|b5| b6|b7| b8|b9| bA|bB| bC|bD| bE| bF
目的寄存器排列结果:   b8|a8| b9| a9| bA|aA| bB|aB| bC|aC| bD|aD| bE|aE| bF| aF

9. SSE2指令集
http://www.yibei.com/book/4df5ae4d7e021e33400728e6




-------------------------------------------------

SSE/AVX Intrinsics简介
1.头文件
SSE/AVX指令主要定义于以下一些头文件中：
<xmmintrin.h> : SSE, 支持同时对4个32位单精度浮点数的操作。
<emmintrin.h> : SSE 2, 支持同时对2个64位双精度浮点数的操作。
<pmmintrin.h> : SSE 3, 支持对SIMD寄存器的水平操作(horizontal operation)，如hadd, hsub等…。
<tmmintrin.h> : SSSE 3, 增加了额外的instructions。
<smmintrin.h> : SSE 4.1, 支持点乘以及更多的整形操作。
<nmmintrin.h> : SSE 4.2, 增加了额外的instructions。（这个支持之前所有版本的SEE）
<immintrin.h> : AVX, 支持同时操作8个单精度浮点数或4个双精度浮点数。

2.命名规则（很重要）
SSE/AVX提供的数据类型和函数的命名规则如下：
a.数据类型通常以_mxxx(T)的方式进行命名，其中xxx代表数据的位数，如SSE提供的__m128为128位，AVX提供的__m256为256位。T为类型，若为单精度浮点型则省略，若为整形则为i，如__m128i，若为双精度浮点型则为d，如__m256d。
b.操作浮点数的内置函数命名方式为：_mm(xxx)_name_PT。 xxx为SIMD寄存器的位数，若为128m则省略，如_mm_addsub_ps，若为_256m则为256，如_mm256_add_ps。 name为函数执行的操作的名字，如加法为_mm_add_ps，减法为_mm_sub_ps。 P代表的是对矢量(packed data vector)还是对标量(scalar)进行操作，如_mm_add_ss是只对最低位的32位浮点数执行加法，而_mm_add_ps则是对4个32位浮点数执行加法操作。 T代表浮点数的类型，若为s则为单精度浮点型，若为d则为双精度浮点，如_mm_add_pd和_mm_add_ps。
c.操作整形的内置函数命名方式为：_mm(xxx)_name_epUY。xxx为SIMD寄存器的位数，若为128位则省略。 name为函数的名字。U为整数的类型，若为无符号类型则为u，否在为i，如_mm_adds_epu16和_mm_adds_epi16。Y为操作的数据类型的位数，如_mm_cvtpd_pi32。

3.内置函数（instructions）
1).存取操作(load/store/set)
```c++
    __attribute__((aligned(32))) int d1[8] = {-1,-2,-3,-4,-5,-6,-7,-8};
    __m256i d = _mm256_load_si256((__m256i*)d1);//装在int可以使用指针类型转换 必须32位对齐
```
这里说明一下，使用load函数要保证数组的起始地址32位字节对齐。在linux下就需要__attribute__((aligned(32)))，Windows下要用__declspec(align(32))
这里有没有疑问，为什么要字节对齐呢？
现代计算机中内存空间都是按照byte划分的，从理论上讲似乎对任何类型的变量的访问可以从任何地址开始，但实际情况是在访问特定类型变量的时候经常在特定的内存地址访问，这就需要各种类型数据按照一定的规则在空间上排列，而不是顺序的一个接一个的排放，这就是对齐。
对齐的作用和原因：各个硬件平台对存储空间的处理上有很大的不同。一些平台对某些特定类型的数据只能从某些特定地址开始存取。比如有些架构的CPU在访问一个没有进行对齐的变量的时候会发生错误,那么在这种架构下编程必须保证字节对齐.其他平台可能没有这种情况，但是最常见的是如果不按照适合其平台要求对数据存放进行对齐，会在存取效率上带来损失。比如有些平台每次读都是从偶地址开始，如果一个int型（假设为32位系统）如果存放在偶地址开始的地方，那么一个读周期就可以读出这32bit，而如果存放在奇地址开始的地方，就需要2个读周期，并对两次读出的结果的高低字节进行拼凑才能得到该32bit数据。显然在读取效率上下降很多。

上面是从手册查询到的load系列的函数。其中，
_mm_load_ss用于scalar的加载，所以，加载一个单精度浮点数到暂存器的低字节，其它三个字节清0，（r0 := *p, r1 := r2 := r3 := 0.0）。
_mm_load_ps用于packed的加载（下面的都是用于packed的），要求p的地址是16字节对齐，否则读取的结果会出错，（r0 := p[0], r1 := p[1], r2 := p[2], r3 := p[3]）。
_mm_load1_ps表示将p地址的值，加载到暂存器的四个字节，需要多条指令完成，所以，从性能考虑，在内层循环不要使用这类指令。（r0 := r1 := r2 := r3 := *p）。
_mm_loadh_pi和_mm_loadl_pi分别用于从两个参数高底字节等组合加载。具体参考手册。
_mm_loadr_ps表示以_mm_load_ps反向的顺序加载，需要多条指令完成，当然，也要求地址是16字节对齐。（r0 := p[3], r1 := p[2], r2 := p[1], r3 := p[0]）。
_mm_loadu_ps和_mm_load_ps一样的加载，但是不要求地址是16字节对齐，对应指令为movups。

store系列可以将SSE/AVX提供的类型中的数据存储到内存中，如：

```c++
void test() 
{
	__declspec(align(16)) float p[] = { 1.0f, 2.0f, 3.0f, 4.0f };
	__m128 v = _mm_load_ps(p);

	__declspec(align(16)) float a[] = { 1.0f, 2.0f, 3.0f, 4.0f };
	_mm_store_ps(a, v);
}
```

_mm_store_ps可以__m128中的数据存储到16字节对齐的内存。
_mm_storeu_ps不要求存储的内存对齐。
_mm_store_ps1则是把__m128中最低位的浮点数存储为4个相同的连续的浮点数，即：p[0] = m[0], p[1] = m[0], p[2] = m[0], p[3] = m[0]。
_mm_store_ss是存储__m128中最低位的位浮点数到内存中。
_mm_storer_ps是按相反顺序存储__m128中的4个浮点数。

set系列可以直接设置SSE/AVX提供的类型中的数据，如：

```c++
__m128 v = _mm_set_ps(0.5f, 0.2f, 0.3f, 0.4f);
```
_mm_set_ps可以将4个32位浮点数按相反顺序赋值给__m128中的4个浮点数，即：_mm_set_ps(a, b, c, d) : m[0] = d, m[1] = c, m[2] = b, m[3] = a。
_mm_set_ps1则是将一个浮点数赋值给__m128中的四个浮点数。
_mm_set_ss是将给定的浮点数设置到__m128中的最低位浮点数中，并将高三位的浮点数设置为0.
_mm_setzero_ps是将__m128中的四个浮点数全部设置为0.

2). 算术运算
SSE/AVX提供的算术运算操作包括：
_mm_add_ps，_mm_add_ss 等加法系列
_mm_sub_ps，_mm_sub_pd 等减法系列
_mm_mul_ps，_mm_mul_epi32 等乘法系列
_mm_div_ps，_mm_div_ss 等除法系列
_mm_sqrt_pd，_mm_rsqrt_ps 等开平方系列
_mm_rcp_ps，_mm_rcp_ss 等求倒数系列
_mm_dp_pd，_mm_dp_ps 计算点乘
此外还有向下取整，向上取整等运算，这里只列出了浮点数支持的算术运算类型，还有一些整形的算术运算类型未列出。

3).比较运算
SSE/AVX提供的比较运算操作包括：
_mm_max_ps逐分量对比两个数据，并将较大的分量存储到返回类型的对应位置中。
_mm_min_ps逐分量对比两个数据，并将较小的分量存储到返回类型的对应位置中。
_mm_cmpeq_ps逐分量对比两个数据是否相等。
_mm_cmpge_ps逐分量对比一个数据是否大于等于另一个是否相等。
_mm_cmpgt_ps逐分量对比一个数据是否大于另一个是否相等。
_mm_cmple_ps逐分量对比一个数据是否小于等于另一个是否相等。
_mm_cmplt_ps逐分量对比一个数据是否小于另一个是否相等。
_mm_cmpneq_ps逐分量对比一个数据是否不等于另一个是否相等。
_mm_cmpnge_ps逐分量对比一个数据是否不大于等于另一个是否相等。
_mm_cmpngt_ps逐分量对比一个数据是否不大于另一个是否相等。
_mm_cmpnle_ps逐分量对比一个数据是否不小于等于另一个是否相等。
_mm_cmpnlt_ps逐分量对比一个数据是否不小于另一个是否相等。

4).逻辑运算
SSE/AVX提供的逻辑运算操作包括：
_mm_and_pd对两个数据逐分量and
_mm_andnot_ps先对第一个数进行not，然后再对两个数据进行逐分量and
_mm_or_pd对两个数据逐分量or
_mm_xor_ps对两个数据逐分量xor

详情可查Intel的Intrinsics Guide


