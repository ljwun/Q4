package sse_test

// Message 表示一個 SSE 訊息，包含資料字段。
type Message struct {
	Data string `json:"data"`
}
