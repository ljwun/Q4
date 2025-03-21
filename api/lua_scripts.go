package api

import (
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// BidInfoUser represents a user
type BidInfoUser struct {
	ID   uuid.UUID
	Name string
}

// BidInfo represents the bid information
type BidInfo struct {
	ItemID    uuid.UUID
	User      BidInfoUser
	Amount    uint32
	CreatedAt time.Time
}

// BidScript 用於執行競價腳本
//
//	KEYS[1] - 競價商品鍵
//	KEYS[2] - 競價的 stream
//	ARGV[1] - 競價金額
//	ARGV[2] - 競價資訊(結構參考BidInfo，會進行msgpack和base64的處理)
//	ARGV[3] - 過期時間(秒)
//	ARGV[4] - 預設最高競價金額
//
// 返回值:
//
//	1 - 競價成功
//	0 - 競價失敗
//
// 流程:
//   - 1. 取得當前最高競價，如果不存在則使用預設值
//   - 2a. 如果新競價金額不高於當前最高競價，返回0
//   - 2b. 如果新競價金額高於當前最高競價，更新最高競價金額
//   - 3. 將出價資訊寫入stream
//   - 4. 返回1
var BidScript = redis.NewScript(`
-- 取得當前最高競價，如果不存在則使用預設值
local current_bid = tonumber(redis.call('GET', KEYS[1])) or tonumber(ARGV[4])
local new_bid = tonumber(ARGV[1])

-- 檢查新競價是否高於當前最高價
if new_bid <= current_bid then
    return 0
end

-- 更新最高競價
redis.call('SET', KEYS[1], new_bid, 'EX', ARGV[3])

-- 將競價記錄寫入 stream
redis.call('XADD', KEYS[2], '*', 'data', ARGV[2])

return 1
`)
