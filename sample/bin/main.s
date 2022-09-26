"".main STEXT size=453 args=0x0 locals=0xf0 funcid=0x0 align=0x0
	0x0000 00000 (main.go:8)	TEXT	"".main(SB), ABIInternal, $240-0
	0x0000 00000 (main.go:8)	LEAQ	-112(SP), R12
	0x0005 00005 (main.go:8)	CMPQ	R12, 16(R14)
	0x0009 00009 (main.go:8)	PCDATA	$0, $-2
	0x0009 00009 (main.go:8)	JLS	442
	0x000f 00015 (main.go:8)	PCDATA	$0, $-1
	0x000f 00015 (main.go:8)	SUBQ	$240, SP
	0x0016 00022 (main.go:8)	MOVQ	BP, 232(SP)
	0x001e 00030 (main.go:8)	LEAQ	232(SP), BP
	0x0026 00038 (main.go:8)	FUNCDATA	$0, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0026 00038 (main.go:8)	FUNCDATA	$1, gclocals·abcd98c9569535db9f3522d983659dbd(SB)
	0x0026 00038 (main.go:8)	FUNCDATA	$2, "".main.stkobj(SB)
	0x0026 00038 (main.go:9)	LEAQ	go.string."qq"(SB), CX
	0x002d 00045 (main.go:9)	MOVQ	CX, "".x+96(SP)
	0x0032 00050 (main.go:9)	MOVQ	$2, "".x+104(SP)
	0x003b 00059 (main.go:11)	MOVUPS	X15, "".i+112(SP)
	0x0041 00065 (main.go:13)	MOVQ	"".x+96(SP), AX
	0x0046 00070 (main.go:13)	MOVQ	"".x+104(SP), BX
	0x004b 00075 (main.go:13)	PCDATA	$1, $0
	0x004b 00075 (main.go:13)	CALL	runtime.convTstring(SB)
	0x0050 00080 (main.go:13)	MOVQ	AX, ""..autotmp_10+88(SP)
	0x0055 00085 (main.go:13)	LEAQ	type.string(SB), CX
	0x005c 00092 (main.go:13)	MOVQ	CX, "".i+112(SP)
	0x0061 00097 (main.go:13)	MOVQ	AX, "".i+120(SP)
	0x0066 00102 (main.go:15)	LEAQ	go.string."%+v"(SB), CX
	0x006d 00109 (main.go:15)	MOVQ	CX, fmt.format+128(SP)
	0x0075 00117 (main.go:15)	MOVQ	$3, fmt.format+136(SP)
	0x0081 00129 (main.go:15)	MOVUPS	X15, ""..autotmp_12+176(SP)
	0x008a 00138 (main.go:15)	LEAQ	""..autotmp_12+176(SP), CX
	0x0092 00146 (main.go:15)	MOVQ	CX, ""..autotmp_11+80(SP)
	0x0097 00151 (main.go:15)	TESTB	AL, (CX)
	0x0099 00153 (main.go:15)	MOVQ	"".i+120(SP), DX
	0x009e 00158 (main.go:15)	MOVQ	"".i+112(SP), SI
	0x00a3 00163 (main.go:15)	MOVQ	SI, ""..autotmp_12+176(SP)
	0x00ab 00171 (main.go:15)	MOVQ	DX, ""..autotmp_12+184(SP)
	0x00b3 00179 (main.go:15)	TESTB	AL, (CX)
	0x00b5 00181 (main.go:15)	JMP	183
	0x00b7 00183 (main.go:15)	MOVQ	CX, fmt.a+208(SP)
	0x00bf 00191 (main.go:15)	MOVQ	$1, fmt.a+216(SP)
	0x00cb 00203 (main.go:15)	MOVQ	$1, fmt.a+224(SP)
	0x00d7 00215 (main.go:15)	MOVQ	$0, fmt.n+56(SP)
	0x00e0 00224 (main.go:15)	MOVUPS	X15, fmt.err+144(SP)
	0x00e9 00233 (<unknown line number>)	NOP
	0x00e9 00233 ($GOROOT/src/fmt/print.go:213)	MOVQ	$0, fmt..autotmp_0+72(SP)
	0x00f2 00242 ($GOROOT/src/fmt/print.go:213)	MOVUPS	X15, fmt..autotmp_1+192(SP)
	0x00fb 00251 ($GOROOT/src/fmt/print.go:213)	MOVUPS	X15, ""..autotmp_9+160(SP)
	0x0104 00260 ($GOROOT/src/fmt/print.go:213)	MOVQ	fmt.format+128(SP), CX
	0x010c 00268 ($GOROOT/src/fmt/print.go:213)	MOVQ	fmt.a+208(SP), SI
	0x0114 00276 ($GOROOT/src/fmt/print.go:213)	MOVQ	fmt.format+136(SP), DI
	0x011c 00284 ($GOROOT/src/fmt/print.go:213)	MOVQ	fmt.a+216(SP), R8
	0x0124 00292 ($GOROOT/src/fmt/print.go:213)	MOVQ	fmt.a+224(SP), R9
	0x012c 00300 ($GOROOT/src/fmt/print.go:213)	MOVQ	os.Stdout(SB), BX
	0x0133 00307 ($GOROOT/src/fmt/print.go:213)	LEAQ	go.itab.*os.File,io.Writer(SB), AX
	0x013a 00314 ($GOROOT/src/fmt/print.go:213)	CALL	fmt.Fprintf(SB)
	0x013f 00319 ($GOROOT/src/fmt/print.go:213)	MOVQ	AX, ""..autotmp_8+64(SP)
	0x0144 00324 ($GOROOT/src/fmt/print.go:213)	MOVQ	BX, ""..autotmp_9+160(SP)
	0x014c 00332 ($GOROOT/src/fmt/print.go:213)	MOVQ	CX, ""..autotmp_9+168(SP)
	0x0154 00340 ($GOROOT/src/fmt/print.go:213)	MOVQ	""..autotmp_8+64(SP), DX
	0x0159 00345 ($GOROOT/src/fmt/print.go:213)	MOVQ	DX, fmt..autotmp_0+72(SP)
	0x015e 00350 ($GOROOT/src/fmt/print.go:213)	MOVQ	""..autotmp_9+160(SP), DX
	0x0166 00358 ($GOROOT/src/fmt/print.go:213)	MOVQ	""..autotmp_9+168(SP), R10
	0x016e 00366 ($GOROOT/src/fmt/print.go:213)	MOVQ	DX, fmt..autotmp_1+192(SP)
	0x0176 00374 ($GOROOT/src/fmt/print.go:213)	MOVQ	R10, fmt..autotmp_1+200(SP)
	0x017e 00382 (main.go:15)	MOVQ	fmt..autotmp_0+72(SP), DX
	0x0183 00387 (main.go:15)	MOVQ	DX, fmt.n+56(SP)
	0x0188 00392 (main.go:15)	MOVQ	fmt..autotmp_1+192(SP), DX
	0x0190 00400 (main.go:15)	MOVQ	fmt..autotmp_1+200(SP), R10
	0x0198 00408 (main.go:15)	MOVQ	DX, fmt.err+144(SP)
	0x01a0 00416 (main.go:15)	MOVQ	R10, fmt.err+152(SP)
	0x01a8 00424 (main.go:15)	JMP	426
	0x01aa 00426 (main.go:16)	PCDATA	$1, $-1
	0x01aa 00426 (main.go:16)	MOVQ	232(SP), BP
	0x01b2 00434 (main.go:16)	ADDQ	$240, SP
	0x01b9 00441 (main.go:16)	RET
	0x01ba 00442 (main.go:16)	NOP
	0x01ba 00442 (main.go:8)	PCDATA	$1, $-1
	0x01ba 00442 (main.go:8)	PCDATA	$0, $-2
	0x01ba 00442 (main.go:8)	CALL	runtime.morestack_noctxt(SB)
	0x01bf 00447 (main.go:8)	PCDATA	$0, $-1
	0x01bf 00447 (main.go:8)	NOP
	0x01c0 00448 (main.go:8)	JMP	0
	0x0000 4c 8d 64 24 90 4d 3b 66 10 0f 86 ab 01 00 00 48  L.d$.M;f.......H
	0x0010 81 ec f0 00 00 00 48 89 ac 24 e8 00 00 00 48 8d  ......H..$....H.
	0x0020 ac 24 e8 00 00 00 48 8d 0d 00 00 00 00 48 89 4c  .$....H......H.L
	0x0030 24 60 48 c7 44 24 68 02 00 00 00 44 0f 11 7c 24  $`H.D$h....D..|$
	0x0040 70 48 8b 44 24 60 48 8b 5c 24 68 e8 00 00 00 00  pH.D$`H.\$h.....
	0x0050 48 89 44 24 58 48 8d 0d 00 00 00 00 48 89 4c 24  H.D$XH......H.L$
	0x0060 70 48 89 44 24 78 48 8d 0d 00 00 00 00 48 89 8c  pH.D$xH......H..
	0x0070 24 80 00 00 00 48 c7 84 24 88 00 00 00 03 00 00  $....H..$.......
	0x0080 00 44 0f 11 bc 24 b0 00 00 00 48 8d 8c 24 b0 00  .D...$....H..$..
	0x0090 00 00 48 89 4c 24 50 84 01 48 8b 54 24 78 48 8b  ..H.L$P..H.T$xH.
	0x00a0 74 24 70 48 89 b4 24 b0 00 00 00 48 89 94 24 b8  t$pH..$....H..$.
	0x00b0 00 00 00 84 01 eb 00 48 89 8c 24 d0 00 00 00 48  .......H..$....H
	0x00c0 c7 84 24 d8 00 00 00 01 00 00 00 48 c7 84 24 e0  ..$........H..$.
	0x00d0 00 00 00 01 00 00 00 48 c7 44 24 38 00 00 00 00  .......H.D$8....
	0x00e0 44 0f 11 bc 24 90 00 00 00 48 c7 44 24 48 00 00  D...$....H.D$H..
	0x00f0 00 00 44 0f 11 bc 24 c0 00 00 00 44 0f 11 bc 24  ..D...$....D...$
	0x0100 a0 00 00 00 48 8b 8c 24 80 00 00 00 48 8b b4 24  ....H..$....H..$
	0x0110 d0 00 00 00 48 8b bc 24 88 00 00 00 4c 8b 84 24  ....H..$....L..$
	0x0120 d8 00 00 00 4c 8b 8c 24 e0 00 00 00 48 8b 1d 00  ....L..$....H...
	0x0130 00 00 00 48 8d 05 00 00 00 00 e8 00 00 00 00 48  ...H...........H
	0x0140 89 44 24 40 48 89 9c 24 a0 00 00 00 48 89 8c 24  .D$@H..$....H..$
	0x0150 a8 00 00 00 48 8b 54 24 40 48 89 54 24 48 48 8b  ....H.T$@H.T$HH.
	0x0160 94 24 a0 00 00 00 4c 8b 94 24 a8 00 00 00 48 89  .$....L..$....H.
	0x0170 94 24 c0 00 00 00 4c 89 94 24 c8 00 00 00 48 8b  .$....L..$....H.
	0x0180 54 24 48 48 89 54 24 38 48 8b 94 24 c0 00 00 00  T$HH.T$8H..$....
	0x0190 4c 8b 94 24 c8 00 00 00 48 89 94 24 90 00 00 00  L..$....H..$....
	0x01a0 4c 89 94 24 98 00 00 00 eb 00 48 8b ac 24 e8 00  L..$......H..$..
	0x01b0 00 00 48 81 c4 f0 00 00 00 c3 e8 00 00 00 00 90  ..H.............
	0x01c0 e9 3b fe ff ff                                   .;...
	rel 3+0 t=23 type.string+0
	rel 3+0 t=23 type.*os.File+0
	rel 41+4 t=14 go.string."qq"+0
	rel 76+4 t=7 runtime.convTstring+0
	rel 88+4 t=14 type.string+0
	rel 105+4 t=14 go.string."%+v"+0
	rel 303+4 t=14 os.Stdout+0
	rel 310+4 t=14 go.itab.*os.File,io.Writer+0
	rel 315+4 t=7 fmt.Fprintf+0
	rel 443+4 t=7 runtime.morestack_noctxt+0
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
gclocals·abcd98c9569535db9f3522d983659dbd SRODATA dupok size=11
	0x0000 01 00 00 00 13 00 00 00 00 00 00                 ...........
"".main.stkobj SRODATA static size=24
	0x0000 01 00 00 00 00 00 00 00 c8 ff ff ff 10 00 00 00  ................
	0x0010 10 00 00 00 00 00 00 00                          ........
	rel 20+4 t=5 runtime.gcbits.02+0
