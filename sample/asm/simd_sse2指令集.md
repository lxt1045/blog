# SSE2指令集
http://www.yibei.com/book/4df5ae4d7e021e33400728e6

1. MMX/SSE类扩展引入了SIMD（单指令多数据）的执行模式，可用于加速多媒体应用。 下面简要介绍一下这些指令的执行环境和特征。\
  8个32位通用寄存器可为各个SIMD扩展所使用；\
  MMX：8个64位MMX寄存器（mm0 - mm7），也可为各SSE扩展所使用；\
      数据为整数，最多支持两个32位\
      运算中没有寄存器能够进行溢出指示\
  SSE：8个128位xmm寄存器，MXSCR寄存器，EFLAGS寄存器\
      支持单精度浮点\
      MXSCR含有rounding, overflow标志\
      支持64位SIMD整数\
  SSE2：执行环境同sse\
      双精度浮点\
      128位整数\
      双—单精度转换\
  SSE3：与Inte Prescott处理器一同发布不久，共13条指令\
      主要增强了视频解码、3D图形优化和超线程性能

2. MMX技术出现最早，目前几乎所有的X86处理器都提供支持，包括嵌入式X86， 所以下面的讨论主要基于MMX，但方法完全适用于SSEn， 包括像AMD的3D Now等其它SIMD扩展。\
MMX指令又分为以下几种：\
  数据传送：movd, movq\
  数据转换：packsswb, packssdw, packuswb, punpckhbw, punpckhwd, punpckhdq, punpcklbw, punpcklwd, punpckldq\
  并行算术：paddb, paddw, paddd, paddsb, paddsw, paddusb, paddusw, psubb, psubw, psubd, psubsb, psubsw, psubusb, psubusb, psubusw, pmulhw, pmullw, pmaddwd\
  并行比较：pcmpeqb, pcmpeqw, pcmpeqd, pcmpgtb, pcmpgtw, pcmpgtd\
  并行逻辑：pand, pandn, por, pxor\
  移位与旋转：psllw, pslld, psllq, psrlw, psrld, psrlq, psraw, psrad\
  状态管理：emms

3. 


1\
Movaps\
movaps XMM,XMM/m128\
movaps XMM/128,XMM\
把源存储器内容值送入目的寄存器,当有m128时,必须对齐内存16字节,也就是内存地址低4位为0.\

2\
Movups\
movups XMM,XMM/m128\
movaps XMM/128,XMM\
把源存储器内容值送入目的寄存器,但不必对齐内存16字节\

3\
Movlps\
movlps XMM,m64\
把源存储器64位内容送入目的寄存器低64位,高64位不变,内存变量不必对齐内存16字节\

4\
Movhps\
movhps XMM,m64\
把源存储器64位内容送入目的寄存器高64位,低64位不变,内存变量不必对齐内存16字节.\

5\
Movhlps\
movhlps XMM,XMM\
把源寄存器高64位送入目的寄存器低64位,高64位不变.\

6\
Movlhps\
movlhps XMM,XMM\
把源寄存器低64位送入目的寄存器高64位,低64位不变.\

7\
movss\
movss XMM,m32/XMM\
原操作数为m32时：dest[31-00] <== m32 dest[127-32] <== 0\
原操作数为XMM时: dest[31-00] <== src[31-00] dest[127-32]不变\

8\
movmskpd\
movmskpd r32,XMM\
取64位操作数符号位\
r32[0] <== XMM[63] r32[1] <== XMM[127] r32[31-2] <== 0\

9\
movmskps\
movmskps r32,XMM\
取32位操作数符号位\
r32[0] <== XMM[31] r32[1] <== XMM[63] r32[2] <== XMM[95] r32[3] <== XMM[127] r32[31-4] <== 0\

10\
pmovmskb\
pmovmskb r32,XMM\
取16位操作数符号位 具体操作同前


11 - 20, 共 150 个条目\
11\
movntps\
movntps m128,XMM\
m128 <== XMM 直接把XMM中的值送入m128，不经过cache,必须对齐16字节.\

12\
Movntpd\
movntpd m128,XMM\
m128 <== XMM 直接把XMM中的值送入m128，不经过cache,必须对齐16字节.\

13\
Movnti\
movnti m32,r32\
m32 <== r32 把32寄存器的值送入m32，不经过cache.\

14\
Movapd\
movapd XMM,XMM/m128\
movapd XMM/m128,XMM\
把源存储器内容值送入目的寄存器,当有m128时,必须对齐内存16字节\

15\
Movupd\
movupd XMM,XMM/m128\
movapd XMM/m128,XMM\
把源存储器内容值送入目的寄存器,但不必对齐内存16字节.\

16\
Movlpd\
movlpd XMM,m64\
movlpd m64,XMM\
把源存储器64位内容送入目的寄存器低64位,高64位不变,内存变量不必对齐内存16字节\

17\
Movhpd\
movhpd XMM,m64\
movhpd m64,XMM\
把源存储器64位内容送入目的寄存器高64位,低64位不变,内存变量不必对齐内存16字节.\

18\
Movdqa\
movdqa XMM,XMM/m128\
movdqa XMM/m128,XMM\
把源存储器内容值送入目的寄存器,当有m128时,必须对齐内存16字节.\

19\
Movdqu\
movdqu XMM,XMM/m128\
movdqu XMM/m128,XMM\
把源存储器内容值送入目的寄存器,但不必对齐内存16字节.\

20\
movq2dq\
movq2dq XMM,MM\
把源寄存器内容送入目的寄存器的低64位,高64位清零.\
 21 - 30, 共 150 个条目\
21\
movdq2q\
movdq2q MM,XMM\
把源寄存器低64位内容送入目的寄存器.\

22\
Movd\
movd XMM,r32/m32\
movd MM,r32/m32\
把源存储器32位内容送入目的寄存器的低32位,高96位清零.\

23\
Movd\
movd r32/m32,XMM\
movd r32/m32,MM\
把源寄存器的低32位内容送入目的存储器32位.\

24\
Movq\
movq XMM,XMM/m64\
movq MM,MM/m64\
把源存储器低64位内容送入目的寄存器的低64位,高64位清零.\

25\
Movq\
movq m64,XMM\
把源寄存器的低64位内容送入目的存储器.\

26\
addps\
addps XMM,XMM/m128\
源存储器内容按双字对齐,共4个单精度浮点数与目的寄存器相加,结果送入目的寄存器,内存变量必须对齐内存16字节\

27\
ADDS\
addss XMM,XMM/m32\
源存储器的低32位1个单精度浮点数与目的寄存器的低32位1个单精度浮点数相加,结果送入目的寄存器的低32位高96位不变,内存变量不必对齐内存16字节\

28\
addpd\
addpd XMM,XMM/m128\
源存储器内容按四字对齐,共两个双精度浮点数与目的寄存器相加,结果送入目的寄存器,内存变量必须对齐内存16字节.\

29\
addsd\
addsd XMM,XMM/m64\
源存储器的低64位1个双精度浮点数与目的寄存器的低64位1个双精度浮点数相加,结果送入目的寄存器的低64位, 高64位不变,内存变量不必对齐内存16字节\

30\
paddd\
paddd XMM,XMM/m128\
把源存储器与目的寄存器按双字对齐无符号整数普通相加,结果送入目的寄存器,内存变量必须对齐内存16字节.\
 \
31 - 40, 共 150 个条目\
31\
Paddq\
paddq XMM,XMM/m128\
把源存储器与目的寄存器按四字对齐无符号整数普通相加,结果送入目的寄存器,内存变量必须对齐内存16字节.\

32\
Paddq\
paddq MM,MM/m64\
把源存储器与目的寄存器四字无符号整数普通相加,结果送入目的寄存器.\

33\
Pmaddwd\
pmaddwd XMM,XMM/m128\
把源存储器与目的寄存器分4组进行向量点乘(有符号补码操作),内存变量必须对齐内存16字节.\

34\
Paddsb\
paddsb XMM,XMM/m128\
paddsb MM,MM/m64\
源存储器与目的寄存器按字节对齐有符号补码饱和相加,内存变量必须对齐内存16字节.\

35\
paddsw\
paddsw XMM,XMM/m128\
源存储器与目的寄存器按字对齐有符号补码饱和相加,内存变量必须对齐内存16字节.\

36\
paddusb\
paddusb XMM,XMM/m128\
源存储器与目的寄存器按字节对齐无符号饱和相加,内存变量必须对齐内存16字节.\

37\
Paddusw\
paddusw XMM,XMM/m128\
源存储器与目的寄存器按字对齐无符号饱和相加,内存变量必须对齐内存16字节.\

38\
Paddb\
paddb XMM,XMM/m128\
源存储器与目的寄存器按字节对齐无符号普通相加,内存变量必须对齐内存16字节.\

39\
Paddw\
paddw XMM,XMM/m128\
源存储器与目的寄存器按字对齐无符号普通相加,内存变量必须对齐内存16字节.\

40\
Paddd\
paddd XMM,XMM/m128\
源存储器与目的寄存器按双字对齐无符号普通相加,内存变量必须对齐内存16字节.\
 \
41 - 50, 共 150 个条目\
41\
Paddq\
paddq XMM,XMM/m128\
源存储器与目的寄存器按四字对齐无符号普通相加,内存变量必须对齐内存16字节.\

42\
subps\
subps XMM,XMM/m128\
源存储器内容按双字对齐,共4个单精度浮点数与目的寄存器相减(目的减去源),结果送入目的寄存器, 内存变量必须对齐内存16字节.\

43\
Subss\
subss XMM,XMM/m32\
源存储器的低32位1个单精度浮点数与目的寄存器的低32位1个单精度浮点数相减(目的减去源), 结果送入目的寄存器的低32位,高96位不变,内存变量不必对齐内存16字节\

44\
Subpd\
subpd XMM,XMM/m128\
把目的寄存器内容按四字对齐,两个双精度浮点数,减去源存储器两个双精度浮点数, 结果送入目的寄存器,内存变量必须对齐内存16字节.\

45\
subsd\
subsd XMM,XMM/m128\
把目的寄存器的低64位1个双精度浮点数,减去源存储器低64位1个双精度浮点数,结果送入目的寄存器的低64位, 高64位不变,内存变量不必对齐内存16字节\

46\
Psubd\
psubd XMM,XMM/m128\
把目的寄存器与源存储器按双字对齐无符号整数普通相减,结果送入目的寄存器, 内存变量必须对齐内存16字节.(目的减去源)\

47\
Psubq\
psubq XMM,XMM/m128\
把目的寄存器与源存储器按四字对齐无符号整数普通相减,结果送入目的寄存器, 内存变量必须对齐内存16字节.(目的减去源)\

48\
Psubq\
psubq MM,MM/m64\
把目的寄存器与源存储器四字无符号整数普通相减,结果送入目的寄存器.(目的减去源)\

49\
psubsb\
psubsb XMM,XMM/m128\
源存储器与目的寄存器按字节对齐有符号补码饱和相减(目的减去源),内存变量必须对齐内存16字节.\

50\
Psubsw\
psubsw XMM,XMM/m128\
源存储器与目的寄存器按字对齐有符号补码饱和相减(目的减去源),内存变量必须对齐内存16字节.\
 51 - 60, 共 150 个条目\
51\
Psubusb\
psubusb XMM,XMM/m128\
源存储器与目的寄存器按字节对齐无符号饱和相减(目的减去源),内存变量必须对齐内存16字节\

52\
Psubusw\
psubusw XMM,XMM/m128\
源存储器与目的寄存器按字对齐无符号饱和相减(目的减去源),内存变量必须对齐内存16字节.\

53\
psubb\
psubb XMM,XMM/m128\
源存储器与目的寄存器按字节对齐无符号普通相减(目的减去源),内存变量必须对齐内存16字节\

54\
Psubw\
psubw XMM,XMM/m128\
源存储器与目的寄存器按字对齐无符号普通相减(目的减去源),内存变量必须对齐内存16字节\

55\
Psubd\
psubd XMM,XMM/m128\
源存储器与目的寄存器按双字对齐无符号普通相减(目的减去源),内存变量必须对齐内存16字节\

56\
Psubq\
psubq XMM,XMM/m128\
源存储器与目的寄存器按四字对齐无符号普通相减(目的减去源),内存变量必须对齐内存16字节\

57\
Maxps\
maxps XMM,XMM/m128\
源存储器4个单精度浮点数与目的寄存器4个单精度浮点数比较,较大数放入对应目的寄存器,内存变量必须对齐内存16字节.\

58\
Maxss\
maxss XMM,XMM/m32\
源存储器低32位1个单精度浮点数与目的寄存器低32位1个单精度浮点数比较,较大数放入目的寄存器低32位,高96位不变内存变量不必对齐内存16字节\

59\
Minps\
minps XMM,XMM/m128\
源存储器4个单精度浮点数与目的寄存器4个单精度浮点数比较,较小数放入对应目的寄存器,内存变量必须对齐内存16字节.\

60\
minss\
minss XMM,XMM/m32\
源存储器低32位1个单精度浮点数与目的寄存器低32位1个单精度浮点数比较,较小数放入目的寄存器低32位,高96位不变内存变量不必对齐内存16字节\
 61 - 70, 共 150 个条目\
61\
cmpps\
cmpps XMM0,XMM1,imm8\
imm8是立即数范围是0~7\
根据imm8的值进行4对单精度浮点数的比较，符合imm8的就置目的寄存器对应的32位全1,否则全0\
当imm8 = 0时,目的寄存器等于原寄存器数时，置目的寄存器对应的32位全1,否则全0\
imm8 = 1 时,目的寄存器小于原寄存器数时，置目的寄存器对应的32位全1,否则全0\
imm8 = 2 时,目的寄存器小于等于原寄存器数时，置目的寄存器对应的32位全1,否则全0\
imm8 = 4 时,目的寄存器不等于原寄存器数时，置目的寄存器对应的32位全1,否则全0\
imm8 = 5 时,目的寄存器大于等于原寄存器数时，置目的寄存器对应的32位全1,否则全0\
imm8 = 6 时,目的寄存器大于原寄存器数时，置目的寄存器对应的32位全1,否则全0\

62\
pcmpeqb\
pcmpeqb XMM,XMM/m128\
目的寄存器与源存储器按字节比较,如果对应字节相等,就置目的寄存器对应字节为0ffh,否则为00h内存变量必须对齐内存16字节.\

63\
Pcmpeqw\
pcmpeqw XMM,XMM/m128\
目的寄存器与源存储器按字比较,如果对应字相等,就置目的寄存器对应字为0ffffh,否则为0000h, 内存变量必须对齐内存16字节\

64\
Pcmpeqd\
pcmpeqd XMM,XMM/m128\
目的寄存器与源存储器按双字比较,如果对应双字相等,就置目的寄存器对应双字为0ffffffffh,否则为00000000h内存变量必须对齐内存16字节\

65\
Pcmpgtb\
pcmpgtb XMM,XMM/m128\
目的寄存器与源存储器按字节(有符号补码)比较,如果目的寄存器对应字节大于源存储器,就置目的寄存器对应字节为0ffh, 否则为00h,内存变量必须对齐内存16字节\

66\
Pcmpgtw\
pcmpgtw XMM,XMM/m128\
目的寄存器与源存储器按字(有符号补码)比较,如果目的寄存器对应字大于源存储器,就置目的寄存器对应字为0ffffh, 否则为0000h,内存变量必须对齐内存16字节.\

67\
Pcmpgtd\
pcmpgtd XMM,XMM/m128\
目的寄存器与源存储器按双字(有符号补码)比较,如果目的寄存器对应双字大于源存储器, 就置目的寄存器对应双字为0ffffffffh,否则为00000000h,内存变量必须对齐内存16字节.\

68\
rcpps\
rcpps XMM,XMM/m128\
源存储器4个单精度浮点数的倒数放入对应目的寄存器,内存变量必须对齐内存16字节\
注:比如2.0E0的倒数为1÷2.0E0 = 5.0E-1, 这操作只有12bit的精度\

69\
rcpss\
rcpss XMM,XMM/32\
源存储器低32位1个单精度浮点数的倒数放入目的寄存器低32位,高96位不变\

70\
rsqrtps\
rsqrtps XMM,XMM/m128\
源存储器4个单精度浮点数的开方的倒数放入对应目的寄存器,内存变量必须对齐内存16字节. 比如2.0E0的开方的倒数为1÷√2.0E0 ≈ 7.0711E-1, 这操作只有12bit的精度.\
 71 - 80, 共 150 个条目\
71\
Rsqrtss\
rsqrtss XMM,XMM/32\
源存储器低32位1个单精度浮点数的开方的倒数放入目的寄存器低32位,高96位不变,内存变量不必对齐内存16字节.\

72\
Pavgb\
pavgb MM,MM/m64\
pavgb XMM,XMM/m128\
把源存储器与目的寄存器按字节无符号整数相加,再除以2,结果四舍五入为整数放入目的寄存器, 源存储器为m128时,内存变量必须对齐内存16字节. 注:此运算不会产生溢出.\

73\
Pavgw\
pavgw MM,MM/m64\
pavgw XMM,XMM/m128\
把源存储器与目的寄存器按字无符号整数相加,再除以2,结果四舍五入为整数放入目的寄存器, 源存储器为m128时,内存变量必须对齐内存16字节.\

74\
Sqrtpd\
sqrtpd XMM,XMM/m128\
源存储器两个双精度浮点数的开方放入对应目的寄存器,内存变量必须对齐内存16字节.\

75\
Mulps\
mulps XMM,XMM/m128\
源存储器内容按双字对齐,共4个单精度浮点数与目的寄存器相乘,结果送入目的寄存器,内存变量必须对齐内存16字节.\

76\
Mulss\
mulss XMM,XMM/32\
源存储器的低32位1个单精度浮点数与目的寄存器的低32位1个单精度浮点数相乘,结果送入目的寄存器的低32位, 高96位不变,内存变量不必对齐内存16字节\

77\
Mulpd\
mulpd XMM,XMM/m128\
源存储器内容按四字对齐,共两个双精度浮点数与目的寄存器相乘,结果送入目的寄存器,内存变量必须对齐内存16字节\

78\
Mulsd\
mulsd XMM,XMM/m128\
源存储器的低64位1个双精度浮点数与目的寄存器的低64位1个双精度浮点数相乘,结果送入目的寄存器的低64位, 高64位不变,内存变量不必对齐内存16字节\

79\
Pmuludq\
pmuludq XMM,XMM/m128\
把源存储器与目的寄存器的低32位无符号整数相乘,结果变为64位,送入目的寄存器低64位, 把源存储器与目的寄存器的高64位的低32位无符号整数相乘,结果变为64位,送入目的寄存器高64位内存变量必须对齐内存16字节.\
高64位 | 低64位\
目的寄存器: a0 | a1 | a2 | a3\
源存储器: b0 | b1 | b2 | b3\
目的寄存器结果: b1*a1 | b3*a3\

80\
Pmuludq\
pmuludq MM,MM/m64\
把源存储器与目的寄存器的低32位无符号整数相乘,结果变为64位,送入目的寄存器.\
 81 - 90, 共 150 个条目\
81\
pmulhw\
pmulhw XMM,XMM/m128\
源存储器与目的寄存器按字对齐有符号补码饱和相乘,取结果的高16位放入目的寄存器对应字中. 内存变量必须对齐内存16字节\

82\
pmullw\
pmullw XMM,XMM/m128\
源存储器与目的寄存器按字对齐有符号补码饱和相乘,取结果的低16位放入目的寄存器对应字中. 内存变量必须对齐内存16字节.\

83\
Divps\
divps XMM,XMM/m128\
目的寄存器共4个单精度浮点数除以源存储器4个单精度浮点数,结果送入目的寄存器,内存变量必须对齐内存16字节.\

84\
Divss\
divss XMM,XMM/32\
目的寄存器低32位1个单精度浮点数除以源存储器低32位1个单精度浮点数,结果送入目的寄存器的低32位, 高96位不变,内存变量不必对齐内存16字节\

85\
Divpd\
divpd XMM,XMM/m128\
目的寄存器共两个双精度浮点数除以源存储器两个双精度浮点数,结果送入目的寄存器,内存变量必须对齐内存16字节\

86\
Divsd\
divsd XMM,XMM/m128\
目的寄存器低64位1个双精度浮点数除以源存储器低64位1个双精度浮点数,结果送入目的寄存器的低64位, 高64位不变,内存变量不必对齐内存16字节.\

87\
Andps\
andps XMM,XMM/m128\
源存储器128个二进制位'与'目的寄存器128个二进制位,结果送入目的寄存器,内存变量必须对齐内存16字节.\

88\
Orps\
orps XMM,XMM/m128\
源存储器128个二进制位'或'目的寄存器128个二进制位,结果送入目的寄存器,内存变量必须对齐内存16字节.\

89\
Xorps\
xorps XMM,XMM/m128\
源存储器128个二进制位'异或'目的寄存器128个二进制位,结果送入目的寄存器,内存变量必须对齐内存16字节.\

90\
Unpckhps\
unpckhps XMM,XMM/m128\
源存储器与目的寄存器高64位按双字交错排列,结果送入目的寄存器,内存变量必须对齐内存16字节.\
高64位 | 低64位\
目的寄存器: a0 | a1 | a2 | a3\
源存储器: b0 | b1 | b2 | b3\
目的寄存器结果: b0 | a0 | b1 | a1\
 91 - 100, 共 150 个条目\
91\
Unpcklps\
unpcklps XMM,XMM/m128\
源存储器与目的寄存器低64位按双字交错排列,结果送入目的寄存器,内存变量必须对齐内存16字节\
高64位 | 低64位\
目的寄存器: a0 | a1 | a2 | a3\
源存储器: b0 | b1 | b2 | b3\
目的寄存器结果: b2 | a2 | b3 | a3\

92\
Pextrw\
pextrw r32,MM,imm8\
pextrw r32,XMM,imm8\
imm8为8位立即数(无符号)\
从源寄存器中选第imm8(0~3 或 0~7)个字送入目的寄存器的低16位,高16位清零.\
注:imm8范围为 0~255,当源寄存器为'MM'时,有效值= imm8 mod 4,当目的寄存器为'XMM'时,有效值= imm8 mod 8\

93\
Pinsrw\
pinsrw MM,r32/m32,imm8\
pinsrw XMM,r32/m32,imm8\
把源存储器的低16位内容送入目的寄存器第imm8(0~3 或 0~7)个字,其余字不变\
注:imm8范围为 0~255,当目的寄存器为'MM'时,有效值= imm8 mod 4,当目的寄存器为'XMM'时,有效值= imm8 mod 8\

94\
Pmaxsw\
pmaxsw MM,MM/m64\
pmaxsw XMM,XMM/m128\
把源存储器与目的寄存器按字有符号(补码)整数比较,大数放入目的寄存器对应字, 源存储器为m128时,内存变量必须对齐内存16字节\

95\
Pmaxub\
pmaxub MM,MM/m64\
pmaxsw XMM,XMM/m128\
把源存储器与目的寄存器按字节无符号整数比较,大数放入目的寄存器对应字节, 源存储器为m128时,内存变量必须对齐内存16字节.\

96\
pminsw\
pminsw MM,MM/m64\
pmaxsw XMM,XMM/m128\
把源存储器与目的寄存器按字有符号(补码)整数比较,较小数放入目的寄存器对应字, 源存储器为m128时,内存变量必须对齐内存16字节.\

97\
Pminub\
pminub MM,MM/m64\
pmaxsw XMM,XMM/m128\
把源存储器与目的寄存器按字节无符号整数比较,较小数放入目的寄存器对应字节, 源存储器为m128时,内存变量必须对齐内存16字节\

98\
Maxpd\
maxpd XMM,XMM/m128\
源存储器两个双精度浮点数与目的寄存器两个双精度浮点数比较,较大数放入对应目的寄存器,内存变量必须对齐内存16字节.\

99\
Maxsd\
maxsd XMM,XMM/m128\
源存储器低64位1个双精度浮点数与目的寄存器低64位1个双精度浮点数比较,较大数放入目的寄存器低64位,高64位不变内存变量不必对齐内存16字节.\

100\
Minpd\
minpd XMM,XMM/m128\
源存储器两个双精度浮点数与目的寄存器两个双精度浮点数比较,较小数放入对应目的寄存器,内存变量必须对齐内存16字节.\
 \
101 - 110, 共 150 个条目\
101\
Minsd\
minsd XMM,XMM/m128\
源存储器低64位1个双精度浮点数与目的寄存器低64位1个双精度浮点数比较,较小数放入目的寄存器低64位,高64位不变内存变量不必对齐内存16字节.\

102\
Andpd\
andpd XMM,XMM/m128\
源存储器128个二进制位'与'目的寄存器128个二进制位,结果送入目的寄存器,内存变量必须对齐内存16字节.\

103\
Andnpd\
andnpd XMM,XMM/m128\
目的寄存器128个二进制位先取'非',再'与'源存储器128个二进制位,结果送入目的寄存器,内存变量必须对齐内存16字节\

104\
Orpd\
orpd XMM,XMM/m128\
源存储器128个二进制位'或'目的寄存器128个二进制位,结果送入目的寄存器,内存变量必须对齐内存16字节.\

105\
Xorpd\
xorpd XMM,XMM/m128\
源存储器128个二进制位'异或'目的寄存器128个二进制位,结果送入目的寄存器,内存变量必须对齐内存16字节.\

106\
Pslldq\
pslldq XMM,imm8\
把目的寄存器128位按imm8(立即数)指定字节数逻辑左移,移出的字节丢失.\
imm8 == 1时,代表左移8位,imm8 == 2时,代表左移16位.\

107\
Psrldq\
psrldq XMM,imm8\
把目的寄存器128位按imm8(立即数)指定字节数逻辑右移,移出的字节丢失.\
imm8 == 1时,代表右移8位,imm8 == 2时,代表右移16位.\

108\
Psllw\
psllw XMM,XMM/m128\
psllw XMM,imm8\
把目的寄存器按字由源存储器(或imm8 立即数)指定位数逻辑左移,移出的位丢失. 低字移出的位不会移入高字,内存变量必须对齐内存16字节.\

109\
Psrlw\
psrlw XMM,XMM/m128\
psrlw XMM,imm8\
把目的寄存器按字由源存储器(或imm8 立即数)指定位数逻辑右移,移出的位丢失.\
高字移出的位不会移入低字,内存变量必须对齐内存16字节.\

110\
Pslld\
pslld XMM,XMM/m128\
pslld XMM,XMM imm8\
把目的寄存器按双字由源存储器(或imm8 立即数)指定位数逻辑左移,移出的位丢失. 低双字移出的位不会移入高双字,内存变量必须对齐内存16字节.\
 111 - 120, 共 150 个条目\
111\
Psrld\
psrld XMM,XMM/m128\
psrld XMM,imm8\
把目的寄存器按双字由源存储器(或imm8 立即数)指定位数逻辑右移,移出的位丢失.\
高双字移出的位不会移入低双字,内存变量必须对齐内存16字节.\
pand\
pand XMM,XMM/m128\
源存储器128个二进制位'与'目的寄存器128个二进制位,结果送入目的寄存器,内存变量必须对齐内存16字节. 我发现与andpd功能差不多,就不知其它特性是否一样\

112\
Pandn\
pandn XMM,XMM/m128\
目的寄存器128个二进制位先取'非',再'与'源存储器128个二进制位,结果送入目的寄存器,内存变量必须对齐内存16字节\

113\
Por\
por XMM,XMM/m128\
源存储器128个二进制位'或'目的寄存器128个二进制位,结果送入目的寄存器,内存变量必须对齐内存16字节.\

114\
Pxor\
pxor XMM,XMM/m128\
源存储器128个二进制位'异或'目的寄存器128个二进制位,结果送入目的寄存器,内存变量必须对齐内存16字节.\

115\
packuswb\
packuswb XMM,XMM/m128\
packuswb MM,MM/m64\
把目的寄存器按字有符号数压缩为字节无符号数放入目的寄存器低64位\
把源寄存器按字有符号数压缩为字节无符号数放入目的寄存器高64位\
压缩时负数变为00h,大于255的正数变为0ffh,内存变量必须对齐内存16字节.\
高64位 | 低64位\
目的寄存器: a0 | a1 | a2 | a3 | a4 | a5 | a6 | a7\
源寄存器: b0 | b1 | b2 | b3 | b4 | b5 | b6 | b7\
目的寄存器压缩结果: b0|b1| b2| b3| b4|b5| b6|b7| a0|a1| a2|a3| a4|a5| a6| a7\

116\
packsswb\
packsswb XMM,XMM/m128\
packsswb MM,MM/m64\
把目的寄存器按字有符号数压缩为字节有符号数放入目的寄存器低64位\
把源寄存器按字有符号数压缩为字节有符号数放入目的寄存器高64位\
压缩时小于-128负数变为80h,大于127的正数变为7fh,内存变量必须对齐内存16字节.\
高64位 | 低64位\
目的寄存器: a0 | a1 | a2 | a3 | a4 | a5 | a6 | a7\
源寄存器: b0 | b1 | b2 | b3 | b4 | b5 | b6 | b7\
目的寄存器压缩结果: b0|b1| b2| b3| b4|b5| b6|b7| a0|a1| a2|a3| a4|a5| a6| a7\

117\
packssdw\
packssdw XMM,XMM/m128\
把目的寄存器按双字有符号数压缩为字有符号数放入目的寄存器低64位\
把源寄存器按双字有符号数压缩为字有符号数放入目的寄存器高64位\
压缩时小于-32768负数变为8000h,大于32767的正数变为7fffh,内存变量必须对齐内存16字节.\
高64位 | 低64位\
目的寄存器: a0 | a1 | a2 | a3\
源寄存器: b0 | b1 | b2 | b3\
目的寄存器压缩结果: b0 | b1 | b2 | b3 | a0 | a1 | a2 | a3\

118\
punpckldq\
punpckldq XMM,XMM/m128\
把源存储器与目的寄存器低64位按双字交错排列,内存变量必须对齐内存16字节.\
高64位 | 低64位\
目的寄存器: a0 | a1 | a2 | a3\
源寄存器: b0 | b1 | b2 | b3\
目的寄存器排列结果: b2 | a2 | b3 | a3\

119\
punpckhdq\
把源存储器与目的寄存器高64位按双字交错排列\
内存变量必须对齐内存16字节.\
高64位 | 低64位\
目的寄存器: a0 | a1 | a2 | a3\
源寄存器: b0 | b1 | b2 | b3\
目的寄存器排列结果: b0 | a0 | b1 | a1\

120\
punpcklwd\
把源存储器与目的寄存器低64位按字交错排列\
内存变量必须对齐内存16字节.\
高64位 | 低64位\
目的寄存器: a0 | a1 | a2 | a3 | a4 | a5 | a6 | a7\
源寄存器: b0 | b1 | b2 | b3 | b4 | b5 | b6 | b7\
目的寄存器排列结果: b4 | a4 | b5 | a5 | b6 | a6 | b7 | a7\
 \
121 - 130, 共 150 个条目\
121\
punpckhwd\
punpckhwd XMM,XMM/m128\
把源存储器与目的寄存器高64位按字交错排列,内存变量必须对齐内存16字节.\
高64位 | 低64位\
目的寄存器: a0 | a1 | a2 | a3 | a4 | a5 | a6 | a7\
源寄存器: b0 | b1 | b2 | b3 | b4 | b5 | b6 | b7\
目的寄存器排列结果: b0 | a0 | b1 | a1 | b2 | a2 | b3 | a3\

122\
punpcklbw\
punpcklbw XMM,XMM/m128\
把源存储器与目的寄存器低64位按字节交错排列,内存变量必须对齐内存16字节.\
高64位 | 低64位\
目的寄存器: a0|a1| a2| a3| a4|a5| a6|a7| a8|a9| aA|aB| aC|aD| aE| aF\
源寄存器: b0|b1| b2| b3| b4|b5| b6|b7| b8|b9| bA|bB| bC|bD| bE| bF\
目的寄存器排列结果: b8|a8| b9| a9| bA|aA| bB|aB| bC|aC| bD|aD| bE|aE| bF| aF\

123\
punpckhbw\
把源存储器与目的寄存器高64位按字节交错排列\
内存变量必须对齐内存16字节.\
高64位 | 低64位\
目的寄存器: a0|a1| a2| a3| a4|a5| a6|a7| a8|a9| aA|aB| aC|aD| aE| aF\
源寄存器: b0|b1| b2| b3| b4|b5| b6|b7| b8|b9| bA|bB| bC|bD| bE| bF\
目的寄存器排列结果: b0|a0| b1| a1| b2|a2| b3|a3| b4|a4| b5|a5| b6|a6| b7| a7\

124\
shufps\
shufps XMM,XMM/m128,imm8\
把源存储器与目的寄存器按双字划分,由imm8(立即数)八个二进制位(00~11,00^11,00~11,00~11)指定排列, 内存变量必须对齐内存16字节.目的寄存器高64位放源存储器被指定数,目的寄存器低64位放目的寄存器被指定数. '( )'中的都是二进制数\
目的寄存器: a(11) | a(10) | a(01) | a(00)\
源寄存器: b(11) | b(10) | b(01) | b(00)\
目的寄存器排列结果: b(00~11) | b(00~11) | a(00~11) | a(00~11)\
目的寄存器压缩结果'( )'中的值由imm8对应的两位二进制位指定.\

125\
shufpd\
shufpd XMM,XMM/m128,imm8(0~255) imm8(操作值) = imm8(输入值) mod 4\
把源存储器与目的寄存器按四字划分,由imm8(立即数)4个二进制位(0~1,0^1,0~1,0~1)指定排列, 内存变量必须对齐内存16字节.目的寄存器高64位放源存储器被指定数,目的寄存器低64位放目的寄存器被指定数.\
当XMM0 = 1111111122222222 3333333344444444 h\
XMM1 = 5555555566666666 aaaaaaaacccccccc h,执行shufpd XMM0,XMM1,101001 1 0 b\
则XMM0 = 5555555566666666 3333333344444444 h\

126\
pshuflw\
pshuflw XMM,XMM/m128,imm8(0~255)\
先把源存储器的高64位内容送入目的寄存器的高64位,然后用imm8将源存储器的低64位4个字选入目的寄存器的低64位,内存变量必须对齐内存16字节.\
源寄存器低64位: b(11) | b(10) | b(01) | b(00)\
目的寄存器低64位排列结果: b(00~11) | b(00~11) | b(00~11) | b(00~11)\
当XMM0 = 1111111122222222 3333 4444 5555 6666 h\
XMM1 = 5555555566666666 7777 8888 9999 cccc h,执行pshuflw XMM0,XMM1,10 10 01 10 b\
则XMM0 = 5555555566666666 8888 8888 9999 8888 h\

127\
pshufhw\
pshufhw XMM,XMM/m128,imm8(0~255)\
先把源存储器的低64位内容送入目的寄存器的低64位,然后用imm8将源存储器的高64位4个字选入目的寄存器的高64位,内存变量必须对齐内存16字节.\
源寄存器高64位: b(11) | b(10) | b(01) | b(00)\
目的寄存器高64位排列结果: b(00~11) | b(00~11) | b(00~11) | b(00~11)\
当XMM0 = 3333 4444 5555 6666 1111111122222222 h\
XMM1 = 7777 8888 9999 cccc 5555555566666666 h,执行pshufhw XMM0,XMM1,10 10 01 10 b\
则XMM0 = 8888 8888 9999 8888 5555555566666666 h\

128\
pshufd\
pshufd XMM,XMM/m128,imm8(0~255)\
将源存储器的4个双字由imm8指定选入目的寄存器,内存变量必须对齐内存16字节.\
源寄存器: b(11) | b(10) | b(01) | b(00)\
目的寄存器排列结果: b(00~11) | b(00~11) | b(00~11) | b(00~11)\
当XMM1 = 11111111 22222222 33333333 44444444 h,执行pshufd XMM0,XMM1,11 01 01 10b\
则XMM0 = 11111111 33333333 33333333 22222222 h\

129\
cvtpi2ps\
cvtpi2ps XMM,MM/m64\
源存储器64位两个32位有符号(补码)整数转为两个单精度浮点数,放入目的寄存器低64中,高64位不变.\

130\
cvtsi2ss\
cvtsi2ss XMM,r32/m32\
源存储器1个32位有符号(补码)整数转为1个单精度浮点数,放入目的寄存器低32中,高96位不变\
 131 - 140, 共 150 个条目\
131\
cvtps2pi\
cvtps2pi MM,XMM/m64\
把源存储器低64位两个32位单精度浮点数转为两个32位有符号(补码)整数,放入目的寄存器\

132\
cvttps2pi\
cvttps2pi MM,XMM/m64\
类似于cvtps2pi，截断取整.\

133\
cvtss2si\
cvtss2si r32,XMM/m32\
把源存储器低32位1个单精度浮点数转为1个32位有符号(补码)整数,放入目的寄存器.\

134\
cvttss2si\
cvttss2si r32,XMM/m32\
类似cvtss2si,截断取整.\

135\
cvtps2pd\
cvtps2pd XMM,XMM/m64\
把源存储器低64位两个单精度浮点数变成两个双精度浮点数,结果送入目的寄存器.\

136\
cvtss2sd\
cvtss2sd XMM,XMM/m32\
把源存储器低32位1个单精度浮点数变成1个双精度浮点数,结果送入目的寄存器的低64位,高64位不变.\

137\
cvtpd2ps\
把源存储器两个双精度浮点数变成两个单精度浮点数,结果送入目的寄存器的低64位,高64位清零\
内存变量必须对齐内存16字节.\
＾特殊状态 ＾3.14E5 (表示负无穷大)\

138\
cvtsd2ss\
cvtsd2ss XMM,XMM/m64\
把源存储器低64位1个双精度浮点数变成1个单精度浮点数,结果送入目的寄存器的低32位,高96位不变.\

139\
cvtpd2pi\
cvtpd2pi MM,XMM/m128\
把源存储器两个双精度浮点数变成两个双字有符号整数,结果送入目的寄存器,内存变量必须对齐内存16字节. 如果结果大于所能表示的范围,那么转化为\
80000000h(正数也转为此值).\

140\
cvttpd2pi\
cvttpd2pi MM,XMM/m128\
类似于cvtpd2pi,截断取整.\
 141 - 150, 共 150 个条目\
141\
cvtpi2pd\
cvtpi2pd XMM,MM/m64\
把源存储器两个双字有符号整数变成两个双精度浮点数,结果送入目的寄存器.\

142\
cvtpd2dq\
cvtpd2dq XMM,XMM/m128\
把源存储器两个双精度浮点数变成两个双字有符号整数(此运算与cvtpd2pi类似但目的寄存器变为XMM), 结果送入目的寄存器的低64位,高64位清零,内存变\
量必须对齐内存16字节.\

143\
cvttpd2dq\
cvttpd2dq XMM,XMM/m128\
类似于cvtpd2dq，为截断取整.\

144\
cvtdq2pd\
cvtdq2pd XMM,XMM/m128\
把源存储器低64位两个双字有符号整数变成两个双精度浮点数,结果送入目的寄存器,内存变量必须对齐内存16字节.\

145\
cvtsd2si\
cvtsd2si r32,XMM/m64\
把源存储器低64位1个双精度浮点数变成1个双字有符号整数,结果送入目的寄存器.\

146\
cvttsd2si\
cvttsd2si r32,XMM/m64\
类似于cvtsd2si，截断取整.\

147\
cvtsi2sd\
cvtsi2sd XMM,r32/m32\
把源存储器1个双字有符号整数变成1个双精度浮点数,结果送入目的寄存器的低64位,高64位不变\

148\
cvtps2dq\
cvtps2dq XMM,XMM/m128\
把源存储器4个单精度浮点数变成4个双字有符号整数,结果送入目的寄存器,内存变量必须对齐内存16字节.\

149\
cvttps2dq\
cvttps2dq XMM,XMM/m128\
类似于cvtps2dq，截断取整\

150\
cvtdq2ps\
cvtdq2ps XMM,XMM/m128\
把源存储器4个双字有符号整数变成4个单精度浮点数,结果送入目的寄存器,内存变量必须对齐内存16字节.\
 

++++++++++++++++++++++++++++++\
SSE3：\
SSE3 — OpCode List\
(under construction - need pictures :)\
Arithmetic:\
addsubpd - Adds the top two doubles and subtracts the bottom two.\
addsubps - Adds top singles and subtracts bottom singles.\
haddpd - Top double is sum of top and bottom, bottom double is sum of second operand's top and bottom.\
haddps - Horizontal addition of single-precision values.\
hsubpd - Horizontal subtraction of double-precision values.\
hsubps - Horizontal subtraction of single-precision values.

Load/Store:\
lddqu - Loads an unaligned 128bit value.\
movddup - Loads 64bits and duplicates it in the top and bottom halves of a 128bit register.\
movshdup - Duplicates the high singles into high and low singles.\
movsldup - Duplicates the low singles into high and low singles.\
fisttp - Converts a floating-point value to an integer using truncation.

Process Control:\
monitor - Sets up a region to monitor for activity.\
mwait - Waits until activity happens in a region specified by monitor.\
# 共通性的指令

## 算术指令（Arithmetic）\
  ADDSUBPD - （Add-Subtract-Packed-Double）\
    输入： - { A0, A1 }, { B0, B1 }\
    输出： - { A0 - B0, A1 + B1 }\
  ADDSUBPS - （Add-Subtract-Packed-Single）\
    输入： { A0, A1, A2, A3 }, { B0, B1, B2, B3 }\
    输出： { A0 - B0, A1 + B1, A2 - B2, A3 + B3 }

## 数组结构指令（Array Of Structures；AOS）\
  HADDPD - （Horizontal-Add-Packed-Double）\
    输入： { A0, A1 }, { B0, B1 }\
    输出： { B0 + B1, A0 + A1 }\
  HADDPS （Horizontal-Add-Packed-Single）\
    输入： { A0, A1, A2, A3 }, { B0, B1, B2, B3 }\
    输出： { B0 + B1, B2 + B3, A0 + A1, A2 + A3 }\
  HSUBPD - （Horizontal-Subtract-Packed-Double）\
    输入： { A0, A1 }, { B0, B1 }\
    输出： { A0 - A1, B0 - B1 }\
  HSUBPS - （Horizontal-Subtract-Packed-Single）\
    输入： { A0, A1, A2, A3 }, { B0, B1, B2, B3 }\
    输出： { A0 - A1, A2 - A3, B0 - B1, B2 - B3 }\
  LDDQU - 如上所述，这是有交替需求时所用的指令，可以加载（load）不整齐排列的整数向量值，此指令对视频压缩的运算工作有帮助。\
  MOVDDUP、MOVSHDUP、MOVSLDUP - 此三个指令是针对复杂数目需求时所用，对波形信号的运算有帮助，例如音频的声波波形处理。\
  FISTTP - 类似过去x87浮点运算中的FISTP指令，不过此指令的运算执行或忽略掉浮点控制寄存器的rounding（溢绕）模式的设置，并且用“chop”（truncate，截切）模式[6]取代。允许控制寄存器忽略繁重的加载及再加载，例如C语言中将浮点数转换成整数就需要使用此种截切效果，且此种截切程序已成为C语言中的标准作法。\
# Intel针对SSE3所额外设计的自用指令\
  MONITOR 、 MWAIT - 此二个指令能针对多线程的应用程序进行执行优化，使处理器原有的超线程功效获得更佳的发挥。


SSSE3:\
在以下的列表中，satsw(X)（饱和为有符号字（saturate to signed word）的简写），任取有号整数X，如果X小于-32768时就代表-32768，X大于32767时就代表32767 ，其余数值不变。在一般的Intel架构上，字节（byte）表示8位，字（word）是16位，而双字（dword）是32位；寄存器表示MMX或是XMM向量寄存器。\
    PSIGNB, PSIGNW, PSIGND\
         包裹式有符号整型取反     如果另一个寄存器中的整形为负，那么将目标寄存器中的数取反。\
    PABSB, PABSW, PABSD\
          包裹式绝对值    将源寄存器中的数取绝对值并放到目标寄存器中。\
    PALIGNR\
        包裹式右移    将两个寄存器的值串起来，然后根据编码到指令中的立即数将寄存器中的值右移。\
    PSHUFB\
        包裹式将任意字节重新排布到目的寄存器    如果源寄存器高位被置1，就把目的寄存器赋值为0,否则根据源操作数的低4位选择目的操作数，将其拷贝到目的操作数的相应位置。\
        \
        字节混选指令: PSHUFB (https://www.officedaytime.com/simd512e/simdimg/si.php?f=pshufb)\
           这条指令非常的灵活，用于选取源寄存器的任意字节重新排布到目的寄存器。用伪代码描述如下：\
        　　char a[16]; // input a\
        　　char b[16]; // input b\
        　　char r[16]; // output r\
        　　for (i=0; i < 16; i++)\
        　　    r[i] = (b[i] < 0) ? 0 : a[b[i] % 16];\
        如果源操作数的该字节高位符号位为1，即源字节操作数小于零的时候，把目的寄存器赋值为零。\
        否则，根据源操作数的低4位选择目的操作数，将其拷贝到目的操作数的相应位。\
    PMULHRSW\
        包裹式舍入相乘    将两个寄存器中的16位word处理成-1到1间的15位定点数(例如0x4000被处理成0.5，0xa000 处理成−0.75), 并且将他们舍入相乘。\
    PMADDUBSW\
        相乘并相加包裹式整型然后饱和    将两个寄存器中的8位整型相乘并相加，然后饱和成有符号整型。（也就是 [a0 a1 a2 …] pmaddubsw [b0 b1 b2 …] = [satsw(a0b0+a1b1) satsw(a2b2+a3b3) …]）\
    PHSUBW, PHSUBD\
        包裹式水平相减    将两个寄存器 A = [a0 a1 a2 …] 和 B = [b0 b1 b2 …] 相减输出 [a0−a1 a2−a3 … b0−b1 b2−b3 …]\
    PHSUBSW\
        包裹式水平相减并且饱和为有符号字    类似PHSUBW, 但是输出的是[satsw(a0−a1) satsw(a2−a3) … satsw(b0−b1) satsw(b2−b3) …]\
    PHADDW, PHADDD\
        包裹式有符号相加    将两个寄存器 A = [a0 a1 a2 …] 和 B = [b0 b1 b2 …] 相加然后输出 [a0+a1 a2+a3 … b0+b1 b2+b3 …]\
    PHADDSW\
        包裹式水平相加并且饱和为有符号字    类似PHADDW, 但是输出的是[satsw(a0+a1) satsw(a2+a3) … satsw(b0+b1) satsw(b2+b3) …]


SSE4：\
与之前SSE的所有迭代不同，SSE4包含执行不特定于多媒体应用的操作的指令。它具有许多指令，其操作由一个常量字段和一组将XMM0作为隐式第三操作数的指令决定。\
enryn公司的单周期shuffle引擎激活了其中的几条指令。（随机操作重新排序寄存器中的字节被称为shuffle。）\
(SSE4.1 ?)\
    指令    描述\
    MPSADBW\
        計算絕對差的八個偏移和，每次四個（即：|x0−y0|+|x1−y1|+|x2−y2|+|x3−y3|, |x0−y1|+|x1−y2|+|x2−y3|+|x3−y4|, …, |x0−y7|+|x1−y8|+|x2−y9|+|x3−y10|）。這個操作對一些HD 编解码器来说很重要。并且允许在少于七个周期内计算8×8块的差异。[9]三位直接操作数的一个位指示是否应从目标操作数中使用y0 .. y10或 y4 .. y14, 另外两种方法是否应从源中使用x0..x3, x4..x7, x8..x11或x12..x15。\
    PHMINPOSUW\
        将目标的底部无符号16位字设置为源中最小的无符号16位字，将底部的下一个字设置为源中该字的索引。\
    PMULDQ\
        在两组四个压缩整数中的两组中进行压缩有符号乘法，第一个和第三个压缩4，给出两个打包的64位结果。\
    PMULLD\
        打包有符号乘法，四个打包的32位整数组相乘，得到4个打包的32位结果。\
    DPPS；DPPD\
        AOS（结构数组）数据的点积。这需要一个立即操作数，它由四个（或两个DPPD）位组成，用于选择输入中的哪个条目进行乘法和累加，另外四个（或两个DPPD）选择是将0还是点积输出的相应字段。\
    BLENDPS；BLENDPD； BLENDVPS；BLENDVPD；PBLENDVB；PBLENDW\
        基于（对于非V形式）立即操作数中的位以及（对于V形式）寄存器XMM0中的位的条件复制一个位置中的元素与另一个位置中的元素。\
    PMINSB；PMAXSB；PMINUW；PMAXUW；PMINUD；PMAXUD；PMINSD；PMAXSD\
        不同整型操作数类型的最小/最大值压缩。\
    ROUNDPS；ROUNDSS；ROUNDPD；ROUNDSD\
        使用立即数操作数指定的四种舍入模式中的一种将浮点寄存器中的值整数到整数。\
    INSERTPS；PINSRB；PINSRD / PINSRQ；EXTRACTPS；PEXTRB；PEXTRD / PEXTRQ\
        NSERTPS和PINSR指令从x86寄存器或存储器位置读取8,16或32位，并将其插入由立即数操作数给定的目标寄存器中的字段。EXTRACTPS和PEXTR从源寄存器中读取一个字段，并将其插入x86寄存器或存储器位置。例如，PEXTRD eax，[xmm0]，1; EXTRACTPS [addr + 4 * eax]，xmm1，1将xmm1的第一个字段存储在由xmm0的第一个字段给出的地址中。\
    PMOVSXBW；PMOVZXBW；PMOVSXBD；PMOVZXBD；PMOVSXBQ；PMOVZXBQ；PMOVSXWD；PMOVZXWD；PMOVSXWQ ；PMOVZXWQ；PMOVSXDQ；PMOVZXDQ\
        打包标志/零扩展到更广泛的类型。\
    PTEST\
        这与TEST指令相似，因为它将Z标志设置为其操作数之间的AND结果：如果DEST AND SRC等于0，则设置ZF。另外，如果（NOT DEST）AND SRC等于零。\
        这相当于如果没有设置SRC掩码的位，则设置Z标志，如果设置了SRC掩码的所有位，则设置C标志。\
    PCMPEQQ\
        四字节（64位）相等比较。\
    PACKUSDW\
        将带符号的DWORD转换为饱和的无符号WORD。\
    MOVNTDQA\
        从写入组合存储区有效读取到SSE寄存器; 这对于从连接到存储器总线的外设检索结果很有用。\
    \
SSE4.2\
SSE4.2添加了STTNI（字符串和文本新指令）[10]，和每次对16个字节的两个操作数执行字符搜索和比较的几个新指令。这些设计（除其他外）旨在加快解析XML文档。[11]这也增加了一个CRC32指令来计算循环冗余校验，比如可以在某些数据传输协议使用。这些指令首先在基于Nehalem的Intel Core i7产品系列中实现，并完成SSE4指令集。支持通过CPUID.01H：ECX.SSE42 [bit20]标志指示。\
    指令    描述\
    CRC32\
        使用多项式0x11EDC6F41（或没有高位，0x1EDC6F41）累加CRC32C值。\
    PCMPESTRI\
        打包比较显式长度字符串，返回索引。\
    PCMPESTRM\
        打包比较显式长度字符串，返回掩码。\
    PCMPISTRI \
       打包比较隐式长度字符串，返回索引。\
    PCMPISTRM\
        打包比较隐式长度字符串，返回掩码。\
    PCMPGTQ\
        比较已打包签名的64位数据。For Greater Than\
    \
POPCNT和LZCNT\
这些指令在整数而不是SSE寄存器上运行，因为它们不是SIMD指令，而是同时出现的指令。虽然它们是由AMD通过SSE4a指令集引入的，但却往往被视为单独的扩展，并且带有自己的专用CPUID位以指示对其的支持。Intel以Nehalem微体系架构和LZCNT开始，实现了从Haswell微架构开始的POPCNT 。AMD从Barcelona微体系架构开始实施。\
AMD称这一对高级位操作Advanced Bit Manipulation （ABM）指令。\
    指令    描述\
    POPCNT\
        汉明权重（计数字数设置为1）。支持通过CPUID.01H：ECX.POPCNT [位23]标志指示。\
    LZCNT\
        Find First Set。支持通过CPUID.80000001H：ECX.ABM [位5]标志指示。\
除非输入为0，否则lzcnt的结果等于bsr（位扫描反转）。lzcnt产生32的结果，而bsr产生未定义的结果（并设置零标志）。lzcnt的编码与bsr的编码相似，如果lzcnt在不支持它的CPU上执行，比如Haswell之前的Intel CPU，它将执行bsr操作，而不是产生无效的指令错误。\
Trailing zeros可以使用现有的bsf指令进行计数。

SSE4a\
AMD公司的Barcelona微体系架构中引入了SSE4a指令组。这些说明在英特尔处理器中不可用。支持通过CPUID.80000001H：ECX.SSE4A [Bit 6]标志指示。\
    指令    描述\
    EXTRQ / INSERTQ\
        组合掩码移位指令。\
    MOVNTSD / MOVNTSS\
        标量流存储指令。

++++++++++++++++++++++++++++++\
其他：\
clflush 从所有级别的 cache 刷新 cache line

++++++++++++++++++++++++++++++\
SSE\
 2 CVTSI2SS – 把一个64位的有符号整型转换为一个浮点值，并把它插入到一个128位的参数中。内部指令：_mm_cvtsi64_ss\
 3 CVTSS2SI – 取出一个32位的浮点值，并取整（四舍五入）为一个64位的整型。内部指令：_mm_cvtss_si64\
 4 CVTTSS2SI – 取出一个32位的浮点值，并截断为一个64位的整型。内部指令：_mm_cvttss_si64\
 5 SSE2\
 6 CVTSD2SI – 取出最低位的64位浮点值，并取整为一个整型。内部指令：_mm_cvtsd_si64\
 7 CVTSI2SD – 取出最低位的64位整型，并将其转换为一个浮点值。内部指令：_mm_cvtsi64_sd\
 8 CVTTSD2SI – 取出一个64位的浮点值，并截断为一个64位的整型。内部指令：_mm_cvttsd_si64\
 9 MOVNTI – 写64位数据到特定内存位置。内部指令：_mm_stream_si64\
10 MOVQ – 移动一个64位的整型到一个128位的参数中，或从128位的参数中移动一个64位的整型。内部指令：_mm_cvtsi64_si128、_mm_cvtsi128_si64\
11 SSSE3\
12 PABSB / PABSW / PABSD – 取有符号整型的绝对值。内部指令：_mm_abs_epi8、_mm_abs_epi16、_mm_abs_epi32、_mm_abs_pi8、_mm_abs_pi16、_mm_abs_pi32\
13 PALIGNR – 结合两个参数并右移结果。内部指令：_mm_alignr_epi8、_mm_alignr_pi8\
14 PHADDSW – 将两个包含16位有符号整型的参数相加，并尽量使结果为16位可表示的最大值。内部指令：_mm_hadds_epi16、_mm_hadds_pi16\
15 PHADDW / PHADDD – 将两个包含有符号整型的参数相加。内部指令：_mm_hadd_epi16、_mm_hadd_epi32、_mm_hadd_pi16、_mm_hadd_pi32\
16 PHSUBSW – 将两个包含16位有符号整型的参数相减，并尽量使结果为16位可表示的最大值。内部指令：_mm_hsubs_epi16、_mm_shubs_pi16\
17 PHSUBW / PHSUBD – 将两个包含有符号整型的参数相减。内部指令：_mm_hsub_epi16、_mm_hsub_epi32、_mm_hsub_pi16、_mm_hsub_pi32\
18 PMADDUBSW – 相乘并相加8位整型。内部指令：_mm_maddubs_epi16、_mm_maddubs_pi16\
19 PMULHRSW – 乘以16位有符号整型，并右移结果。内部指令：_mm_mulhrs_epi16、_mm_mulhrs_pi16\
20 PSHUFB – 从一个128位的参数中选取并乱序其中8位的数据块。内部指令：_mm_shuffle_epi8、_mm_shuffle_pi8\
21 PSIGNB / PSIGNW / PSIGND – 求反（取非）、取零、或保留有符号整型。内部指令：_mm_sign_epi8、_mm_sign_epi16、_mm_sign_epi32、_mm_sign_pi8、_mm_sign_pi16、_mm_sign_pi32\
22 SSE4A\
23 EXTRQ – 从参数中取特定位。内部指令：_mm_extract_si64、_mm_extracti_si64\
24 INSERTQ – 插入特定位到给定参数中。内部指令：_mm_insert_si64、_mm_inserti_si64\
25 MOVNTSD / MOVNTSS – 不使用缓存，直接把数据位写到特定内存位置。内部指令：_mm_stream_sd、_mm_stream_ss\
26 SSE4.1\
27 DPPD / DPPS – 计算两参数的点结果。内部指令：_mm_dp_pd、_mm_dp_ps\
28 EXTRACTPS – 从参数中取出一个特定的32位浮点值。内部指令：_mm_extract_ps\
29 INSERTPS – 把一个32位整型插入到一个128位参数中，并把某些位置零。内部指令：_mm_insert_ps\
30 MOVNTDQA – 从特定内存位置加载128位数据。内部指令：_mm_stream_load_si128\
31 MPSADBW – 计算绝对差分的八个偏移总和。内部指令：_mm_mpsadbw_epu8\
32 PACKUSDW – 使用16位饱和度，把32位有符号整型转换为有符号16位整型。内部指令：_mm_packus_epi32\
33 PBLENDW / BLENDPD / BLENDPS / PBLENDVB / BLENDVPD / BLENDVPS – 把两个不同块大小的参数混合在一起。内部指令：_mm_blend_epi16、_mm_blend_pd、_mm_blend_ps、_mm_blendv_epi8、_mm_blendv_pd、_mm_blendv_ps\
34 PCMPEQQ － 比较64位整型是否相等。内部指令：_mm_cmpeq_epi64\
35 PEXTRB / PEXTRW / PEXTRD / PEXTRQ － 从输入的参数中取出一个整型。内部指令：_mm_extract_epi8、_mm_extract_epi16、_mm_extract_epi32、_mm_extract_epi64\
36 PHMINPOSUW － 选择最小的16位无符号整型并确定它的下标。内部指令：_mm_minpos_epu16\
37 PINSRB / PINSRD / PINSRQ － 把一个整型插入到一个128位参数中。内部指令：_mm_insert_epi8、_mm_insert_epi32、_mm_insert_epi64\
38 PMAXSB / PMAXSD － 接受两个参数中的有符号整型，并选择其中的最大者。内部指令：_mm_max_epi8、_mm_max_epi32\
39 PMAXUW / PMAXUD － 接受两个参数中的无符号整型，并选择其中的最大者。内部指令：_mm_max_epu16、_mm_max_epu32\
40 PMINSB / PMINSD － 接受两个参数中的有符号整型，并选择其中的最小者。内部指令：_mm_min_epi8、_mm_min_epi32\
41 PMINUW / PMINUD － 接受两个参数中的无符号整型，并选择其中的最小者。内部指令：_mm_min_epu16、_mm_min_epu32\
42 PMOVSXBW / PMOVSXBD / PMOVSXBQ / PMOVSXWD / PMOVSXWQ / PMOVSXDQ － 把一有符号整型转换到更大的尺寸。内部指令：_mm_cvtepi8_epi16、_mm_cvtepi8_epi32、_mm_cvtepi8_epi64、_mm_cvtepi16_epi32、_mm_cvtepi16_epi64、_mm_cvtepi32_epi64\
43 PMOVZXBW / PMOVZXBD / PMOVZXBQ / PMOVZXWD / PMOVZXWQ / PMOVZXDQ － 把一无符号整型转换到更大的尺寸。内部指令：_mm_cvtepu8_epi16、_mm_cvtepu8_epi32、_mm_cvtepu8_epi64、_mm_cvtepu16_epi32、_mm_cvtepu16_epi64、_mm_cvtepu32_epi64\
44 PMULDQ － 32位有符号整型相乘，并把结果存储为64位有符号整型。内部指令：_mm_mul_epi32\
45 PMULLUD － 32位有符号整型相乘。内部指令：_mm_mullo_epi32\
46 PTEST － 按位计算两个128位参数，并基于CC标志寄存器的CF与ZF位返回值。内部指令：_mm_testc_si128、_mm_testnzc_si128、_mm_testz_si128\
47 ROUNDPD / ROUNDPS － 取整浮点数值。内部指令：_mm_ceil_pd、_mm_ceil_ps、_mm_floor_pd、_mm_floor_ps、_mm_round_pd、_mm_round_ps\
48 ROUNDSD / ROUNDSS － 结合两个参数，从其一取整到一个浮点数值。内部指令：_mm_ceil_sd、_mm_ceil_ss、_mm_floor_sd、_mm_floor_ss、_mm_round_sd、_mm_round_ss\
49 SSE4.2\
50 CRC32 － 计算参数的CRC-32C检验和。内部指令：_mm_crc32_u8、_mm_crc32_u16、_mm_crc32_u32、_mm_crc32_u64\
51 PCMPESTRI / PCMPESTRM －比较特定长度的两个参数。内部指令：_mm_cmpestra、_mm_cmpestrc、_mm_cmpestri、_mm_cmpestrm、_mm_cmpestro、_mm_cmpestrs、_mm_cmpestrz\
52 PCMPGTQ － 比较两个参数。内部指令：_mm_cmpgt_epi64\
53 PCMPISTRI / PCMPISTRM － 比较两个参数。内部指令：_mm_cmpistra、_mm_cmpistrc、_mm_cmpistri、_mm_cmpistrm、_mm_cmpistro、_mm_cmpistrs、_mm_cmpistrz\
54 POPCNT － 统计位集中1的数量。内部指令：_mm_popcnt_u32、_mm_popcnt_u64、__popcnt16、__popcnt、__popcnt64\
55 高级位操纵\
56 LZCNT － 统计参数中零的数量。内部指令：__lzcnt16、 __lzcnt、__lzcnt64\
57 POPCNT － 统计位集中1的数量。内部指令：_mm_popcnt_u32、_mm_popcnt_u64、__popcnt16、__popcnt、__popcnt64\
58 其他新指令\
59 _InterlockedCompareExchange128 － 对比两个参数。\
60 _mm_castpd_ps / _mm_castpd_si128 / _mm_castps_pd / _mm_castps_si128 / _mm_castsi128_pd / _mm_castsi128_ps － 对32位浮点值（ps）、64位浮点值（pd）及32位整型值（si128）重新解释。\
61 _mm_cvtsd_f64 － 从参数中取出最低的64位浮点值。\
62 _mm_cvtss_f32 － 取出一个32位的浮点值。\
63 _rdtscp － 生成RDTSCP。把TSC AUX[31:0]写到内存，并返回64位时间戳计数器结果。
