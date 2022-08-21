package ret

import (
	"log"
	"runtime"
	"testing"
)

func TestRet(t *testing.T) {
	t.Run("NewWrapper", func(t *testing.T) {
		defer func() {
			t.Log("defer out")
		}()
		func() {
			defer func() {
				t.Log("defer")
			}()
			pcs := make([]uintptr, 32)
			// n := runtime.Callers(1, pcs)
			n := GetStack(pcs)
			pcs = pcs[:n]
			for i, c := range toCallers(pcs) {
				log.Printf("%d:%x:%s", i, c.pc, c.line)
			}

			funcInfo := runtime.FuncForPC(pcs[1])
			continpc := funcInfo.Entry() + uintptr(raw(funcInfo).deferreturn) //+ 1

			for i, c := range toCallers([]uintptr{funcInfo.Entry(), continpc}) {
				log.Printf("continpc,%d:%x:%s", i, c.pc, c.line)
			}
			t.Log("end")
		}()
	})
}
