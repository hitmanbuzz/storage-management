package storage

// this file contain helper functions/methods

import (
	"fmt"
	"io"
	"mime/multipart"
)

func (hu *HandleUpload) GetForm(part *multipart.Part) (string, error) {
	formName := part.FormName()
	if len(formName) == 0 {
		return "", fmt.Errorf("form name is empty")
	}
	return formName, nil
}

func (hu *HandleUpload) GetFile(part *multipart.Part) (string, error) {
	fileName := part.FileName()
	if len(fileName) == 0 {
		return "", fmt.Errorf("file name is empty")
	}
	return fileName, nil
}

func (hu *HandleUpload) GetUser(part *multipart.Part) (string, error) {
	userByte, err := io.ReadAll(part)
	part.Close()
	if err != nil {
		return "", err
	}

	return string(userByte), nil
}

func (hu *HandleUpload) GetUpload() *Upload {
	return hu.upload
}
