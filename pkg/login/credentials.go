package login

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/gommon/log"
	"github.com/pablitovicente/auth_server/pkg/db"
	"golang.org/x/crypto/bcrypt"
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
// Commenting out for now
// type sanitizedLogin struct {
// 	Username string
// 	Password string
// 	isAdmin  bool
// }
// @TODO generate random keys for JWT signing for each login
func (c *Credentials) ValidateOut(dbp *pgxpool.Pool) (bool, db.User) {
	sql := "SELECT u.id, username, g.id AS gId, g.name, g.description FROM users u INNER JOIN groups g ON u.groupid = g.id WHERE username = $1 AND password = $2"
	row, err := dbp.Query(context.Background(), sql, c.Username, c.Password)
	var found db.User

	if err != nil {
		log.Error("Failed to query DB:", err)
	}
	defer row.Close()

	if row.Err() != nil {
		log.Error("Can't read DB row error was: ", err)
	}

	if !row.Next() {
		log.Error("Login failed")
		return false, found
	}

	err = row.Scan(&found.Id, &found.Username, &found.GroupId, &found.GroupName, &found.GroupDescription)
	if err != nil {
		log.Error("Error parsing authenticated user:", err)
	}

	if err != nil {
		return false, found
	}

	return true, found
}

// TODO: Normalize this struct for User in a single place and replace all over the code
type User struct {
	Id               int
	Username         string
	Password         string
	GroupId          int    `db:"groupid"`
	GroupName        string `db:"groupname"`
	GroupDescription string `db:"groupdescription"`
}

func (c *Credentials) Validate(dbp *pgxpool.Pool) (bool, db.User) {
	query := "SELECT u.id as id, username, password, g.id AS groupid, g.name AS groupname, g.description AS groupdescription FROM users u INNER JOIN groups g ON u.groupid = g.id WHERE username = $1 LIMIT 1"
	var user db.User

	var found []*User
	err := DBPool.Select(nil, &found, query, c.Username)
	// Check for scan error
	if err != nil {
		return false, user
	}
	// Check if no user found
	if len(found) < 1 {
		return false, user
	}
	// Check if the provided password match
	if passwordMatch := CheckPasswordHash(c.Password, found[0].Password); !passwordMatch {
		return false, user
	}
	// If we reached this point user exists and password is valid
	user = db.User{
		Username:         found[0].Username,
		Id:               found[0].Id,
		GroupName:        found[0].GroupName,
		GroupId:          found[0].GroupId,
		GroupDescription: found[0].GroupDescription,
	}

	return true, user
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
