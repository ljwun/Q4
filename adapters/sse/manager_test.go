package sse_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"q4/adapters/sse"
)

func setupRedis(t *testing.T) (*redis.Client, func()) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return client, func() {
		client.Close()
		mr.Close()
	}
}

func TestConnectionManager(t *testing.T) {
	testingLogger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	t.Run("basic publish and subscribe", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		rdb, closeRedis := setupRedis(t)
		defer closeRedis()
		cm := sse.NewConnectionManager[Message](rdb, "test_stream", testingLogger)
		cm.Start()
		defer cm.Done()

		ch, err := cm.Subscribe("test_channel")
		require.NoError(t, err)
		require.NotNil(t, ch)

		msg := Message{Data: "test message"}
		err = cm.Publish("test_channel", msg)
		require.NoError(t, err)

		select {
		case received := <-ch:
			assert.Equal(t, msg, received)
		case <-time.After(time.Second):
			t.Fatal("did not receive message in time")
		}
	})

	t.Run("multiple subscribers receive same message", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		rdb, closeRedis := setupRedis(t)
		defer closeRedis()
		cm := sse.NewConnectionManager[Message](rdb, "test_stream", testingLogger)
		cm.Start()
		defer cm.Done()

		ch1, err1 := cm.Subscribe("test_channel")
		ch2, err2 := cm.Subscribe("test_channel")
		require.NoError(t, err1)
		require.NoError(t, err2)

		msg := Message{Data: "broadcast test"}
		err := cm.Publish("test_channel", msg)
		require.NoError(t, err)

		for i, ch := range []<-chan Message{ch1, ch2} {
			select {
			case received := <-ch:
				assert.Equal(t, msg, received, "subscriber %d", i+1)
			case <-time.After(time.Second):
				t.Fatalf("subscriber %d did not receive message in time", i+1)
			}
		}
	})

	t.Run("manager done stops all operations", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		rdb, closeRedis := setupRedis(t)
		defer closeRedis()
		cm := sse.NewConnectionManager[Message](rdb, "test_stream", testingLogger)
		cm.Start()

		ch, err := cm.Subscribe("test_channel")
		require.NoError(t, err)

		// 正確關閉 manager
		cm.Done()

		// 檢查訂閱操作
		_, err = cm.Subscribe("test_channel")
		assert.ErrorIs(t, err, context.Canceled)

		// 檢查發布操作
		err = cm.Publish("test_channel", Message{Data: "test"})
		assert.ErrorIs(t, err, context.Canceled)

		// 確認訂閱通道會被正確關閉
		select {
		case _, ok := <-ch:
			assert.False(t, ok, "channel should be closed")
		case <-time.After(time.Second):
			t.Fatal("channel was not closed")
		}
	})

	t.Run("unsubscribe stops message receiving", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		rdb, closeRedis := setupRedis(t)
		defer closeRedis()
		cm := sse.NewConnectionManager[Message](rdb, "test_stream", testingLogger)
		cm.Start()
		defer cm.Done()

		ch, err := cm.Subscribe("test_channel")
		require.NoError(t, err)

		cm.Unsubscribe("test_channel", ch)

		// 發送訊息確認已取消訂閱的通道不會收到訊息
		err = cm.Publish("test_channel", Message{Data: "test"})
		require.NoError(t, err)
		for msg := range ch {
			t.Fatalf("should not receive message after unsubscribe, got: %v", msg)
		}
	})
}
