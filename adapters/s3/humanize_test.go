package s3_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"q4/adapters/s3"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name  string
		bytes int64
		want  string
	}{
		{
			name:  "bytes",
			bytes: 500,
			want:  "500 bytes",
		},
		{
			name:  "KB",
			bytes: 1024 * 2,
			want:  "2.00 KB",
		},
		{
			name:  "MB",
			bytes: 1024 * 1024 * 3,
			want:  "3.00 MB",
		},
		{
			name:  "GB",
			bytes: 1024 * 1024 * 1024 * 4,
			want:  "4.00 GB",
		},
		{
			name:  "TB",
			bytes: 1024 * 1024 * 1024 * 1024 * 5,
			want:  "5.00 TB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s3.FormatBytes(tt.bytes)
			assert.Equal(t, tt.want, got)
		})
	}
}
