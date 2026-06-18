package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"storage-management/internal/util"

	"github.com/google/uuid"
)

type FileStorage struct {
	Filename string
	Ext      string
	Path     string
	Size     int64
	Header   []byte
}

func NewFileStorage(fName, ext, path string, size int64, header []byte) FileStorage {
	return FileStorage{
		Filename: fName,
		Ext:      ext,
		Size:     size,
		Path:     path,
		Header:   header,
	}
}

func SaveFile(headerBytes []byte, src_data io.Reader, src_name string) (FileStorage, error) {
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
		return FileStorage{}, err
	}
	defer dst.Close()

	var totalWriteSize int64

	if len(headerBytes) > 0 {
		n, err := dst.Write(headerBytes)
		if err != nil {
			os.Remove(fullPath)
			return FileStorage{}, err
		}
		totalWriteSize += int64(n)
		fmt.Println("Total Write Size:", int64(n))
	}

	copiedSize, err := io.Copy(dst, src_data)
	if err != nil {
		os.Remove(fullPath)
		return FileStorage{}, err
	}

	fmt.Println("Copy Size:", copiedSize)
	totalWriteSize += copiedSize
	return NewFileStorage(fileName, ext, fullPath, totalWriteSize, headerBytes), nil
}
