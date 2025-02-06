package s3_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"q4/adapters/s3"
)

func TestCheckSecureImageAndGetExtension(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
		wantOk   bool
		wantExt  string
	}{
		{
			name:     "valid JPEG image",
			mimeType: "image/jpeg",
			wantOk:   true,
			wantExt:  "jpeg",
		},
		{
			name:     "valid PNG image",
			mimeType: "image/png",
			wantOk:   true,
			wantExt:  "png",
		},
		{
			name:     "invalid image type",
			mimeType: "application/pdf",
			wantOk:   false,
			wantExt:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOk, gotExt := s3.CheckSecureImageAndGetExtension(tt.mimeType)
			assert.Equal(t, tt.wantOk, gotOk)
			assert.Equal(t, tt.wantExt, gotExt)
		})
	}
}
