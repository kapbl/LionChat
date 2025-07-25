package middlewares

import (
	"cchat/config"
	"cchat/pkg/token"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func JwtMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取token
		currToken := c.GetHeader("Authorization")
		if currToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
			c.Abort()
			return
		}
		currToken = strings.TrimPrefix(currToken, "Bearer ")
		// 解析token
		claims := &token.Claims{}
		token, err := jwt.ParseWithClaims(currToken, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.AppConfig.JWT.ApiSecret), nil
		})
		if err != nil {
			// 401 未授权
			c.JSON(401, gin.H{"error": "未授权"})
			c.Abort()
			return
		}
		// 验证token
		if !token.Valid {
			// 401 未授权
			c.JSON(401, gin.H{"error": "未授权"})
			c.Abort()
			return
		}
	}
}

func JwtParse(c *gin.Context) {
	// 从请求头中获取token
	currToken := c.GetHeader("Authorization")
	if currToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		c.Abort()
		return
	}
	currToken = strings.TrimPrefix(currToken, "Bearer ")
	// 解析token
	claims := &token.Claims{}
	token, err := jwt.ParseWithClaims(currToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.AppConfig.JWT.ApiSecret), nil
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		c.Abort()
		return
	}
	if !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		c.Abort()
		return
	}
	c.Set("userId", claims.UserID)
	c.Set("userUuid", claims.UserUUID)
	c.Set("username", claims.UserName)
	c.Next()
}
