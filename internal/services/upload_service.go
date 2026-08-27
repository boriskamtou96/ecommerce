package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"ecommerce/internal/interfaces"
)

const (
	// sniffLen is what http.DetectContentType reads.
	sniffLen = 512
	// keyBytes is the entropy of a generated object name.
	keyBytes = 16
)

var ErrFileTooLarge = errors.New("file exceeds the maximum allowed size")

// allowedImageTypes maps a detected MIME type to the extension used for
// the stored object. SVG is deliberately absent: it can embed scripts and
// would run in the CDN origin.
var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

type UploadService struct {
	provider    interfaces.UploadProvider
	maxFileSize int64
}

func NewUploadService(provider interfaces.UploadProvider, maxFileSize int64) *UploadService {
	return &UploadService{
		provider:    provider,
		maxFileSize: maxFileSize,
	}
}

// UploadProductImage stores an image and returns its storage key.
// The key is generated server side, so an attacker controlled filename
// can neither escape the products/ prefix nor overwrite an existing
// object, which is what makes long lived CDN caching safe.
func (s *UploadService) UploadProductImage(
	ctx context.Context,
	productID uint,
	file *multipart.FileHeader,
) (string, error) {
	if s.maxFileSize > 0 && file.Size > s.maxFileSize {
		return "", fmt.Errorf("%w: %d bytes (max %d)", ErrFileTooLarge, file.Size, s.maxFileSize)
	}

	contentType, ext, err := detectImageType(file)
	if err != nil {
		return "", err
	}

	name, err := randomName()
	if err != nil {
		return "", err
	}

	key := fmt.Sprintf("products/%d/%s%s", productID, name, ext)

	return s.provider.UploadFile(ctx, file, key, contentType)
}

func (s *UploadService) DeleteFile(ctx context.Context, key string) error {
	return s.provider.DeleteFile(ctx, key)
}

// detectImageType inspects the actual bytes instead of trusting the
// filename extension sent by the client.
func detectImageType(file *multipart.FileHeader) (contentType, ext string, err error) {
	src, err := file.Open()
	if err != nil {
		return "", "", err
	}
	defer src.Close()

	buf := make([]byte, sniffLen)
	n, err := io.ReadFull(src, buf)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return "", "", err
	}

	detected := http.DetectContentType(buf[:n])

	ext, ok := allowedImageTypes[detected]
	if !ok {
		return "", "", fmt.Errorf("unsupported file type: %s", detected)
	}

	return detected, ext, nil
}

func randomName() (string, error) {
	buf := make([]byte, keyBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
