package utils

import "github.com/gin-gonic/gin"

// GetUserAgent 获取设备信息
func GetUserAgent(c *gin.Context) string {
	return c.Request.Header.Get("User-Agent")
}

// GetClientIP 获取设备IP
func GetClientIP(c *gin.Context) string {
	return c.ClientIP()
}
