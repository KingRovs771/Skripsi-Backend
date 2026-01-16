package utils

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type CustomClaims struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	UserType string `json:"user_type"`
	jwt.RegisteredClaims
}

var BlacklistedTokens = make(map[string]time.Time)
var mu sync.Mutex

func BlacklistToken(token string, expiry time.Time) error {
	mu.Lock()
	defer mu.Unlock()

	BlacklistedTokens[token] = expiry

	go cleanupExpiredTokens()

	return nil
}

func IsTokenBlacklisted(token string) bool {
	mu.Lock()
	defer mu.Unlock()

	exp, exists := BlacklistedTokens[token]
	if !exists {
		return false
	}

	if time.Now().After(exp) {
		delete(BlacklistedTokens, token)
		return false
	}

	return true
}

func cleanupExpiredTokens() {
	mu.Lock()
	defer mu.Unlock()

	now := time.Now()
	for token, expiry := range BlacklistedTokens {
		if now.After(expiry) {
			delete(BlacklistedTokens, token)
		}
	}
}
func GenerateJWT(id, email, userType string) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", errors.New("variabel lingkungan JWT_SECRET belum di-set")
	}

	tokenLifespanStr := os.Getenv("TOKEN_HOUR_LIFESPAN")
	tokenLifespan, err := strconv.Atoi(tokenLifespanStr)
	if err != nil {
		tokenLifespan = 1 // Default 1 jam
	}

	claims := CustomClaims{
		ID:       id,
		Email:    email,
		UserType: userType, // Tipe user: "admin", "student", "pakar", "teacher"
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(tokenLifespan))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func GetTokenFromRequest(c *gin.Context) string {
	bearerToken := c.Request.Header.Get("Authorization")
	splitToken := strings.Split(bearerToken, " ")
	if len(splitToken) == 2 && strings.ToLower(splitToken[0]) == "bearer" {
		return splitToken[1]
	}
	return ""
}

func ValidateJWT(c *gin.Context) (*CustomClaims, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET belum di-set")
	}

	tokenString := GetTokenFromRequest(c)
	if tokenString == "" {
		return nil, errors.New("token otorisasi tidak disediakan")
	}
	if IsTokenBlacklisted(tokenString) {
		return nil, errors.New("token telah logout dan tidak valid")
	}
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("metode signing tidak terduga")
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("token tidak valid, silakan ulangi lagi")
}
