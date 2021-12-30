package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Pool struct {
	ConnString string
	Pool       *pgxpool.Pool
}

type User struct {
	Id               int
	Username         string
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

func (dbp *Pool) GetConnection() (conn *pgxpool.Conn, err error) {
	return dbp.Pool.Acquire(context.TODO())
}

func (dbp *Pool) NewTx(conn *pgxpool.Conn) (tx pgx.Tx, err error) {
	return conn.BeginTx(context.TODO(), pgx.TxOptions{})
}

func (dbp *Pool) CommitOrRollback(tx pgx.Tx, err *error) (er error) {
	if *err != nil {
		er = tx.Rollback(context.TODO())
	} else {
		er = tx.Commit(context.TODO())
	}
	return
}

func (dbp *Pool) Exec(tx pgx.Tx, query string, values ...interface{}) (tag pgconn.CommandTag, err error) {
	commandTag, err := tx.Exec(context.TODO(), query, values...)

	if err != nil {
		return nil, err
	}

	return commandTag, nil
}

// Store should be an slice of pointers to a type. For example if we have a type Foo store will be []*Foo
func (dbp *Pool) Select(tx pgx.Tx, store interface{}, query string, values ...interface{}) (err error) {
	var er error
	if values == nil {
		er = pgxscan.Select(context.Background(), dbp.Pool, store, query)
	} else {
		er = pgxscan.Select(context.Background(), dbp.Pool, store, query, values...)
	}

	if er != nil {
		// TODO: replace this fmt.print with better logging
		fmt.Println("error on pgxscan", er)
		return er
	}
	return nil
}

// TODO: remove seeding and move somewhere else
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
