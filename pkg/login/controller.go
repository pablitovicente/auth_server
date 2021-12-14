package login

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pablitovicente/auth_server/pkg/db"
)

// This package scoped globals don't look nice
// What is a better way of doing this?
var DBPool *db.Pool
var J *JWT

func Handler(c echo.Context) error {
	// Create an empty struct
	credentials := new(Credentials)
	// Try to get data from request
	if err := c.Bind(credentials); err != nil {
		return err
	}
	// Validate the login
	loginOk, dbUser := credentials.Validate(DBPool.Pool)

	if loginOk {
		signedToken, err := J.Generate(&dbUser)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, "Error generating token")
		}
		dbUser.JWT = signedToken

		c.Response().Header().Set("Authorization", "Bearer "+dbUser.JWT)
		return c.JSON(http.StatusOK, dbUser)
	}
	return c.JSON(http.StatusUnauthorized, "Invalid username or password!")
}
