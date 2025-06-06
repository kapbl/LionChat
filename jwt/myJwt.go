package myjwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type JWTClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func JWTEncoder(secret string, email string) string {
	claims := JWTClaims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 12)),
			Issuer:    "my_app",
		},
	}
	// 创建Token，使用HS256算法
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 签名并获取完整Token字符串
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		fmt.Println("生成Token失败:", err)
		return ""
	}
	return tokenString
}

func JWTUnencoder(secretKey []byte, tokenString string) *JWTClaims {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&JWTClaims{},
		func(token *jwt.Token) (interface{}, error) {
			// 验证签名算法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
			}
			return secretKey, nil
		},
	)
	// 处理解析错误
	if err != nil {
		return nil
	}
	// 验证Claims有效性
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims
	}
	return nil
}
