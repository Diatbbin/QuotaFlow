package main

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/diatbbin/QuotaFlow/util"
	_ "github.com/lib/pq"
	db "github.com/diatbbin/QuotaFlow/db/sqlc"
)

var testQueries *db.Queries
var testDB *sql.DB

func TestMain(m *testing.M) {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("Error loading config:", err)
	}

	testConn, err = sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	testQueries = db.New(testConn)
	os.Exit(m.Run())
}
