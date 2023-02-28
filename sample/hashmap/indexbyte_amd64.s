// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"
#include "funcdata.h"


// handle for lengths < 16
//func IndexByte(bs []byte, c byte) (idx int)
TEXT	·IndexByte(SB), NOSPLIT, $0-40
	MOVQ b_base+0(FP), SI  	// bs.p
	MOVQ b_len+8(FP), BX	// bs.len
					//16(FB)   bs.cap
	MOVB c+24(FP), AL		// c
	LEAQ ret+32(FP), R8		// &idx
	// Shuffle X0 around so that each byte contains
	// the character we're looking for.
	MOVD AX, X0
	PUNPCKLBW X0, X0
	PUNPCKLBW X0, X0
	PSHUFL $0, X0, X0


	MOVOU	(SI), X1 // Load data
	PCMPEQB	X0, X1	// Compare target byte with each byte in data.
	PMOVMSKB X1, DX	// Move result bits to integer register.
	BSFL	DX, DX	// Find first set bit.
	JZ	failure	// No set bit, failure.
	CMPL	DX, BX
	JAE	failure	// Match is past end of data.
	MOVQ	DX, (R8)
	RET

failure:
	MOVQ $-1, (R8)
	RET

// handle for lengths < 16
//func IndexBytes(bs []byte, cs []byte) (idx int)
TEXT	·IndexBytes(SB), NOSPLIT, $0-56
	MOVQ b_base+0(FP), SI  	// bs.p
	MOVQ b_len+8(FP), R13	// bs.len
					//16(FB)   bs.cap
					//24(FB)   cs.p
					//32(FB)   cs.len
					//40(FB)   cs.cap
	LEAQ ret+48(FP), R8		// &idx

	MOVQ c_len+32(FP), R14	// len	
	MOVQ $0, CX			// i
	MOVQ c+24(FP), BX		// p
	// CMPQ 	BX,$0			
	// JA	loop
	// RET
loop:
	CMPQ	CX, R14			// if s.len >= s.cap { return }
	JAE	return
	MOVB 0(BX)(CX*1), AL		// c
	// Shuffle X0 around so that each byte contains
	// the character we're looking for.
	MOVD AX, X0
	PUNPCKLBW X0, X0
	PUNPCKLBW X0, X0
	PSHUFL $0, X0, X0


	MOVOU	(SI), X1 // Load data
	PCMPEQB	X0, X1	// Compare target byte with each byte in data.
	PMOVMSKB X1, DX	// Move result bits to integer register.
	BSFL	DX, DX	// Find first set bit.
	JZ	failure	// No set bit, failure.
	// CMPL	DX, R13
	// JAE	failure	// Match is past end of data.
	
	
	//found 找到目标值
	ADDQ	$1, CX			// CX++ / i++
	JMP loop  // JB loop

return:
	MOVQ	CX, (R8)
	RET

back:
	MOVQ $-2, (R8)
	RET

failure:
	MOVQ CX, (R8)
	RET


// handle for lengths < 16
//func IndexBytes(bs []byte, cs []byte) (idx int)
TEXT	·IndexBytes1(SB), NOSPLIT, $0-56
	MOVQ b_base+0(FP), SI  	// bs.p
	MOVQ b_len+8(FP), R13	// bs.len
					//16(FB)   bs.cap
					//24(FB)   cs.p
					//32(FB)   cs.len
					//40(FB)   cs.cap
	LEAQ ret+48(FP), R8		// &idx

	MOVQ c_len+32(FP), R14	// len	
	MOVQ $0, CX			// i
	MOVQ $0, R12			// i
	MOVQ c+24(FP), BX		// p
	// CMPQ 	BX,$0			
	// JA	loop
	// RET
	MOVOU	(SI), X8 // Load data
loop:
	CMPQ	CX, R14			// if s.len >= s.cap { return }
	JAE	return
	MOVB 0(BX)(CX*1), AL		// c
	// Shuffle X0 around so that each byte contains
	// the character we're looking for.
	MOVD AX, X0
	PUNPCKLBW X0, X0
	PUNPCKLBW X0, X0
	PSHUFL $0, X0, X0

	PCMPEQB	X8, X0	// Compare target byte with each byte in data.
	PMOVMSKB X0, DX	// Move result bits to integer register.
	BSFL	DX, DX	// Find first set bit.
	JZ	failure	// No set bit, failure.
	// CMPL	DX, R13
	// JAE	failure	// Match is past end of data.
	

	//found 找到目标值
	ADDQ	$1, R12		
	ADDQ	$1, CX			// CX++ / i++
	JMP loop  // JB loop

return:
	// MOVQ	CX, (R8)
	MOVQ	R12, (R8)
	RET

back:
	MOVQ $-2, (R8)
	RET

failure:
	ADDQ	$1, CX			// CX++ / i++
	JMP loop  // JB loop

	MOVQ CX, (R8)
	RET



// handle for lengths < 16
//func IndexBytes(bs []byte, cs []byte) (idx int)
TEXT	·IndexBytes2(SB), NOSPLIT, $0-56
	MOVQ b_base+0(FP), SI  	// bs.p
	// MOVQ b_len+8(FP), R13	// bs.len
					//16(FB)   bs.cap
					//24(FB)   cs.p
					//32(FB)   cs.len
					//40(FB)   cs.cap
	LEAQ ret+48(FP), R8		// &idx

	MOVQ c_len+32(FP), R14	// len	
	MOVQ $0, DX			// i
	MOVQ $0, R12			// i
	MOVQ c+24(FP), BX		// p
	//
	LEAQ    ·SpaceBytes(SB), CX
	MOVOU	0(CX),X0
	MOVOU	16(CX),X1
	MOVOU	32(CX),X2
	MOVOU	48(CX),X3
	MOVOU	64(CX),X4
	MOVOU	80(CX),X5
	MOVOU	96(CX),X6
	MOVOU	112(CX),X7

	CMPQ	DX, R14			// if s.len >= s.cap { return }
	JAE	return

loop:
	MOVOU	(BX)(DX*1), X9 

	MOVOU	X9, X8 
	PCMPEQB	X0, X8	// Compare target byte with each byte in data.
	PMOVMSKB X8, CX	// Move result bits to integer register.
	ORQ CX,AX
x1:
	MOVOU	X9, X8 
	PCMPEQB	X1, X8	// Compare target byte with each byte in data.
	PMOVMSKB X8, CX	// Move result bits to integer register.
	ORQ CX,AX
x2:
	MOVOU	X9, X8 
	PCMPEQB	X2, X8	// Compare target byte with each byte in data.
	PMOVMSKB X8, CX	// Move result bits to integer register.
	ORQ CX,AX
x3:
	MOVOU	X9, X8 
	PCMPEQB	X3, X8	// Compare target byte with each byte in data.
	PMOVMSKB X8, CX	// Move result bits to integer register.
	ORQ CX,AX
x4:
	MOVOU	X9, X8 
	PCMPEQB	X4, X8	// Compare target byte with each byte in data.
	PMOVMSKB X8, CX	// Move result bits to integer register.
	ORQ CX,AX
x5:
	MOVOU	X9, X8 
	PCMPEQB	X5, X8	// Compare target byte with each byte in data.
	PMOVMSKB X8, CX	// Move result bits to integer register.
	ORQ CX,AX
x6:
	MOVOU	X9, X8 
	PCMPEQB	X6, X8	// Compare target byte with each byte in data.
	PMOVMSKB X8, CX	// Move result bits to integer register.
	ORQ CX,AX
x7:
	MOVOU	X9, X8 
	PCMPEQB	X7, X8	// Compare target byte with each byte in data.
	PMOVMSKB X8, CX	// Move result bits to integer register.
	ORQ CX,AX

	TESTQ AX,AX 
	JZ next
	JMP loop1
again:
	BTRW CX,AX
	ADDQ $1,R12
loop1:
	BSFW AX,CX
	JNZ	again

next:
	ADDQ	$16, DX			// CX++ / i++

	CMPQ	DX, R14			// if s.len >= s.cap { return }
	JB	loop

return:
	// MOVQ	CX, (R8)
	MOVQ	R12, (R8)
	RET

//func Test1(x, y int) (a, b int)
TEXT	·Test1(SB), NOSPLIT, $0-32
	MOVQ x+0(FP), AX  	// x
	MOVQ y+8(FP), BX  	// y
	MOVQ $0,CX
	
	TESTQ AX,AX 
	JZ return
loop1:
	BSFQ AX,BX
	ADDQ $1,CX
	BTRQ BX,AX
	TESTQ AX,AX
	JNZ loop1


return:
	MOVQ AX,a+16(FP)
	MOVQ CX,b+24(FP)
	RET



//func Test2(i int,xs []byte) (n int)
TEXT	·Test2(SB), NOSPLIT, $0-40
	// NO_LOCAL_POINTERS
	MOVQ 	p+8(FP), AX		// s.ptr

	MOVOU	·SpaceQ(SB),X0

	MOVOU	X0,0(AX)
	
	RET

// func InSpaceQ(b byte) bool
TEXT	·InSpaceQ(SB), NOSPLIT, $0-16
	MOVQ $0,DX
	MOVQ    ·SpaceQ(SB), X1

	MOVB 	p+0(FP), AL		// c
	// Shuffle X0 around so that each byte contains
	// the character we're looking for.
	MOVD AX, X0
	PUNPCKLBW X0, X0
	PUNPCKLBW X0, X0
	PSHUFL $0, X0, X0

	// MOVB 	p+0(FP), X0 
	PCMPEQB	X1, X0	// Compare target byte with each byte in data.
	PMOVMSKB X0, DX	// Move result bits to integer register.
	BSFL	DX, DX	// Find first set bit.
	JZ	reture_false	// No set bit, failure.
	
	// MOVB ·TrueB(SB), AX
	// MOVB AL, p+1(FP)
	MOVL $1,p+8(FP)
	RET
	
reture_false:
	// MOVB ·FalseB(SB), AX
	// MOVB AX, p+1(FP)
	MOVL $0,p+8(FP)
	RET
	PAND	X1, X0
	PSLLW X2, X11

TEXT	·InSpaceQ2(SB), NOSPLIT, $0-16


	RETFL



// handle for lengths < 16
//func IndexBytes(bs []byte, cs []byte) (idx int)
TEXT	·Hash(SB), NOSPLIT, $0-56
	MOVQ b_base+0(FP), SI  	// bs.p
	MOVQ b_len+8(FP), R13	// bs.len
					//16(FB)   bs.cap
	MOVQ c_len+24(FP), DI	// cs.p	
	// MOVQ c_len+32(FP), R14	// cs.len	
					//40(FB)   cs.cap
	LEAQ ret+48(FP), R8		// &idx

	MOVQ $0, CX				// i

loop:
	MOVOU	0(SI)(CX*1), X0
	MOVOU	0(DI)(CX*1), X1

	ADDQ	$16, CX

	PAND	X1, X0
	// MOVHPS  X0, BX
	// MOVLPS  X0, DX
	MOVQ  X0, BX
	MOVHLPS X0,X3
	MOVQ  X3, DX
	// MOVLPS  X0, DX

	// if l <n
	CMPQ R13,CX
	JLT small

h_bit:
	BSFL	BX, BX	// Find first set bit.
	JZ	l_bit	// No set bit, failure.

	JMP h_bit
l_bit:
	BSFL	DX, DX	// Find first set bit.
	JZ	loop	// No set bit, failure.

	JMP l_bit

small:
h_bit1:
	BSFL	BX, BX	// Find first set bit.
	JZ	l_bit1	// No set bit, failure.

	JMP h_bit1
l_bit1:
	BSFL	DX, DX	// Find first set bit.
	JZ	return	// No set bit, failure.

	JMP l_bit1

return:
	MOVQ	CX, (R8)
	RET

back:
	MOVQ $-2, (R8)
	RET

failure:
	MOVQ CX, (R8)
	RET


// handle for lengths < 16
/*
struct s{
	pick   [1024]byte
	mask   [1024]byte
	shiftL [128]int64
}
 加速宣告失败： 	执行代码 1ns； 函数调用 5ns
*/
//func Hashx(bs []byte, cs []s) (idx int)
TEXT	·Hashx(SB), NOSPLIT, $0-64
	MOVQ b_base+0(FP), SI  	// bs.p
	MOVQ b_len+8(FP), R13	// bs.len
					//16(FB)   bs.cap
	MOVQ c_base+24(FP), DI	// cs.p	
	// MOVQ c_len+32(FP), R14	// cs.len	
					//40(FB)   cs.cap
	LEAQ ret+48(FP), R8		// &idx

	PXOR 	X4, X4
	LEAQ    ·Pick08(SB), CX
	MOVOU 	(CX), X5
	MOVQ 	$0, BX				// i

loop:
	MOVOU	0(SI)(BX*1), X0
	MOVOU	0(DI)(BX*1), X1  // pick
	MOVOU	8(DI)(BX*1), X2  // mask
	MOVQ	16(DI)(BX*1), CX // shiftL
	MOVQ	$16, BX

	PSHUFB 	X0, X1
	PAND	X0, X2
	PSADBW	X0, X4  // Sum, 高 64bit 和低 64bit 分开处理
	PSHUFB	X0, X5  // 合并高低 64bit
	MOVQ	X0, AX
	SHLQ	CX, AX
	ADDQ	AX,(R8)

	// if l <n
	CMPQ R13,BX
	JLT small
	JMP	loop	// No set bit, failure.

small:

return:
	RET

