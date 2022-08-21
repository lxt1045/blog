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
	"fmt"
	"testing"

	myfmt "github.com/lxt1045/blog/sample/fmt"
)

var escapeStr string

func BenchmarkLineNO1(b *testing.B) {
	b.Run("+", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			escapeStr = "1234567890:" + "cxxxx"
		}
	})
	b.Run("fmt.Sprint", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			escapeStr = fmt.Sprint("1234567890:", "cxxxx")
		}
	})
	b.Run("fmt.Sprintf", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			escapeStr = fmt.Sprintf("1234567890:%s", ":")
		}
	})
	b.Run(`fmt.Sprintf("%f")`, func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			escapeStr = fmt.Sprintf("%f", 1.00000001)
		}
	})
	b.Run(`fmt.Sprintf("%d")`, func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			escapeStr = fmt.Sprintf("%d", 100000001)
		}
	})
	b.Run("myfmt.Sprintf", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			escapeStr = myfmt.Sprintf("1234567890:%s", ":")
		}
	})
	b.Run("myfmt.Sprintf2", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			escapeStr = myfmt.Sprintf2("1234567890:%s", ":")
		}
	})
}

func BenchmarkLineNO2(b *testing.B) {
	b.Run("fmt.Sprintf", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			escapeStr = fmt.Sprintf("1234567890:%s", ":")
		}
	})
	b.Run("myfmt.Sprintf", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			escapeStr = myfmt.Sprintf("1234567890:%s", ":")
		}
	})
	b.Run("myfmt.Sprintf2", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			escapeStr = myfmt.Sprintf2("1234567890:%s", ":")
		}
	})
}
