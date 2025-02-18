package redis

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"go.uber.org/mock/gomock"
)

func TestNewGroupConsumer(t *testing.T) {
	tests := []struct {
		name     string
		client   *redis.Client
		stream   string
		group    string
		consumer string
		opts     []GroupConsumerOption[TestMessage]
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid configuration",
			client:   redis.NewClient(&redis.Options{}),
			stream:   "test-stream",
			group:    "test-group",
			consumer: "test-consumer",
			wantErr:  false,
		},
		{
			name:     "nil client",
			client:   nil,
			stream:   "test-stream",
			group:    "test-group",
			consumer: "test-consumer",
			wantErr:  true,
			errMsg:   "redis client cannot be nil",
		},
		{
			name:     "empty stream",
			client:   redis.NewClient(&redis.Options{}),
			stream:   "",
			group:    "test-group",
			consumer: "test-consumer",
			wantErr:  true,
			errMsg:   "stream, group and consumer cannot be empty",
		},
		{
			name:     "with strict ordering and mutex",
			client:   redis.NewClient(&redis.Options{}),
			stream:   "test-stream",
			group:    "test-group",
			consumer: "test-consumer",
			opts: []GroupConsumerOption[TestMessage]{
				WithGroupConsumerLogger[TestMessage](slog.Default()),
				WithGroupConsumerParseFunc[TestMessage](DefaultParseFromMessage[TestMessage]),
				WithGroupConsumerBufferSize[TestMessage](1),
				WithGroupConsumerBlockTimeout[TestMessage](time.Second),
				WithGroupConsumerStrictOrdering[TestMessage](true),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer goleak.VerifyNone(t)

			consumer, err := NewGroupConsumer(
				tt.client,
				tt.stream,
				tt.group,
				tt.consumer,
				tt.opts...,
			)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, consumer)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, consumer)
			}

			if tt.client != nil {
				tt.client.Close()
			}
		})
	}
}

func TestGroupConsumer_StartStop(t *testing.T) {
	t.Run("normal start and stop", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		client, mock, cleanup := setupTest(t)
		defer cleanup()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockMutex := NewMockIAutoRenewMutex(ctrl)
		mockMutex.EXPECT().Lock(gomock.Any()).Return(nil)
		mockMutex.EXPECT().Unlock().Return(true, nil)

		mock.ExpectXPendingExt(&redis.XPendingExtArgs{
			Stream: "test-stream",
			Group:  "test-group",
			Start:  "-",
			End:    "+",
			Count:  100,
		}).SetVal([]redis.XPendingExt{})

		consumer, err := NewGroupConsumer[TestMessage](
			client,
			"test-stream",
			"test-group",
			"test-consumer",
			WithGroupConsumerStrictOrdering[TestMessage](true),
			WithGroupConsumerMutex[TestMessage](mockMutex),
		)
		require.NoError(t, err)

		err = consumer.Start()
		assert.NoError(t, err)

		err = consumer.Close()
		assert.NoError(t, err)
	})

	t.Run("start with lock error", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		client, _, cleanup := setupTest(t)
		defer cleanup()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockMutex := NewMockIAutoRenewMutex(ctrl)
		mockMutex.EXPECT().Lock(gomock.Any()).Return(errors.New("lock error"))

		consumer, err := NewGroupConsumer[TestMessage](
			client,
			"test-stream",
			"test-group",
			"test-consumer",
			WithGroupConsumerStrictOrdering[TestMessage](true),
			WithGroupConsumerMutex[TestMessage](mockMutex),
		)
		require.NoError(t, err)

		err = consumer.Start()
		assert.NoError(t, err) // Start不會返回錯誤，因為錯誤會在goroutine中處理

		time.Sleep(100 * time.Millisecond)
		err = consumer.Close()
		assert.NoError(t, err)
	})

	t.Run("multiple starts", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		client, _, cleanup := setupTest(t)
		defer cleanup()

		consumer, err := NewGroupConsumer[TestMessage](
			client,
			"test-stream",
			"test-group",
			"test-consumer",
		)
		require.NoError(t, err)

		// 第一次啟動
		err = consumer.Start()
		assert.NoError(t, err)

		// 第二次啟動應該不會出錯
		err = consumer.Start()
		assert.NoError(t, err)

		err = consumer.Close()
		assert.NoError(t, err)
	})

	t.Run("multiple closes", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		client, _, cleanup := setupTest(t)
		defer cleanup()

		consumer, err := NewGroupConsumer[TestMessage](
			client,
			"test-stream",
			"test-group",
			"test-consumer",
		)
		require.NoError(t, err)

		err = consumer.Start()
		assert.NoError(t, err)

		// 第一次關閉
		err = consumer.Close()
		assert.NoError(t, err)

		// 第二次關閉不應該出錯
		err = consumer.Close()
		assert.NoError(t, err)
	})
}

func TestGroupConsumer_MessageProcessing(t *testing.T) {
	t.Run("successful message processing", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		client, mock, cleanup := setupTest(t)
		defer cleanup()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockMutex := NewMockIAutoRenewMutex(ctrl)
		mockMutex.EXPECT().Lock(gomock.Any()).Return(nil)
		mockMutex.EXPECT().Unlock().Return(true, nil).AnyTimes()

		// Setup test message
		testMsg := TestMessage{ID: "1", Data: "test"}
		msgData, err := DefaultParseToMessage(testMsg)
		require.NoError(t, err)

		// Set expectations
		mock.ExpectXPendingExt(&redis.XPendingExtArgs{
			Stream: "test-stream",
			Group:  "test-group",
			Start:  "-",
			End:    "+",
			Count:  100,
		}).SetVal([]redis.XPendingExt{})

		mock.ExpectXReadGroup(&redis.XReadGroupArgs{
			Group:    "test-group",
			Consumer: "test-consumer",
			Streams:  []string{"test-stream", ">"},
			Count:    1,
			Block:    time.Second,
		}).SetVal([]redis.XStream{
			{
				Stream: "test-stream",
				Messages: []redis.XMessage{
					{
						ID:     "1234-0",
						Values: msgData,
					},
				},
			},
		})

		mock.ExpectXAck("test-stream", "test-group", "1234-0").SetVal(1)

		consumer, err := NewGroupConsumer[TestMessage](
			client,
			"test-stream",
			"test-group",
			"test-consumer",
			WithGroupConsumerStrictOrdering[TestMessage](true),
			WithGroupConsumerMutex[TestMessage](mockMutex),
		)
		require.NoError(t, err)

		err = consumer.Start()
		require.NoError(t, err)

		// Subscribe and wait for message
		msgChan := consumer.Subscribe()
		select {
		case msg := <-msgChan:
			assert.Equal(t, testMsg.ID, msg.Data.ID)
			assert.Equal(t, testMsg.Data, msg.Data.Data)
			err = msg.Done(context.Background())
			assert.NoError(t, err)
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for message")
		}

		err = consumer.Close()
		assert.NoError(t, err)
	})

	t.Run("message parse error handling", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		client, mock, cleanup := setupTest(t)
		defer cleanup()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockMutex := NewMockIAutoRenewMutex(ctrl)
		mockMutex.EXPECT().Lock(gomock.Any()).Return(nil)
		mockMutex.EXPECT().Unlock().Return(true, nil).AnyTimes()

		// Set expectations for invalid message
		mock.ExpectXPendingExt(&redis.XPendingExtArgs{
			Stream: "test-stream",
			Group:  "test-group",
			Start:  "-",
			End:    "+",
			Count:  100,
		}).SetVal([]redis.XPendingExt{})

		mock.ExpectXReadGroup(&redis.XReadGroupArgs{
			Group:    "test-group",
			Consumer: "test-consumer",
			Streams:  []string{"test-stream", ">"},
			Count:    1,
			Block:    time.Second,
		}).SetVal([]redis.XStream{
			{
				Stream: "test-stream",
				Messages: []redis.XMessage{
					{
						ID:     "1234-0",
						Values: map[string]interface{}{"data": "invalid"},
					},
				},
			},
		})

		// Expect message to be moved to dead letter queue
		mock.ExpectXAdd(&redis.XAddArgs{
			Stream: "test-stream:dead-letter",
			Values: map[string]interface{}{"data": "invalid"},
		}).SetVal("1234-0")

		mock.ExpectXAck("test-stream", "test-group", "1234-0").SetVal(1)

		consumer, err := NewGroupConsumer[TestMessage](
			client,
			"test-stream",
			"test-group",
			"test-consumer",
			WithGroupConsumerStrictOrdering[TestMessage](true),
			WithGroupConsumerMutex[TestMessage](mockMutex),
			WithGroupConsumerParseFunc(func(data map[string]any) (TestMessage, error) {
				return TestMessage{}, errors.New("parse error")
			}), // 模擬解析錯誤
		)
		require.NoError(t, err)

		err = consumer.Start()
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		err = consumer.Close()
		assert.NoError(t, err)
	})

	t.Run("concurrent messages", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		client, mock, cleanup := setupTest(t)
		defer cleanup()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockMutex := NewMockIAutoRenewMutex(ctrl)
		mockMutex.EXPECT().Lock(gomock.Any()).Return(nil)
		mockMutex.EXPECT().Unlock().Return(true, nil).AnyTimes()

		// Setup multiple test messages
		testMsg1 := TestMessage{ID: "1", Data: "test1"}
		testMsg2 := TestMessage{ID: "2", Data: "test2"}
		msgData1, err := DefaultParseToMessage(testMsg1)
		require.NoError(t, err)
		msgData2, err := DefaultParseToMessage(testMsg2)
		require.NoError(t, err)

		mock.ExpectXPendingExt(&redis.XPendingExtArgs{
			Stream: "test-stream",
			Group:  "test-group",
			Start:  "-",
			End:    "+",
			Count:  100,
		}).SetVal([]redis.XPendingExt{})

		// Expect multiple messages in order
		mock.ExpectXReadGroup(&redis.XReadGroupArgs{
			Group:    "test-group",
			Consumer: "test-consumer",
			Streams:  []string{"test-stream", ">"},
			Count:    1,
			Block:    time.Second,
		}).SetVal([]redis.XStream{
			{
				Stream: "test-stream",
				Messages: []redis.XMessage{
					{
						ID:     "1234-0",
						Values: msgData1,
					},
				},
			},
		})

		mock.ExpectXAck("test-stream", "test-group", "1234-0").SetVal(1)

		mock.ExpectXReadGroup(&redis.XReadGroupArgs{
			Group:    "test-group",
			Consumer: "test-consumer",
			Streams:  []string{"test-stream", ">"},
			Count:    1,
			Block:    time.Second,
		}).SetVal([]redis.XStream{
			{
				Stream: "test-stream",
				Messages: []redis.XMessage{
					{
						ID:     "1234-1",
						Values: msgData2,
					},
				},
			},
		})

		mock.ExpectXAck("test-stream", "test-group", "1234-1").SetVal(1)

		consumer, err := NewGroupConsumer[TestMessage](
			client,
			"test-stream",
			"test-group",
			"test-consumer",
			WithGroupConsumerStrictOrdering[TestMessage](true),
			WithGroupConsumerMutex[TestMessage](mockMutex),
		)
		require.NoError(t, err)

		err = consumer.Start()
		require.NoError(t, err)

		// Verify messages are received in order
		msgChan := consumer.Subscribe()

		// First message
		select {
		case msg := <-msgChan:
			assert.Equal(t, testMsg1.ID, msg.Data.ID)
			assert.Equal(t, testMsg1.Data, msg.Data.Data)
			err = msg.Done(context.Background())
			assert.NoError(t, err)
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for first message")
		}

		// Second message
		select {
		case msg := <-msgChan:
			assert.Equal(t, testMsg2.ID, msg.Data.ID)
			assert.Equal(t, testMsg2.Data, msg.Data.Data)
			err = msg.Done(context.Background())
			assert.NoError(t, err)
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for second message")
		}

		err = consumer.Close()
		assert.NoError(t, err)
	})

	t.Run("dead letter queue error", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		client, mock, cleanup := setupTest(t)
		defer cleanup()

		consumer, err := NewGroupConsumer[TestMessage](
			client,
			"test-stream",
			"test-group",
			"test-consumer",
		)
		require.NoError(t, err)

		// 設置一個無效的消息格式
		mock.ExpectXReadGroup(&redis.XReadGroupArgs{
			Group:    "test-group",
			Consumer: "test-consumer",
			Streams:  []string{"test-stream", ">"},
			Count:    1,
			Block:    time.Second,
		}).SetVal([]redis.XStream{
			{
				Stream: "test-stream",
				Messages: []redis.XMessage{
					{
						ID:     "1234-0",
						Values: map[string]interface{}{"data": "invalid"},
					},
				},
			},
		})

		// Dead letter queue寫入失敗
		mock.ExpectXAdd(&redis.XAddArgs{
			Stream: "test-stream:dead-letter",
			Values: map[string]interface{}{"data": "invalid"},
		}).SetErr(errors.New("dead letter queue error"))

		err = consumer.Start()
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		err = consumer.Close()
		assert.NoError(t, err)
	})
}

func TestGroupConsumer_PendingMessages(t *testing.T) {
	t.Run("process pending messages", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		client, mock, cleanup := setupTest(t)
		defer cleanup()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockMutex := NewMockIAutoRenewMutex(ctrl)
		mockMutex.EXPECT().Lock(gomock.Any()).Return(nil)
		mockMutex.EXPECT().Unlock().Return(true, nil).AnyTimes()

		testMsg := TestMessage{ID: "1", Data: "test"}
		msgData, err := DefaultParseToMessage(testMsg)
		require.NoError(t, err)

		// Set up pending messages
		mock.ExpectXPendingExt(&redis.XPendingExtArgs{
			Stream: "test-stream",
			Group:  "test-group",
			Start:  "-",
			End:    "+",
			Count:  100,
		}).SetVal([]redis.XPendingExt{
			{
				ID: "1234-0",
			},
		})

		mock.ExpectXRangeN("test-stream", "1234-0", "1234-0", 1).
			SetVal([]redis.XMessage{
				{
					ID:     "1234-0",
					Values: msgData,
				},
			})

		mock.ExpectXAck("test-stream", "test-group", "1234-0").SetVal(1)

		consumer, err := NewGroupConsumer[TestMessage](
			client,
			"test-stream",
			"test-group",
			"test-consumer",
			WithGroupConsumerStrictOrdering[TestMessage](true),
			WithGroupConsumerMutex[TestMessage](mockMutex),
		)
		require.NoError(t, err)

		err = consumer.Start()
		assert.NoError(t, err)

		msgChan := consumer.Subscribe()
		select {
		case msg := <-msgChan:
			assert.Equal(t, testMsg.ID, msg.Data.ID)
			assert.Equal(t, testMsg.Data, msg.Data.Data)
			err = msg.Done(context.Background())
			assert.NoError(t, err)
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for pending message")
		}

		err = consumer.Close()
		assert.NoError(t, err)
	})

	t.Run("pending messages fetch error", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		client, mock, cleanup := setupTest(t)
		defer cleanup()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockMutex := NewMockIAutoRenewMutex(ctrl)
		mockMutex.EXPECT().Lock(gomock.Any()).Return(nil)
		mockMutex.EXPECT().Unlock().Return(true, nil)

		// 模擬 XPendingExt 返回錯誤
		mock.ExpectXPendingExt(&redis.XPendingExtArgs{
			Stream: "test-stream",
			Group:  "test-group",
			Start:  "-",
			End:    "+",
			Count:  100,
		}).SetErr(errors.New("pending messages fetch error"))

		consumer, err := NewGroupConsumer[TestMessage](
			client,
			"test-stream",
			"test-group",
			"test-consumer",
			WithGroupConsumerStrictOrdering[TestMessage](true),
			WithGroupConsumerMutex[TestMessage](mockMutex),
		)
		require.NoError(t, err)

		err = consumer.Start()
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		err = consumer.Close()
		assert.NoError(t, err)
	})
}

func TestGroupConsumer_NonOrderingModes(t *testing.T) {
	t.Run("non-strict ordering mode", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		client, mock, cleanup := setupTest(t)
		defer cleanup()

		// Setup test message
		testMsg := TestMessage{ID: "1", Data: "test"}
		msgData, err := DefaultParseToMessage(testMsg)
		require.NoError(t, err)

		mock.ExpectXReadGroup(&redis.XReadGroupArgs{
			Group:    "test-group",
			Consumer: "test-consumer",
			Streams:  []string{"test-stream", ">"},
			Count:    1,
			Block:    time.Second,
		}).SetVal([]redis.XStream{
			{
				Stream: "test-stream",
				Messages: []redis.XMessage{
					{
						ID:     "1234-0",
						Values: msgData,
					},
				},
			},
		})

		mock.ExpectXAck("test-stream", "test-group", "1234-0").SetVal(1)

		consumer, err := NewGroupConsumer[TestMessage](
			client,
			"test-stream",
			"test-group",
			"test-consumer",
			WithGroupConsumerStrictOrdering[TestMessage](false), // 非嚴格順序模式
		)
		require.NoError(t, err)

		err = consumer.Start()
		require.NoError(t, err)

		// Subscribe and wait for message
		msgChan := consumer.Subscribe()
		select {
		case msg := <-msgChan:
			assert.Equal(t, testMsg.ID, msg.Data.ID)
			assert.Equal(t, testMsg.Data, msg.Data.Data)
			err = msg.Done(context.Background())
			assert.NoError(t, err)
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for message")
		}

		err = consumer.Close()
		assert.NoError(t, err)
	})
}

func TestMessage_Done(t *testing.T) {
	t.Run("multiple done calls", func(t *testing.T) {
		client, mock, cleanup := setupTest(t)
		defer cleanup()

		msg := &Message[TestMessage]{
			Data:      TestMessage{ID: "1", Data: "test"},
			messageID: "1234-0",
			stream:    "test-stream",
			group:     "test-group",
			client:    client,
		}

		// 只應該呼叫一次XAck
		mock.ExpectXAck("test-stream", "test-group", "1234-0").SetVal(1)

		// 第一次呼叫
		err := msg.Done(context.Background())
		assert.NoError(t, err)

		// 第二次呼叫應該直接返回nil
		err = msg.Done(context.Background())
		assert.NoError(t, err)
	})

	t.Run("ack error", func(t *testing.T) {
		client, mock, cleanup := setupTest(t)
		defer cleanup()

		msg := &Message[TestMessage]{
			Data:      TestMessage{ID: "1", Data: "test"},
			messageID: "1234-0",
			stream:    "test-stream",
			group:     "test-group",
			client:    client,
		}

		mock.ExpectXAck("test-stream", "test-group", "1234-0").
			SetErr(errors.New("ack error"))

		err := msg.Done(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ack error")
	})
}
