package e

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/crochee/lirity/logger"
)

// Code gin response with ErrorCode
func Code(c *gin.Context, code ErrorCode) {
	c.Abort()
	c.JSON(code.StatusCode(), code)
}

// Error gin Response with error
func Error(c *gin.Context, err error) {
	logger.From(c.Request.Context()).Sugar().Errorf("%+v", err)
	for err != nil {
		u, ok := err.(interface {
			Unwrap() error
		})
		if !ok {
			break
		}
		err = u.Unwrap()
	}
	if err == nil {
		Code(c, ErrInternalServerError)
		return
	}
	var errorCode *ErrCode
	if errors.As(err, &errorCode) {
		Code(c, errorCode)
		return
	}
	Code(c, ErrInternalServerError.WithResult(err))
}
