package session

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewSession(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		id      string
		store   IStore
		wantNil bool
	}{
		{
			name:    "valid parameters",
			ctx:     context.Background(),
			id:      "test-id",
			store:   &MockIStore{},
			wantNil: false,
		},
		{
			name:    "nil context",
			ctx:     nil,
			id:      "test-id",
			store:   &MockIStore{},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := NewSession(tt.ctx, tt.id, tt.store)
			if tt.wantNil {
				assert.Nil(t, session)
			} else {
				assert.NotNil(t, session)
			}
		})
	}
}

func TestSession_Load(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name      string
		mockSetup func(*MockIStore)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "successful load",
			mockSetup: func(mockStore *MockIStore) {
				mockStore.EXPECT().
					Load(gomock.Any(), "test-id").
					Return(map[string]string{"key": "value"}, nil)
			},
			wantErr: false,
		},
		{
			name: "load error",
			mockSetup: func(mockStore *MockIStore) {
				mockStore.EXPECT().
					Load(gomock.Any(), "test-id").
					Return(nil, errors.New("load error"))
			},
			wantErr: true,
			errMsg:  "load error",
		},
		{
			name: "already loaded",
			mockSetup: func(mockStore *MockIStore) {
				// 不應該呼叫 Load
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := NewMockIStore(ctrl)
			tt.mockSetup(mockStore)

			s := &sessionImpl{
				id:    "test-id",
				ctx:   context.Background(),
				store: mockStore,
			}

			if tt.name == "already loaded" {
				s.data = map[string]string{"existing": "data"}
			}

			err := s.Load()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSession_Save(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name      string
		data      map[string]string
		mockSetup func(*MockIStore)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "successful save",
			data: map[string]string{"key": "value"},
			mockSetup: func(mockStore *MockIStore) {
				mockStore.EXPECT().
					Save(gomock.Any(), "test-id", map[string]string{"key": "value"}).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "save error",
			data: map[string]string{"key": "value"},
			mockSetup: func(mockStore *MockIStore) {
				mockStore.EXPECT().
					Save(gomock.Any(), "test-id", gomock.Any()).
					Return(errors.New("save error"))
			},
			wantErr: true,
			errMsg:  "save error",
		},
		{
			name:      "nil data",
			data:      nil,
			mockSetup: func(mockStore *MockIStore) {},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := NewMockIStore(ctrl)
			tt.mockSetup(mockStore)

			s := &sessionImpl{
				id:    "test-id",
				ctx:   context.Background(),
				store: mockStore,
				data:  tt.data,
			}

			err := s.Save()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSession_Get(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]string
		key      string
		expected string
	}{
		{
			name:     "get existing key",
			data:     map[string]string{"key1": "value1"},
			key:      "key1",
			expected: "value1",
		},
		{
			name:     "get non-existent key",
			data:     map[string]string{"key1": "value1"},
			key:      "key2",
			expected: "",
		},
		{
			name:     "nil data",
			data:     nil,
			key:      "key1",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sessionImpl{
				data: tt.data,
			}
			assert.Equal(t, tt.expected, s.Get(tt.key))
		})
	}
}

func TestSession_Set(t *testing.T) {
	tests := []struct {
		name         string
		initialData  map[string]string
		key          string
		value        string
		expectedData map[string]string
	}{
		{
			name:         "set to existing map",
			initialData:  map[string]string{"key1": "value1"},
			key:          "key2",
			value:        "value2",
			expectedData: map[string]string{"key1": "value1", "key2": "value2"},
		},
		{
			name:         "set to nil map",
			initialData:  nil,
			key:          "key1",
			value:        "value1",
			expectedData: map[string]string{"key1": "value1"},
		},
		{
			name:         "override existing key",
			initialData:  map[string]string{"key1": "value1"},
			key:          "key1",
			value:        "new value",
			expectedData: map[string]string{"key1": "new value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sessionImpl{
				data: tt.initialData,
			}
			s.Set(tt.key, tt.value)
			assert.Equal(t, tt.expectedData, s.data)
		})
	}
}

func TestSession_Delete(t *testing.T) {
	tests := []struct {
		name         string
		initialData  map[string]string
		key          string
		expectedData map[string]string
	}{
		{
			name:         "delete existing key",
			initialData:  map[string]string{"key1": "value1", "key2": "value2"},
			key:          "key1",
			expectedData: map[string]string{"key2": "value2"},
		},
		{
			name:         "delete non-existent key",
			initialData:  map[string]string{"key1": "value1"},
			key:          "key2",
			expectedData: map[string]string{"key1": "value1"},
		},
		{
			name:         "delete from nil map",
			initialData:  nil,
			key:          "key1",
			expectedData: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sessionImpl{
				data: tt.initialData,
			}
			s.Delete(tt.key)
			assert.Equal(t, tt.expectedData, s.data)
		})
	}
}

func TestSession_Clear(t *testing.T) {
	tests := []struct {
		name        string
		initialData map[string]string
	}{
		{
			name:        "clear non-empty map",
			initialData: map[string]string{"key1": "value1", "key2": "value2"},
		},
		{
			name:        "clear empty map",
			initialData: map[string]string{},
		},
		{
			name:        "clear nil map",
			initialData: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sessionImpl{
				data: tt.initialData,
			}
			s.Clear()
			assert.NotNil(t, s.data)
			assert.Empty(t, s.data)
		})
	}
}
