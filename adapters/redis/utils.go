package redis

import (
	"encoding/base64"
	"errors"
	"fmt"
	"reflect"

	"github.com/vmihailenco/msgpack/v5"
)

var (
	ErrPointerType = errors.New("pointer type is not allowed")
)

// DefaultParseToMessage 將struct轉換為map[string]any
func DefaultParseToMessage[T any](data T) (map[string]any, error) {
	// 檢查是否為指標類型
	if reflect.TypeOf(data).Kind() == reflect.Ptr {
		return nil, ErrPointerType
	}

	// 使用 msgpack 序列化
	bytes, err := msgpack.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("msgpack marshal error: %w", err)
	}

	// base64編碼
	encoded := base64.StdEncoding.EncodeToString(bytes)

	// 封裝成map
	return map[string]any{
		"data": encoded,
	}, nil
}

// DefaultParseFromMessage 將map[string]any轉換為struct
func DefaultParseFromMessage[T any](message map[string]any) (T, error) {
	var result T

	// 檢查是否為指標類型
	if reflect.TypeOf(result).Kind() == reflect.Ptr {
		return result, ErrPointerType
	}

	if len(message) == 0 {
		return result, nil
	}

	// 獲取data字段
	dataStr, ok := message["data"].(string)
	if !ok {
		return result, fmt.Errorf("data field not found or invalid type")
	}

	// base64解碼
	bytes, err := base64.StdEncoding.DecodeString(dataStr)
	if err != nil {
		return result, fmt.Errorf("base64 decode error: %w", err)
	}

	// msgpack反序列化
	err = msgpack.Unmarshal(bytes, &result)
	if err != nil {
		return result, fmt.Errorf("msgpack unmarshal error: %w", err)
	}

	return result, nil
}
