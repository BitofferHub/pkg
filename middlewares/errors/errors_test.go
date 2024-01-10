package errors

import (
	"testing"
)

func TestError(t *testing.T) {
	// new 一个error
	err1 := NewError(1, "err1")
	err2 := NewError(2, "err2")
	var err3 Error

	if err1 != nil {
		t.Error(err1.Error())
	}
	if err2 != nil {
		t.Error(err2.Error())
	}
	if err3 != nil {
		t.Error(err3.Error())
	} else {
		t.Log("err3 is nil")
	}
	// 修改error1的code
	err1.SetCode(11)
	err1.SetMessage("err11")

	err2.SetCode(22)
	err2.SetMessage("err22")

	t.Log(err1.Error(), err2.Error())
	t.Log(err1.Code(), err2.Code())
}
