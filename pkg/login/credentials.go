package login

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/gommon/log"
	"github.com/pablitovicente/auth_server/pkg/db"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type jwtCustomClaims struct {
	User db.User
	jwt.StandardClaims
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
	sql := "SELECT u.id, username, g.id AS gId, g.name, g.description FROM users u INNER JOIN groups g ON u.groupid = g.id WHERE username = $1 AND password = $2"
	row, err := dbp.Query(context.Background(), sql, c.Username, c.Password)

	if err != nil {
		log.Error("Failed to query DB:", err)
	}
	defer row.Close()

	if row.Err() != nil {
		log.Error("Can't read DB row error was: ", err)
	}

	if !row.Next() {
		log.Error("Login failed")
		emptyUser := db.User{}
		return false, emptyUser
	}

	var found db.User
	err = row.Scan(&found.Id, &found.Username, &found.GroupId, &found.GroupName, &found.GroupDescription)
	if err != nil {
		log.Error("Error parsing authenticated user:", err)
	}

	// Set custom claims
	claims := &jwtCustomClaims{
		found,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte("THIS_SECRET_SHOULD_BE_A_COMMAND_LINE_ARGUMENT_INJECTED_TO_THE_OWNER_STRUCT"))
	found.JWT = t

	if err != nil {
		emptyUser := db.User{}
		return false, emptyUser
	}

	return true, found
}
