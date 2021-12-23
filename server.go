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

	// JWT Object
	jwto := login.JWT{
		ExpirationHours: cfg.GetInt("jwt.expirationHours"),
		MiddleWareConfig: middleware.JWTConfig{
			Claims:                  &login.JwtCustomClaims{},
			SigningKey:              []byte(cfg.GetString("jwt.secret")),
			ErrorHandlerWithContext: login.JWTError,
		},
	}
	// Need to find idiomatic way of sharing this...
	login.DBPool = &db
	login.J = &jwto
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

	// Echo Group of JWT protected routes
	r := e.Group("/api")
	// Attach JWT middleware just to this route group
	r.Use(middleware.JWTWithConfig(jwto.MiddleWareConfig))

	// Routes
	e.POST("/api/login", login.Handler)

	r.GET("/test", func(c echo.Context) error {
		// Get the JWT from the Context
		permissions := jwto.Decode(c.Get("user").(*jwt.Token))
		// Use one of the claims as example
		return c.JSON(http.StatusOK, "Welcome "+permissions.User.Username+" from "+permissions.User.GroupName)
	})

	// Start server
	if cfg.GetBool("http.sslEnabled") {
		if err := e.StartTLS(":3000", cfg.GetString("http.certFile"), cfg.GetString("http.certKey")); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	} else {
		e.Logger.Fatal(e.Start(":" + cfg.GetString("http.port")))
	}
}
