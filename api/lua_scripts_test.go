package api

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

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

	tests := []struct {
		name        string
		setupFunc   func()
		itemKey     string
		streamKey   string
		itemID      string
		bidderID    string
		bidAmount   string
		timestamp   string
		want        int
		checkStream bool
	}{
		{
			name:      "商品不存在時應返回-1",
			setupFunc: func() {},
			itemKey:   "item:nonexistent",
			streamKey: "stream:bids",
			itemID:    "1",
			bidderID:  "user1",
			bidAmount: "100",
			timestamp: "1640995200000",
			want:      -1,
		},
		{
			name: "出價金額不足時應返回0",
			setupFunc: func() {
				mr.Set("item:1", "200")
			},
			itemKey:   "item:1",
			streamKey: "stream:bids",
			itemID:    "1",
			bidderID:  "user1",
			bidAmount: "100",
			timestamp: "1640995200000",
			want:      0,
		},
		{
			name: "競價成功時應返回1且寫入stream",
			setupFunc: func() {
				mr.Set("item:1", "100")
			},
			itemKey:     "item:1",
			streamKey:   "stream:bids",
			itemID:      "1",
			bidderID:    "user1",
			bidAmount:   "200",
			timestamp:   "1640995200000",
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

			// 執行腳本
			result, err := BidScript.Run(ctx, client,
				[]string{tt.itemKey, tt.streamKey},
				tt.itemID, tt.bidderID, tt.bidAmount, tt.timestamp,
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

				// 檢查stream記錄
				streams, err := client.XRange(ctx, tt.streamKey, "-", "+").Result()
				assert.NoError(t, err)
				assert.Equal(t, 1, len(streams))
				assert.Equal(t, tt.itemID, streams[0].Values["item_id"])
				assert.Equal(t, tt.bidderID, streams[0].Values["user_id"])
				assert.Equal(t, tt.bidAmount, streams[0].Values["bid"])
				assert.Equal(t, tt.timestamp, streams[0].Values["time"])
			}
		})
	}
}
