package storage

// this file contain helper functions/methods

import (
	"fmt"
	"io"
	"mime/multipart"
	"storage-management/internal/util"
)

func (hu *UploadHandler) GetForm(part *multipart.Part) (string, error) {
	formName := part.FormName()
	if len(formName) == 0 {
		return "", fmt.Errorf("form name is empty")
	}
	return formName, nil
}

func (hu *UploadHandler) GetFile(part *multipart.Part) (string, error) {
	fileName := part.FileName()
	if len(fileName) == 0 {
		return "", fmt.Errorf("file name is empty")
	}
	return fileName, nil
}

func (hu *UploadHandler) GetUpload() *Upload {
	return hu.upload
}

// read the first provided bytes from the upload file (before saving it)
func (hu *UploadHandler) ReadHeader(part *multipart.Part) ([]byte, error) {
	headerBuf := make([]byte, util.MAX_BYTE_READ)
	n, err := io.ReadFull(part, headerBuf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return headerBuf, err
	}

	return headerBuf[:n], nil
}
