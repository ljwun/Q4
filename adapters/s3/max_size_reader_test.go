package s3_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"q4/adapters/s3"
)

func TestMaxSizeReader(t *testing.T) {
	tests := []struct {
		name       string
		input      []byte
		maxSize    int64
		wantN      int
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:    "讀取小於限制的內容",
			input:   []byte("hello"),
			maxSize: 10,
			wantN:   5,
			wantErr: false,
		},
		{
			name:       "讀取超過限制的內容",
			input:      []byte("hello world"),
			maxSize:    5,
			wantN:      5,
			wantErr:    true,
			wantErrMsg: "reach limit of 5 bytes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := s3.NewMaxSizeReader(bytes.NewReader(tt.input), tt.maxSize)
			buf := make([]byte, len(tt.input))
			n, err := reader.Read(buf)

			assert.Equal(t, tt.wantN, n)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErrMsg, err.Error())
			} else {
				assert.True(t, err == nil || err == io.EOF)
			}
		})
	}
}
