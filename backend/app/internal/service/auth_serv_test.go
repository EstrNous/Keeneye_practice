package service

import (
	"testing"
	"time"

	"keeneye_practice/app/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

// Мы не мокаем JWT! Мы тестируем его реальную работу.
func TestJWTGenerationAndParsing(t *testing.T) {
	secret := "test-secret-key"

	// 1. СОЗДАЕМ РЕАЛЬНЫЙ ТОКЕН (как это делает твой сервис)
	claims := domain.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID:    42,
		Role:      "admin",
		ProfileID: 99,
	}

	realToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := realToken.SignedString([]byte(secret))
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// 2. ПАРСИМ РЕАЛЬНЫЙ ТОКЕН ОБРАТНО (как это делает твой мидлварь)
	parsedClaims := &domain.Claims{}
	parsedToken, err := jwt.ParseWithClaims(tokenString, parsedClaims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	// 3. ПРОВЕРЯЕМ, ЧТО РЕАЛЬНЫЙ ТОКЕН РАБОТАЕТ ПРАВИЛЬНО
	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	// САМОЕ ГЛАВНОЕ: проверяем, что данные не исказились
	assert.Equal(t, int32(42), parsedClaims.UserID, "UserID в токене должен быть 42")
	assert.Equal(t, "admin", parsedClaims.Role, "Роль должна быть admin")
	assert.Equal(t, int32(99), parsedClaims.ProfileID, "ProfileID должен быть 99")
}
