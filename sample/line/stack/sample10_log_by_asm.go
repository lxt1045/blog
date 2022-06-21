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

package stack

import (
	"runtime"
	"strconv"
	"sync/atomic"
	"unsafe"
)

var (
	mapLogByAsm unsafe.Pointer = func() unsafe.Pointer {
		m := make(map[uintptr]string, 1024)
		return unsafe.Pointer(&m)
	}()
)

type Log struct {
	PC   uintptr
	Code int64
	Msg  string
}

func NewLog(code int, msg string) Log

func (l Log) GetCode() (line int64) {
	return l.Code
}

func (l Log) GetMsg() (msg string) {
	return l.Msg
}

func (l Log) LineNO() (line string) {
	mPCs := *(*map[uintptr]string)(atomic.LoadPointer(&mapLogByAsm))
	line, ok := mPCs[l.PC]
	if !ok {
		file, n := runtime.FuncForPC(l.PC).FileLine(l.PC)
		line = file + ":" + strconv.Itoa(n)
		mPCs2 := make(map[uintptr]string, len(mPCs)+10)
		mPCs2[l.PC] = line
		for {
			p := atomic.LoadPointer(&mapLogByAsm)
			mPCs = *(*map[uintptr]string)(p)
			for k, v := range mPCs {
				mPCs2[k] = v
			}
			swapped := atomic.CompareAndSwapPointer(&mapLogByAsm, p, unsafe.Pointer(&mPCs2))
			if swapped {
				break
			}
		}
	}
	return
}
