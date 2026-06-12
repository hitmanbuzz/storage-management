package storage

import (
	"io"
	"os"
	"path/filepath"
	"storage-management/internal/util"

	"github.com/google/uuid"
)

func SaveFile(src_data io.Reader, src_name string) (int64, string, error) {
	var fileName string

	ext, hasExt := util.GetExtension(src_name)
	if !hasExt {
		fileName = uuid.NewString()
		ext = "unknown"
	} else {
		fileName = uuid.NewString() + ext
		ext = ext[1:]
	}

	subDir := ext
	dir := filepath.Join(util.BASE_PATH, subDir)
	fullPath := filepath.Join(dir, fileName)

	os.MkdirAll(dir, 0750)
	dst, err := os.Create(fullPath)
	if err != nil {
		return 0, "", err
	}
	defer dst.Close()
	var writeSize int64

	if writeSize, err = io.Copy(dst, src_data); err != nil {
		os.Remove(fullPath)
		return 0, "", err
	}

	return writeSize, filepath.Join(util.BASE_PATH, subDir, fileName), nil // return the path without the `BASE_PATH`
}
