package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserIdentity 代表使用者的身份
// 包含基本的身份資訊，如 SSO 提供者 ID、使用者 ID 以及識別字串，用來識別使用者在 SSO 提供者的身份
type UserIdentity struct {
	gorm.Model

	ID            uuid.UUID `gorm:"type:uuid;default:public.uuid_generate_v7();primaryKey;<-:false"`
	SsoProviderID uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_user_identity_sso_provider_id_user_id,where:deleted_at IS NULL;uniqueIndex:idx_user_identity_sso_provider_id_identity,where:deleted_at IS NULL;not null;<-:create"`
	UserID        uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_user_identity_sso_provider_id_user_id,where:deleted_at IS NULL;not null;<-:create"`
	Identity      string    `gorm:"type:text;uniqueIndex:idx_user_identity_sso_provider_id_identity,where:deleted_at IS NULL;not null;<-:create"`

	SsoProvider *SsoProvider `gorm:"foreignKey:SsoProviderID"`
	User        *User        `gorm:"foreignKey:UserID"`
}
