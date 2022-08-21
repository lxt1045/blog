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



// func Ret1(p uintptr)
TEXT ·Retxx(SB), NOSPLIT, $0-16
	CMPQ	p+0(FP), $0
	JHI	unwind
	RET
unwind:
	MOVQ	BP, SP
	ADDQ	$16,SP
	MOVQ	+0(BP), BP
	//CALL runtime·deferreturn(SB)
	JMP	-8(SP)


//try 级的 derfer ok
TEXT ·Ret(SB), NOSPLIT, $0-16
	CMPQ	p+0(FP), $0
	JHI	unwind
	RET
unwind:
	ADDQ	$8,SP  // Call Retx 后需要回退 return address 的 SP 地址（自动入栈的）
	CALL	runtime·deferreturn(SB)
	MOVQ	BP,SP
	ADDQ	$8,SP
	RET
	// ADDQ	$8,SP	//同 RET， JMP 不能想 RET 自动 pop return addr，所以要手动 ADDQ
	// JMP	8(BP)
