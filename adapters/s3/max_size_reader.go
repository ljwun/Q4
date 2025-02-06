package s3

import (
	"fmt"
	"io"
)

var ErrReachLimitType *ReachLimitError

type ReachLimitError struct {
	MaxBytes int64
}

func (e *ReachLimitError) Error() string {
	return fmt.Sprintf("reach limit of %s", FormatBytes(e.MaxBytes))
}

// NewMaxSizeReader 創建一個新的 MaxSizeReader 實例，
// 用於限制讀取的最大長度；如果讀取的長度超過限制，將返
// 回 ReachLimitError。
func NewMaxSizeReader(r io.Reader, maxSize int64) io.Reader {
	return &maxSizeReader{r, maxSize, maxSize}
}

type maxSizeReader struct {
	reader io.Reader
	i      int64 // 限制的總長度
	n      int64 // 還可以讀取的長度
}

func (r *maxSizeReader) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	// 如果請求的長度大於剩餘需要讀取的長度
	// 不需要讀取請求長度的內容，只需要讀取
	// 剩餘數量加1的內容，就能判斷是否超過
	// 最大長度
	if int64(len(p)) > r.n+1 {
		p = p[:r.n+1]
	}
	n, err = r.reader.Read(p)

	// 如果讀取的長度小於等於剩餘需要讀取的
	// 長度代表沒有超過限制的長度，直接返回
	// 原始資料
	if int64(n) <= r.n {
		r.n -= int64(n)
		return n, err
	}

	// 如果讀取的長度大於剩餘需要讀取的長度
	// 代表超過限制的長度，返回超過限制錯誤
	n = int(r.n)
	r.n = 0
	return n, &ReachLimitError{r.i}
}
