package interfaces

import (
	"context"
	"mime/multipart"
)

// UploadProvider abstracts the object storage backing the CDN.
// Implementations receive a storage key that is already sanitised by the
// caller and the exact Content-Type to persist alongside the object, so
// that the CDN never has to guess or rewrite it.
type UploadProvider interface {
	UploadFile(ctx context.Context, file *multipart.FileHeader, key, contentType string) (string, error)
	DeleteFile(ctx context.Context, key string) error
}
