package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/diatbbin/QuotaFlow/server"
	db "github.com/diatbbin/QuotaFlow/db/sqlc"
)

const (
	dbDriver = "postgres"
	dbSource = "postgres://root:test@localhost:5432/QuotaFlow?sslmode=disable"
	serverAddress = "0.0.0.0:8080"
)

func main() {
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := db.NewStore(conn)
	server := server.NewServer(store)
	err = server.Start(serverAddress)

	if err != nil {
		log.Fatal("Server did not start successfully:", err)
	}
}