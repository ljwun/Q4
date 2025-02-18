package redis

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"log/slog"

	"github.com/redis/go-redis/v9"
)

var (
	ErrConsumerClosed = errors.New("consumer is closed")
)

// Message 封裝消息和ack所需資料
type Message[T any] struct {
	Data T

	client    *redis.Client
	done      bool
	messageID string
	stream    string
	group     string
}

// Done 確認消息已處理完成
func (m *Message[T]) Done(ctx context.Context) error {
	if m.done {
		return nil
	}
	err := m.client.XAck(ctx, m.stream, m.group, m.messageID).Err()
	if err != nil {
		return fmt.Errorf("ack error: %w", err)
	}
	m.done = true
	return nil
}

type GroupConsumer[T any] struct {
	client        *redis.Client
	stream        string
	group         string
	consumer      string
	downStream    chan *Message[T]
	cancelFunc    context.CancelFunc
	wg            sync.WaitGroup
	closed        bool
	logger        *slog.Logger
	mutex         IAutoRenewMutex
	pendingMsgIds []string
	options       groupConsumerOptions[T]
}

type groupConsumerOptions[T any] struct {
	logger         *slog.Logger
	parseFunc      func(map[string]any) (T, error)
	bufferSize     int
	blockTimeout   time.Duration
	mutex          IAutoRenewMutex
	strictOrdering bool // 嚴格順序模式
}

type GroupConsumerOption[T any] func(*groupConsumerOptions[T])

// WithGroupConsumerLogger 設置日誌記錄器
func WithGroupConsumerLogger[T any](logger *slog.Logger) GroupConsumerOption[T] {
	return func(o *groupConsumerOptions[T]) {
		o.logger = logger
	}
}

// WithGroupConsumerParseFunc 設置消息解析函數
func WithGroupConsumerParseFunc[T any](fn func(map[string]any) (T, error)) GroupConsumerOption[T] {
	return func(o *groupConsumerOptions[T]) {
		o.parseFunc = fn
	}
}

// WithGroupConsumerBufferSize 設置下游channel的緩衝大小
func WithGroupConsumerBufferSize[T any](size int) GroupConsumerOption[T] {
	return func(o *groupConsumerOptions[T]) {
		o.bufferSize = size
	}
}

// WithGroupConsumerBlockTimeout 設置阻塞讀取超時時間
func WithGroupConsumerBlockTimeout[T any](d time.Duration) GroupConsumerOption[T] {
	return func(o *groupConsumerOptions[T]) {
		o.blockTimeout = d
	}
}

// WithGroupConsumerMutex 注入mutex (主要用於測試)
func WithGroupConsumerMutex[T any](mutex IAutoRenewMutex) GroupConsumerOption[T] {
	return func(o *groupConsumerOptions[T]) {
		o.mutex = mutex
	}
}

// WithGroupConsumerStrictOrdering 設置是否使用嚴格順序模式
func WithGroupConsumerStrictOrdering[T any](strict bool) GroupConsumerOption[T] {
	return func(o *groupConsumerOptions[T]) {
		o.strictOrdering = strict
	}
}

func NewGroupConsumer[T any](
	client *redis.Client,
	stream, group, consumer string,
	opts ...GroupConsumerOption[T],
) (IGroupConsumer[T], error) {
	if client == nil {
		return nil, errors.New("redis client cannot be nil")
	}
	if stream == "" || group == "" || consumer == "" {
		return nil, errors.New("stream, group and consumer cannot be empty")
	}

	// 默認選項
	options := groupConsumerOptions[T]{
		logger:         slog.Default(),
		parseFunc:      DefaultParseFromMessage[T],
		bufferSize:     1,
		blockTimeout:   time.Second,
		strictOrdering: false,
	}

	// 應用自定義選項
	for _, opt := range opts {
		opt(&options)
	}

	gc := &GroupConsumer[T]{
		wg:         sync.WaitGroup{},
		downStream: make(chan *Message[T], options.bufferSize),
		logger:     options.logger.With(slog.String("caller", "GroupConsumer"), slog.String("stream", stream), slog.String("group", group), slog.String("consumer", consumer)),
		client:     client,
		stream:     stream,
		group:      group,
		consumer:   consumer,
		closed:     true,
		options:    options,
	}

	// 只在嚴格順序模式下設置mutex
	if options.strictOrdering {
		if options.mutex != nil {
			gc.mutex = options.mutex
		} else {
			gc.mutex = NewAutoRenewMutex(client, fmt.Sprintf("lock:%s:%s", stream, group), WithAutoRenewMutexSkipLockError(true))
		}
	}

	return gc, nil
}

func (s *GroupConsumer[T]) Start() error {
	if !s.closed {
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFunc = cancel
	s.closed = false
	s.logger.Info("starting group consumer")

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer s.logger.Info("consumer goroutine stopped")
		defer close(s.downStream)

		// 如果是嚴格順序模式，拿到鎖後先獲取所有pending消息ID
		if s.options.strictOrdering {
			if err := s.mutex.Lock(ctx); err != nil {
				s.logger.Error("failed to acquire lock", slog.Any("error", err))
				s.cancelFunc()
				return
			}
			defer s.mutex.Unlock()

			if err := s.fetchPendingMessageIds(ctx); err != nil {
				s.logger.Error("initial pending messages fetch failed", slog.Any("error", err))
				s.cancelFunc()
				return
			}
		}

		for {
			select {
			case <-ctx.Done():
				return
			default:
				message, err := s.fetchNextMessage(ctx)
				if err != nil {
					if errors.Is(err, redis.Nil) {
						continue // 沒有新消息，正常情況
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

					if err := s.moveToDeadLetter(ctx, message); err != nil {
						if errors.Is(err, context.Canceled) {
							return
						}
						s.logger.Error("error moving message to dead letter",
							slog.String("messageId", message.ID),
							slog.Any("error", err))
					}
					continue
				}

				// 準備發送到下游的消息
				msg := &Message[T]{
					Data:      data,
					messageID: message.ID,
					stream:    s.stream,
					group:     s.group,
					client:    s.client,
				}

				if err := s.moveToDownStream(ctx, msg); err != nil {
					if errors.Is(err, context.Canceled) {
						return
					}
					s.logger.Error("error moving message to downstream",
						slog.String("messageId", message.ID),
						slog.Any("error", err))
				}
			}
		}
	}()

	return nil
}

// Subscribe 訂閱Stream，返回Message通道
func (s *GroupConsumer[T]) Subscribe() <-chan *Message[T] {
	return s.downStream
}

func (s *GroupConsumer[T]) Close() error {
	if s.closed {
		return nil
	}
	s.logger.Info("closing group consumer")
	s.closed = true
	s.cancelFunc()

	s.wg.Wait()
	s.logger.Info("group consumer closed gracefully")
	return nil
}

func (s *GroupConsumer[T]) fetchPendingMessageIds(ctx context.Context) error {
	s.pendingMsgIds = make([]string, 0, 1000)
	lastId := "-"

	for {
		pending, err := s.client.XPendingExt(ctx, &redis.XPendingExtArgs{
			Stream: s.stream,
			Group:  s.group,
			Start:  lastId,
			End:    "+",
			Count:  100, // 每次獲取100條
		}).Result()

		if err != nil {
			if errors.Is(err, redis.Nil) {
				break
			}
			return fmt.Errorf("error getting pending messages: %w", err)
		}

		if len(pending) == 0 {
			break
		}

		// 保存ID
		for _, p := range pending {
			s.pendingMsgIds = append(s.pendingMsgIds, p.ID)
		}

		// 更新lastId為最後一條消息的ID
		lastId = pending[len(pending)-1].ID

		// 如果返回的消息數量少於請求的數量，說明已經沒有更多消息了
		if len(pending) < 100 {
			break
		}
	}

	s.logger.Info("fetched all pending message IDs",
		slog.Int("count", len(s.pendingMsgIds)))
	return nil
}

func (s *GroupConsumer[T]) fetchNextMessage(ctx context.Context) (redis.XMessage, error) {
	var message redis.XMessage
	var err error

	if len(s.pendingMsgIds) > 0 {
		// 讀取pending消息
		var messages []redis.XMessage
		messages, err = s.client.XRangeN(ctx, s.stream, s.pendingMsgIds[0], s.pendingMsgIds[0], 1).Result()
		s.pendingMsgIds = s.pendingMsgIds[1:]
		if len(messages) > 0 {
			message = messages[0]
		}
	} else {
		// 讀取新消息
		var streams []redis.XStream
		streams, err = s.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    s.group,
			Consumer: s.consumer,
			Streams:  []string{s.stream, ">"},
			Count:    1,
			Block:    s.options.blockTimeout,
		}).Result()
		if len(streams) > 0 && len(streams[0].Messages) > 0 {
			message = streams[0].Messages[0]
		}
	}

	return message, err
}

// 添加死信處理
func (s *GroupConsumer[T]) moveToDeadLetter(ctx context.Context, message redis.XMessage) error {
	deadLetterStream := s.stream + ":dead-letter"

	err := s.client.XAdd(ctx, &redis.XAddArgs{
		Stream: deadLetterStream,
		Values: message.Values,
	}).Err()

	if err != nil {
		return fmt.Errorf("failed to move message to dead letter queue: %w", err)
	}

	// 確認原消息
	return s.client.XAck(ctx, s.stream, s.group, message.ID).Err()
}

// moveToDownStream 處理發送消息到下游channel
func (s *GroupConsumer[T]) moveToDownStream(ctx context.Context, message *Message[T]) error {
	select {
	case <-ctx.Done():
		return context.Canceled
	case s.downStream <- message:
		return nil
	}
}
