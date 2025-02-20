package sse_test

import (
	"io"
	"log"
)

func init() {
	// 將日誌輸出重定向到io.Discard
	log.SetOutput(io.Discard)
}

// Message 表示一個 SSE 訊息，包含資料字段。
type Message struct {
	Data string `json:"data"`
}
