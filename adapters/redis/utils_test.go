package redis

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// 測試結構
type TestStruct struct {
	Name      string    `json:"name"`
	Age       int       `json:"age"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// 無標籤結構
type NoTagStruct struct {
	Name     string
	Age      int
	IsActive bool
}

// 空結構
type EmptyStruct struct{}

// 複雜結構
type ComplexStruct struct {
	ID        int64          `json:"id"`
	Nested    TestStruct     `json:"nested"`
	Tags      []string       `json:"tags"`
	Map       map[string]any `json:"map"`
	Interface any            `json:"interface"`
}

// 指標結構
type PointerStruct struct {
	Data *TestStruct `json:"data"`
}

// compareTime 比較兩個時間是否相等，忽略單調時鐘和位置信息
func compareTime(t1, t2 time.Time) bool {
	return t1.UTC().Equal(t2.UTC())
}

// compareTestStruct 比較兩個TestStruct，特別處理時間比較
func compareTestStruct(t *testing.T, expected, actual TestStruct) {
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.Age, actual.Age)
	assert.Equal(t, expected.IsActive, actual.IsActive)
	assert.True(t, compareTime(expected.CreatedAt, actual.CreatedAt),
		"time mismatch: expected %v, got %v", expected.CreatedAt, actual.CreatedAt)
}

// compareComplexStruct 比較兩個ComplexStruct，特別處理嵌套結構和map
func compareComplexStruct(t *testing.T, expected, actual ComplexStruct) {
	assert.Equal(t, expected.ID, actual.ID)
	compareTestStruct(t, expected.Nested, actual.Nested)
	assert.Equal(t, expected.Tags, actual.Tags)
	assert.Equal(t, expected.Interface, actual.Interface)

	// 比較 map
	assert.Equal(t, len(expected.Map), len(actual.Map))
	for k, v := range expected.Map {
		actualVal, ok := actual.Map[k]
		assert.True(t, ok, "key %s not found in actual map", k)
		assert.EqualValues(t, v, actualVal, "value mismatch for key %s", k)
	}
}

func TestDefaultParseToMessage(t *testing.T) {
	t.Run("normal struct", func(t *testing.T) {
		input := TestStruct{
			Name:      "test",
			Age:       25,
			IsActive:  true,
			CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		result, err := DefaultParseToMessage(input)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Contains(t, result, "data")
		assert.NotEmpty(t, result["data"])
	})

	t.Run("empty struct", func(t *testing.T) {
		input := EmptyStruct{}

		result, err := DefaultParseToMessage(input)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Contains(t, result, "data")
		assert.NotEmpty(t, result["data"])
	})

	t.Run("struct with no tags", func(t *testing.T) {
		input := NoTagStruct{
			Name:     "test",
			Age:      25,
			IsActive: true,
		}

		result, err := DefaultParseToMessage(input)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Contains(t, result, "data")
		assert.NotEmpty(t, result["data"])
	})

	t.Run("complex struct", func(t *testing.T) {
		input := ComplexStruct{
			ID: 1,
			Nested: TestStruct{
				Name:      "nested",
				Age:       30,
				IsActive:  true,
				CreatedAt: time.Now(),
			},
			Tags: []string{"tag1", "tag2"},
			Map: map[string]any{
				"key1": "value1",
				"key2": 123,
			},
			Interface: "test",
		}

		result, err := DefaultParseToMessage(input)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Contains(t, result, "data")
		assert.NotEmpty(t, result["data"])
	})

	t.Run("pointer type error", func(t *testing.T) {
		input := &TestStruct{
			Name: "test",
		}

		_, err := DefaultParseToMessage(input)
		assert.ErrorIs(t, err, ErrPointerType)
	})

	t.Run("nil pointer struct", func(t *testing.T) {
		var input *TestStruct

		_, err := DefaultParseToMessage(input)
		assert.ErrorIs(t, err, ErrPointerType)
	})

	t.Run("zero values", func(t *testing.T) {
		input := TestStruct{} // 全部為零值

		message, err := DefaultParseToMessage(input)
		assert.NoError(t, err)

		result, err := DefaultParseFromMessage[TestStruct](message)
		assert.NoError(t, err)
		compareTestStruct(t, input, result)
	})
}

func TestDefaultParseFromMessage(t *testing.T) {
	t.Run("normal struct", func(t *testing.T) {
		input := TestStruct{
			Name:      "test",
			Age:       25,
			IsActive:  true,
			CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		// 先轉換為message
		message, err := DefaultParseToMessage(input)
		assert.NoError(t, err)

		// 再轉換回struct
		result, err := DefaultParseFromMessage[TestStruct](message)
		assert.NoError(t, err)
		compareTestStruct(t, input, result)
	})

	t.Run("empty struct", func(t *testing.T) {
		input := EmptyStruct{}

		message, err := DefaultParseToMessage(input)
		assert.NoError(t, err)

		result, err := DefaultParseFromMessage[EmptyStruct](message)
		assert.NoError(t, err)
		assert.Equal(t, input, result)
	})

	t.Run("complex struct", func(t *testing.T) {
		input := ComplexStruct{
			ID: 1,
			Nested: TestStruct{
				Name:      "nested",
				Age:       30,
				IsActive:  true,
				CreatedAt: time.Now().UTC(), // 使用UTC時間避免時區問題
			},
			Tags: []string{"tag1", "tag2"},
			Map: map[string]any{
				"key1": "value1",
				"key2": 123,
			},
			Interface: "test",
		}

		message, err := DefaultParseToMessage(input)
		assert.NoError(t, err)

		result, err := DefaultParseFromMessage[ComplexStruct](message)
		assert.NoError(t, err)
		compareComplexStruct(t, input, result)
	})

	t.Run("empty map", func(t *testing.T) {
		input := map[string]any{}

		result, err := DefaultParseFromMessage[TestStruct](input)
		assert.NoError(t, err)
		assert.Empty(t, result.Name)
		assert.Zero(t, result.Age)
		assert.False(t, result.IsActive)
	})

	t.Run("nil map", func(t *testing.T) {
		var input map[string]any

		result, err := DefaultParseFromMessage[TestStruct](input)
		assert.NoError(t, err)
		assert.Empty(t, result.Name)
		assert.Zero(t, result.Age)
		assert.False(t, result.IsActive)
	})

	t.Run("pointer type error", func(t *testing.T) {
		input := map[string]any{"data": "some base64 data"}

		_, err := DefaultParseFromMessage[*TestStruct](input)
		assert.ErrorIs(t, err, ErrPointerType)
	})

	t.Run("invalid base64", func(t *testing.T) {
		input := map[string]any{
			"data": "invalid base64",
		}

		_, err := DefaultParseFromMessage[TestStruct](input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "base64 decode error")
	})

	t.Run("missing data field", func(t *testing.T) {
		input := map[string]any{
			"wrong_field": "some data",
		}

		_, err := DefaultParseFromMessage[TestStruct](input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "data field not found")
	})

	t.Run("invalid data type", func(t *testing.T) {
		input := map[string]any{
			"data": 123, // 錯誤的類型
		}

		_, err := DefaultParseFromMessage[TestStruct](input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid type")
	})
}
