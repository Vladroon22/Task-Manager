package database

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"
)

type DataBase struct {
	pool *sql.DB
}

func NewDB() *DataBase {
	return &DataBase{}
}

func (d *DataBase) Connect(ctx context.Context) (*sql.DB, error) {
	dbURL := os.Getenv("DB")
	if dbURL == "" {
		log.Fatalln("dbURL doesn't set")
	}

	pool, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}
	if err := ping(pool); err != nil {
		return nil, err
	}
	d.pool = pool

	log.Println("Database connection is valid")
	return pool, nil
}

func ping(cn *sql.DB) error {
	var err error
	for i := 0; i < 5; i++ {
		if err = cn.Ping(); err == nil {
			return nil
		}
		time.Sleep(time.Millisecond * 500)
	}
	return err
}

func (d *DataBase) Close() {
	d.pool.Close()
}
