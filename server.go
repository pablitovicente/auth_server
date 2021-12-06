package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/pablitovicente/auth_server/pkg/db"
	"github.com/pablitovicente/auth_server/pkg/login"
)

func main() {
	db := db.Pool{
		ConnString: "postgres://test:test1234@db:5432/auth_server",
	}

	db.Connect()
	defer db.Pool.Close()
	db.SeedDB()

	// Echo instance
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	// Middleware
	// Logger disabled as it hit peformance bad
	// e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// This level configuration affects
	// calls to e.Logger.<Debug, Infor, etc.>
	e.Logger.SetLevel(log.OFF)

	e.POST("/login", func(c echo.Context) error {
		// Create an empty struct
		credentials := new(login.Credentials)
		// Try to get data from request
		if err := c.Bind(credentials); err != nil {
			return err
		}
		// Execute the login
		loginOk, dbUser := credentials.Execute(db.Pool)

		if loginOk {
			c.Response().Header().Set("Authorization", "Bearer 12314")
			return c.JSON(http.StatusOK, dbUser)
		}
		return c.JSON(http.StatusUnauthorized, "Invalid username or password!")
	})

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}
