package storage

import (
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"storage-management/internal/util"

	"github.com/gin-gonic/gin"
)

type User struct {
	Name  string `json:"username"`
	Token string `json:"-"`
}

type File struct {
	Name   string `json:"filename"`
	Size   int64  `json:"filesize"`
	Ext    string `json:"-"`
	IsErr  bool   `json:"-"`
	Status bool   `json:"status"`
}

func NewFile(name string, ext string) *File {
	return &File{
		Name:   name,
		Size:   0,
		Ext:    ext,
		IsErr:  false,
		Status: true,
	}
}

type Upload struct {
	User     User    `json:"user"`
	Files    []*File `json:"files"`
	CurrFile *File   `json:"-"`
}

func NewUpload() *Upload {
	return &Upload{
		Files: make([]*File, 0),
	}
}

func (u *Upload) SetUsername(name string) {
	u.User.Name = name
}

func (u *Upload) SetToken(token string) {
	u.User.Token = token
}

func (u *Upload) AddFile(file *File) {
	u.Files = append(u.Files, file)
}

func (u *Upload) UpdateCurrFile(file *File) {
	if u.CurrFile == nil {
		u.CurrFile = file
		return
	}

	if u.CurrFile.Name != file.Name {
		u.AddFile(u.CurrFile)
		u.CurrFile = file
		return
	}
}

type HandleUpload struct {
	reader *multipart.Reader
	upload *Upload
	logger *slog.Logger
}

func NewHandleUpload(mr *multipart.Reader, logger *slog.Logger) *HandleUpload {
	return &HandleUpload{
		reader: mr,
		upload: NewUpload(),
		logger: logger,
	}
}

func (hu *HandleUpload) Do(ginCtx *gin.Context) {
	for {
		part, err := hu.reader.NextPart()
		if err == io.EOF {
			if hu.upload.CurrFile != nil {
				hu.logger.Debug("eof reached", "add file", hu.upload.CurrFile.Name)
				hu.upload.AddFile(hu.upload.CurrFile)
			}
			break
		}

		if err != nil {
			hu.logger.Error("error", "reading part", err)
			ginCtx.JSON(http.StatusInternalServerError, gin.H{"status": "multipart error"})
		}

		if hu.upload.CurrFile != nil {
			if hu.upload.CurrFile.IsErr {
				hu.upload.AddFile(hu.upload.CurrFile)
				hu.upload.CurrFile = nil
			}
		}

		form, err := hu.GetForm(part)
		if err != nil {
			ginCtx.JSON(http.StatusBadRequest, gin.H{"status": "form name is empty"})
			part.Close()
			continue
		}

		hu.HandleForm(ginCtx, form, part)
	}
}

func (hu *HandleUpload) HandleForm(ginCtx *gin.Context, formName string, part *multipart.Part) {
	switch formName {
	case "user":
		hu.HandleUser(ginCtx, part)
	case "file":
		hu.HandleFile(ginCtx, part)
	}
}

func (hu *HandleUpload) HandleUser(ginCtx *gin.Context, part *multipart.Part) {
	userName, err := hu.GetUser(part)
	part.Close()
	if err != nil {
		hu.logger.Error("error", "failed to read username", err)
		return
	}

	if len(userName) == 0 {
		hu.logger.Error("username is empty")
		ginCtx.JSON(http.StatusBadRequest, gin.H{"status": "username is empty"})
		return
	}

	if len(hu.upload.User.Name) == 0 {
		hu.upload.SetUsername(userName)
	}

	hu.logger.Info("uploaded user", "name", userName)
}

func (hu *HandleUpload) HandleFile(ginCtx *gin.Context, part *multipart.Part) {
	fileName, err := hu.GetFile(part)
	if err != nil {
		hu.logger.Warn("empty filename")
		part.Close()
		return
	}

	ext, _ := util.GetExtension(fileName)
	if hu.upload.CurrFile == nil {
		hu.upload.CurrFile = NewFile(fileName, ext)
	} else if hu.upload.CurrFile.Name != fileName {
		hu.upload.AddFile(hu.upload.CurrFile)
		hu.upload.CurrFile = NewFile(fileName, ext)
	}

	writeSize, fullPath, err := SaveFile(part, fileName)
	part.Close()
	if err != nil {
		hu.upload.CurrFile.IsErr = true
		hu.upload.CurrFile.Status = false
		hu.logger.Error("error", "upload file chunk", err)
		ginCtx.JSON(http.StatusInternalServerError, gin.H{"status": "failed to upload file chunk"})
		return
	}

	hu.upload.CurrFile.Size += writeSize
	hu.logger.Info("saved file chunk", "size", writeSize, "path", fullPath)
}

func (hu *HandleUpload) HandleAuth() {
	// TODO: Still need to think how a simple auth need to be implemented
}

// --- Helper Methods ---

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
