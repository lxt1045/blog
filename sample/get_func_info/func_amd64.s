// MIT License
//
// Copyright (c) 2021 Xiantu Li
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.


//go:build (amd64 || amd64p32) && gc && go1.5
// +build amd64 amd64p32
// +build gc
// +build go1.5

#include "go_asm.h"
#include "textflag.h"
#include "funcdata.h"


// func GetStack(s []uintptr) int
TEXT ·GetStack(SB), NOSPLIT, $24-8
	// GO_ARGS
	// GO_RESULTS_INITIALIZED
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
	// SUBQ	$1,DX
	// JZ 	return

	MOVQ	+0(BP), BP 		// last BP; 展开调用栈至上一层
	CMPQ	BP, $0 			// if (BP) <= 0 { return }
	JA loop					// 无符号大于就跳转

return:
	MOVQ	CX,n+24(FP) 	// ret n
	RET


