package sse

import (
	"sync"
)

// Channel 用於管理針對某個主題 (Topic) 的所有訂閱者，
// 並將接收到的訊息廣播給所有訂閱者。
type Channel[T any] struct {
	subscribers map[chan T]struct{}
	mu          sync.RWMutex
}

// NewChannel creates a new SSE channel.
func NewChannel[T any]() *Channel[T] {
	return &Channel[T]{
		subscribers: make(map[chan T]struct{}),
	}
}

// Subscribe 建立一個新的 chan T，將其加入 subscribers，並回傳給呼叫者。
func (c *Channel[T]) Subscribe() chan T {
	c.mu.Lock()
	defer c.mu.Unlock()
	newCh := make(chan T)
	c.subscribers[newCh] = struct{}{}
	return newCh
}

// Unsubscribe 從 subscribers 中移除指定的 chan T，並關閉該通道。
func (c *Channel[T]) Unsubscribe(ch chan T) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.subscribers, ch)
	close(ch)
}

// UnsubscribeAll 關閉所有訂閱者的通道並清空訂閱清單。
func (c *Channel[T]) UnsubscribeAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 關閉所有訂閱者的通道
	for ch := range c.subscribers {
		close(ch)
	}
	// 清空訂閱清單
	clear(c.subscribers)
}

// Broadcast 將訊息廣播給所有仍在訂閱清單中的通道。
func (c *Channel[T]) Broadcast(message T) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for ch := range c.subscribers {
		ch <- message
	}
}

// IsIdle 判斷 subscribers 是否為空。
func (c *Channel[T]) IsIdle() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.subscribers) == 0
}
