package redis

import (
	"context"
	"errors"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestStore_Load(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(mock redismock.ClientMock)
		session  string
		expected map[string]string
		wantErr  bool
	}{
		{
			name: "success",
			setup: func(mock redismock.ClientMock) {
				mock.ExpectHGetAll("test:session1").SetVal(map[string]string{
					"key1": "value1",
					"key2": "value2",
				})
			},
			session: "session1",
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "empty_session",
			setup: func(mock redismock.ClientMock) {
				mock.ExpectHGetAll("test:empty").SetVal(map[string]string{})
			},
			session:  "empty",
			expected: map[string]string{},
		},
		{
			name: "redis_error",
			setup: func(mock redismock.ClientMock) {
				mock.ExpectHGetAll("test:session1").
					SetErr(errors.New("redis connection error"))
			},
			session:  "session1",
			wantErr:  true,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 準備測試環境
			client, mock, cleanup := setupTest(t)
			defer cleanup()

			tt.setup(mock)

			store := NewStore(client, WithStorePrefix("test:"))

			// 執行測試
			got, err := store.Load(context.Background(), tt.session)

			// 驗證結果
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestStore_Save(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(mock redismock.ClientMock)
		session string
		data    map[string]string
		wantErr bool
	}{
		{
			name: "success",
			setup: func(mock redismock.ClientMock) {
				mock.ExpectEvalSha(
					saveScript.Hash(),
					[]string{"test:session1"},
					[]interface{}{"key1", "value1"},
				).SetVal(1)
			},
			session: "session1",
			data: map[string]string{
				"key1": "value1",
			},
		},
		{
			name: "empty_data",
			setup: func(mock redismock.ClientMock) {
				mock.ExpectEvalSha(
					saveScript.Hash(),
					[]string{"test:session1"},
					[]interface{}{},
				).SetVal(1)
			},
			session: "session1",
			data:    map[string]string{},
		},
		{
			name: "nil_data",
			setup: func(mock redismock.ClientMock) {
				mock.ExpectEvalSha(
					saveScript.Hash(),
					[]string{"test:session1"},
					[]interface{}{},
				).SetVal(1)
			},
			session: "session1",
			data:    nil,
		},
		{
			name: "redis_error",
			setup: func(mock redismock.ClientMock) {
				mock.ExpectEvalSha(
					saveScript.Hash(),
					[]string{"test:session1"},
					[]interface{}{"key1", "value1"},
				).SetErr(redis.ErrClosed)
			},
			session: "session1",
			data: map[string]string{
				"key1": "value1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 準備測試環境
			client, mock, cleanup := setupTest(t)
			defer cleanup()

			tt.setup(mock)

			store := NewStore(client, WithStorePrefix("test:"))

			// 執行測試
			err := store.Save(context.Background(), tt.session, tt.data)

			// 驗證結果
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
