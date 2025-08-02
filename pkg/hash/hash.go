package hash

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword 使用bcrypt算法加密密码
func HashPassword(password string) (string, error) {
	// 使用默认的cost值(10)生成哈希
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword 验证密码是否匹配
func VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}