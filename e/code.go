package e

import "fmt"

type ErrorCode interface {
	error
	StatusCode() int
	ErrorCode() int
	Msg() string
	WithMsg(string)
}

const CodeBit = 100000

// ErrCode 规定组成部分为http状态码+5位错误码
type ErrCode struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Result  interface{} `json:"result"`
}

func (e *ErrCode) Error() string {
	return e.Message
}

func (e *ErrCode) StatusCode() int {
	return e.Code / CodeBit
}

func (e *ErrCode) ErrorCode() int {
	return e.Code % CodeBit
}

func (e *ErrCode) Msg() string {
	return e.Message
}

func (e *ErrCode) WithMsg(msg string) {
	e.Message = msg
}

var (
	// 00~99为服务级别错误码

	ErrInternalServerError = &ErrCode{Code: 50010000, Message: "服务器内部错误"}
	ErrInvalidParam        = &ErrCode{Code: 40010001, Message: "请求参数不正确"}
	ErrNotFound            = &ErrCode{Code: 40410002, Message: "资源不存在"}
	ErrMethodNotAllow      = &ErrCode{Code: 40510003, Message: "方法不允许"}
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
		code := errorCode.ErrorCode()
		value, ok := temp[code]
		if ok {
			return fmt.Errorf("error code %d(%s) already exists", code, value)
		}
		temp[code] = errorCode.Msg()
	}
	for errorCode := range m {
		code := errorCode.ErrorCode()
		value, ok := temp[code]
		if ok {
			return fmt.Errorf("error code %d(%s) already exists", code, value)
		}
		temp[code] = errorCode.Msg()
	}
	return nil
}
