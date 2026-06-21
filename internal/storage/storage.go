package storage

import (
	"io"
	"os"
	"path/filepath"
	"storage-management/internal/util"

	"github.com/google/uuid"
)

func NewFileStorage(fName, ext, path string, size int64, header []byte) util.FileStorage {
	return util.FileStorage{
		Filename: fName,
		Ext:      ext,
		Size:     size,
		Path:     path,
		Header:   header,
	}
}

type UserStorage struct {
	Username string
	Password string // hashed password
}

func SaveFile(headerBytes []byte, src_data io.Reader, src_name string) (util.FileStorage, error) {
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
		return util.FileStorage{}, err
	}
	defer dst.Close()

	var totalWriteSize int64

	if len(headerBytes) > 0 {
		n, err := dst.Write(headerBytes)
		if err != nil {
			os.Remove(fullPath)
			return util.FileStorage{}, err
		}
		totalWriteSize += int64(n)
	}

	copiedSize, err := io.Copy(dst, src_data)
	if err != nil {
		os.Remove(fullPath)
		return util.FileStorage{}, err
	}

	totalWriteSize += copiedSize
	return NewFileStorage(fileName, ext, fullPath, totalWriteSize, headerBytes), nil
}
