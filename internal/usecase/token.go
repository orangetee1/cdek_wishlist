package usecase

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenManager interface {
	SignUserToken(userID int64) (string, error)
	ParseUserToken(token string) (int64, error)
}

type JWTManager struct {
	secret []byte
	expiry time.Duration
}

type userClaims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

func NewJWTManager(secret string, expiry time.Duration) JWTManager {
	return JWTManager{secret: []byte(secret), expiry: expiry}
}

func (j JWTManager) SignUserToken(userID int64) (string, error) {
	claims := userClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j JWTManager) ParseUserToken(tokenString string) (int64, error) {
	token, err := jwt.ParseWithClaims(tokenString, &userClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.secret, nil
	})
	if err != nil {
		return 0, err
	}
	claims, ok := token.Claims.(*userClaims)
	if !ok || !token.Valid {
		return 0, jwt.ErrTokenInvalidClaims
	}
	return claims.UserID, nil
}
