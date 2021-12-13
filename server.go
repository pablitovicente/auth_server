package main

import (
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/pablitovicente/auth_server/pkg/db"
	"github.com/pablitovicente/auth_server/pkg/login"
	config "github.com/spf13/viper"
)

func main() {
	// Configs
	config.SetConfigType("json")
	config.SetConfigName("config")
	config.AddConfigPath("./")

	err := config.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	httpPort := config.GetString("http.port")

	// DB connection and seed
	db := db.Pool{
		ConnString: "postgres://" + config.GetString("db.username") + ":" + config.GetString("db.password") + "@" + config.GetString("db.host") + ":" + config.GetString("db.port") + "/" + config.GetString("db.name"),
	}

	db.Connect()
	defer db.Pool.Close()
	db.SeedDB()
	// JWT Configuration
	jwto := login.JWT{
		Key:             config.GetString("jwt.secret"),
		ExpirationHours: config.GetInt("jwt.expirationHours"),
	}

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

	e.POST("/api/login", func(c echo.Context) error {
		// Create an empty struct
		credentials := new(login.Credentials)
		// Try to get data from request
		if err := c.Bind(credentials); err != nil {
			return err
		}
		// Validate the login
		loginOk, dbUser := credentials.Validate(db.Pool)

		if loginOk {
			signedToken, err := jwto.Generate(dbUser)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, "Error generating token")
			}
			dbUser.JWT = signedToken

			c.Response().Header().Set("Authorization", "Bearer "+dbUser.JWT)
			return c.JSON(http.StatusOK, dbUser)
		}
		return c.JSON(http.StatusUnauthorized, "Invalid username or password!")
	})

	// Echo Group of JWT protected routes
	r := e.Group("/api")
	// Configure middleware with the custom claims type
	config := middleware.JWTConfig{
		Claims:                  &login.JwtCustomClaims{},
		SigningKey:              []byte(jwto.Key),
		ErrorHandlerWithContext: jwtError,
	}
	// Attach JWT middleware to route group
	r.Use(middleware.JWTWithConfig(config))

	r.GET("/test", func(c echo.Context) error {
		// Get the JWT from the Context
		decodedJWT := c.Get("user").(*jwt.Token)
		// Extract Claims
		claims := decodedJWT.Claims.(*login.JwtCustomClaims)
		// Use one of the claims as example
		name := claims.User.Username
		group := claims.User.GroupName

		return c.JSON(http.StatusOK, "Welcome "+name+" from "+group)
	})

	// Start server
	e.Logger.Fatal(e.Start(":" + httpPort))
}

// JWT custom error handler
func jwtError(err error, c echo.Context) error {
	return c.JSON(http.StatusUnauthorized, "JWT validation error: "+err.Error())
}
