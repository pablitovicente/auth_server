package users

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pablitovicente/auth_server/pkg/db"
)

// This package scoped globals don't look nice
// What is a better way of doing this?
var DBPool *db.Pool

type User struct {
	Username string `json:"name"`
	Password string `json:"password"`
	Groupid  int    `json:"group"`
}

func AddHandler(c echo.Context) error {
	// Create an empty User
	user := new(User)
	// Try to get data from request
	if err := c.Bind(user); err != nil {
		return err
	}
	// Acquire connection from pool
	conn, err := DBPool.GetConnection()
	if err != nil {
		return err
	}
	// Defer the release of the client
	defer conn.Release()
	// Begin transaction
	tx, err := DBPool.NewTx(conn)
	if err != nil {
		return err
	}
	// Defer Commit/Rollback the way this works
	// is that if the err variable is set the TX
	// will be rolled back and if not it will
	// commit it. We use a pointer so the deferred
	// function sees the last value of err
	defer DBPool.CommitOrRollback(tx, &err)
	// Prepare the query and execute
	query := "INSERT INTO users (username, password, groupid) VALUES ($1, $2, $3)"
	commandTag, err := DBPool.Exec(tx, query, user.Username, user.Password, user.Groupid)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, commandTag.RowsAffected())
}
