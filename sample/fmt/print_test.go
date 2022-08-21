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

package fmt

import (
	stdfmt "fmt"
	"strconv"
	"testing"
)

var escapeStr string

func BenchmarkLineNO1(b *testing.B) {
	b.Run("+", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			escapeStr = "1234567890:" + "cxxxx"
		}
	})
	b.Run("+(10)", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			escapeStr = "1234567890" + ":cxxxx" + ":cxxxx" + ":cxxxx" + ":cxxxx" + ":cxxxx" + ":cxxxx" + ":cxxxx" + ":cxxxx" + ":cxxxx"
		}
	})
	b.Run("fmt.Sprint", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			escapeStr = stdfmt.Sprint("1234567890:", "cxxxx")
		}
	})
	b.Run("fmt.Sprintf", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			escapeStr = stdfmt.Sprintf("1234567890:%s", "cxxxx")
		}
	})
	b.Run("fmt.Sprintf(10)", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			escapeStr = stdfmt.Sprintf("1234567890:%s:%s:%s:%s:%s:%s:%s:%s:%s:%s",
				"cxxxx", "cxxxx", "cxxxx", "cxxxx", "cxxxx", "cxxxx", "cxxxx", "cxxxx", "cxxxx", "cxxxx")
		}
	})
	b.Run(`fmt.Sprintf("%f")`, func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			escapeStr = stdfmt.Sprintf("%f", 1.00000001)
		}
	})
	b.Run(`fmt.Sprintf("%d")`, func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			escapeStr = stdfmt.Sprintf("%d", 100000001)
		}
	})
	b.Run("myfmt.Sprintf", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			escapeStr = Sprintf("1234567890:%s", ":")
		}
	})
	b.Run("myfmt.fmtFloat", func(b *testing.B) {
		f := fmt{}
		buf := buffer(make([]byte, 0, 16*1024))
		f.init(&buf)
		f32 := 1.000245678
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			f.fmtFloat(f32, 32, 'g', -1)
		}
	})
	b.Run("strconv.AppendFloat", func(b *testing.B) {
		buf := buffer(make([]byte, 0, 16*1024))
		f32 := 1.000245678
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf = strconv.AppendFloat(buf, f32, 'f', 1, 64)
		}
	})
	b.Run("strconv.FormatFloat", func(b *testing.B) {
		f32 := 1.000245678
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			strconv.FormatFloat(f32, 'f', 1, 64)
		}
	})
	b.Run("myfmt.fmtInteger", func(b *testing.B) {
		f := fmt{}
		buf := buffer(make([]byte, 0, 16*1024))
		f.init(&buf)
		var i32 uint64 = 1000245678
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			f.fmtInteger(i32, 10, true, 'd', ldigits)
		}
	})
	b.Run("strconv.AppendUint", func(b *testing.B) {
		buf := buffer(make([]byte, 0, 16*1024))
		var i32 uint64 = 1000245678
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf = strconv.AppendUint(buf, i32, 10)
		}
	})
	b.Run("strconv.FormatUint", func(b *testing.B) {
		var i32 uint64 = 1000245678
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			strconv.FormatUint(i32, 10)
		}
	})
}
