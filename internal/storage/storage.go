package storage

import (
	"io"
	"os"
	"path/filepath"
	"storage-management/internal/util"

	"github.com/google/uuid"
)

func hasExtension(fileName string) (string, bool) {
	ext := filepath.Ext(fileName)
	if ext == "" {
		return "unknown", false
	} else {
		return ext, true
	}
}

func SaveFile(src_data io.Reader, src_name string) (string, error) {
	var fileName string

	ext, hasExt := hasExtension(src_name)
	if !hasExt {
		fileName = uuid.NewString()
		ext = "unknown"
	} else {
		fileName = uuid.NewString() + ext
		ext = ext[1:]
	}

	subDir := ext
	dir := filepath.Join(util.BASE_PATH, subDir)

	os.MkdirAll(dir, 0750)
	dst, err := os.Create(filepath.Join(dir, fileName))
	if err != nil {
		return "", err
	}
	defer dst.Close()
	if _, err := io.Copy(dst, src_data); err != nil {
		return "", err
	}

	return filepath.Join(subDir, fileName), nil // return the path without the `BASE_PATH`
}
