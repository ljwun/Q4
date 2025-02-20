package sse_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"go.uber.org/mock/gomock"

	"q4/adapters/redis"
	"q4/adapters/sse"
)

type testConnectionManagerSetup struct {
	ctrl       *gomock.Controller
	subscriber *redis.MockIConsumer[sse.PublishRequest[Message]]
	publisher  *redis.MockIProducer[sse.PublishRequest[Message]]
	channel    *sse.MockIChannel[Message]
}

func setupConnectionManagerTest(t *testing.T) testConnectionManagerSetup {
	ctrl := gomock.NewController(t)
	return testConnectionManagerSetup{
		ctrl:       ctrl,
		subscriber: redis.NewMockIConsumer[sse.PublishRequest[Message]](ctrl),
		publisher:  redis.NewMockIProducer[sse.PublishRequest[Message]](ctrl),
		channel:    sse.NewMockIChannel[Message](ctrl),
	}
}

func TestNewConnectionManager(t *testing.T) {
	tests := []struct {
		name    string
		opts    []sse.ConnectionManagerOption[Message]
		wantErr error
	}{
		{
			name: "with subscriber",
			opts: []sse.ConnectionManagerOption[Message]{
				sse.WithSubscriber[Message](redis.NewMockIConsumer[sse.PublishRequest[Message]](gomock.NewController(t))),
			},
			wantErr: nil,
		},
		{
			name: "all options",
			opts: []sse.ConnectionManagerOption[Message]{
				sse.WithSubscriber[Message](redis.NewMockIConsumer[sse.PublishRequest[Message]](gomock.NewController(t))),
				sse.WithPublisher[Message](redis.NewMockIProducer[sse.PublishRequest[Message]](gomock.NewController(t))),
				sse.WithLogger[Message](slog.Default()),
			},
			wantErr: nil,
		},
		{
			name:    "no option",
			opts:    []sse.ConnectionManagerOption[Message]{},
			wantErr: sse.ErrSubscriberRequired,
		},
		{
			name: "nil publisher",
			opts: []sse.ConnectionManagerOption[Message]{
				sse.WithSubscriber[Message](redis.NewMockIConsumer[sse.PublishRequest[Message]](gomock.NewController(t))),
			},
			wantErr: nil,
		},
		{
			name: "custom channel creator",
			opts: []sse.ConnectionManagerOption[Message]{
				sse.WithSubscriber[Message](redis.NewMockIConsumer[sse.PublishRequest[Message]](gomock.NewController(t))),
				sse.WithCreateChannelFunc[Message](func() sse.IChannel[Message] {
					return sse.NewChannel[Message]()
				}),
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer goleak.VerifyNone(t)
			cm, err := sse.NewConnectionManager(tt.opts...)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, cm)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cm)
			}
		})
	}
}

func TestConnectionManager_Start(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		setup := setupConnectionManagerTest(t)
		defer setup.ctrl.Finish()

		msgChan := make(chan sse.PublishRequest[Message])
		setup.subscriber.EXPECT().Subscribe().Return(msgChan)

		cm, err := sse.NewConnectionManager[Message](
			sse.WithSubscriber[Message](setup.subscriber),
		)
		assert.NoError(t, err)

		cm.Start()
		close(msgChan)
		cm.Done()
	})

	t.Run("multiple start", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		setup := setupConnectionManagerTest(t)
		defer setup.ctrl.Finish()

		msgChan := make(chan sse.PublishRequest[Message])
		setup.subscriber.EXPECT().Subscribe().Return(msgChan)

		cm, err := sse.NewConnectionManager[Message](
			sse.WithSubscriber[Message](setup.subscriber),
		)
		assert.NoError(t, err)

		cm.Start()
		cm.Start()
		close(msgChan)
		cm.Done()
	})

	t.Run("process messages", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		setup := setupConnectionManagerTest(t)
		defer setup.ctrl.Finish()

		msgChan := make(chan sse.PublishRequest[Message])
		setup.subscriber.EXPECT().Subscribe().Return(msgChan)

		cm, err := sse.NewConnectionManager[Message](
			sse.WithSubscriber[Message](setup.subscriber),
		)
		assert.NoError(t, err)

		cm.Start()

		// Subscribe to create a channel
		subCh, err := cm.Subscribe("test-channel")
		assert.NoError(t, err)

		// Send test message
		select {
		case msgChan <- sse.PublishRequest[Message]{
			Channel: "test-channel",
			Message: Message{Data: "test"},
		}:
		case <-time.After(time.Second):
			t.Fatal("timeout sending message")
		}

		// Verify message received
		select {
		case msg := <-subCh:
			assert.Equal(t, "test", msg.Data)
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for message")
		}

		// Clean up
		close(msgChan)
		cm.Done()
	})
}

func TestConnectionManager_Subscribe(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		setup := setupConnectionManagerTest(t)
		defer setup.ctrl.Finish()

		msgChan := make(chan sse.PublishRequest[Message])
		setup.subscriber.EXPECT().Subscribe().Return(msgChan)

		cm, _ := sse.NewConnectionManager[Message](
			sse.WithSubscriber[Message](setup.subscriber),
		)
		cm.Start()

		ch, err := cm.Subscribe("test")
		assert.NoError(t, err)
		assert.NotNil(t, ch)

		// Clean up
		cm.Unsubscribe("test", ch)
		close(msgChan)
		cm.Done()
	})

	t.Run("inactive manager", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		setup := setupConnectionManagerTest(t)
		defer setup.ctrl.Finish()

		msgChan := make(chan sse.PublishRequest[Message])
		setup.subscriber.EXPECT().Subscribe().Return(msgChan)

		cm, _ := sse.NewConnectionManager[Message](
			sse.WithSubscriber[Message](setup.subscriber),
		)
		cm.Start()
		cm.Done()

		ch, err := cm.Subscribe("test")
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
		assert.Nil(t, ch)

		close(msgChan)
	})

	t.Run("custom channel creator", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		setup := setupConnectionManagerTest(t)
		defer setup.ctrl.Finish()

		msgChan := make(chan sse.PublishRequest[Message])
		setup.subscriber.EXPECT().Subscribe().Return(msgChan)

		customCreator := func() sse.IChannel[Message] {
			return setup.channel
		}
		setup.channel.EXPECT().Subscribe().Return(make(chan Message))
		setup.channel.EXPECT().UnsubscribeAll()

		cm, _ := sse.NewConnectionManager[Message](
			sse.WithSubscriber[Message](setup.subscriber),
			sse.WithCreateChannelFunc[Message](customCreator),
		)
		cm.Start()

		ch, err := cm.Subscribe("test")
		assert.NoError(t, err)
		assert.NotNil(t, ch)

		close(msgChan)
		cm.Done()
	})

}

func TestConnectionManager_Publish(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		setup := setupConnectionManagerTest(t)
		defer setup.ctrl.Finish()

		msgChan := make(chan sse.PublishRequest[Message])
		setup.subscriber.EXPECT().Subscribe().Return(msgChan)

		testMsg := Message{Data: "test"}
		setup.publisher.EXPECT().Publish(sse.PublishRequest[Message]{
			Channel: "test",
			Message: testMsg,
		}).Return(nil)

		cm, _ := sse.NewConnectionManager[Message](
			sse.WithSubscriber[Message](setup.subscriber),
			sse.WithPublisher[Message](setup.publisher),
		)
		cm.Start()

		err := cm.Publish("test", testMsg)
		assert.NoError(t, err)

		// Clean up
		close(msgChan)
		cm.Done()
	})

	t.Run("no publisher", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		setup := setupConnectionManagerTest(t)
		defer setup.ctrl.Finish()

		msgChan := make(chan sse.PublishRequest[Message])
		setup.subscriber.EXPECT().Subscribe().Return(msgChan)

		cm, _ := sse.NewConnectionManager[Message](
			sse.WithSubscriber[Message](setup.subscriber),
		)
		cm.Start()

		err := cm.Publish("test", Message{Data: "test"})
		assert.Error(t, err)
		assert.ErrorIs(t, err, sse.ErrPublisherNotConfigured)

		// Clean up
		close(msgChan)
		cm.Done()
	})

	t.Run("inactive manager", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		setup := setupConnectionManagerTest(t)
		defer setup.ctrl.Finish()

		msgChan := make(chan sse.PublishRequest[Message])
		setup.subscriber.EXPECT().Subscribe().Return(msgChan)

		cm, _ := sse.NewConnectionManager[Message](
			sse.WithSubscriber[Message](setup.subscriber),
			sse.WithPublisher[Message](setup.publisher),
		)
		cm.Start()
		cm.Done()

		err := cm.Publish("test", Message{Data: "test"})
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)

		close(msgChan)
	})
}

func TestConnectionManager_Unsubscribe(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		setup := setupConnectionManagerTest(t)
		defer setup.ctrl.Finish()

		msgChan := make(chan sse.PublishRequest[Message])
		setup.subscriber.EXPECT().Subscribe().Return(msgChan)

		cm, _ := sse.NewConnectionManager[Message](
			sse.WithSubscriber[Message](setup.subscriber),
		)
		cm.Start()

		ch, _ := cm.Subscribe("test")
		cm.Unsubscribe("test", ch)

		// Clean up
		close(msgChan)
		cm.Done()
	})

	t.Run("nonexistent channel", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		setup := setupConnectionManagerTest(t)
		defer setup.ctrl.Finish()

		msgChan := make(chan sse.PublishRequest[Message])
		setup.subscriber.EXPECT().Subscribe().Return(msgChan)

		cm, _ := sse.NewConnectionManager[Message](
			sse.WithSubscriber[Message](setup.subscriber),
		)
		cm.Start()

		ch := make(chan Message)
		cm.Unsubscribe("non-existent", ch)

		// Clean up
		close(msgChan)
		cm.Done()
	})
}

func TestConnectionManager_Done(t *testing.T) {
	t.Run("cleanup", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		setup := setupConnectionManagerTest(t)
		defer setup.ctrl.Finish()

		msgChan := make(chan sse.PublishRequest[Message])
		setup.subscriber.EXPECT().Subscribe().Return(msgChan)

		cm, _ := sse.NewConnectionManager[Message](
			sse.WithSubscriber[Message](setup.subscriber),
		)
		cm.Start()

		// Subscribe to create some channels
		ch1, _ := cm.Subscribe("test1")
		ch2, _ := cm.Subscribe("test2")

		// Call Done
		cm.Done()

		// Verify channels are closed
		_, ok1 := <-ch1
		_, ok2 := <-ch2
		assert.False(t, ok1)
		assert.False(t, ok2)

		// Call Done again to test idempotency
		cm.Done()

		close(msgChan)
	})
}
