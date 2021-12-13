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
	// Might use this later...
	// password         string
	GroupId          int
	GroupName        string
	GroupDescription string
	JWT              string
}

func (dbp *Pool) Connect() {
	// HACK to wait for PG for quick POC
	// this should be a propper reconnection strategy
	time.Sleep(4 * time.Second)
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
	sql := "CREATE TABLE IF NOT EXISTS users ( id serial NOT NULL, PRIMARY KEY (id), username character varying(255) NOT NULL, password character varying(255) NOT NULL, groupid integer NOT NULL)"
	if _, err := db.Pool.Exec(context.Background(), sql); err != nil {
		log.Fatal(err)
	}
	// Create the "groups" table.
	sql = "CREATE TABLE IF NOT EXISTS groups (id integer NOT NULL, name character varying(255) NOT NULL, enabled integer NOT NULL, description character varying(1024) NOT NULL)"
	if _, err := db.Pool.Exec(context.Background(), sql); err != nil {
		log.Fatal(err)
	}

	// Insert some rows into groups
	sql = "INSERT INTO users (username, password, groupid) VALUES ('paul', 'testtest', 1963), ('george', 'testtest', 1963), ('john', 'liverpool', 1963), ('ringo', 'liverpool', 1963), ('greenwood', 'london', 1993)"
	if _, err := db.Pool.Exec(context.Background(), sql); err != nil {
		log.Fatal(err)
	}

	// Insert some rows into groups
	sql = "INSERT INTO groups (id, name, enabled, description) VALUES (1963, 'The Beatles', 1, 'Greatest band ever'), (1993, 'Radiohead', 1, 'Greatest alternative rock band ever')"
	if _, err := db.Pool.Exec(context.Background(), sql); err != nil {
		log.Fatal(err)
	}
}
