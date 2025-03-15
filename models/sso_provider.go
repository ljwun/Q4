package models

import (
	"q4/api/openapi"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SsoProvider 代表支援的 SSO 提供者
// 包含基本的 SSO 提供者資訊，如名稱
type SsoProvider struct {
	gorm.Model

	ID   uuid.UUID           `gorm:"type:uuid;default:public.uuid_generate_v7();primaryKey;<-:false"`
	Name openapi.SSOProvider `gorm:"type:text;not null;unique;<-:create"`
}
