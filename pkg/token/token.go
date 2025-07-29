package token

import (
	"cchat/config"
	"cchat/internal/dao/model"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	UserID   int    `json:"user_id"`
	UserUUID string `json:"user_uuid"`
	UserName string `json:"user_name"`
	jwt.RegisteredClaims
}

// todo 等待实现
func GEnToken(user *model.Users) (string, error) {
	// Jwt载荷: user_id, user_name, user_uuid
	claims := Claims{
		UserID:   int(user.Id),
		UserUUID: user.Uuid,
		UserName: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)), // 过期时间
			Issuer:    "chatLion",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secretKey := []byte(config.AppConfig.JWT.ApiSecret)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ParseToken(inputToken string) (*Claims, error) {
	if inputToken == "" {
		return nil, errors.New("token 不能为空")
	}
	// 解析token
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(inputToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.AppConfig.JWT.ApiSecret), nil
	})
	if err != nil {
		return nil, err
	}
	// 验证token
	if !token.Valid {
		// 401 未授权
		return nil, errors.New("token 无效")
	}
	return claims, nil
}
