package stack

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
}
