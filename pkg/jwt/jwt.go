package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/reusedev/uportal-api/pkg/config"
)

var (
	ErrInvalidToken = errors.New("invalid token")
)

// Claims 自定义的 JWT 声明
type Claims struct {
	UserID   int64  `json:"user_id"`
	Sub      string `json:"sub"`
	IsAdmin  bool   `json:"is_admin"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken 生成 JWT token
func GenerateToken(userID int64, isAdmin bool, userName, password, role string) (string, error) {
	claims := Claims{
		UserID:   userID,
		IsAdmin:  isAdmin,
		Username: userName,
		Password: password,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // token 有效期 24 小时
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Get().JWT.Secret))
}

// ParseToken 解析 JWT token
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Get().JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}
