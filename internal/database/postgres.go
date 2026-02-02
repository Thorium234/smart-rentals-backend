package database

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq" //called for side effects
)

type Database struct {
	*sql.DB
}

func NewDatabase(connectionString string) (*Database, error) {
	//Open database connection
	db, err := sql.Open("postgres", connectionString) //sql driver
	if err != nil {
		return nil, err
	}

	//configure connection pool
	db.SetMaxOpenConns(25)                 //limit maximum sumultaneous connections
	db.SetMaxIdleConns(4)                  //keep some connections ready
	db.SetConnMaxLifetime(5 * time.Minute) //refresh connections periodically

	//verify connection is working
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Database{db}, nil
}
