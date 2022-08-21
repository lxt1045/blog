package ret

import (
	"fmt"
	"testing"
)

func TestRet(t *testing.T) {
	t.Run("NewWrapper", func(t *testing.T) {
		func() {
			defer func() {
				t.Log("outer defer")
			}()
			err := func() (err error) {
				defer func() {
					t.Log("inner defer")
				}()
				t.Log("1")
				t.Log("2")
				err3 := fmt.Errorf("error 3")

				Ret(err)
				t.Log("3")
				err = err3
				Ret(err)
				t.Log("4")
				return
			}()
			t.Log("outer err:", err)
		}()
	})
}
