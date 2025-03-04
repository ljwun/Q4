package openapi

import (
	"crypto"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type JWT struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func ParseAndValidateJWT(tokenString string, secret crypto.Signer) (*JWT, error) {
	const op = "ParseJWT"
	token, err := jwt.ParseWithClaims(tokenString, &JWT{}, func(token *jwt.Token) (interface{}, error) {
		return secret.Public(), nil
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("%s: token is invalid", op)
	}
	claims, ok := token.Claims.(*JWT)
	if !ok {
		return nil, fmt.Errorf("%s: token claims are invalid", op)
	}
	return claims, nil
}
