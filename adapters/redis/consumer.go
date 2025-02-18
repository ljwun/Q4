package redis

import (
	"context"
	"errors"
	"sync"
	"time"

	"log/slog"

	"github.com/redis/go-redis/v9"
)

type consumerOptions[T any] struct {
	logger       *slog.Logger
	bufferSize   int
	blockTimeout time.Duration
	parseFunc    func(map[string]any) (T, error)
}

type ConsumerOption[T any] func(*consumerOptions[T])

// WithConsumerLogger 設置日誌記錄器
func WithConsumerLogger[T any](logger *slog.Logger) ConsumerOption[T] {
	return func(o *consumerOptions[T]) {
		o.logger = logger
	}
}

// WithConsumerBufferSize 設置下游channel的緩衝大小
func WithConsumerBufferSize[T any](size int) ConsumerOption[T] {
	return func(o *consumerOptions[T]) {
		o.bufferSize = size
	}
}

// WithConsumerBlockTimeout 設置阻塞讀取超時時間
func WithConsumerBlockTimeout[T any](d time.Duration) ConsumerOption[T] {
	return func(o *consumerOptions[T]) {
		o.blockTimeout = d
	}
}

// WithConsumerParseFunc 設置自定義解析函數
func WithConsumerParseFunc[T any](fn func(map[string]any) (T, error)) ConsumerOption[T] {
	return func(o *consumerOptions[T]) {
		o.parseFunc = fn
	}
}

type Consumer[T any] struct {
	client     *redis.Client
	stream     string
	lastID     string
	downStream chan T
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup
	closed     bool
	logger     *slog.Logger
	options    consumerOptions[T]
}

func NewConsumer[T any](client *redis.Client, stream string, opts ...ConsumerOption[T]) (IConsumer[T], error) {
	if client == nil {
		return nil, errors.New("redis client cannot be nil")
	}
	if stream == "" {
		return nil, errors.New("stream cannot be empty")
	}

	// 默認選項
	options := consumerOptions[T]{
		logger:       slog.Default(),
		bufferSize:   100,
		blockTimeout: time.Second,
		parseFunc:    DefaultParseFromMessage[T],
	}

	// 應用自定義選項
	for _, opt := range opts {
		opt(&options)
	}

	consumer := &Consumer[T]{
		client:  client,
		stream:  stream,
		lastID:  "$",
		closed:  true,
		logger:  options.logger.With(slog.String("caller", "Consumer"), slog.String("stream", stream)),
		options: options,
	}

	return consumer, nil
}

func (s *Consumer[T]) Start() {
	if !s.closed {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.downStream = make(chan T, s.options.bufferSize)
	s.closed = false
	s.cancelFunc = cancel
	s.logger.Info("starting stream consumer")

	// 啟動消費者 goroutine
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer s.logger.Info("consumer goroutine stopped")
		defer close(s.downStream)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				message, err := s.fetchNextMessage(ctx)
				if err != nil {
					if errors.Is(err, redis.Nil) {
						continue
					}
					s.logger.Error("fetch message error", slog.Any("error", err))
					continue
				}

				// 解析消息
				data, err := s.options.parseFunc(message.Values)
				if err != nil {
					s.logger.Error("failed to parse message",
						slog.String("messageId", message.ID),
						slog.Any("error", err))
					continue
				}

				// 發送到下游
				select {
				case <-ctx.Done():
					return
				case s.downStream <- data:
					s.logger.Debug("message sent to downstream",
						slog.String("messageId", message.ID))
				}
			}
		}
	}()
}

func (s *Consumer[T]) fetchNextMessage(ctx context.Context) (redis.XMessage, error) {
	streams, err := s.client.XRead(ctx, &redis.XReadArgs{
		Streams: []string{s.stream, s.lastID},
		Count:   1,
		Block:   s.options.blockTimeout,
	}).Result()

	if err != nil {
		return redis.XMessage{}, err
	}

	if len(streams) > 0 && len(streams[0].Messages) > 0 {
		message := streams[0].Messages[0]
		s.lastID = message.ID
		s.logger.Debug("received message", slog.String("messageId", message.ID))
		return message, nil
	}

	return redis.XMessage{}, redis.Nil
}

// Subscribe 訂閱數據流
func (s *Consumer[T]) Subscribe() <-chan T {
	return s.downStream
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
	s.logger.Info("stream consumer closed")
}
