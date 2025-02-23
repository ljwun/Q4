//go:generate mockgen -package=redis -destination=mock.go -source=interfaces.go

package redis

import (
	"context"
)

// IProducer 定義了 Producer 的操作介面
type IProducer[T any] interface {
	Start()
	Publish(data T) error
	Close()
}

// IGroupConsumer 定義了 GroupConsumer 的操作介面
type IGroupConsumer[T any] interface {
	Start() error
	Subscribe() <-chan *Message[T]
	Close() error
}

// IConsumer 定義了 Consumer 的操作介面
type IConsumer[T any] interface {
	Start()
	Subscribe() <-chan T
	Close()
}

// IAutoRenewMutex 定義了 AutoRenewMutex 的操作介面
type IAutoRenewMutex interface {
	Lock(ctx context.Context) (context.Context, error)
	Unlock() (bool, error)
	Valid() bool
}
