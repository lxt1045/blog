"".main STEXT size=571 args=0x0 locals=0x108 funcid=0x0 align=0x0
	0x0000 00000 (main.go:10)	TEXT	"".main(SB), ABIInternal, $264-0
	0x0000 00000 (main.go:10)	LEAQ	-136(SP), R12
	0x0008 00008 (main.go:10)	CMPQ	R12, 16(R14)
	0x000c 00012 (main.go:10)	PCDATA	$0, $-2
	0x000c 00012 (main.go:10)	JLS	561
	0x0012 00018 (main.go:10)	PCDATA	$0, $-1
	0x0012 00018 (main.go:10)	SUBQ	$264, SP
	0x0019 00025 (main.go:10)	MOVQ	BP, 256(SP)
	0x0021 00033 (main.go:10)	LEAQ	256(SP), BP
	0x0029 00041 (main.go:10)	FUNCDATA	$0, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0029 00041 (main.go:10)	FUNCDATA	$1, gclocals·95286601ae45d090b3c9e9ccf89a5648(SB)
	0x0029 00041 (main.go:10)	FUNCDATA	$2, "".main.stkobj(SB)
	0x0029 00041 (main.go:11)	MOVQ	$1, "".a+72(SP)
	0x0032 00050 (main.go:12)	MOVQ	"".XXX(SB), CX
	0x0039 00057 (main.go:12)	TESTQ	CX, CX
	0x003c 00060 (main.go:12)	JGE	69
	0x003e 00062 (main.go:12)	NOP
	0x0040 00064 (main.go:12)	JMP	555
	0x0045 00069 (main.go:12)	CMPQ	CX, $64
	0x0049 00073 (main.go:12)	SBBQ	DX, DX
	0x004c 00076 (main.go:12)	MOVL	$1, SI
	0x0051 00081 (main.go:12)	SHLQ	CX, SI
	0x0054 00084 (main.go:12)	ANDQ	DX, SI
	0x0057 00087 (main.go:12)	MOVQ	SI, "".shlx+56(SP)
	0x005c 00092 (main.go:13)	LEAQ	go.string."qq"(SB), DX
	0x0063 00099 (main.go:13)	MOVQ	DX, "".x+120(SP)
	0x0068 00104 (main.go:13)	MOVQ	$2, "".x+128(SP)
	0x0074 00116 (main.go:15)	MOVUPS	X15, "".i+136(SP)
	0x007d 00125 (main.go:17)	MOVQ	"".x+120(SP), AX
	0x0082 00130 (main.go:17)	MOVQ	"".x+128(SP), BX
	0x008a 00138 (main.go:17)	PCDATA	$1, $0
	0x008a 00138 (main.go:17)	CALL	runtime.convTstring(SB)
	0x008f 00143 (main.go:17)	MOVQ	AX, ""..autotmp_12+112(SP)
	0x0094 00148 (main.go:17)	LEAQ	type.string(SB), DX
	0x009b 00155 (main.go:17)	MOVQ	DX, "".i+136(SP)
	0x00a3 00163 (main.go:17)	MOVQ	AX, "".i+144(SP)
	0x00ab 00171 (main.go:18)	MOVQ	"".shlx+56(SP), AX
	0x00b0 00176 (main.go:18)	CALL	runtime.convT64(SB)
	0x00b5 00181 (main.go:18)	MOVQ	AX, ""..autotmp_13+104(SP)
	0x00ba 00186 (main.go:18)	LEAQ	type.int(SB), DX
	0x00c1 00193 (main.go:18)	MOVQ	DX, "".i+136(SP)
	0x00c9 00201 (main.go:18)	MOVQ	AX, "".i+144(SP)
	0x00d1 00209 (main.go:20)	LEAQ	go.string."%+v"(SB), DX
	0x00d8 00216 (main.go:20)	MOVQ	DX, fmt.format+152(SP)
	0x00e0 00224 (main.go:20)	MOVQ	$3, fmt.format+160(SP)
	0x00ec 00236 (main.go:20)	MOVUPS	X15, ""..autotmp_15+184(SP)
	0x00f5 00245 (main.go:20)	LEAQ	""..autotmp_15+184(SP), DX
	0x00fd 00253 (main.go:20)	MOVQ	DX, ""..autotmp_14+96(SP)
	0x0102 00258 (main.go:20)	TESTB	AL, (DX)
	0x0104 00260 (main.go:20)	MOVQ	"".i+144(SP), SI
	0x010c 00268 (main.go:20)	MOVQ	"".i+136(SP), DI
	0x0114 00276 (main.go:20)	MOVQ	DI, ""..autotmp_15+184(SP)
	0x011c 00284 (main.go:20)	MOVQ	SI, ""..autotmp_15+192(SP)
	0x0124 00292 (main.go:20)	TESTB	AL, (DX)
	0x0126 00294 (main.go:20)	JMP	296
	0x0128 00296 (main.go:20)	MOVQ	DX, fmt.a+232(SP)
	0x0130 00304 (main.go:20)	MOVQ	$1, fmt.a+240(SP)
	0x013c 00316 (main.go:20)	MOVQ	$1, fmt.a+248(SP)
	0x0148 00328 (main.go:20)	MOVQ	$0, fmt.n+64(SP)
	0x0151 00337 (main.go:20)	MOVUPS	X15, fmt.err+168(SP)
	0x015a 00346 (<unknown line number>)	NOP
	0x015a 00346 ($GOROOT/src/fmt/print.go:213)	MOVQ	$0, fmt..autotmp_0+88(SP)
	0x0163 00355 ($GOROOT/src/fmt/print.go:213)	MOVUPS	X15, fmt..autotmp_1+216(SP)
	0x016c 00364 ($GOROOT/src/fmt/print.go:213)	MOVUPS	X15, ""..autotmp_11+200(SP)
	0x0175 00373 ($GOROOT/src/fmt/print.go:213)	MOVQ	fmt.format+152(SP), CX
	0x017d 00381 ($GOROOT/src/fmt/print.go:213)	MOVQ	fmt.a+232(SP), SI
	0x0185 00389 ($GOROOT/src/fmt/print.go:213)	MOVQ	fmt.format+160(SP), DI
	0x018d 00397 ($GOROOT/src/fmt/print.go:213)	MOVQ	fmt.a+240(SP), R8
	0x0195 00405 ($GOROOT/src/fmt/print.go:213)	MOVQ	fmt.a+248(SP), R9
	0x019d 00413 ($GOROOT/src/fmt/print.go:213)	MOVQ	os.Stdout(SB), BX
	0x01a4 00420 ($GOROOT/src/fmt/print.go:213)	LEAQ	go.itab.*os.File,io.Writer(SB), AX
	0x01ab 00427 ($GOROOT/src/fmt/print.go:213)	CALL	fmt.Fprintf(SB)
	0x01b0 00432 ($GOROOT/src/fmt/print.go:213)	MOVQ	AX, ""..autotmp_10+80(SP)
	0x01b5 00437 ($GOROOT/src/fmt/print.go:213)	MOVQ	BX, ""..autotmp_11+200(SP)
	0x01bd 00445 ($GOROOT/src/fmt/print.go:213)	MOVQ	CX, ""..autotmp_11+208(SP)
	0x01c5 00453 ($GOROOT/src/fmt/print.go:213)	MOVQ	""..autotmp_10+80(SP), DX
	0x01ca 00458 ($GOROOT/src/fmt/print.go:213)	MOVQ	DX, fmt..autotmp_0+88(SP)
	0x01cf 00463 ($GOROOT/src/fmt/print.go:213)	MOVQ	""..autotmp_11+200(SP), DX
	0x01d7 00471 ($GOROOT/src/fmt/print.go:213)	MOVQ	""..autotmp_11+208(SP), R10
	0x01df 00479 ($GOROOT/src/fmt/print.go:213)	MOVQ	DX, fmt..autotmp_1+216(SP)
	0x01e7 00487 ($GOROOT/src/fmt/print.go:213)	MOVQ	R10, fmt..autotmp_1+224(SP)
	0x01ef 00495 (main.go:20)	MOVQ	fmt..autotmp_0+88(SP), DX
	0x01f4 00500 (main.go:20)	MOVQ	DX, fmt.n+64(SP)
	0x01f9 00505 (main.go:20)	MOVQ	fmt..autotmp_1+216(SP), DX
	0x0201 00513 (main.go:20)	MOVQ	fmt..autotmp_1+224(SP), R10
	0x0209 00521 (main.go:20)	MOVQ	DX, fmt.err+168(SP)
	0x0211 00529 (main.go:20)	MOVQ	R10, fmt.err+176(SP)
	0x0219 00537 (main.go:20)	JMP	539
	0x021b 00539 (main.go:21)	PCDATA	$1, $-1
	0x021b 00539 (main.go:21)	MOVQ	256(SP), BP
	0x0223 00547 (main.go:21)	ADDQ	$264, SP
	0x022a 00554 (main.go:21)	RET
	0x022b 00555 (main.go:12)	PCDATA	$1, $0
	0x022b 00555 (main.go:12)	CALL	runtime.panicshift(SB)
	0x0230 00560 (main.go:12)	XCHGL	AX, AX
	0x0231 00561 (main.go:12)	NOP
	0x0231 00561 (main.go:10)	PCDATA	$1, $-1
	0x0231 00561 (main.go:10)	PCDATA	$0, $-2
	0x0231 00561 (main.go:10)	CALL	runtime.morestack_noctxt(SB)
	0x0236 00566 (main.go:10)	PCDATA	$0, $-1
	0x0236 00566 (main.go:10)	JMP	0
	0x0000 4c 8d a4 24 78 ff ff ff 4d 3b 66 10 0f 86 1f 02  L..$x...M;f.....
	0x0010 00 00 48 81 ec 08 01 00 00 48 89 ac 24 00 01 00  ..H......H..$...
	0x0020 00 48 8d ac 24 00 01 00 00 48 c7 44 24 48 01 00  .H..$....H.D$H..
	0x0030 00 00 48 8b 0d 00 00 00 00 48 85 c9 7d 07 66 90  ..H......H..}.f.
	0x0040 e9 e6 01 00 00 48 83 f9 40 48 19 d2 be 01 00 00  .....H..@H......
	0x0050 00 48 d3 e6 48 21 d6 48 89 74 24 38 48 8d 15 00  .H..H!.H.t$8H...
	0x0060 00 00 00 48 89 54 24 78 48 c7 84 24 80 00 00 00  ...H.T$xH..$....
	0x0070 02 00 00 00 44 0f 11 bc 24 88 00 00 00 48 8b 44  ....D...$....H.D
	0x0080 24 78 48 8b 9c 24 80 00 00 00 e8 00 00 00 00 48  $xH..$.........H
	0x0090 89 44 24 70 48 8d 15 00 00 00 00 48 89 94 24 88  .D$pH......H..$.
	0x00a0 00 00 00 48 89 84 24 90 00 00 00 48 8b 44 24 38  ...H..$....H.D$8
	0x00b0 e8 00 00 00 00 48 89 44 24 68 48 8d 15 00 00 00  .....H.D$hH.....
	0x00c0 00 48 89 94 24 88 00 00 00 48 89 84 24 90 00 00  .H..$....H..$...
	0x00d0 00 48 8d 15 00 00 00 00 48 89 94 24 98 00 00 00  .H......H..$....
	0x00e0 48 c7 84 24 a0 00 00 00 03 00 00 00 44 0f 11 bc  H..$........D...
	0x00f0 24 b8 00 00 00 48 8d 94 24 b8 00 00 00 48 89 54  $....H..$....H.T
	0x0100 24 60 84 02 48 8b b4 24 90 00 00 00 48 8b bc 24  $`..H..$....H..$
	0x0110 88 00 00 00 48 89 bc 24 b8 00 00 00 48 89 b4 24  ....H..$....H..$
	0x0120 c0 00 00 00 84 02 eb 00 48 89 94 24 e8 00 00 00  ........H..$....
	0x0130 48 c7 84 24 f0 00 00 00 01 00 00 00 48 c7 84 24  H..$........H..$
	0x0140 f8 00 00 00 01 00 00 00 48 c7 44 24 40 00 00 00  ........H.D$@...
	0x0150 00 44 0f 11 bc 24 a8 00 00 00 48 c7 44 24 58 00  .D...$....H.D$X.
	0x0160 00 00 00 44 0f 11 bc 24 d8 00 00 00 44 0f 11 bc  ...D...$....D...
	0x0170 24 c8 00 00 00 48 8b 8c 24 98 00 00 00 48 8b b4  $....H..$....H..
	0x0180 24 e8 00 00 00 48 8b bc 24 a0 00 00 00 4c 8b 84  $....H..$....L..
	0x0190 24 f0 00 00 00 4c 8b 8c 24 f8 00 00 00 48 8b 1d  $....L..$....H..
	0x01a0 00 00 00 00 48 8d 05 00 00 00 00 e8 00 00 00 00  ....H...........
	0x01b0 48 89 44 24 50 48 89 9c 24 c8 00 00 00 48 89 8c  H.D$PH..$....H..
	0x01c0 24 d0 00 00 00 48 8b 54 24 50 48 89 54 24 58 48  $....H.T$PH.T$XH
	0x01d0 8b 94 24 c8 00 00 00 4c 8b 94 24 d0 00 00 00 48  ..$....L..$....H
	0x01e0 89 94 24 d8 00 00 00 4c 89 94 24 e0 00 00 00 48  ..$....L..$....H
	0x01f0 8b 54 24 58 48 89 54 24 40 48 8b 94 24 d8 00 00  .T$XH.T$@H..$...
	0x0200 00 4c 8b 94 24 e0 00 00 00 48 89 94 24 a8 00 00  .L..$....H..$...
	0x0210 00 4c 89 94 24 b0 00 00 00 eb 00 48 8b ac 24 00  .L..$......H..$.
	0x0220 01 00 00 48 81 c4 08 01 00 00 c3 e8 00 00 00 00  ...H............
	0x0230 90 e8 00 00 00 00 e9 c5 fd ff ff                 ...........
	rel 3+0 t=23 type.string+0
	rel 3+0 t=23 type.int+0
	rel 3+0 t=23 type.*os.File+0
	rel 53+4 t=14 "".XXX+0
	rel 95+4 t=14 go.string."qq"+0
	rel 139+4 t=7 runtime.convTstring+0
	rel 151+4 t=14 type.string+0
	rel 177+4 t=7 runtime.convT64+0
	rel 189+4 t=14 type.int+0
	rel 212+4 t=14 go.string."%+v"+0
	rel 416+4 t=14 os.Stdout+0
	rel 423+4 t=14 go.itab.*os.File,io.Writer+0
	rel 428+4 t=7 fmt.Fprintf+0
	rel 556+4 t=7 runtime.panicshift+0
	rel 562+4 t=7 runtime.morestack_noctxt+0
"".toEface STEXT nosplit size=1 args=0x0 locals=0x0 funcid=0x0 align=0x0
	0x0000 00000 (main.go:23)	TEXT	"".toEface(SB), NOSPLIT|ABIInternal, $0-0
	0x0000 00000 (main.go:23)	FUNCDATA	$0, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (main.go:23)	FUNCDATA	$1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (main.go:25)	RET
	0x0000 c3                                               .
"".init STEXT nosplit size=1 args=0x0 locals=0x0 funcid=0x0 align=0x0
	0x0000 00000 (main.go:8)	TEXT	"".init(SB), NOSPLIT|ABIInternal, $0-0
	0x0000 00000 (main.go:8)	FUNCDATA	$0, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (main.go:8)	FUNCDATA	$1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (main.go:8)	RET
	0x0000 c3                                               .
go.cuinfo.packagename. SDWARFCUINFO dupok size=0
	0x0000 6d 61 69 6e                                      main
go.info.fmt.Printf$abstract SDWARFABSFCN dupok size=54
	0x0000 05 66 6d 74 2e 50 72 69 6e 74 66 00 01 01 13 66  .fmt.Printf....f
	0x0010 6f 72 6d 61 74 00 00 00 00 00 00 13 61 00 00 00  ormat.......a...
	0x0020 00 00 00 13 6e 00 01 00 00 00 00 13 65 72 72 00  ....n.......err.
	0x0030 01 00 00 00 00 00                                ......
	rel 0+0 t=22 type.[]interface {}+0
	rel 0+0 t=22 type.error+0
	rel 0+0 t=22 type.int+0
	rel 0+0 t=22 type.string+0
	rel 23+4 t=31 go.info.string+0
	rel 31+4 t=31 go.info.[]interface {}+0
	rel 39+4 t=31 go.info.int+0
	rel 49+4 t=31 go.info.error+0
""..inittask SNOPTRDATA size=32
	0x0000 00 00 00 00 00 00 00 00 01 00 00 00 00 00 00 00  ................
	0x0010 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	rel 24+8 t=1 fmt..inittask+0
go.string."qq" SRODATA dupok size=2
	0x0000 71 71                                            qq
go.string."%+v" SRODATA dupok size=3
	0x0000 25 2b 76                                         %+v
go.itab.*os.File,io.Writer SRODATA dupok size=32
	0x0000 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0010 44 b5 f3 33 00 00 00 00 00 00 00 00 00 00 00 00  D..3............
	rel 0+8 t=1 type.io.Writer+0
	rel 8+8 t=1 type.*os.File+0
	rel 24+8 t=-32767 os.(*File).Write+0
"".XXX SNOPTRDATA size=8
	0x0000 08 00 00 00 00 00 00 00                          ........
runtime.nilinterequal·f SRODATA dupok size=8
	0x0000 00 00 00 00 00 00 00 00                          ........
	rel 0+8 t=1 runtime.nilinterequal+0
runtime.memequal64·f SRODATA dupok size=8
	0x0000 00 00 00 00 00 00 00 00                          ........
	rel 0+8 t=1 runtime.memequal64+0
runtime.gcbits.01 SRODATA dupok size=1
	0x0000 01                                               .
type..namedata.*interface {}- SRODATA dupok size=15
	0x0000 00 0d 2a 69 6e 74 65 72 66 61 63 65 20 7b 7d     ..*interface {}
type.*interface {} SRODATA dupok size=56
	0x0000 08 00 00 00 00 00 00 00 08 00 00 00 00 00 00 00  ................
	0x0010 4f 0f 96 9d 08 08 08 36 00 00 00 00 00 00 00 00  O......6........
	0x0020 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0030 00 00 00 00 00 00 00 00                          ........
	rel 24+8 t=1 runtime.memequal64·f+0
	rel 32+8 t=1 runtime.gcbits.01+0
	rel 40+4 t=5 type..namedata.*interface {}-+0
	rel 48+8 t=1 type.interface {}+0
runtime.gcbits.02 SRODATA dupok size=1
	0x0000 02                                               .
type.interface {} SRODATA dupok size=80
	0x0000 10 00 00 00 00 00 00 00 10 00 00 00 00 00 00 00  ................
	0x0010 e7 57 a0 18 02 08 08 14 00 00 00 00 00 00 00 00  .W..............
	0x0020 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0030 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0040 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	rel 24+8 t=1 runtime.nilinterequal·f+0
	rel 32+8 t=1 runtime.gcbits.02+0
	rel 40+4 t=5 type..namedata.*interface {}-+0
	rel 44+4 t=-32763 type.*interface {}+0
	rel 56+8 t=1 type.interface {}+80
type..namedata.*[]interface {}- SRODATA dupok size=17
	0x0000 00 0f 2a 5b 5d 69 6e 74 65 72 66 61 63 65 20 7b  ..*[]interface {
	0x0010 7d                                               }
type.*[]interface {} SRODATA dupok size=56
	0x0000 08 00 00 00 00 00 00 00 08 00 00 00 00 00 00 00  ................
	0x0010 f3 04 9a e7 08 08 08 36 00 00 00 00 00 00 00 00  .......6........
	0x0020 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0030 00 00 00 00 00 00 00 00                          ........
	rel 24+8 t=1 runtime.memequal64·f+0
	rel 32+8 t=1 runtime.gcbits.01+0
	rel 40+4 t=5 type..namedata.*[]interface {}-+0
	rel 48+8 t=1 type.[]interface {}+0
type.[]interface {} SRODATA dupok size=56
	0x0000 18 00 00 00 00 00 00 00 08 00 00 00 00 00 00 00  ................
	0x0010 70 93 ea 2f 02 08 08 17 00 00 00 00 00 00 00 00  p../............
	0x0020 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0030 00 00 00 00 00 00 00 00                          ........
	rel 32+8 t=1 runtime.gcbits.01+0
	rel 40+4 t=5 type..namedata.*[]interface {}-+0
	rel 44+4 t=-32763 type.*[]interface {}+0
	rel 48+8 t=1 type.interface {}+0
type..importpath.fmt. SRODATA dupok size=5
	0x0000 00 03 66 6d 74                                   ..fmt
gclocals·33cdeccccebe80329f1fdbee7f5874cb SRODATA dupok size=8
	0x0000 01 00 00 00 00 00 00 00                          ........
gclocals·95286601ae45d090b3c9e9ccf89a5648 SRODATA dupok size=11
	0x0000 01 00 00 00 14 00 00 00 00 00 00                 ...........
"".main.stkobj SRODATA static size=24
	0x0000 01 00 00 00 00 00 00 00 b8 ff ff ff 10 00 00 00  ................
	0x0010 10 00 00 00 00 00 00 00                          ........
	rel 20+4 t=5 runtime.gcbits.02+0
