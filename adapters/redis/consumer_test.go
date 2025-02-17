package redis

import (
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestNewConsumer(t *testing.T) {
	client, _, cleanup := setupTest(t)
	defer cleanup()

	tests := []struct {
		name    string
		client  *redis.Client
		stream  string
		opts    []ConsumerOption[TestMessage]
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid configuration",
			client:  client,
			stream:  "test-stream",
			wantErr: false,
		},
		{
			name:    "nil client",
			client:  nil,
			stream:  "test-stream",
			wantErr: true,
			errMsg:  "redis client cannot be nil",
		},
		{
			name:    "empty stream",
			client:  client,
			stream:  "",
			wantErr: true,
			errMsg:  "stream cannot be empty",
		},
		{
			name:   "with all options",
			client: client,
			stream: "test-stream",
			opts: []ConsumerOption[TestMessage]{
				WithConsumerLogger[TestMessage](slog.Default()),
				WithConsumerBufferSize[TestMessage](200),
				WithConsumerBlockTimeout[TestMessage](2 * time.Second),
				WithConsumerParseFunc[TestMessage](func(m map[string]any) (TestMessage, error) {
					return TestMessage{}, nil
				}),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer goleak.VerifyNone(t)

			consumer, err := NewConsumer[TestMessage](tt.client, tt.stream, tt.opts...)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, consumer)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, consumer)
				consumer.Close()
			}
		})
	}
}

func TestConsumer_StartStop(t *testing.T) {
	t.Run("normal start and stop", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		client, mock, cleanup := setupTest(t)
		defer cleanup()

		mock.ExpectXRead(&redis.XReadArgs{
			Streams: []string{"test-stream", "$"},
			Count:   1,
			Block:   time.Second,
		}).SetErr(redis.Nil)

		consumer, err := NewConsumer[TestMessage](
			client,
			"test-stream",
		)
		require.NoError(t, err)

		consumer.Start()
		time.Sleep(100 * time.Millisecond)
		consumer.Close()

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("multiple start calls", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		client, mock, cleanup := setupTest(t)
		defer cleanup()

		mock.ExpectXRead(&redis.XReadArgs{
			Streams: []string{"test-stream", "$"},
			Count:   1,
			Block:   time.Second,
		}).SetErr(redis.Nil)

		consumer, err := NewConsumer[TestMessage](
			client,
			"test-stream",
		)
		require.NoError(t, err)

		consumer.Start()
		consumer.Start() // Should be no-op
		time.Sleep(100 * time.Millisecond)
		consumer.Close()

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("multiple stop calls", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		client, mock, cleanup := setupTest(t)
		defer cleanup()

		mock.ExpectXRead(&redis.XReadArgs{
			Streams: []string{"test-stream", "$"},
			Count:   1,
			Block:   time.Second,
		}).SetErr(redis.Nil)

		consumer, err := NewConsumer[TestMessage](
			client,
			"test-stream",
		)
		require.NoError(t, err)

		consumer.Start()
		time.Sleep(100 * time.Millisecond)
		consumer.Close()
		consumer.Close() // Should be no-op

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestConsumer_MessageConsumption(t *testing.T) {
	t.Run("successful message consumption", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		client, mock, cleanup := setupTest(t)
		defer cleanup()

		testMsg := TestMessage{
			ID:   "1",
			Data: "test data",
		}
		msgValues, err := DefaultParseToMessage(testMsg)
		require.NoError(t, err)

		mock.ExpectXRead(&redis.XReadArgs{
			Streams: []string{"test-stream", "$"},
			Count:   1,
			Block:   time.Second,
		}).SetVal([]redis.XStream{
			{
				Stream: "test-stream",
				Messages: []redis.XMessage{
					{
						ID:     "1234-0",
						Values: msgValues,
					},
				},
			},
		})

		consumer, err := NewConsumer[TestMessage](
			client,
			"test-stream",
			WithConsumerBlockTimeout[TestMessage](time.Second),
		)
		require.NoError(t, err)

		consumer.Start()
		defer consumer.Close()

		select {
		case msg := <-consumer.Subscribe():
			assert.Equal(t, testMsg.ID, msg.ID)
			assert.Equal(t, testMsg.Data, msg.Data)
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for message")
		}

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("redis error handling", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		client, mock, cleanup := setupTest(t)
		defer cleanup()

		mock.ExpectXRead(&redis.XReadArgs{
			Streams: []string{"test-stream", "$"},
			Count:   1,
			Block:   time.Second,
		}).SetErr(redis.ErrClosed)

		consumer, err := NewConsumer[TestMessage](
			client,
			"test-stream",
			WithConsumerBlockTimeout[TestMessage](time.Second),
		)
		require.NoError(t, err)

		consumer.Start()
		defer consumer.Close()

		time.Sleep(100 * time.Millisecond)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("invalid message format", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		client, mock, cleanup := setupTest(t)
		defer cleanup()

		mock.ExpectXRead(&redis.XReadArgs{
			Streams: []string{"test-stream", "$"},
			Count:   1,
			Block:   time.Second,
		}).SetVal([]redis.XStream{
			{
				Stream: "test-stream",
				Messages: []redis.XMessage{
					{
						ID: "1234-0",
						Values: map[string]interface{}{
							"id":      123,       // wrong type
							"data":    true,      // wrong type
							"time":    "invalid", // invalid time
							"counter": "wrong",   // wrong type
						},
					},
				},
			},
		})

		consumer, err := NewConsumer[TestMessage](
			client,
			"test-stream",
			WithConsumerBlockTimeout[TestMessage](time.Second),
			WithConsumerParseFunc[TestMessage](func(m map[string]any) (TestMessage, error) {
				return TestMessage{}, fmt.Errorf("failed to parse message")
			}),
		)
		require.NoError(t, err)

		consumer.Start()
		defer consumer.Close()

		select {
		case <-consumer.Subscribe():
			t.Fatal("should not receive invalid message")
		case <-time.After(300 * time.Millisecond):
			// Expected timeout
		}

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty stream response", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		client, mock, cleanup := setupTest(t)
		defer cleanup()

		mock.ExpectXRead(&redis.XReadArgs{
			Streams: []string{"test-stream", "$"},
			Count:   1,
			Block:   time.Second,
		}).SetVal([]redis.XStream{})

		consumer, err := NewConsumer[TestMessage](
			client,
			"test-stream",
			WithConsumerBlockTimeout[TestMessage](time.Second),
		)
		require.NoError(t, err)

		consumer.Start()
		defer consumer.Close()

		select {
		case <-consumer.Subscribe():
			t.Fatal("should not receive message from empty stream")
		case <-time.After(300 * time.Millisecond):
			// Expected timeout
		}

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
