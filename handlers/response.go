package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mongowapi/models"
)

// ok 返回成功响应
func ok(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, models.APIResponse{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// fail 返回失败响应
func fail(c *gin.Context, status int, code int, msg string) {
	c.JSON(status, models.APIResponse{
		Code:    code,
		Message: msg,
	})
}

// bindJSON 解析请求体并处理解析错误
func bindJSON(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		fail(c, http.StatusBadRequest, 400, "请求参数无效: "+err.Error())
		return false
	}
	return true
}