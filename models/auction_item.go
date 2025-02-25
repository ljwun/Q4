package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuctionItem 代表拍賣系統中的商品
// 包含商品資訊、起標價、目前最高出價、拍賣時間等資訊
type AuctionItem struct {
	gorm.Model

	ID            uuid.UUID  `gorm:"type:uuid;default:public.uuid_generate_v7();primaryKey;<-:false"`
	UserID        uuid.UUID  `gorm:"type:uuid;<-:create"`
	Title         string     `gorm:"type:varchar(255);not null"`
	Description   string     `gorm:"type:text;not null"`
	StartingPrice uint32     `gorm:"type:integer;not null"`
	CurrentBidID  *uuid.UUID `gorm:"type:uuid;"`
	StartTime     time.Time  `gorm:"type:timestamp with time zone;not null"`
	EndTime       time.Time  `gorm:"type:timestamp with time zone;not null"`
	Carousels     []string   `gorm:"type:text[];default:'{}'"`

	// 外鍵關聯
	User       User
	CurrentBid *Bid `gorm:"foreignKey:CurrentBidID"`
	BidRecords []Bid
}
