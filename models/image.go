package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Image 代表拍賣系統中的圖片
// 包含基本的圖片資訊，如圖片 URL 以及上傳者的使用者 ID
type Image struct {
	gorm.Model

	ID         uuid.UUID `gorm:"type:uuid;default:public.uuid_generate_v7();primaryKey;<-:false"`
	UploaderID uuid.UUID `gorm:"type:uuid;not null;<-:create"`
	Url        string    `gorm:"type:text;not null;<-:create"`

	Uploader *User `gorm:"foreignKey:UploaderID"`
}
