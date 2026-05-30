package worker

import (
	"context"

	"github.com/hibiken/asynq"
	log "github.com/rs/zerolog/log"
	db "github.com/diatbbin/QuotaFlow/db/sqlc"
	"github.com/diatbbin/QuotaFlow/mail"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
)

type TaskProcessor interface {
	Start() error
	ProcessTaskSendTransferEmail(ctx context.Context, task *asynq.Task) error
}

type RedisTaskProcessor struct {
	server     *asynq.Server
	store      *db.Store
	mailSender mail.MailSender
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, store *db.Store, mailSender mail.MailSender) *RedisTaskProcessor {
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Queues: map[string]int{
				QueueCritical: 10,
				QueueDefault:  5,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Error().
					Err(err).
					Str("type", task.Type()).
					Bytes("payload", task.Payload()).
					Msg("error processing task")
			}),
		},
	)

	return &RedisTaskProcessor{
		server:     server,
		store:      store,
		mailSender: mailSender,
	}
}

func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskSendTransferEmail, processor.ProcessTaskSendTransferEmail)
	return processor.server.Start(mux)
}
