package users

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pablitovicente/auth_server/pkg/db"
	"golang.org/x/crypto/bcrypt"
)

// This package scoped globals don't look nice
// What is a better way of doing this?
var DBPool *db.Pool

type User struct {
	Id       int
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
	// Hash the password
	err := HashPassword(user)
	if err != nil {
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

func ReadOneHandler(c echo.Context) error {
	id := c.Param("id")
	// Create an slice of User for Select to store into
	var users []*User
	// Get result
	DBPool.Select(nil, &users, "SELECT * FROM users WHERE id = $1", id)

	return c.JSON(http.StatusOK, users)
}

// TODO: add pagination and configuration for pagination
// or change to chunk transfer encoding and use pgxscan rowscaner
func ReadAllHandler(c echo.Context) error {
	// Create an slice of User for Select to store into
	var users []*User
	// Get result
	DBPool.Select(nil, &users, "SELECT * FROM users")

	return c.JSON(http.StatusOK, users)
}

func HashPassword(user *User) (err error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.MinCost)
	user.Password = string(bytes)
	return err
}
