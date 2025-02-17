package redis

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/redis/go-redis/v9"
	"github.com/smallnest/chanx"
)

type producerOptions[T any] struct {
	logger     *slog.Logger
	bufferSize int
	parseFunc  func(T) (map[string]any, error)
}

type ProducerOption[T any] func(*producerOptions[T])

// WithProducerLogger 設置日誌記錄器
func WithProducerLogger[T any](logger *slog.Logger) ProducerOption[T] {
	return func(o *producerOptions[T]) {
		o.logger = logger
	}
}

// WithProducerBufferSize 設置緩衝大小
func WithProducerBufferSize[T any](size int) ProducerOption[T] {
	return func(o *producerOptions[T]) {
		o.bufferSize = size
	}
}

// WithProducerParseFunc 設置消息序列化函數
func WithProducerParseFunc[T any](fn func(T) (map[string]any, error)) ProducerOption[T] {
	return func(o *producerOptions[T]) {
		o.parseFunc = fn
	}
}

type Producer[T any] struct {
	client     *redis.Client
	stream     string
	upstream   *chanx.UnboundedChan[map[string]any]
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup
	closed     bool
	logger     *slog.Logger
	options    producerOptions[T]
}

func NewProducer[T any](client *redis.Client, stream string, opts ...ProducerOption[T]) (*Producer[T], error) {
	if client == nil {
		return nil, errors.New("redis client cannot be nil")
	}
	if stream == "" {
		return nil, errors.New("stream cannot be empty")
	}

	// 默認選項
	options := producerOptions[T]{
		logger:     slog.Default(),
		bufferSize: 100,
		parseFunc:  DefaultParseToMessage[T],
	}

	// 應用自定義選項
	for _, opt := range opts {
		opt(&options)
	}

	producer := &Producer[T]{
		client:  client,
		stream:  stream,
		closed:  true,
		logger:  options.logger.With(slog.String("caller", "Producer"), slog.String("stream", stream)),
		options: options,
	}

	return producer, nil
}

func (p *Producer[T]) Start() {
	if !p.closed {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	p.upstream = chanx.NewUnboundedChan[map[string]any](ctx, p.options.bufferSize)
	p.cancelFunc = cancel
	p.closed = false
	p.logger.Info("starting stream producer")

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		defer p.logger.Info("producer goroutine stopped")

		for {
			select {
			case <-ctx.Done():
				return
			case message := <-p.upstream.Out:
				id, err := p.client.XAdd(ctx, &redis.XAddArgs{
					Stream: p.stream,
					Values: message,
				}).Result()

				if err != nil {
					if errors.Is(err, context.Canceled) {
						return
					}
					p.logger.Error("publish message error", slog.Any("error", err))
					continue
				}

				p.logger.Debug("message published", slog.String("messageId", id))
			}
		}
	}()
}

func (p *Producer[T]) Publish(data T) error {
	if p.closed {
		return ErrConsumerClosed
	}

	message, err := p.options.parseFunc(data)
	if err != nil {
		return fmt.Errorf("parse message error: %w", err)
	}

	p.upstream.In <- message
	return nil
}

func (p *Producer[T]) Close() {
	if p.closed {
		return
	}
	p.logger.Info("closing stream producer")
	p.closed = true
	p.cancelFunc()
	p.wg.Wait()
	p.logger.Info("stream producer closed")
}
