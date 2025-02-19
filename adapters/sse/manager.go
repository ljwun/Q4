package sse

import (
	"context"
	"log/slog"
	"sync"
)

type Subscriber[T any] interface {
	Subscribe() <-chan T
}

type Publisher[T any] interface {
	Publish(data T) error
}

// connectionManager 管理多個 SSE 頻道的訂閱與發布。
// 透過 Redis Stream 實現跨節點的訊息廣播，讓多個服務實例能夠協同運作。
type connectionManager[T any] struct {
	cancel context.CancelFunc
	logger *slog.Logger

	mu     sync.RWMutex   // 保護 active 和 channels 的讀寫
	wg     sync.WaitGroup // 用於等待所有 goroutine 完成
	active bool           // 標記 manager 是否正在運作中

	channels   map[string]*Channel[T]        // 儲存所有活躍的頻道
	subscriber Subscriber[PublishRequest[T]] // 用於訂閱上游的訊息
	publisher  Publisher[PublishRequest[T]]  // 用於發送訊息到上游
}

// NewConnectionManager 建立一個新的連線管理器
func NewConnectionManager[T any](sub Subscriber[PublishRequest[T]], pub Publisher[PublishRequest[T]], logger *slog.Logger) ConnectionManager[T] {
	if logger == nil {
		logger = slog.Default()
	}

	return &connectionManager[T]{
		logger:     logger.With("Caller", "ConnectionManager"),
		channels:   make(map[string]*Channel[T]),
		subscriber: sub,
		publisher:  pub,
		active:     true,
	}
}

// Start 啟動連線管理器，開始處理訊息的接收與廣播。
// 應在呼叫其他方法前先呼叫此方法。
func (cm *connectionManager[T]) Start() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if cm.active {
		return
	}
	cm.active = true
	ctx, cancel := context.WithCancel(context.Background())
	cm.cancel = cancel
	// 啟動訊息處理的 goroutine
	cm.wg.Add(1)
	go func() {
		defer cm.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-cm.subscriber.Subscribe():
				if !ok {
					continue
				}
				cm.mu.RLock()
				if channel, ok := cm.channels[msg.Channel]; ok {
					channel.Broadcast(msg.Message)
				}
				cm.mu.RUnlock()
			}
		}
	}()
}

// Done 停止連線管理器的運作。
func (cm *connectionManager[T]) Done() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if !cm.active {
		return
	}

	cm.active = false
	cm.cancel()
	cm.wg.Wait()
	for _, channel := range cm.channels {
		channel.UnsubscribeAll()
	}
	clear(cm.channels)
}

// Subscribe 訂閱指定的頻道。
// channelName: 要訂閱的頻道名稱
// 返回: 用於接收訊息的唯讀通道，以及可能的錯誤
func (cm *connectionManager[T]) Subscribe(channelName string) (<-chan T, error) {
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

// Publish 發布訊息到指定的頻道。
// channelName: 目標頻道名稱
// data: 要發布的訊息內容
func (cm *connectionManager[T]) Publish(channelName string, data T) error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if !cm.active {
		return context.Canceled
	}

	return cm.publisher.Publish(PublishRequest[T]{
		Channel: channelName,
		Message: data,
	})
}

// Unsubscribe 取消訂閱指定的頻道。
func (cm *connectionManager[T]) Unsubscribe(channelName string, ch <-chan T) {
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
