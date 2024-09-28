package jwt_service

import (
	"fadacontrol/internal/schema"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

const ExpirationTime = 7 * 24 * time.Hour
const JwtSecretKey = "test"

type JwtService struct {
}

func NewJwtService() *JwtService {
	return &JwtService{}
}
func (j *JwtService) GenerateToken(username string) (string, error) {
	claims := schema.JwtClaims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ExpirationTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(JwtSecretKey))
}
func (j *JwtService) ValidateToken(tokenString string) (*schema.JwtClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &schema.JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(JwtSecretKey), nil
	})

	if claims, ok := token.Claims.(*schema.JwtClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}
