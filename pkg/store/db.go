package store

import (
	"context"
	"github.com/jackc/pgx/v5"
	"log"
)

func OpenDB(connString string) (*pgx.Conn, error) {
	conn, err := pgx.Connect(context.Background(), connString)

	if err != nil {
		return nil, err
	}

	return conn, nil
}

func CloseDB(conn *pgx.Conn) {
	if err := conn.Close(context.Background()); err != nil {
		log.Printf("Failed to close connection %v\n", err)
	}
}
