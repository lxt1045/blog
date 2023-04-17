// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"
#include "funcdata.h"


TEXT	·FirstBitIdx(SB), NOSPLIT, $0-16
	MOVQ b_base+0(FP), AX  	// in
	BSRQ	AX, AX
	MOVQ 	AX, ret+8(FP)
	RET
