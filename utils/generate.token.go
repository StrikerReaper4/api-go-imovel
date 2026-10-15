package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// A chave vem da variável JWT_SECRET (systemd). O valor fixo é só fallback.
func jwtSecret() []byte {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return []byte(s)
	}
	return []byte("293829452039stghsetg")
}

func GenerateToken(userId int, email string, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userId,
		"role":    role,
		"email":   email,
		"exp":     time.Now().Add(time.Hour * 8).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenStr, err := token.SignedString(jwtSecret())

	if err != nil {
		return "", err
	}

	return tokenStr, nil
}

func ValidateToken(tokenStr string) (jwt.MapClaims, error) {

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("token inválido")
		}
		return jwtSecret(), nil
	})

	if err != nil {
		return nil, fmt.Errorf("token inválido")
	}

	if !token.Valid || token == nil {
		return nil, fmt.Errorf("token inválido")
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return nil, fmt.Errorf("token invalido")
	}

	return claims, nil

}
