package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Bid 代表拍賣商品的出價紀錄
// 記錄每次競標的金額、競標者和競標商品
type Bid struct {
	*gorm.Model

	ID            uuid.UUID `gorm:"type:uuid;default:public.uuid_generate_v7();primaryKey;<-:false"`
	Amount        uint32    `gorm:"type:integer;not null;<-:create"`
	UserID        uuid.UUID `gorm:"type:uuid;not null;<-:create"`
	AuctionItemID uuid.UUID `gorm:"type:uuid;not null;<-:create"`

	// 外鍵關聯
	User        User
	AuctionItem AuctionItem
}
