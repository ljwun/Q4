package sse_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"q4/adapters/sse"
)

func TestChannel(t *testing.T) {
	ch := sse.NewChannel[Message]()

	// 測試訂閱
	sub := ch.Subscribe()
	assert.NotNil(t, sub)

	// 測試廣播訊息
	msg := Message{Data: "test message"}
	go ch.Broadcast(msg)

	select {
	case received := <-sub:
		assert.Equal(t, msg, received)
	case <-time.After(time.Second):
		t.Fatal("did not receive message in time")
	}

	// 測試取消訂閱
	ch.Unsubscribe(sub)
	_, ok := <-sub
	assert.False(t, ok, "channel should be closed")

	// 測試 IsIdle
	assert.True(t, ch.IsIdle(), "channel should be idle")
}

func TestChannel_UnsubscribeAll(t *testing.T) {
	ch := sse.NewChannel[Message]()

	// 建立多個訂閱者
	subs := make([]chan Message, 3)
	for i := range subs {
		subs[i] = ch.Subscribe()
	}

	// 確認通道都建立成功
	assert.False(t, ch.IsIdle(), "channel should not be idle with subscribers")

	// 執行 UnsubscribeAll
	ch.UnsubscribeAll()

	// 驗證所有通道都已關閉
	for i, sub := range subs {
		_, ok := <-sub
		assert.False(t, ok, "subscriber %d should be closed", i)
	}

	// 驗證訂閱清單已清空
	assert.True(t, ch.IsIdle(), "channel should be idle after UnsubscribeAll")
}
