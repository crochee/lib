package e

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/crochee/lirity/log"
)

// GinErrorCode gin response with ErrorCode
func GinErrorCode(c *gin.Context, code ErrorCode) {
	c.Abort()
	c.JSON(code.StatusCode(), code)
}

// GinError gin Response with error
func GinError(c *gin.Context, err error) {
	log.FromContext(c.Request.Context()).Errorf("%+v", err)
	for err != nil {
		wrapper, ok := err.(UnwrapHandle)
		if !ok {
			break
		}
		err = wrapper.Unwrap()
	}
	if err == nil {
		GinErrorCode(c, ErrInternalServerError)
		return
	}
	var errorCode *ErrCode
	if errors.As(err, &errorCode) {
		GinErrorCode(c, errorCode)
		return
	}
	GinErrorCode(c, ErrInternalServerError)
}
