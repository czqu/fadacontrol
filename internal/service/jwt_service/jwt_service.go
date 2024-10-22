package jwt_service

import (
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/entity"
	"fadacontrol/internal/schema"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
	"time"
)

const ExpirationTime = 7 * 24 * time.Hour

const DefaultJwtSecretKey = "undefined"

var jwtSecretKey = DefaultJwtSecretKey

type JwtService struct {
	_db *gorm.DB
}

func NewJwtService(_db *gorm.DB) *JwtService {
	return &JwtService{_db}
}
func (j *JwtService) RefreshKey() {
	if jwtSecretKey != DefaultJwtSecretKey {
		return
	}
	user := &entity.User{}
	if err := j._db.Where(&entity.User{Username: "root"}).Select("password").First(user).Error; err != nil {
		logger.Error(err)
	} else {
		jwtSecretKey = user.Password
	}
}
func (j *JwtService) GenerateToken(username string) (string, error) {
	j.RefreshKey()

	claims := schema.JwtClaims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ExpirationTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecretKey))
}
func (j *JwtService) ValidateToken(tokenString string) (*schema.JwtClaims, error) {
	j.RefreshKey()
	token, err := jwt.ParseWithClaims(tokenString, &schema.JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecretKey), nil
	})

	if claims, ok := token.Claims.(*schema.JwtClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}
