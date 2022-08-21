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



// func getg() unsafe.Pointer
TEXT ·Getg(SB), NOSPLIT, $0-8
    MOVQ (TLS), AX
	ADDQ ·g_goid_offset(SB),AX
    MOVQ (AX), BX
    MOVQ BX, ret+0(FP)
    RET

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

    // convert (*g) to interface{}
    // MOVQ AX, 8(SP)
    // MOVQ BX, 0(SP)
    // CALL runtime·convT2E(SB)
    // MOVQ 16(SP), AX
    // MOVQ 24(SP), BX

    // return interface{}
    MOVQ BX, ret_type+0(FP)
    MOVQ AX, ret_data+8(FP)
    RET
