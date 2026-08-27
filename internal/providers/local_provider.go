package providers

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

type LocalUploadProvider struct {
	basePath string
}

func NewLocalUploadProvider(basePath string) *LocalUploadProvider {
	abs, err := filepath.Abs(basePath)
	if err != nil {
		abs = basePath
	}
	return &LocalUploadProvider{basePath: abs}
}

// resolve joins the key to the base directory and guarantees the result
// stays inside it. Without this check a key containing ".." would let a
// caller write anywhere on the filesystem.
func (p *LocalUploadProvider) resolve(key string) (string, error) {
	cleaned := filepath.Clean("/" + strings.ReplaceAll(key, "\\", "/"))
	fullPath := filepath.Join(p.basePath, cleaned)

	if fullPath != p.basePath && !strings.HasPrefix(fullPath, p.basePath+string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid storage key: %s", key)
	}
	return fullPath, nil
}

func (p *LocalUploadProvider) UploadFile(
	_ context.Context,
	file *multipart.FileHeader,
	key, _ string,
) (string, error) {
	fullPath, err := p.resolve(key)
	if err != nil {
		return "", err
	}

	if mkErr := os.MkdirAll(filepath.Dir(fullPath), 0o755); mkErr != nil {
		return "", mkErr
	}

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, copyErr := dst.ReadFrom(src); copyErr != nil {
		return "", copyErr
	}

	return key, nil
}

func (p *LocalUploadProvider) DeleteFile(_ context.Context, key string) error {
	fullPath, err := p.resolve(key)
	if err != nil {
		return err
	}
	if removeErr := os.Remove(fullPath); removeErr != nil && !os.IsNotExist(removeErr) {
		return removeErr
	}
	return nil
}
