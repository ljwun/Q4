package api

import (
	"context"
	"encoding/base64"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/vmihailenco/msgpack/v5"

	redisAdapter "q4/adapters/redis"
)

// compareBidInfo compares two BidInfo structs with proper time comparison
func compareBidInfo(t *testing.T, expected, actual BidInfo) {
	assert.Equal(t, expected.ItemID, actual.ItemID)
	assert.Equal(t, expected.BidderID, actual.BidderID)
	assert.Equal(t, expected.Amount, actual.Amount)
	assert.True(t, expected.CreatedAt.Equal(actual.CreatedAt),
		"CreatedAt times are not equal. Expected: %v, Got: %v",
		expected.CreatedAt, actual.CreatedAt)
}

func TestBidScript(t *testing.T) {
	// 設置 miniredis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	// 建立 Redis 客戶端
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer client.Close()

	ctx := context.Background()
	now := time.Now()
	itemID := uuid.New()
	bidderID := uuid.New()

	tests := []struct {
		name        string
		setupFunc   func()
		itemKey     string
		streamKey   string
		bidAmount   string
		bidInfo     BidInfo
		expireTime  string
		want        int
		checkStream bool
	}{
		{
			name:      "商品不存在時應返回-1",
			setupFunc: func() {},
			itemKey:   "item:nonexistent",
			streamKey: "stream:bids",
			bidAmount: "100",
			bidInfo: BidInfo{
				ItemID:    itemID,
				BidderID:  bidderID,
				Amount:    100,
				CreatedAt: now,
			},
			expireTime: "3600",
			want:       -1,
		},
		{
			name: "出價金額不足時應返回0",
			setupFunc: func() {
				mr.Set("item:1", "200")
			},
			itemKey:    "item:1",
			streamKey:  "stream:bids",
			bidAmount:  "100",
			expireTime: "3600",
			bidInfo: BidInfo{
				ItemID:    itemID,
				BidderID:  bidderID,
				Amount:    100,
				CreatedAt: now,
			},
			want: 0,
		},
		{
			name: "競價成功時應返回1且寫入stream",
			setupFunc: func() {
				mr.Set("item:1", "100")
			},
			itemKey:    "item:1",
			streamKey:  "stream:bids",
			bidAmount:  "200",
			expireTime: "3600",
			bidInfo: BidInfo{
				ItemID:    itemID,
				BidderID:  bidderID,
				Amount:    200,
				CreatedAt: now,
			},
			want:        1,
			checkStream: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置 Redis
			mr.FlushAll()

			// 設置測試資料
			tt.setupFunc()

			// 序列化競價資訊
			bidInfoBytes, err := msgpack.Marshal(tt.bidInfo)
			assert.NoError(t, err)
			bidInfo := base64.StdEncoding.EncodeToString(bidInfoBytes)

			// 執行腳本
			result, err := BidScript.Run(ctx, client,
				[]string{tt.itemKey, tt.streamKey},
				tt.bidAmount, bidInfo, tt.expireTime,
			).Int()

			// 驗證結果
			assert.NoError(t, err)
			assert.Equal(t, tt.want, result)

			// 如果需要檢查stream
			if tt.checkStream && result == 1 {
				// 檢查最新競價金額
				val, err := client.Get(ctx, tt.itemKey).Result()
				assert.NoError(t, err)
				assert.Equal(t, tt.bidAmount, val)

				// 檢查過期時間
				ttl, err := client.TTL(ctx, tt.itemKey).Result()
				assert.NoError(t, err)
				assert.True(t, ttl > 0)

				// 檢查stream記錄
				streams, err := client.XRange(ctx, tt.streamKey, "-", "+").Result()
				assert.NoError(t, err)
				assert.Equal(t, 1, len(streams))

				// 解析stream中的競價資訊
				var streamBidInfo BidInfo
				streamBidInfo, err = redisAdapter.DefaultParseFromMessage[BidInfo](map[string]any{"data": streams[0].Values["data"]})
				assert.NoError(t, err)
				compareBidInfo(t, tt.bidInfo, streamBidInfo)
			}
		})
	}
}
