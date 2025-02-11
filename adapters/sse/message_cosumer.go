package sse

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"time"
	"unsafe"

	"log/slog"

	"github.com/redis/go-redis/v9"
	"github.com/smallnest/chanx"
	"github.com/vmihailenco/msgpack/v5"
)

// ErrConsumerClosed 表示消費者已關閉
var ErrConsumerClosed = errors.New("consumer is closed")

// string轉[]byte，零拷貝
func stringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// []byte轉string，零拷貝
func bytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

type Consumer[T any] struct {
	client     *redis.Client
	stream     string
	lastID     string
	downStream chan T
	upstream   *chanx.UnboundedChan[T]
	ctx        context.Context
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup
	closed     bool
	logger     *slog.Logger
}

func NewConsumer[T any](client *redis.Client, stream string, logger *slog.Logger) *Consumer[T] {
	if logger == nil {
		logger = slog.Default()
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &Consumer[T]{
		client:     client,
		stream:     stream,
		lastID:     "$", // 只讀取最新的消息
		downStream: make(chan T, 100),
		upstream:   chanx.NewUnboundedChan[T](ctx, 100),
		ctx:        ctx,
		cancelFunc: cancel,
		wg:         sync.WaitGroup{},
		closed:     false,
		logger:     logger.With(slog.String("caller", "Consumer"), slog.String("stream", stream)),
	}
}

func (s *Consumer[T]) Start() {
	s.logger.Info("starting stream consumer")
	// 啟動消費者 goroutine
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer s.logger.Info("consumer goroutine stopped")

		for {
			select {
			case <-s.ctx.Done():
				return
			default:
				streams, err := s.client.XRead(s.ctx, &redis.XReadArgs{
					Streams: []string{s.stream, s.lastID},
					Count:   1,
					Block:   time.Second * 1,
				}).Result()

				if err != nil {
					if err == context.Canceled {
						return
					}
					if err == redis.Nil {
						continue
					}
					s.logger.Error("error reading from stream",
						slog.Any("error", err))
					time.Sleep(time.Millisecond * 100)
					continue
				}

				for _, stream := range streams {
					for _, message := range stream.Messages {
						s.lastID = message.ID
						s.logger.Debug("received message",
							slog.String("messageId", message.ID))

						for key, value := range message.Values {
							strValue, ok := value.(string)
							if !ok {
								s.logger.Error("type assertion failed",
									slog.String("messageId", message.ID),
									slog.String("key", key),
									slog.String("actualType", fmt.Sprintf("%T", value)))
								continue
							}
							byteValue := stringToBytes(strValue)
							var data T
							if err := msgpack.Unmarshal(byteValue, &data); err != nil {
								s.logger.Error("unmarshal error",
									slog.String("messageId", message.ID),
									slog.String("messageData", base64.StdEncoding.EncodeToString(byteValue)),
									slog.Any("error", err))
								continue
							}
							select {
							case <-s.ctx.Done():
								return
							case s.downStream <- data:
							}
						}
					}
				}
			}
		}
	}()

	// 啟動發布者 goroutine
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer s.logger.Info("publisher goroutine stopped")

		for {
			select {
			case <-s.ctx.Done():
				return
			case data := <-s.upstream.Out:
				bytes, err := msgpack.Marshal(data)
				if err != nil {
					s.logger.Error("marshal error",
						slog.Any("error", err))
					continue
				}

				result, err := s.client.XAdd(s.ctx, &redis.XAddArgs{
					Stream: s.stream,
					Values: map[string]interface{}{
						"data": bytesToString(bytes),
					},
				}).Result()

				if err != nil {
					s.logger.Error("publish error",
						slog.Any("error", err))
				} else {
					s.logger.Debug("message published",
						slog.String("messageId", result))
				}
			}
		}
	}()
}

// Subscribe 訂閱數據流
func (s *Consumer[T]) Subscribe() <-chan T {
	return s.downStream
}

// Publish 發布數據到stream，如果消費者已關閉則返回錯誤
func (s *Consumer[T]) Publish(data T) error {
	if s.closed {
		return ErrConsumerClosed
	}

	select {
	case s.upstream.In <- data:
		return nil
	case <-s.ctx.Done():
		return ErrConsumerClosed
	}
}

// Close 關閉消費者
func (s *Consumer[T]) Close() {
	if s.closed {
		return
	}
	s.logger.Info("closing stream consumer")
	s.closed = true
	s.cancelFunc()
	s.wg.Wait()
	close(s.downStream)
	s.logger.Info("stream consumer closed")
}
