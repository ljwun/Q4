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

	raw map[string]any
}

// Done 確認消息已處理完成
func (m *Message[T]) Done(ctx context.Context) error {
	const op = "Message.Done"
	if m.done {
		return nil
	}
	err := m.client.XAck(ctx, m.stream, m.group, m.messageID).Err()
	if err != nil {
		return fmt.Errorf("[%s] failed to ack message: %w", op, err)
	}
	m.done = true
	return nil
}

// Fail 確認消息處理失敗
func (m *Message[T]) Fail(ctx context.Context, failErr error) error {
	const op = "Message.Fail"
	if m.done {
		return nil
	}

	m.raw["error"] = failErr.Error()
	err := m.client.XAdd(ctx, &redis.XAddArgs{
		Stream: m.stream + ":dead-letter",
		Values: m.raw,
	}).Err()
	if err != nil {
		return fmt.Errorf("[%s] failed to move message to dead letter queue: %w", op, err)
	}

	err = m.client.XAck(ctx, m.stream, m.group, m.messageID).Err()
	if err != nil {
		return fmt.Errorf("[%s] failed to ack failed message: %w", op, err)
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
		logger:   options.logger.With(slog.String("caller", "GroupConsumer"), slog.String("stream", stream), slog.String("group", group), slog.String("consumer", consumer)),
		client:   client,
		stream:   stream,
		group:    group,
		consumer: consumer,
		closed:   true,
		options:  options,
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
	s.downStream = make(chan *Message[T], s.options.bufferSize)
	s.cancelFunc = cancel
	s.closed = false
	s.logger.Info("starting group consumer")

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer s.logger.Info("group consumer goroutine stopped")
		defer close(s.downStream)
		defer func() {
			if s.options.strictOrdering {
				s.mutex.Unlock()
			}
		}()

		for {
			workloadContext := ctx

			// 如果是嚴格順序模式下，會先拿鎖，然後再處理消息
			if s.options.strictOrdering {
				var err error
				// workloadContext在嚴格順序模式下會被修改成帶鎖狀態的child context，可以接收到鎖的釋放信號
				workloadContext, err = s.mutex.Lock(ctx)
				if err != nil {
					s.logger.Error("failed to acquire lock", slog.Any("error", err))
					if errors.Is(err, context.Canceled) {
						break
					}
					continue
				}
			}
			if err := s.messagesWorkflow(workloadContext); err != nil {
				// 如果是context.Canceled，且是因為外部context取消，則退出循環
				if errors.Is(err, context.Canceled) && ctx.Err() != nil {
					break
				}
				if s.options.strictOrdering && errors.Is(err, context.Canceled) && ctx.Err() == nil {
					// 如果是context.Canceled，且是因為鎖的context取消，則繼續循環
					s.logger.Error("lock context cancelled, stopping current processing, restarting group consumer")
				} else {
					// 其他錯誤情況，重啟group consumer
					s.logger.Error("error processing messages, stopping current processing, restarting group consumer", slog.Any("error", err))
				}
				continue
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

// messagesWorkflow 處理消息的工作流程
func (s *GroupConsumer[T]) messagesWorkflow(ctx context.Context) error {
	if s.options.strictOrdering {
		if err := s.fetchPendingMessageIds(ctx); err != nil {
			s.logger.Error("initial pending messages fetch failed", slog.Any("error", err))
			return err
		}
	}
	for {
		message, err := s.fetchNextMessage(ctx)
		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			}
			s.logger.Error("fetch message error", slog.Any("error", err))
			if errors.Is(err, context.Canceled) {
				return err
			}
			// 其他的錯誤一般是server跟redis之間的通訊異常，重試即可
			continue
		}
		data, err := s.options.parseFunc(message.Values)
		if err != nil {
			// 解析失敗的問題一個是原始資料，一個是解析方案，不管哪種都是需要額外處理的
			// 不會因為重試就成功，因此先將消息移動到dead-letter，系統繼續處理下一條消息
			s.logger.Error("failed to parse message",
				slog.String("messageId", message.ID),
				slog.Any("error", err),
			)
			if deadLetterErr := s.moveToDeadLetter(ctx, message); deadLetterErr != nil {
				s.logger.Error("error moving message to dead letter",
					slog.String("messageId", message.ID),
					slog.Any("error", deadLetterErr),
				)
				// 如果在移動到dead-letter的過程中發生了異常，這個訊息會以pending的形式留在stream中
				// WARN: 目前在嚴格順序模式下，這種狀況會在下一輪開始時優先處理這種訊息
				// 		 但是在非嚴格順序模式下，會跳過pending訊息，一旦遇到這種狀況，訊息會一直以pending的形式留在stream中，需要手動對stream處理
				return deadLetterErr
			}
			continue
		}
		msg := &Message[T]{
			Data:      data,
			messageID: message.ID,
			stream:    s.stream,
			group:     s.group,
			client:    s.client,
			raw:       message.Values,
		}
		if err := s.moveToDownStream(ctx, msg); err != nil {
			s.logger.Error("error moving message to downstream",
				slog.String("messageId", message.ID),
				slog.Any("error", err),
			)
			// 如果在移動到downstream的過程中發生了異常(只有可能是context.Canceled)，這個訊息會以pending的形式留在stream中
			// WARN: 目前在嚴格順序模式下，這種狀況會在下一輪開始時優先處理這種訊息
			// 		 但是在非嚴格順序模式下，會跳過pending訊息，一旦遇到這種狀況，訊息會一直以pending的形式留在stream中，需要手動對stream處理
			return err
		}
	}
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
	if ctx.Err() != nil {
		return context.Canceled
	}
	select {
	case <-ctx.Done():
		return context.Canceled
	case s.downStream <- message:
		return nil
	}
}
