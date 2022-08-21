package errors

import (
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"testing"

	lxterrs "github.com/lxt1045/errors"
	pkgerrs "github.com/pkg/errors"
)

func TestCacheStack(t *testing.T) {
	t.Run("NewStack", func(t *testing.T) {
		stack := NewStack(0)
		t.Logf("stack:%+v", stack)
	})
	t.Run("NewStack2", func(t *testing.T) {
		stack := NewStack2(0)
		t.Logf("stack:%+v", stack)
	})
}

var str string
var GlobalE string

func deepCall(depth int, f func()) {
	if depth <= 0 {
		f()
		return
	}
	deepCall(depth-1, f)
}

func BenchmarkPkg(b *testing.B) {
	b.Run("pkg/errors", func(b *testing.B) {
		b.ReportAllocs()
		var err error
		deepCall(10, func() {
			for i := 0; i < b.N; i++ {
				err = pkgerrs.New("ye error")
				GlobalE = fmt.Sprintf("%+v", err)
			}
			b.StopTimer()
		})
	})
	b.Run("errors-fmt", func(b *testing.B) {
		b.ReportAllocs()
		var err error
		deepCall(10, func() {
			for i := 0; i < b.N; i++ {
				err = errors.New("ye error")
				GlobalE = fmt.Sprintf("%+v", err)
			}
			b.StopTimer()
		})
	})
	b.Run("errors-Errors", func(b *testing.B) {
		b.ReportAllocs()
		var err error
		deepCall(10, func() {
			for i := 0; i < b.N; i++ {
				err = errors.New("ye error")
				GlobalE = err.Error()
			}
			b.StopTimer()
		})
	})
}

func BenchmarkStack(b *testing.B) {
	b.Run("runtime.Callers", func(b *testing.B) {
		deepCall(10, func() {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				pcs := pool.Get().(*[DefaultDepth]uintptr)
				n := runtime.Callers(2, pcs[:DefaultDepth])
				var cs []string
				traces, more, f := runtime.CallersFrames(pcs[:n]), true, runtime.Frame{}
				for more {
					f, more = traces.Next()
					cs = append(cs, f.File+":"+strconv.Itoa(f.Line))
				}
				pool.Put(pcs)
			}
		})
	})
}

func BenchmarkCacheStack(b *testing.B) {
	b.Run("pkg.Sprintf", func(b *testing.B) {
		deepCall(10, func() {
			err := pkgerrs.WithStack(errors.New("test"))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				str = fmt.Sprintf("%+v", err)
			}
		})
	})
	b.Run("NewStack", func(b *testing.B) {
		deepCall(10, func() {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				NewStack(0)
			}
		})
	})
	b.Run("NewStack2", func(b *testing.B) {
		deepCall(10, func() {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				NewStack2(0)
			}
		})
	})
	b.Run("runtime.Callers", func(b *testing.B) {
		deepCall(10, func() {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				pcs := pool.Get().(*[DefaultDepth]uintptr)
				n := runtime.Callers(2, pcs[:DefaultDepth])
				var cs []string
				traces, more, f := runtime.CallersFrames(pcs[:n]), true, runtime.Frame{}
				for more {
					f, more = traces.Next()
					cs = append(cs, f.File+":"+strconv.Itoa(f.Line))
				}
				pool.Put(pcs)
			}
		})
	})
}

func BenchmarkErrors(b *testing.B) {
	b.Run("lxt1045/errors-fmt", func(b *testing.B) {
		deepCall(10, func() {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				err := lxterrs.NewErr(1, "test")
				str = fmt.Sprintf("%+v", err)
			}
		})
	})
	b.Run("lxt1045/errors-Errors", func(b *testing.B) {
		deepCall(10, func() {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				err := lxterrs.NewErr(1, "test")
				str = err.Error()
			}
		})
	})
	b.Run("pkg/errors", func(b *testing.B) {
		deepCall(10, func() {
			err := errors.New("test")
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				err := pkgerrs.WithStack(err)
				str = fmt.Sprintf("%+v", err)
			}
		})
	})
}

func TestErrors(t *testing.T) {
	t.Run("lxt1045/errors", func(t *testing.T) {
		deepCall(3, func() {
			err := lxterrs.NewErr(1, "test")
			str = fmt.Sprintf("%+v", err)
			t.Log(str)
		})
	})
	t.Run("lxt1045/errors", func(t *testing.T) {
		deepCall(3, func() {
			err := lxterrs.NewErr(1, "test")
			str = err.Error()
			t.Log(str)
		})
	})
	t.Run("pkg/errors", func(t *testing.T) {
		deepCall(3, func() {
			err := errors.New("test")
			err = pkgerrs.WithStack(err)
			str = fmt.Sprintf("%+v", err)
			t.Log(str)
		})
	})
}
