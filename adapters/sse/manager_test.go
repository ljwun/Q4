package sse_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"

	"q4/adapters/sse"
)

func TestConnectionManager(t *testing.T) {
	defer goleak.VerifyNone(t)

	cm := sse.NewConnectionManager[Message]()
	defer cm.Done()

	// 測試訂閱
	ch, err := cm.Subscribe("test_channel")
	assert.NoError(t, err)
	assert.NotNil(t, ch)

	// 測試發布訊息
	msg := Message{Data: "test message"}
	err = cm.Publish("test_channel", msg)
	assert.NoError(t, err)

	select {
	case received := <-ch:
		assert.Equal(t, msg, received)
	case <-time.After(time.Second):
		t.Fatal("did not receive message in time")
	}

	// 測試取消訂閱
	cm.Unsubscribe("test_channel", ch)
	_, ok := <-ch
	assert.False(t, ok, "channel should be closed")
}
