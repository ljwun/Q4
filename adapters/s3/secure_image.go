package s3

// SecureMIMETypesExtension 定義了允許上傳的安全圖片類型及其對應的副檔名
var SecureMIMETypesExtension = map[string]string{
	"image/jpeg": "jpeg",
	"image/png":  "png",
	"image/gif":  "gif",
	"image/bmp":  "bmp",
	"image/tiff": "tiff",
	"image/webp": "webp",
}

// CheckSecureImageAndGetExtension 檢查給定的 MIME 類型是否為允許的圖片類型，並返回對應的副檔名
func CheckSecureImageAndGetExtension(mimeType string) (bool, string) {
	ext, ok := SecureMIMETypesExtension[mimeType]
	return ok, ext
}
