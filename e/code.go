package e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/json-iterator/go"
)

type ErrorCode interface {
	error
	json.Marshaler
	json.Unmarshaler
	StatusCode() int
	Code() string
	Message() string
	Result() interface{}
	WithStatusCode(int) ErrorCode
	WithCode(string) ErrorCode
	WithMessage(string) ErrorCode
	WithResult(interface{}) ErrorCode
}

type InnerError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Result  interface{} `json:"result"`
}

func From(response *http.Response) ErrorCode {
	decoder := jsoniter.ConfigCompatibleWithStandardLibrary.NewDecoder(response.Body)
	decoder.UseNumber()
	var result ErrCode
	if err := decoder.Decode(&result); err != nil {
		return ErrParseContent.WithResult(err)
	}
	return result.WithStatusCode(response.StatusCode)
}

func Froze(code, message string) ErrorCode {
	return &ErrCode{
		code: code,
		msg:  message,
	}
}

// ErrCode 规定组成部分为http状态码+5位错误码
type ErrCode struct {
	code   string
	msg    string
	result interface{}
}

func (e *ErrCode) Error() string {
	return fmt.Sprintf("code:%s,message:%s,result:%s", e.Code(), e.Message(), e.Result())
}

func (e *ErrCode) MarshalJSON() ([]byte, error) {
	inner := &InnerError{
		Code:    e.code,
		Message: e.msg,
		Result:  e.result,
	}
	return jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(inner)
}

func (e *ErrCode) UnmarshalJSON(bytes []byte) error {
	var result InnerError
	if err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(bytes, &result); err != nil {
		return err
	}
	e.code = result.Code
	e.msg = result.Message
	e.result = result.Result
	return nil
}

func (e *ErrCode) StatusCode() int {
	statusCode, _ := strconv.Atoi(e.code[:3])
	return statusCode
}

func (e *ErrCode) Code() string {
	return e.code
}

func (e *ErrCode) Message() string {
	return e.msg
}

func (e *ErrCode) Result() interface{} {
	return e.result
}

func (e *ErrCode) WithStatusCode(statusCode int) ErrorCode {
	ec := *e
	ec.code = strconv.Itoa(statusCode) + ec.Code()[3:]
	return &ec
}

func (e *ErrCode) WithCode(code string) ErrorCode {
	ec := *e
	ec.code = strconv.Itoa(ec.StatusCode()) + code[3:]
	return &ec
}

func (e *ErrCode) WithMessage(msg string) ErrorCode {
	ec := *e
	ec.msg = msg
	return &ec
}

func (e *ErrCode) WithResult(result interface{}) ErrorCode {
	ec := *e
	ec.result = result
	return &ec
}

var (
	// 00~99为服务级别错误码

	ErrInternalServerError = Froze("5000000000", "服务器内部错误")
	ErrInvalidParam        = Froze("4000000001", "请求参数不正确")
	ErrNotFound            = Froze("4040000002", "资源不存在")
	ErrNotAllowMethod      = Froze("4050000003", "不允许此方法")
	ErrParseContent        = Froze("5000000004", "解析内容失败")
)

// AddCode business code to codeMessageBox
func AddCode(m map[ErrorCode]struct{}) error {
	temp := make(map[string]string)
	for errorCode := range map[ErrorCode]struct{}{
		ErrInternalServerError: {},
		ErrInvalidParam:        {},
		ErrNotFound:            {},
		ErrNotAllowMethod:      {},
		ErrParseContent:        {},
	} {
		if err := validateErrorCode(errorCode); err != nil {
			return err
		}
		code := errorCode.Code()
		value, ok := temp[code]
		if ok {
			return fmt.Errorf("error code %s(%s) already exists", code, value)
		}
		temp[code] = errorCode.Message()
	}
	for errorCode := range m {
		if err := validateErrorCode(errorCode); err != nil {
			return err
		}
		code := errorCode.Code()
		value, ok := temp[code]
		if ok {
			return fmt.Errorf("error code %s(%s) already exists", code, value)
		}
		temp[code] = errorCode.Message()
	}
	return nil
}

// validateErrorCode check err must be 3(http)+3(service)+4(error)
func validateErrorCode(err ErrorCode) error {
	code := err.Code()
	statusCode := err.StatusCode()
	if statusCode < 100 || statusCode >= 600 {
		return fmt.Errorf("error code %s has invalid status code %d", code, statusCode)
	}
	if l := len(code); l != 10 {
		return fmt.Errorf("error code %s is %d,but it must be 10", code, l)
	}
	return nil
}
