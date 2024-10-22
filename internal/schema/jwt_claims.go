package schema

import "github.com/golang-jwt/jwt/v5"

type JwtClaims struct {
	Username string `json:"user_name,omitempty"`
	jwt.RegisteredClaims
}
