package redis

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
)

type AutoRenewMutex struct {
	*redsync.Mutex
	cancel   context.CancelFunc
	renewing bool
	mu       sync.Mutex
	wg       sync.WaitGroup
	options  autoRenewMutexOptions
}

type autoRenewMutexOptions struct {
	renewInterval time.Duration
	retryDelay    time.Duration
	expiry        time.Duration
	skipLockError bool
}

type AutoRenewMutexOption func(*autoRenewMutexOptions)

// WithAutoRenewMutexRenewInterval 設置自動續期間隔
func WithAutoRenewMutexRenewInterval(d time.Duration) AutoRenewMutexOption {
	return func(o *autoRenewMutexOptions) {
		o.renewInterval = d
	}
}

// WithAutoRenewMutexRetryDelay 設置重試延遲
func WithAutoRenewMutexRetryDelay(d time.Duration) AutoRenewMutexOption {
	return func(o *autoRenewMutexOptions) {
		o.retryDelay = d
	}
}

// WithAutoRenewMutexExpiry 設置鎖過期時間
func WithAutoRenewMutexExpiry(d time.Duration) AutoRenewMutexOption {
	return func(o *autoRenewMutexOptions) {
		o.expiry = d
	}
}

// WithAutoRenewMutexSkipLockError 設置是否忽略所有鎖定錯誤
func WithAutoRenewMutexSkipLockError(skip bool) AutoRenewMutexOption {
	return func(o *autoRenewMutexOptions) {
		o.skipLockError = skip
	}
}

// NewAutoRenewMutex 創建一個帶自動續期功能的互斥鎖
func NewAutoRenewMutex(client *redis.Client, key string, opts ...AutoRenewMutexOption) IAutoRenewMutex {
	// 默認選項
	options := autoRenewMutexOptions{
		expiry:        8 * time.Second,
		retryDelay:    500 * time.Millisecond,
		renewInterval: 0, // 會在下面根據expiry計算
		skipLockError: false,
	}

	// 應用自定義選項
	for _, opt := range opts {
		opt(&options)
	}

	// 如果未設置續期間隔，使用過期時間的1/3
	if options.renewInterval <= 0 {
		options.renewInterval = options.expiry / 3
	}

	pool := goredis.NewPool(client)
	rs := redsync.New(pool)

	mutex := rs.NewMutex(
		key,
		redsync.WithExpiry(options.expiry),
		redsync.WithTries(1),
		redsync.WithRetryDelay(options.retryDelay),
	)

	return &AutoRenewMutex{
		Mutex:   mutex,
		options: options,
	}
}

// Lock 獲取鎖並啟動自動續期，支持通過context取消
func (m *AutoRenewMutex) Lock(ctx context.Context) (context.Context, error) {
	timer := time.NewTimer(1)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timer.C:
			err := m.Mutex.LockContext(ctx)
			if err == nil {
				lockCtx, cancel := context.WithCancel(ctx)
				m.cancel = cancel
				m.startAutoRenew(lockCtx)
				return lockCtx, nil
			}
			// 只有在獲取鎖失敗或設置了忽略錯誤(skipLockError)時才重試
			var commErr *redsync.RedisError
			if !m.options.skipLockError && errors.As(err, &commErr) {
				return nil, fmt.Errorf("failed to acquire lock: %w", err)
			}
			// 重置計時器，準備下次重試
			timer.Reset(m.options.retryDelay)
		}
	}
}

// Unlock 停止自動續期並釋放鎖
func (m *AutoRenewMutex) Unlock() (bool, error) {
	m.stopAutoRenew()
	m.wg.Wait()
	return m.Mutex.Unlock()
}

// Valid 檢查鎖是否仍然有效，通過比較當前時間和過期時間判斷
func (m *AutoRenewMutex) Valid() bool {
	return time.Now().Before(m.Mutex.Until()) && m.renewing
}

func (m *AutoRenewMutex) startAutoRenew(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.renewing {
		return
	}

	m.renewing = true
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		ticker := time.NewTicker(m.options.renewInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				success, err := m.Mutex.Extend()
				if err != nil || !success {
					m.stopAutoRenew()
					return
				}
			}
		}
	}()
}

func (m *AutoRenewMutex) stopAutoRenew() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.renewing {
		return
	}

	m.renewing = false
	if m.cancel != nil {
		m.cancel()
	}
}
