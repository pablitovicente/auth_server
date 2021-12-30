package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v4"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/pablitovicente/auth_server/pkg/config"
	"github.com/pablitovicente/auth_server/pkg/db"
	"github.com/pablitovicente/auth_server/pkg/login"
	"github.com/pablitovicente/auth_server/pkg/users"
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
	users.DBPool = &db
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
	// TODO: move this code to a handler
	r.GET("/user/:id", func(c echo.Context) error {
		id := c.Param("id")
		// Acquire connection from pool
		conn, err := db.Pool.Acquire(context.TODO())
		if err != nil {
			fmt.Println("Error aquiring client from DB Pool", err)
			return err
		}
		// Defer the release of the client
		defer conn.Release()
		// Begin transaction
		tx, err := conn.BeginTx(context.TODO(), pgx.TxOptions{})
		if err != nil {
			return err
		}
		// Defer Commit/Rollback the way this works
		// is that if the err variable is set the TX
		// will be rolled back and if not it will
		// commit it
		defer func() {
			if err != nil {
				tx.Rollback(context.TODO())
			} else {
				tx.Commit(context.TODO())
			}
		}()

		query := "SELECT id, username FROM users WHERE id = $1"
		row := tx.QueryRow(context.TODO(), query, id)
		if err != nil {
			return err
		}

		type IdName struct {
			Id       int
			Username string
		}

		var myRow IdName

		err = row.Scan(&myRow.Id, &myRow.Username)
		return c.JSON(http.StatusOK, myRow)
	})

	r.POST("/user", users.AddHandler)

	// Start server
	if cfg.GetBool("http.sslEnabled") {
		if err := e.StartTLS(":3000", cfg.GetString("http.certFile"), cfg.GetString("http.certKey")); err != http.ErrServerClosed {
			e.Logger.Fatal(err)
		}
	} else {
		e.Logger.Fatal(e.Start(":" + cfg.GetString("http.port")))
	}
}
