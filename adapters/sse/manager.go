package sse

import (
	"context"
	"sync"

	"github.com/smallnest/chanx"
)

// connectionManager 用於管理多個 SSE 頻道 (Channel)，
// 提供頻道註冊、取消註冊、訂閱與推送資料的功能。
type connectionManager[T any] struct {
	publisher *chanx.UnboundedChan[PublishRequest[T]]
	channels  map[string]*Channel[T]
	mu        sync.RWMutex
	cancel    context.CancelFunc
	active    bool
}

// NewConnectionManager 創建並返回一個新的 ConnectionManager 實例。
func NewConnectionManager[T any]() ConnectionManager[T] {
	ctx, cancel := context.WithCancel(context.Background())
	cm := &connectionManager[T]{
		channels:  make(map[string]*Channel[T]),
		publisher: chanx.NewUnboundedChan[PublishRequest[T]](ctx, 100),
		cancel:    cancel,
		active:    true,
	}
	go cm.pushData()
	return cm
}

// Done 停止 ConnectionManager，關閉所有頻道並取消上下文。
func (cm *connectionManager[T]) Done() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.active = false
	close(cm.publisher.In)
	cm.cancel()
}

// pushData 從 publisher 中讀取資料並廣播到相應的頻道。
func (cm *connectionManager[T]) pushData() {
	for req := range cm.publisher.Out {
		cm.mu.RLock()
		channel, ok := cm.channels[req.Channel]
		cm.mu.RUnlock()
		if !ok {
			continue
		}
		channel.Broadcast(req.Message)
	}
}

// Subscribe 註冊並訂閱指定頻道，返回一個新的 chan T。
func (cm *connectionManager[T]) Subscribe(channelName string) (chan T, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if !cm.active {
		return nil, context.Canceled
	}
	c, ok := cm.channels[channelName]
	if !ok {
		c = NewChannel[T]()
		cm.channels[channelName] = c
	}
	return c.Subscribe(), nil
}

// Publish 將資料推送到指定頻道。
func (cm *connectionManager[T]) Publish(channelName string, data T) error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	if !cm.active {
		return context.Canceled
	}
	cm.publisher.In <- PublishRequest[T]{Channel: channelName, Message: data}
	return nil
}

// Unsubscribe 取消訂閱指定頻道，並在頻道閒置時釋放頻道資源。
func (cm *connectionManager[T]) Unsubscribe(channelName string, ch chan T) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	c, ok := cm.channels[channelName]
	if !ok {
		return
	}
	c.Unsubscribe(ch)
	if c.IsIdle() {
		delete(cm.channels, channelName)
	}
}
