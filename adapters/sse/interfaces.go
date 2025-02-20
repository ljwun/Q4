//go:generate mockgen -package=sse -destination=mock.go -source=interfaces.go

package sse

// PublishRequest 表示一個發布請求，包含頻道名稱和訊息。
type PublishRequest[T any] struct {
	Channel string `json:"channel"`
	Message T      `json:"message"`
}

// IChannel 定義了 SSE 頻道的介面
type IChannel[T any] interface {
	// Subscribe 建立一個新的訂閱並返回接收訊息的通道
	Subscribe() <-chan T
	// Unsubscribe 取消指定通道的訂閱
	Unsubscribe(ch <-chan T)
	// UnsubscribeAll 取消所有訂閱
	UnsubscribeAll()
	// Broadcast 將訊息廣播給所有訂閱者
	Broadcast(message T)
	// IsIdle 檢查是否沒有訂閱者
	IsIdle() bool
}

// IConnectionManager 定義了 SSE 連線管理員的介面
type IConnectionManager[T any] interface {
	// Start 啟動 ConnectionManager，開始處理訊息的接收與廣播。
	// 應在呼叫其他方法前先呼叫此方法。
	Start()
	// Done 停止 ConnectionManager，釋放所有資源。
	Done()
	// Subscribe 註冊並訂閱指定頻道，返回一個新的 chan Message。
	Subscribe(channelName string) (<-chan T, error)
	// Publish 將資料推送到指定頻道。
	Publish(channelName string, data T) error
	// Unsubscribe 取消訂閱指定頻道。
	Unsubscribe(channelName string, ch <-chan T)
}
