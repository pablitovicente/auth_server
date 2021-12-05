package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

type Login struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type sanitizedLogin struct {
	Username string
	Password string
	isAdmin  bool
}

type dbUser struct {
	Id       int
	Username string
	password string
	GroupId  int
}

func main() {
	// HACK to wait for PG for quick POC
	// this should be a propper reconnection strategy
	time.Sleep(5 * time.Second)
	fmt.Println("going to attempt db connection....")
	// Setup DB Connection Pool
	dbpool, err := pgxpool.Connect(context.Background(), "postgres://test:test1234@db:5432/auth_server")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()
	fmt.Println("connected to db")

	// Create the "users" table.
	if _, err := dbpool.Exec(context.Background(), "CREATE TABLE IF NOT EXISTS users ( id serial NOT NULL, PRIMARY KEY (id), username character varying(255) NOT NULL, password character varying(255) NOT NULL, groupid integer NOT NULL)"); err != nil {
		log.Fatal(err)
	}
	// Insert some rows int users
	if _, err := dbpool.Exec(context.Background(),
		"INSERT INTO users (username, password, groupid) VALUES ('paul', 'testtest', 1963), ('george', 'testtest', 1963)"); err != nil {
		log.Fatal(err)
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

	e.POST("/login", func(c echo.Context) error {
		// Create an empty struct
		login := new(Login)
		// Try to get data from request
		if err := c.Bind(login); err != nil {
			return err
		}
		// For security remap fields (for example imagine the target struct
		// supports an isAdmin field which could cause privilege scalation)
		// not required for read-only operations but using as example
		cleanLogin := sanitizedLogin{
			Username: login.Username,
			Password: login.Password,
			isAdmin:  false,
		}

		loginOk, dbUser := loginUser(cleanLogin, dbpool)

		if loginOk {
			c.Response().Header().Set("Authorization", "Bearer 12314")
			return c.JSON(http.StatusOK, dbUser)
		}
		return c.JSON(http.StatusUnauthorized, dbUser)
	})

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}

func loginUser(credentials sanitizedLogin, db *pgxpool.Pool) (bool, dbUser) {
	row, err := db.Query(context.Background(), "SELECT id, username, groupid FROM users WHERE username = $1 AND password = $2", credentials.Username, credentials.Password)

	if err != nil {
		log.Error("failed to query DB", err)
	}
	defer row.Close()

	if row.Err() != nil {
		log.Error("row has error", err)
	}

	if !row.Next() {
		log.Error("Login failed")
		emptyUser := dbUser{}
		return false, emptyUser
	}

	var found dbUser
	row.Scan(&found.Id, &found.Username, &found.GroupId)

	return true, found
}
