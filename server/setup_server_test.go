package server

import (
	"testing"
	"time"

	"github.com/diatbbin/QuotaFlow/util"
	_ "github.com/lib/pq"
	db "github.com/diatbbin/QuotaFlow/db/sqlc"
	"github.com/stretchr/testify/require"
)

func NewTestServer(t *testing.T, store *db.Store) *Server {
	config := util.Config{
		TokenSymmetricKey: util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := NewServer(store, config)
	require.NoError(t, err)

	return server
}
