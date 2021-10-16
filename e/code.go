package e

import "fmt"

// Code 规定组成部分为http状态码+5位错误码
type Code int

func (c Code) Error() string {
	if c < 10000000 || c >= 60000000 {
		return InvalidError
	}
	message, ok := codeZhMessageBox[c]
	if ok {
		return message
	}
	return UnDefineError
}

func (c Code) Code() int {
	return int(c)
}

func (c Code) StatusCode() int {
	return int(c) / 100000
}

const (
	UnDefineError = "未定义错误码"
	InvalidError  = "无效错误码"
)

const (
	// 00~99为服务级别错误码

	ErrInternalServerError Code = 50000000
	ErrInvalidParam        Code = 40000001
	ErrNotFound            Code = 40400002
	ErrMethodNotAllow      Code = 40500003
)

var codeZhMessageBox = map[Code]string{
	ErrInvalidParam:        "请求参数不正确",
	ErrNotFound:            "资源不存在",
	ErrMethodNotAllow:      "方法不允许",
	ErrInternalServerError: "服务器内部错误",
}

// AddCode business code to codeMessageBox
func AddCode(m map[Code]string) error {
	for code, msg := range m {
		value, ok := codeZhMessageBox[code]
		if ok {
			return fmt.Errorf("error code %d(%s) already exists", code, value)
		}
		codeZhMessageBox[code] = msg
	}
	return nil
}
