package server

import (
	"testing"
	"time"

	"github.com/diatbbin/QuotaFlow/util"
	_ "github.com/lib/pq"
	db "github.com/diatbbin/QuotaFlow/db/sqlc"
	"github.com/stretchr/testify/require"
	worker "github.com/diatbbin/QuotaFlow/worker"
	"github.com/hibiken/asynq"
)

func NewTestServer(t *testing.T, store *db.Store) *Server {
	config := util.Config{
		TokenSymmetricKey: util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	testRedisOpt := asynq.RedisClientOpt{
		Addr: config.RedisAddr,
	}
	testDistributor := worker.NewRedisTaskDistributor(testRedisOpt)

	server, err := NewServer(store, config, testDistributor)
	require.NoError(t, err)

	return server
}
