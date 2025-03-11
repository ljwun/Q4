package session

import (
	"context"
	"fmt"
)

// sessionImpl 實作 ISession 介面，用於管理使用者會話
type sessionImpl struct {
	id    string            // session ID
	ctx   context.Context   // 操作上下文
	data  map[string]string // session 資料
	store IStore            // session 儲存接口
}

// NewSession 建立新的 session 實例
func NewSession(ctx context.Context, id string, store IStore) ISession {
	if ctx == nil {
		ctx = context.Background()
	}
	return &sessionImpl{
		id:    id,
		ctx:   ctx,
		store: store,
		data:  nil,
	}
}

// Load 從儲存層載入 session 資料
func (s *sessionImpl) Load() error {
	const op = "sessionImpl.Load"
	// 如果已經載入過，則直接返回
	if s.data != nil {
		return nil
	}

	data, err := s.store.Load(s.ctx, s.id)
	if err != nil {
		return fmt.Errorf("%s: failed to load session: %w", op, err)
	}

	s.data = data
	if s.data == nil {
		s.data = make(map[string]string)
	}
	return nil
}

// Get 取得指定 key 的值
func (s *sessionImpl) Get(key string) string {
	if s.data == nil {
		return ""
	}
	return s.data[key]
}

// Set 設定 key-value 對
func (s *sessionImpl) Set(key string, value string) {
	if s.data == nil {
		s.data = make(map[string]string)
	}
	s.data[key] = value
}

// Delete 刪除指定 key 的值
func (s *sessionImpl) Delete(key string) {
	if s.data != nil {
		delete(s.data, key)
	}
}

// Clear 清空 session 資料
func (s *sessionImpl) Clear() {
	s.data = make(map[string]string)
}

// Save 保存 session 資料到儲存層
func (s *sessionImpl) Save() error {
	const op = "sessionImpl.Save"
	if s.data == nil {
		return nil
	}
	if err := s.store.Save(s.ctx, s.id, s.data); err != nil {
		return fmt.Errorf("%s: failed to save session: %w", op, err)
	}
	return nil
}
