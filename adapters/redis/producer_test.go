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

func TestNewProducer(t *testing.T) {
	tests := []struct {
		name    string
		client  *redis.Client
		stream  string
		opts    []ProducerOption[TestMessage]
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid configuration",
			client:  redis.NewClient(&redis.Options{}),
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
			client:  redis.NewClient(&redis.Options{}),
			stream:  "",
			wantErr: true,
			errMsg:  "stream cannot be empty",
		},
		{
			name:   "with custom options",
			client: redis.NewClient(&redis.Options{}),
			stream: "test-stream",
			opts: []ProducerOption[TestMessage]{
				WithProducerLogger[TestMessage](slog.Default()),
				WithProducerBufferSize[TestMessage](200),
				WithProducerParseFunc[TestMessage](func(msg TestMessage) (map[string]any, error) {
					return map[string]any{"test": "value"}, nil
				}),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer goleak.VerifyNone(t)

			producer, err := NewProducer[TestMessage](tt.client, tt.stream, tt.opts...)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, producer)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, producer)
				producer.Close()
			}

			if tt.client != nil {
				tt.client.Close()
			}
		})
	}
}

func TestProducer_StartStop(t *testing.T) {
	t.Run("normal start and stop", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		client, _, cleanup := setupTest(t)
		defer cleanup()

		producer, err := NewProducer[TestMessage](client, "test-stream")
		require.NoError(t, err)

		producer.Start()
		time.Sleep(100 * time.Millisecond)
		producer.Close()
	})

	t.Run("multiple start calls", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		client, _, cleanup := setupTest(t)
		defer cleanup()

		producer, err := NewProducer[TestMessage](client, "test-stream")
		require.NoError(t, err)

		producer.Start()
		producer.Start() // Should be no-op
		time.Sleep(100 * time.Millisecond)
		producer.Close()
	})

	t.Run("multiple stop calls", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		client, _, cleanup := setupTest(t)
		defer cleanup()

		producer, err := NewProducer[TestMessage](client, "test-stream")
		require.NoError(t, err)

		producer.Start()
		time.Sleep(100 * time.Millisecond)
		producer.Close()
		producer.Close() // Should be no-op
	})
}

func TestProducer_Publish(t *testing.T) {
	t.Run("successful publish", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		client, mock, cleanup := setupTest(t)
		defer cleanup()

		msg := TestMessage{
			ID:   "1",
			Data: "test data",
		}

		msgValues, err := DefaultParseToMessage(msg)
		require.NoError(t, err)

		mock.ExpectXAdd(&redis.XAddArgs{
			Stream: "test-stream",
			Values: msgValues,
		}).SetVal("1234-0")

		producer, err := NewProducer[TestMessage](client, "test-stream")
		require.NoError(t, err)

		producer.Start()
		err = producer.Publish(msg)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)
		producer.Close()
	})

	t.Run("publish to closed producer", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		client, _, cleanup := setupTest(t)
		defer cleanup()

		producer, err := NewProducer[TestMessage](client, "test-stream")
		require.NoError(t, err)

		producer.Start()
		time.Sleep(100 * time.Millisecond)
		producer.Close()

		msg := TestMessage{
			ID:   "1",
			Data: "test data",
		}

		err = producer.Publish(msg)
		assert.ErrorIs(t, err, ErrConsumerClosed)
	})

	t.Run("publish with custom parse function error", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		client, _, cleanup := setupTest(t)
		defer cleanup()

		producer, err := NewProducer[TestMessage](
			client,
			"test-stream",
			WithProducerParseFunc[TestMessage](func(TestMessage) (map[string]any, error) {
				return nil, fmt.Errorf("parse error")
			}),
		)
		require.NoError(t, err)

		producer.Start()
		err = producer.Publish(TestMessage{})
		assert.Error(t, err)

		producer.Close()
	})

	t.Run("publish with redis connection error", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		client, mock, cleanup := setupTest(t)
		defer cleanup()

		msg := TestMessage{
			ID:   "1",
			Data: "test data",
		}

		msgValues, err := DefaultParseToMessage(msg)
		require.NoError(t, err)

		mock.ExpectXAdd(&redis.XAddArgs{
			Stream: "test-stream",
			Values: msgValues,
		}).SetErr(redis.ErrClosed)

		producer, err := NewProducer[TestMessage](client, "test-stream")
		require.NoError(t, err)

		producer.Start()
		err = producer.Publish(msg)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)
		producer.Close()
	})
}
