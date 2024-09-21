package middleware

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"net/http"
	"qq_demo/model"
	"strings"
)

var Secret = []byte("dech53")

func JWTAuthMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusOK, gin.H{
				"code": 302,
				"msg":  "请求头中auth为空",
			})
			c.Abort()
			return
		}
		// 按空格分割
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusOK, gin.H{
				"code": 2004,
				"msg":  "请求头中auth格式有误",
			})
			c.Abort()
			return
		}
		mc, err := ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": 302,
				"msg":  "无效的Token",
			})
			c.Abort()
			return
		}
		c.Set("username", mc.Username)
		c.Next()
	}
}

// ParseToken 解析JWT
func ParseToken(tokenString string) (*model.MyClaims, error) {
	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &model.MyClaims{}, func(token *jwt.Token) (i interface{}, err error) {
		return Secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*model.MyClaims); ok && token.Valid { // 校验 token
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
