package redis

import (
	"context"
	"io"
	"log"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/go-redsync/redsync/v4"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
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

func TestNewAutoRenewMutex(t *testing.T) {
	tests := []struct {
		name string
		key  string
		opts []AutoRenewMutexOption
	}{
		{
			name: "default options",
			key:  "test-lock",
		},
		{
			name: "custom options",
			key:  "test-lock",
			opts: []AutoRenewMutexOption{
				WithAutoRenewMutexExpiry(5 * time.Second),
				WithAutoRenewMutexRenewInterval(1 * time.Second),
				WithAutoRenewMutexRetryDelay(100 * time.Millisecond),
				WithAutoRenewMutexSkipLockError(true),
			},
		},
		{
			name: "zero expiry",
			key:  "test-lock",
			opts: []AutoRenewMutexOption{
				WithAutoRenewMutexExpiry(0),
			},
		},
		{
			name: "empty key",
			key:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer goleak.VerifyNone(t)
			client, _, cleanup := setupTest(t)
			defer cleanup()

			mutex := NewAutoRenewMutex(client, tt.key, tt.opts...)
			require.NotNil(t, mutex)
		})
	}
}

func TestAutoRenewMutex_Lock(t *testing.T) {
	t.Run("successful lock", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		client, mock, cleanup := setupTest(t)
		defer cleanup()

		mock.Regexp().ExpectSetNX("test-lock", ".*", 8*time.Second).SetVal(true)
		mock.Regexp().ExpectEvalSha(".*", []string{"test-lock"}, []string{".*"}).SetVal(int64(1))

		mutex := NewAutoRenewMutex(client, "test-lock")
		err := mutex.Lock(context.Background())
		assert.NoError(t, err)

		ok, err := mutex.Unlock()
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("lock with context cancellation", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		client, _, cleanup := setupTest(t)
		defer cleanup()

		mutex := NewAutoRenewMutex(client, "test-lock")
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := mutex.Lock(ctx)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("lock with redis error and skip error enabled", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		client, mock, cleanup := setupTest(t)
		defer cleanup()

		mock.Regexp().ExpectSetNX("test-lock", ".*", 8*time.Second).SetErr(redis.ErrClosed)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		mutex := NewAutoRenewMutex(client, "test-lock", WithAutoRenewMutexSkipLockError(true))
		err := mutex.Lock(ctx)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("lock with redis error and skip error disabled", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		client, mock, cleanup := setupTest(t)
		defer cleanup()

		mock.Regexp().ExpectSetNX("test-lock", ".*", 8*time.Second).SetErr(redis.ErrClosed)

		mutex := NewAutoRenewMutex(client, "test-lock")
		err := mutex.Lock(context.Background())
		assert.Error(t, err)
		assert.ErrorIs(t, err, redis.ErrClosed)
	})

	t.Run("double lock", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		client, mock, cleanup := setupTest(t)
		defer cleanup()

		// 第一次鎖定成功
		mock.Regexp().ExpectSetNX("test-lock", ".*", 8*time.Second).SetVal(true)
		// 第二次鎖定失敗
		mock.Regexp().ExpectSetNX("test-lock", ".*", 8*time.Second).SetVal(false)
		mock.Regexp().ExpectEvalSha(".*", []string{"test-lock"}, []string{".*"}).SetVal(int64(0))
		// 解鎖
		mock.Regexp().ExpectEvalSha(".*", []string{"test-lock"}, []string{".*"}).SetVal(int64(1))

		mutex := NewAutoRenewMutex(client, "test-lock", WithAutoRenewMutexRetryDelay(time.Second))
		err := mutex.Lock(context.Background())
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		err = mutex.Lock(ctx)
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.DeadlineExceeded)

		ok, err := mutex.Unlock()
		assert.NoError(t, err)
		assert.True(t, ok)
	})
}

func TestAutoRenewMutex_AutoRenew(t *testing.T) {
	t.Run("successful auto renew", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		client, mock, cleanup := setupTest(t)
		defer cleanup()

		// 初始鎖定
		mock.Regexp().ExpectSetNX("test-lock", ".*", 2*time.Second).SetVal(true)
		// 兩次續期
		mock.Regexp().ExpectEvalSha(".*", []string{"test-lock"}, []string{".*", "2000"}).SetVal(int64(1))
		mock.Regexp().ExpectEvalSha(".*", []string{"test-lock"}, []string{".*", "2000"}).SetVal(int64(1))
		// 解鎖
		mock.Regexp().ExpectEvalSha(".*", []string{"test-lock"}, []string{".*"}).SetVal(int64(1))

		mutex := NewAutoRenewMutex(client, "test-lock",
			WithAutoRenewMutexExpiry(2*time.Second),
			WithAutoRenewMutexRenewInterval(100*time.Millisecond))

		err := mutex.Lock(context.Background())
		require.NoError(t, err)

		time.Sleep(250 * time.Millisecond)
		assert.True(t, mutex.Valid())

		ok, err := mutex.Unlock()
		assert.NoError(t, err)
		assert.True(t, ok)

	})

	t.Run("auto renew fails", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		client, mock, cleanup := setupTest(t)
		defer cleanup()

		// 初始鎖定成功
		mock.Regexp().ExpectSetNX("test-lock", ".*", 2*time.Second).SetVal(true)
		// 續期失敗
		mock.Regexp().ExpectEvalSha(".*", []string{"test-lock"}, []string{".*", "2000"}).SetErr(redis.ErrClosed)
		// 解鎖
		mock.Regexp().ExpectEvalSha(".*", []string{"test-lock"}, []string{".*"}).SetVal(int64(-1))

		mutex := NewAutoRenewMutex(client, "test-lock",
			WithAutoRenewMutexExpiry(2*time.Second),
			WithAutoRenewMutexRenewInterval(100*time.Millisecond))

		err := mutex.Lock(context.Background())
		require.NoError(t, err)

		time.Sleep(150 * time.Millisecond)
		assert.False(t, mutex.Valid())

		ok, err := mutex.Unlock()
		assert.ErrorIs(t, err, redsync.ErrLockAlreadyExpired)
		assert.False(t, ok)
	})
}

func TestAutoRenewMutex_Unlock(t *testing.T) {
	t.Run("unlock after successful lock", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		client, mock, cleanup := setupTest(t)
		defer cleanup()

		// 鎖定成功
		mock.Regexp().ExpectSetNX("test-lock", ".*", 8*time.Second).SetVal(true)
		// 解鎖成功
		mock.Regexp().ExpectEvalSha(".*", []string{"test-lock"}, []string{".*"}).SetVal(int64(1))

		mutex := NewAutoRenewMutex(client, "test-lock")
		err := mutex.Lock(context.Background())
		require.NoError(t, err)

		ok, err := mutex.Unlock()
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("unlock without lock", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		client, mock, cleanup := setupTest(t)
		defer cleanup()

		// 解鎖失敗
		mock.Regexp().ExpectEvalSha(".*", []string{"test-lock"}, []string{".*"}).SetVal(int64(-1))

		mutex := NewAutoRenewMutex(client, "test-lock")
		ok, err := mutex.Unlock()
		assert.Error(t, err)
		assert.ErrorIs(t, err, redsync.ErrLockAlreadyExpired)
		assert.False(t, ok)
	})

	t.Run("double unlock", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		client, mock, cleanup := setupTest(t)
		defer cleanup()

		// 鎖定成功
		mock.Regexp().ExpectSetNX("test-lock", ".*", 8*time.Second).SetVal(true)
		// 解鎖
		mock.Regexp().ExpectEvalSha(".*", []string{"test-lock"}, []string{".*"}).SetVal(int64(1))
		mock.Regexp().ExpectEvalSha(".*", []string{"test-lock"}, []string{".*"}).SetVal(int64(-1))

		mutex := NewAutoRenewMutex(client, "test-lock")
		err := mutex.Lock(context.Background())
		require.NoError(t, err)

		ok, err := mutex.Unlock()
		assert.NoError(t, err)
		assert.True(t, ok)

		ok, err = mutex.Unlock()
		assert.Error(t, err)
		assert.ErrorIs(t, err, redsync.ErrLockAlreadyExpired)
		assert.False(t, ok)
	})
}

func TestAutoRenewMutex_Valid(t *testing.T) {
	t.Run("validity checks", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		client, mock, cleanup := setupTest(t)
		defer cleanup()

		// 鎖定成功
		mock.Regexp().ExpectSetNX("test-lock", ".*", 2*time.Second).SetVal(true)
		// 解鎖成功
		mock.Regexp().ExpectEvalSha(".*", []string{"test-lock"}, []string{".*"}).SetVal(int64(1))

		mutex := NewAutoRenewMutex(client, "test-lock",
			WithAutoRenewMutexExpiry(2*time.Second))

		// 未鎖定時
		assert.False(t, mutex.Valid())

		// 鎖定後
		err := mutex.Lock(context.Background())
		require.NoError(t, err)
		assert.True(t, mutex.Valid())

		// 解鎖後
		ok, err := mutex.Unlock()
		assert.NoError(t, err)
		assert.True(t, ok)
		assert.False(t, mutex.Valid())
	})
}
