package worker

import (
	"context"

	"github.com/hibiken/asynq"
)

type Distributor interface {
	DistributorTaskSendTransferEmail(
		ctx context.Context,
		payload *PayloadSendTransferEmail,
		opts ...asynq.Option,
	) error
}

type RedisTaskDistributor struct {
	client *asynq.Client
}

func NewRedisTaskDistributor(redisOpt asynq.RedisClientOpt) *RedisTaskDistributor {
	client := asynq.NewClient(redisOpt)
	return &RedisTaskDistributor{
		client: client,
	}
}