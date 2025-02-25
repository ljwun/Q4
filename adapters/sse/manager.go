package sse

import (
	"context"
	"errors"
	"log/slog"
	"sync"
)

var (
	// ErrSubscriberRequired 表示建立連線管理器時未提供Subscriber
	ErrSubscriberRequired = errors.New("subscriber is required")
	// ErrPublisherNotConfigured 表示嘗試發布訊息時，Publisher未被設定
	ErrPublisherNotConfigured = errors.New("publisher not configured")
)

// Subscriber 定義訊息訂閱的介面
// T 為訊息的資料型別
type Subscriber[T any] interface {
	// Subscribe 開始訂閱訊息
	// 返回一個用於接收訊息的唯讀通道
	Subscribe() <-chan T
}

// Publisher 定義訊息發布的介面
// T 為訊息的資料型別
type Publisher[T any] interface {
	// Publish 發布一則訊息
	// 返回發布過程中可能發生的錯誤
	Publish(data T) error
}

// ConnectionManager 管理多個 SSE 頻道的訂閱與發布
// 可透過上游的訊息佇列或其他機制接收訊息，並分發到不同的頻道，實現跨伺服器的訊息廣播
type ConnectionManager[T any] struct {
	cancel context.CancelFunc
	logger *slog.Logger

	mu     sync.RWMutex   // 保護 active 和 channels 的讀寫
	wg     sync.WaitGroup // 用於等待所有 goroutine 完成
	active bool           // 標記 manager 是否正在運作中

	channels          map[string]IChannel[T]        // 儲存所有活躍的頻道
	subscriber        Subscriber[PublishRequest[T]] // 用於訂閱上游的訊息
	publisher         Publisher[PublishRequest[T]]  // 用於發送訊息到上游
	createChannelFunc func() IChannel[T]            // 用於建立新頻道的函數
}

// connectionManagerOptions 定義連線管理器的設定選項
type connectionManagerOptions[T any] struct {
	subscriber        Subscriber[PublishRequest[T]] // 必要：用於接收上游訊息的訂閱者
	publisher         Publisher[PublishRequest[T]]  // 選用：用於發送訊息到上游的發布者
	logger            *slog.Logger                  // 選用：用於記錄日誌的 logger
	createChannelFunc func() IChannel[T]            // 選用：用於建立新頻道的函數
}

// ConnectionManagerOption 定義設定選項的函式型別
type ConnectionManagerOption[T any] func(*connectionManagerOptions[T])

// WithSubscriber 設定連線管理器的訂閱者
// sub: 實作 Subscriber 介面的訂閱者實例
func WithSubscriber[T any](sub Subscriber[PublishRequest[T]]) ConnectionManagerOption[T] {
	return func(o *connectionManagerOptions[T]) {
		o.subscriber = sub
	}
}

// WithPublisher 設定連線管理器的發布者
// pub: 實作 Publisher 介面的發布者實例
func WithPublisher[T any](pub Publisher[PublishRequest[T]]) ConnectionManagerOption[T] {
	return func(o *connectionManagerOptions[T]) {
		o.publisher = pub
	}
}

// WithLogger 設定連線管理器的日誌器
// logger: 用於記錄系統事件的 logger，若未設定則使用預設的 logger
func WithLogger[T any](logger *slog.Logger) ConnectionManagerOption[T] {
	return func(o *connectionManagerOptions[T]) {
		o.logger = logger
	}
}

// WithCreateChannelFunc 設定建立新頻道的函數
// createFunc: 用於建立新頻道的函數，主要用於測試時注入mock物件
func WithCreateChannelFunc[T any](createFunc func() IChannel[T]) ConnectionManagerOption[T] {
	return func(o *connectionManagerOptions[T]) {
		o.createChannelFunc = createFunc
	}
}

// NewConnectionManager 建立一個新的連線管理器
// opts: 可變參數，用於設定連線管理器的選項
// 返回:
//   - ConnectionManager: 連線管理器實例
//   - error: 若未提供必要的訂閱者，將返回 ErrSubscriberRequired
func NewConnectionManager[T any](opts ...ConnectionManagerOption[T]) (IConnectionManager[T], error) {
	options := connectionManagerOptions[T]{
		logger:            slog.Default(),
		createChannelFunc: NewChannel[T], // 設定預設的建立頻道函數
	}

	for _, opt := range opts {
		opt(&options)
	}

	if options.subscriber == nil {
		return nil, ErrSubscriberRequired
	}

	return &ConnectionManager[T]{
		logger:            options.logger.With("Caller", "ConnectionManager"),
		channels:          make(map[string]IChannel[T]),
		subscriber:        options.subscriber,
		publisher:         options.publisher,
		active:            false,
		createChannelFunc: options.createChannelFunc,
	}, nil
}

// Start 啟動連線管理器
// 開始監聽並處理上游的訊息，須在呼叫其他方法前先呼叫此方法
func (cm *ConnectionManager[T]) Start() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if cm.active {
		return
	}
	cm.active = true
	ctx, cancel := context.WithCancel(context.Background())
	cm.cancel = cancel
	cm.logger.Info("ConnectionManager started")

	// 啟動訊息處理的 goroutine
	cm.wg.Add(1)
	go func() {
		defer cm.wg.Done()
		defer cm.logger.Info("ConnectionManager stopped")
		upstream := cm.subscriber.Subscribe()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-upstream:
				if !ok {
					continue
				}
				cm.logger.Debug("received message", slog.Any("message", msg))
				cm.mu.RLock()
				if channel, ok := cm.channels[msg.Channel]; ok {
					channel.Broadcast(msg.Message)
				}
				cm.mu.RUnlock()
			}
		}
	}()
}

// Done 停止連線管理器的運作
// 會清理所有活躍的頻道並停止訊息處理
func (cm *ConnectionManager[T]) Done() {
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

// Subscribe 訂閱指定的頻道
// 參數:
//   - channelName: 要訂閱的頻道名稱
//
// 返回:
//   - <-chan T: 用於接收訊息的唯讀通道
//   - error: 若管理器已停止運作，將返回 context.Canceled
func (cm *ConnectionManager[T]) Subscribe(channelName string) (<-chan T, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if !cm.active {
		return nil, context.Canceled
	}

	c, ok := cm.channels[channelName]
	if !ok {
		c = cm.createChannelFunc()
		cm.channels[channelName] = c
	}
	return c.Subscribe(), nil
}

// Publish 發布訊息到指定的頻道
// 參數:
//   - channelName: 目標頻道名稱
//   - data: 要發布的訊息內容
//
// 返回:
//   - error: 可能的錯誤包含：
//   - context.Canceled: 管理器已停止運作
//   - ErrPublisherNotConfigured: 未設定Publisher
func (cm *ConnectionManager[T]) Publish(channelName string, data T) error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if !cm.active {
		return context.Canceled
	}

	if cm.publisher == nil {
		return ErrPublisherNotConfigured
	}

	return cm.publisher.Publish(PublishRequest[T]{
		Channel: channelName,
		Message: data,
	})
}

// Unsubscribe 取消訂閱指定的頻道
// 若該頻道已無訂閱者，會自動移除該頻道
// 參數:
//   - channelName: 要取消訂閱的頻道名稱
//   - ch: 訂閱時取得的訊息通道
func (cm *ConnectionManager[T]) Unsubscribe(channelName string, ch <-chan T) {
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
