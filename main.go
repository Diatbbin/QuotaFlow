package main

import (
	"database/sql"

	log "github.com/rs/zerolog/log"
	_ "github.com/lib/pq"
	"github.com/diatbbin/QuotaFlow/mail"
	"github.com/diatbbin/QuotaFlow/util"
	"github.com/diatbbin/QuotaFlow/server"
	db "github.com/diatbbin/QuotaFlow/db/sqlc"
	worker "github.com/diatbbin/QuotaFlow/worker"
	"github.com/hibiken/asynq"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("error loading config")
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("cannot connect to db")
	}

	redisOpt := asynq.RedisClientOpt{
		Addr: config.RedisAddr,
	}

	distributor := worker.NewRedisTaskDistributor(redisOpt)
	store := db.NewStore(conn)
	mailSender := mail.NewGmailSender(config.EmailSenderName, config.EmailSenderAddr, config.EmailPassword)
	go runTaskProcessor(redisOpt, store, mailSender)
	
	server, err := server.NewServer(store, config, distributor)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("error creating server")
	}
	
	err = server.Start(config.ServerAddress)
	

	if err != nil {
		log.Fatal().
			Err(err).
			Msg("server did not start successfully")
	}
}

func runTaskProcessor(redisOpt asynq.RedisClientOpt, store *db.Store, mailSender mail.MailSender) {
	processor := worker.NewRedisTaskProcessor(redisOpt, store, mailSender)

	log.Info().Msg("Starting task processor")
	err := processor.Start()
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("error starting task processor")
	}
}
