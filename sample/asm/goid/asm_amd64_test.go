package goid

import (
	"testing"

	"github.com/petermattis/goid"
)

var pc uintptr

func TestGetg(t *testing.T) {
	t.Run("Getg", func(t *testing.T) {
		gid := Getg()
		t.Logf("Getg:%d", gid)

		goid := goid.Get()
		t.Logf("GetG:%d", goid)
	})

	t.Run("getgg", func(t *testing.T) {
		gid := add(1, 2)
		t.Logf("Getg:%d", gid)
	})
}

func TestVar(t *testing.T) {
	t.Run("GetVar", func(t *testing.T) {
		t.Logf("Id:%d", Id)
		t.Logf("Name:%s", Name)
		t.Logf("helloworld:%s", Helloworld)
		doSth()
	})
}

func TestPrintln(t *testing.T) {
	t.Run("GetVar", func(t *testing.T) {
		Print("sss")
	})
}

func BenchmarkGoid(b *testing.B) {
	b.Run("Getg", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			Getg()
		}
		b.StopTimer()
	})

	b.Run("goid.Get", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			goid.Get()
		}
		b.StopTimer()
	})
	b.Run("Getg", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			Getg()
		}
		b.StopTimer()
	})
}

//go:noinline
func maxNoinline(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func maxInline(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func BenchmarkInline(b *testing.B) {
	x, y := 1, 2
	b.Run("BenchmarkNoInline", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			maxNoinline(x, y)
		}
	})
	b.Run("BenchmarkInline", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			maxInline(x, y)
		}
	})
}
