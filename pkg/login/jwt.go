package login

import (
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/pablitovicente/auth_server/pkg/db"
)

type JWT struct {
	Key             string
	ExpirationHours int
}

func (j JWT) Generate(user db.User) (string, error) {
	// Perhaps better handling would be better
	if j.Key == "" {
		panic("JWT signing key is empty not safe to continue")
	}
	// Set custom claims
	claims := &JwtCustomClaims{
		user,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * time.Duration(j.ExpirationHours)).Unix(),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	signedToken, err := token.SignedString([]byte(j.Key))
	if err != nil {
		return signedToken, err
	}

	return signedToken, nil
}
