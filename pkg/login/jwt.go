package login

import (
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pablitovicente/auth_server/pkg/db"
)

type JWT struct {
	ExpirationHours  int
	MiddleWareConfig middleware.JWTConfig
}

type JwtCustomClaims struct {
	User db.User
	jwt.StandardClaims
}

func (j JWT) Generate(user *db.User) (string, error) {
	// Perhaps better handling would be better
	if j.MiddleWareConfig.SigningKey == "" {
		panic("JWT signing key is empty not safe to continue")
	}
	// Set custom claims
	claims := &JwtCustomClaims{
		*user,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * time.Duration(j.ExpirationHours)).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Generate encoded token and send it as response.
	signedToken, err := token.SignedString(j.MiddleWareConfig.SigningKey)
	if err != nil {
		return signedToken, err
	}

	return signedToken, nil
}

func (j JWT) Decode(token *jwt.Token) JwtCustomClaims {
	claims := token.Claims.(*JwtCustomClaims)
	return *claims
}
