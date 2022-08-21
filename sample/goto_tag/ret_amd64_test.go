package ret

import (
	"fmt"
	"testing"
)

func TestJump1(t *testing.T) {
	t.Run("NewLine", func(t *testing.T) {
		func() {
			defer func() {
				t.Log("outer defer")
			}()
			err := func() (err error) {
				defer func() {
					t.Log("inner defer")
				}()
				t.Log("1")
				// fJump, err1 := NewTag()
				tag, err1 := NewTag2()
				if err1 != nil {
					t.Log("3")
					err = err1
					t.Log("Tag() get error:", err1)
					tag.Try(err3)
					return
				}
				t.Log("2")
				err3 := fmt.Errorf("error 3")
				// GotoTag(err3)
				// TryJump(err3)
				// fJump(err3)
				tag.Try(nil)
				func() {
					tryTagErr = func(err error) {
						t.Fatal(err)
					}
					tag.Try(err3)
				}()
				tag.Try(err3)
				t.Log("4")
				return
			}()
			t.Log("outer err:", err)
		}()
	})
}

func TestJump(t *testing.T) {
	t.Run("NewLine", func(t *testing.T) {
		func() {
			defer func() {
				t.Log("outer defer")
			}()
			err := func() (err error) {
				defer func() {
					t.Log("inner defer")
				}()
				t.Log("1")
				if err1 := Tag(); err1 != nil {
					t.Log("3")
					err = err1
					t.Log("Tag() get error:", err1)
					return
				}
				t.Log("2")
				err3 := fmt.Errorf("error 3")
				GotoTag(err3)
				t.Log("4")
				return
			}()
			t.Log("outer err:", err)
		}()
	})
}

func TestTag(t *testing.T) {
	t.Run("NewLine", func(t *testing.T) {
		func() {
			defer func() {
				t.Log("outer defer")
			}()
			err := func() (err error) {
				defer func() {
					t.Log("inner defer")
				}()
				t.Log("1")
				if err1 := Tag(); err1 != nil {
					t.Log("3")
					err = err1
					t.Log("Tag() get error:", err1)
					return
				}
				t.Log("2")
				err3 := fmt.Errorf("error 3")
				GotoTag(err3)
				t.Log("4")
				return
			}()
			t.Log("outer err:", err)
		}()
	})
}
