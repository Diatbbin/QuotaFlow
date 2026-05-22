package sqlc_test

import (
	"context"
	"database/sql"
	"testing"

	db "github.com/diatbbin/QuotaFlow/db/sqlc"
	"github.com/diatbbin/QuotaFlow/util"
	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) db.User {
	hashedPassword, err := util.HashPassword(util.RandomString(8))
	require.NoError(t, err)
	
	arg := db.CreateUserParams{
		Email:        util.RandomEmail(),
		Username:     util.RandomString(8),
		PasswordHash: hashedPassword,
	}

	user, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Email, user.Email)
	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.PasswordHash, user.PasswordHash)

	require.NotZero(t, user.ID)
	require.NotZero(t, user.CreatedAt)

	return user
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	user1 := createRandomUser(t)

	user2, err := testQueries.GetUser(context.Background(), user1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.ID, user2.ID)
	require.Equal(t, user1.Email, user2.Email)
	require.Equal(t, user1.Username, user2.Username)
	require.Equal(t, user1.PasswordHash, user2.PasswordHash)
	require.WithinDuration(t, user1.CreatedAt, user2.CreatedAt, 0)
}

func TestGetUserNotFound(t *testing.T) {
	user, err := testQueries.GetUser(context.Background(), 999_999)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, user)
}

func TestCreateUserDuplicateEmail(t *testing.T) {
	user1 := createRandomUser(t)

	_, err := testQueries.CreateUser(context.Background(), db.CreateUserParams{
		Email:        user1.Email,
		Username:     util.RandomString(8),
		PasswordHash: util.RandomString(32),
	})
	require.Error(t, err)
}

func TestCreateUserDuplicateUsername(t *testing.T) {
	user1 := createRandomUser(t)

	_, err := testQueries.CreateUser(context.Background(), db.CreateUserParams{
		Email:        util.RandomString(10) + "@test.com",
		Username:     user1.Username,
		PasswordHash: util.RandomString(32),
	})
	require.Error(t, err)
}
