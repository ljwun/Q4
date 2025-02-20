package api

import (
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// BidInfo represents the bid information
type BidInfo struct {
	ItemID    uuid.UUID
	BidderID  uuid.UUID
	Amount    uint32
	CreatedAt time.Time
}

// BidScript 用於執行競價腳本
//
//	KEYS[1] - 競價商品鍵
//	KEYS[2] - 競價的 stream
//	ARGV[1] - 競價金額
//	ARGV[2] - 競價資訊(結構參考BidInfo，會進行msgpack和base64的處理)
//
// 返回值:
//
//	1  - 競價成功
//	0  - 競價失敗
//	-1 - 拍賣商品ID不存在
//
// 流程:
//   - 1. 檢查商品是否存在
//   - 2a. 如果不存在，返回-1
//   - 2b. 如果存在，檢查競價金額是否高於當前最高競價金額
//   - 3a. 如果不高於，返回0
//   - 3b. 如果高於，更新最高競價金額
//   - 4. 將出價資訊寫入stream
//   - 5. 返回1
var BidScript = redis.NewScript(`
-- 檢查商品是否存在
local item = redis.call('EXISTS', KEYS[1])
if item == 0 then
    return -1
end

-- 取得當前最高競價
local current_bid = tonumber(redis.call('GET', KEYS[1])) or 0
local new_bid = tonumber(ARGV[1])

-- 檢查新競價是否高於當前最高價
if new_bid <= current_bid then
    return 0
end

-- 更新最高競價
redis.call('SET', KEYS[1], new_bid)

-- 將競價記錄寫入 stream
redis.call('XADD', KEYS[2], '*', 'data', ARGV[2])

return 1
`)
