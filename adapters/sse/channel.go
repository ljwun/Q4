package sse

import (
	"sync"
)

// Channel 用於管理針對某個主題 (Topic) 的所有訂閱者，
// 並將接收到的訊息廣播給所有訂閱者。
type Channel[T any] struct {
	subscribers map[<-chan T]chan<- T
	mu          sync.RWMutex
}

// NewChannel creates a new SSE channel.
func NewChannel[T any]() IChannel[T] {
	return &Channel[T]{
		subscribers: make(map[<-chan T]chan<- T),
	}
}

// Subscribe 建立一個新的 chan T，將其加入 subscribers，並回傳唯讀通道給呼叫者。
func (c *Channel[T]) Subscribe() <-chan T {
	c.mu.Lock()
	defer c.mu.Unlock()
	ch := make(chan T)
	c.subscribers[ch] = ch
	return ch
}

// Unsubscribe 從 subscribers 中移除指定的通道，並關閉該通道。
func (c *Channel[T]) Unsubscribe(ch <-chan T) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if writeCh, exists := c.subscribers[ch]; exists {
		delete(c.subscribers, ch)
		close(writeCh)
	}
}

// UnsubscribeAll 關閉所有訂閱者的通道並清空訂閱清單。
func (c *Channel[T]) UnsubscribeAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, writeCh := range c.subscribers {
		close(writeCh)
	}
	clear(c.subscribers)
}

// Broadcast 將訊息廣播給所有仍在訂閱清單中的通道。
func (c *Channel[T]) Broadcast(message T) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, writeCh := range c.subscribers {
		writeCh <- message
	}
}

// IsIdle 判斷 subscribers 是否為空。
func (c *Channel[T]) IsIdle() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.subscribers) == 0
}
