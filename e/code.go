package e

import "fmt"

type ErrorCode interface {
	error
	StatusCode() int
	Code() int
	Message() string
	WithMessage(string)
}

// ErrCode 规定组成部分为http状态码+5位错误码
type ErrCode struct {
	Err int
	Msg string
}

func (e *ErrCode) Error() string {
	return e.Msg
}

func (e *ErrCode) StatusCode() int {
	return e.Err / 100000
}

func (e *ErrCode) Code() int {
	return e.Err % 100000
}

func (e *ErrCode) Message() string {
	return e.Msg
}

func (e *ErrCode) WithMessage(msg string) {
	e.Msg = msg
}

var (
	// 00~99为服务级别错误码

	ErrInternalServerError = &ErrCode{Err: 50010000, Msg: "服务器内部错误"}
	ErrInvalidParam        = &ErrCode{Err: 40010001, Msg: "请求参数不正确"}
	ErrNotFound            = &ErrCode{Err: 40410002, Msg: "资源不存在"}
	ErrMethodNotAllow      = &ErrCode{Err: 40510003, Msg: "方法不允许"}
)

// AddCode business code to codeMessageBox
func AddCode(m map[ErrorCode]struct{}) error {
	temp := make(map[int]string)
	for errorCode := range map[ErrorCode]struct{}{
		ErrInternalServerError: {},
		ErrInvalidParam:        {},
		ErrNotFound:            {},
		ErrMethodNotAllow:      {},
	} {
		code := errorCode.Code()
		value, ok := temp[code]
		if ok {
			return fmt.Errorf("error code %d(%s) already exists", code, value)
		}
		temp[code] = errorCode.Message()
	}
	for errorCode := range m {
		code := errorCode.Code()
		value, ok := temp[code]
		if ok {
			return fmt.Errorf("error code %d(%s) already exists", code, value)
		}
		temp[code] = errorCode.Message()
	}
	return nil
}
