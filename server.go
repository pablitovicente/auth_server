package main

import (
	"net/http"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/pablitovicente/auth_server/pkg/config"
	"github.com/pablitovicente/auth_server/pkg/db"
	"github.com/pablitovicente/auth_server/pkg/login"
)

func main() {
	// Bootstrap configuration
	cfg := config.Bootstrap(".")
	// DB connection and seed
	db := db.Pool{
		ConnString: "postgres://" + cfg.GetString("db.username") + ":" + cfg.GetString("db.password") + "@" + cfg.GetString("db.host") + ":" + cfg.GetString("db.port") + "/" + cfg.GetString("db.name"),
	}

	db.Connect()
	defer db.Pool.Close()
	db.SeedDB()

	// JWT Configuration
	jwto := login.JWT{
		Key:             cfg.GetString("jwt.secret"),
		ExpirationHours: cfg.GetInt("jwt.expirationHours"),
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

	login.DBPool = &db
	login.J = &jwto
	e.POST("/api/login", login.Handler)

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
	e.Logger.Fatal(e.Start(":" + cfg.GetString("http.port")))
}

// JWT custom error handler
func jwtError(err error, c echo.Context) error {
	return c.JSON(http.StatusUnauthorized, "JWT validation error: "+err.Error())
}
