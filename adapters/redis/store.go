package redis

import (
	"context"
	"fmt"
	"q4/adapters/session"

	"github.com/redis/go-redis/v9"
)

// Store 實現了 IStore 介面，提供基於 Redis hash 的資料儲存功能
type Store struct {
	client  *redis.Client // Redis 客戶端連線
	options StoreOptions  // Store 的配置選項
}

// StoreOptions 定義了 Store 的配置選項
type StoreOptions struct {
	Prefix string
}

type StoreOption func(*StoreOptions)

// WithStorePrefix 設定 Store 的 key 前綴
func WithStorePrefix(prefix string) StoreOption {
	return func(o *StoreOptions) {
		o.Prefix = prefix
	}
}

// NewStore 建立一個新的 Store 實例
func NewStore(client *redis.Client, opts ...StoreOption) session.IStore {
	options := &StoreOptions{}
	for _, opt := range opts {
		opt(options)
	}

	return &Store{
		client:  client,
		options: *options,
	}
}

// Load 從 Redis 中載入指定名稱的資料
func (s *Store) Load(ctx context.Context, name string) (map[string]string, error) {
	const op = "redis.Store.Load"
	key := s.options.Prefix + name

	result, err := s.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get hash: %w", op, err)
	}

	// Redis returns empty map when key doesn't exist
	return result, nil
}

// saveScript 是用於原子性地刪除並設定新的 hash 欄位的 Lua 腳本
var saveScript = redis.NewScript(`
local key = KEYS[1]
redis.call('DEL', key)
if #ARGV > 0 then
    redis.call('HSET', key, unpack(ARGV))
end
return 1
`)

// Save 將資料儲存到 Redis 中
// NOTE: 會先刪除舊的資料，再設定新的資料，這個過程是原子性的
func (s *Store) Save(ctx context.Context, name string, data map[string]string) error {
	const op = "redis.Store.Save"
	key := s.options.Prefix + name
	// 準備參數
	args := make([]any, 0, len(data)*2)
	for k, v := range data {
		args = append(args, k, v)
	}
	// 執行 Lua 腳本
	err := saveScript.Run(ctx, s.client, []string{key}, args...).Err()
	if err != nil {
		return fmt.Errorf("%s: failed to execute save script: %w", op, err)
	}

	return nil
}
