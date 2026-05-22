package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/diatbbin/QuotaFlow/util"
	"github.com/diatbbin/QuotaFlow/server"
	db "github.com/diatbbin/QuotaFlow/db/sqlc"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("Error loading config:", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := db.NewStore(conn)
	server := server.NewServer(store)
	err = server.Start(config.ServerAddress)

	if err != nil {
		log.Fatal("Server did not start successfully:", err)
	}
}
