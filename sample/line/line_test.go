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

package main

import (
	"runtime"
	"testing"

	"github.com/lxt1045/blog/sample/line/stack"
	"github.com/stretchr/testify/assert"
)

var line string

// 这里有未初始化的内存?
var _ = GetLineNO3(100)

func BenchmarkLineNO1(b *testing.B) {
	b.Run("LineByRuntime", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			line = LineByRuntime()
		}
	})
	b.Run("Line.Load", func(b *testing.B) {
		var lineNO Line
		for i := 0; i < b.N; i++ {
			line = lineNO.Load()
		}
	})
	pc := stack.GetPC()
	b.Run("runtime.FuncForPC", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = runtime.FuncForPC(pc)
		}
	})
	b.Run("GetLineNO3", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			line = GetLineNO3(99)
		}
	})
	b.Run("GetLineNO4", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			line = GetLineNO4("BenchmarkWrap.id1")
		}
	})
	b.Run("GetLineNO5", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			line = GetLineNO5("BenchmarkWrap.id2")
		}
	})
	b.Run("GetLineByFunc", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			line = GetLineByFunc(func() {})
		}
	})
	b.Run("GetLineByRuntimeCache", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			line = GetLineByRuntimeCache()
		}
	})
	b.Run("GetPCByAsm", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			line = GetPCByAsm(stack.GetPC())
		}
	})
	b.Run("stack.NewLine().LineNO()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			line = stack.NewLine().LineNO()
		}
	})
}

func TestLine(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		l := &Line{}
		line1, line2, line3, line4 := LineByRuntime(), LineByRuntime(), l.Load(), GetLineNO3(1)
		assert.Equal(t, line1, line2)
		assert.Equal(t, line1, line3)
		assert.Equal(t, line1, line4)
		line1, line5, line6 := LineByRuntime(), GetLineNO5("22"), GetLineByFunc(func() {})
		assert.Equal(t, line1, line5)
		assert.Equal(t, line1, line6)
		line1, line7, line8 := LineByRuntime(), GetLineByRuntimeCache(), GetPCByAsm(stack.GetPC())
		assert.Equal(t, line1, line7)
		assert.Equal(t, line1, line8)
		line1, line9 := LineByRuntime(), stack.NewLine().LineNO()
		assert.Equal(t, line1, line9)

		t.Log(line1)
	})

}

func TestLine00(t *testing.T) {

	t.Log(translateNum(624))
}

func translateNum(num int) int {
	sum_1, sum := 1, 1
	idxMax := int('z' - 'a')
	for num >= 10 {
		now := num % 100
		num = num / 10
		if now >= 10 && now <= idxMax {
			sum_0 := sum
			sum = sum_1 + sum
			sum_1 = sum_0
		} else {
			sum_1 = sum
		}
	}
	return sum
}
