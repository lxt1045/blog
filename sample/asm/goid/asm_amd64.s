// MIT License
//
// Copyright (c) 2023 Xiantu Li
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


GLOBL ·runtime_g_type(SB),NOPTR,$8
DATA ·runtime_g_type+0(SB)/8,$type·runtime·g(SB) // # x86

/*
type eface struct {
	_type uint64
	data  unsafe.Pointer
}
*/
GLOBL ·gIface(SB),NOPTR,$16
DATA ·gIface+0(SB)/8,$type·runtime·g(SB) // interface{}.type

GLOBL ·getgg(SB),NOPTR,$8
// DATA ·getgg+0(SB)/8,$syscall·Syscall(SB) 
DATA ·getgg+0(SB)/8,$runtime·add(SB)

// func getg() unsafe.Pointer
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


DATA ·Id+0(SB)/1,$0x37
DATA ·Id+1(SB)/1,$0x25
DATA ·Id+2(SB)/1,$0x00
DATA ·Id+3(SB)/1,$0x00
DATA ·Id+4(SB)/1,$0x00
DATA ·Id+5(SB)/1,$0x00
DATA ·Id+6(SB)/1,$0x00
DATA ·Id+7(SB)/1,$0x00
GLOBL ·Id(SB),NOPTR,$8

/*
Name 的实际存储是
struct{
    Data uintptr    // &self.str
    Len int         // 6
    str [6]byte     // "gopher"
}
*/
DATA ·Name+0(SB)/8,$·Name+16(SB)    // StringHeader.Data
DATA ·Name+8(SB)/8,$6               // StringHeader.Len
DATA ·Name+16(SB)/8,$"gopher"       // [6]byte{'g','o','p','h','e','r'}
GLOBL ·Name(SB),NOPTR,$24           // struct{Data uintptr, Len int, str [6]byte}

DATA str<>+0(SB)/8,$"Hello Wo"      // str[0:8]={'H','e','l','l','o',' ','W','o'}
DATA str<>+8(SB)/8,$"rld!"          // str[9:12]={'r','l','d','!''}
GLOBL str<>(SB),NOPTR,$16           // 定义全局数组 var str<> [16]byte
DATA ·Helloworld+0(SB)/8,$str<>(SB) // StringHeader.Data = &str<>
DATA ·Helloworld+8(SB)/8,$12        // StringHeader.Len = 12
GLOBL ·Helloworld(SB),NOPTR,$16     // struct{Data uintptr, Len int}


TEXT ·Print1(SB), NOSPLIT, $48-16   
  LEAQ strp+0(FP),AX
  MOVQ AX, 0(SP)        // []interface{} slice 的 pointer
  MOVQ $1, BX    
  MOVQ BX, 8(SP)        // slice 的 len
  MOVQ BX, 16(SP)       // slice 的 cap
  CALL fmt·Println(SB)    
  RET

// TEXT ·Print<ABIInternal>(SB), NOSPLIT, $48-16   
//   LEAQ strp+0(FP),AX
//   MOVQ AX, 0(SP)        // []interface{} slice 的 pointer
//   MOVQ $1, BX    
//   MOVQ BX, 8(SP)        // slice 的 len
//   MOVQ BX, 16(SP)       // slice 的 cap
//   CALL fmt·Println(SB)    
//   RET

