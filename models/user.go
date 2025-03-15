package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User 代表拍賣系統中的使用者
// 包含基本的使用者資訊，如使用者名稱
type User struct {
	gorm.Model

	ID       uuid.UUID `gorm:"type:uuid;default:public.uuid_generate_v7();primaryKey;<-:false"`
	Username string    `gorm:"type:varchar(255);not null;<-:create"`
}
