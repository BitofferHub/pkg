package errors

type Error interface {
	Error() string
	Code() int
	SetCode(code int)
	SetMessage(message string)
}

// 自定义错误类型
type bitstormError struct {
	code    int    // 错误码
	message string // 错误消息
}

func (e *bitstormError) Error() string {
	return e.message
}

func (e *bitstormError) Code() int {
	return e.code
}

func (e *bitstormError) SetCode(code int) {
	e.code = code
}

func (e *bitstormError) SetMessage(message string) {
	e.message = message
}

// 创建一个新的自定义错误
func NewError(code int, message string) Error {
	return &bitstormError{code: code, message: message}
}
