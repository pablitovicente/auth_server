package login

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/gommon/log"
	"github.com/pablitovicente/auth_server/pkg/db"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// For security remap fields (for example imagine the target struct
// supports an isAdmin field which could cause privilege scalation)
// not required for read-only operations but using as example
// cleanLogin := sanitizedLogin{
// 	Username: login.Username,
// 	Password: login.Password,
// 	isAdmin:  false,
// }
type sanitizedLogin struct {
	Username string
	Password string
	isAdmin  bool
}

func (c *Credentials) Execute(dbp *pgxpool.Pool) (bool, db.User) {
	row, err := dbp.Query(context.Background(), "SELECT id, username, groupid FROM users WHERE username = $1 AND password = $2", c.Username, c.Password)

	if err != nil {
		log.Error("failed to query DB", err)
	}
	defer row.Close()

	if row.Err() != nil {
		log.Error("row has error", err)
	}

	if !row.Next() {
		log.Error("Login failed")
		emptyUser := db.User{}
		return false, emptyUser
	}

	var found db.User
	row.Scan(&found.Id, &found.Username, &found.GroupId)

	return true, found
}
