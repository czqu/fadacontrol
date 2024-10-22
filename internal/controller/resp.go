package controller

import (
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/schema"
	"fadacontrol/pkg/utils"
	"fmt"
	"github.com/gin-gonic/gin"
)

func GetResult(requestId string, ex *exception.Exception, data interface{}, format interface{}, a ...interface{}) *schema.ResponseData {

	var f string
	if format == nil {
		f = ex.Msg
	}

	switch format.(type) {
	case string:
		f = format.(string)
	case error:
		f = fmt.Sprintf("%v", format.(error))
	}

	return &schema.ResponseData{
		RequestId: requestId,
		Code:      ex.Code,
		Msg:       f,
		Data:      data,
	}
}
func GetSuccess(requestId string) *schema.ResponseData {
	return GetResult(requestId, exception.ErrSuccess, nil, nil)
}
func GetGinSuccess(c *gin.Context) *schema.ResponseData {
	return GetResult(utils.GetRequestId(c), exception.ErrSuccess, nil, nil)

}

// GetSuccessWithData first is requestId , second is data ,
func GetSuccessWithData(requestId string, data ...interface{}) *schema.ResponseData {
	switch len(data) {

	case 1:

		return GetResult(requestId, exception.ErrSuccess, data[1], nil)

	}
	return GetSuccess(requestId)
}
func GetGinSuccessWithData(c *gin.Context, data ...interface{}) *schema.ResponseData {
	requestId := utils.GetRequestId(c)
	switch len(data) {

	case 1:

		return GetResult(requestId, exception.ErrSuccess, data[0], nil)

	}
	return GetSuccess(requestId)
}
func GetError(requestId string, ex *exception.Exception) *schema.ResponseData {
	return GetResult(requestId, ex, nil, nil)
}
func GetGinError(c *gin.Context, ex *exception.Exception) *schema.ResponseData {
	return GetResult(utils.GetRequestId(c), ex, nil, nil)
}
func GetErrorWithData(requestId string, ex *exception.Exception, data ...interface{}) *schema.ResponseData {
	switch len(data) {
	case 1:
		return GetResult(requestId, ex, data[0], nil)

	default:
		return GetError(requestId, exception.ErrUnknownException)
	}
}
func GetGinErrorWithData(c *gin.Context, ex *exception.Exception, data ...interface{}) *schema.ResponseData {
	return GetErrorWithData(utils.GetRequestId(c), ex, data...)
}
