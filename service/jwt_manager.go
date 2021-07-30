package service

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"time"
)

// JWTManager is a json web token manager
type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}

// UserClaims is a custom jwt claims that contains some user's information
type UserClaims struct {
	jwt.StandardClaims
	Username string `json:"username"`
	Role     string `json:"role"`
}

// NewJWTManager returns a new JWT manager
func NewJWTManager(secretkey string, tokenDuration time.Duration) *JWTManager {

	return &JWTManager{
		secretKey:     secretkey,
		tokenDuration: tokenDuration,
	}
}

// Generate generates and signs a new token for a user
func (manager *JWTManager) Generate(user *User) (string, error) {
	claims := UserClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(manager.tokenDuration).Unix(),
		},
		Username: user.UserName,
		Role:     user.Role,
	}
	// ***** consider more complex signing method in production
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(manager.secretKey))
}

// Verify verify a access token and return user claims
func (manager *JWTManager) Verify(accessToken string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(
		accessToken,
		&UserClaims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, fmt.Errorf("unexcept token signing method")
			}
			return []byte(manager.secretKey), nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}
