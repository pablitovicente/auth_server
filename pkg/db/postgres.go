package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Pool struct {
	ConnString string
	Pool       *pgxpool.Pool
}

type User struct {
	Id       int
	Username string
	password string
	GroupId  int
}

func (dbp *Pool) Connect() {
	// HACK to wait for PG for quick POC
	// this should be a propper reconnection strategy
	time.Sleep(5 * time.Second)
	fmt.Println("going to attempt db connection....")
	// Setup DB Connection Pool
	dbpool, err := pgxpool.Connect(context.Background(), dbp.ConnString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("connected to db")
	dbp.Pool = dbpool
}

func (db *Pool) SeedDB() {
	// Create the "users" table.
	if _, err := db.Pool.Exec(context.Background(), "CREATE TABLE IF NOT EXISTS users ( id serial NOT NULL, PRIMARY KEY (id), username character varying(255) NOT NULL, password character varying(255) NOT NULL, groupid integer NOT NULL)"); err != nil {
		log.Fatal(err)
	}
	// Insert some rows int users
	if _, err := db.Pool.Exec(context.Background(),
		"INSERT INTO users (username, password, groupid) VALUES ('paul', 'testtest', 1963), ('george', 'testtest', 1963)"); err != nil {
		log.Fatal(err)
	}
}
