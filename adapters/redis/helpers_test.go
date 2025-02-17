package redis

import (
	"io"
	"log"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func init() {
	// 將日誌輸出重定向到io.Discard
	log.SetOutput(io.Discard)
}

func setupTest(t *testing.T) (*redis.Client, redismock.ClientMock, func()) {
	db, mock := redismock.NewClientMock()
	return db, mock, func() {
		assert.NoError(t, mock.ExpectationsWereMet())
		db.Close()
	}
}

type TestMessage struct {
	ID   string `json:"id"`
	Data string `json:"data"`
}
