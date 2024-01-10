package constant

import (
	"errors"
	"fmt"
)

var (
	ERR_HANDLE_INPUT = errors.New("handle input error")
)

type ErrCode int // 错误码

var (
	Success          ErrCode = 0
	ErrInputInvalid  ErrCode = 8020
	ErrShouldBind    ErrCode = 8021
	ERR_JSON_MARSHAL ErrCode = 8022
)

var errMsgDic = map[ErrCode]string{
	Success:          "ok",
	ErrInputInvalid:  "input invalid",
	ErrShouldBind:    "should bind failed",
	ERR_JSON_MARSHAL: "json marshal failed",
}

// GetErrMsg 获取错误描述
func GetErrMsg(code ErrCode) string {
	if msg, ok := errMsgDic[code]; ok {
		return msg
	}
	return fmt.Sprintf("unknown error code %d", code)
}
