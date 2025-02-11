package sse

// PublishRequest 表示一個發布請求，包含頻道名稱和訊息。
type PublishRequest[T any] struct {
	Channel string `json:"channel"`
	Message T      `json:"message"`
}

// ConnectionManager 定義了 SSE 連線管理員的接口，
// 包含訂閱、取消訂閱、發布訊息和完成的方法。
type ConnectionManager[T any] interface {
	// Start 啟動 ConnectionManager，開始處理訊息的接收與廣播。
	// 應在呼叫其他方法前先呼叫此方法。
	Start()
	// Done 停止 ConnectionManager，釋放所有資源。
	Done()
	// Subscribe 註冊並訂閱指定頻道，返回一個新的 chan Message。
	Subscribe(channelName string) (chan T, error)
	// Publish 將資料推送到指定頻道。
	Publish(channelName string, data T) error
	// Unsubscribe 取消訂閱指定頻道。
	Unsubscribe(channelName string, ch chan T)
}
