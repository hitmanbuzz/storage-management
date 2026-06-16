package storage

import (
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

func (u *Upload) AddFile(file *File) {
	u.Files = append(u.Files, file)
}

type UploadHandler struct {
	reader *multipart.Reader
	upload *Upload
	logger *slog.Logger
}

func NewUploadHandler(mr *multipart.Reader, logger *slog.Logger) *UploadHandler {
	return &UploadHandler{
		reader: mr,
		upload: NewUpload(),
		logger: logger,
	}
}

func (hu *UploadHandler) Do(ginCtx *gin.Context) bool {
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

		errResponse := hu.HandleForm(ginCtx, form, part)
		if errResponse != nil {
			errResponse.Do(ginCtx, hu.logger)
			return false
		}
	}

	return true
}

func (hu *UploadHandler) HandleForm(ginCtx *gin.Context, formName string, part *multipart.Part) *util.ErrorResponse {
	switch formName {
	case "user":
		err := hu.HandleUser(ginCtx, part)
		return err
	case "file":
		err := hu.HandleFile(ginCtx, part)
		return err
	}
	return nil
}

func (hu *UploadHandler) HandleUser(ginCtx *gin.Context, part *multipart.Part) *util.ErrorResponse {
	userName, err := hu.GetUser(part)
	part.Close()
	if err != nil {
		return util.NewErrResponse(
			http.StatusTeapot,
			gin.H{"status": "failed reading username"},
			err.Error(),
		)
	}

	if len(userName) == 0 {
		return util.NewErrResponse(
			http.StatusBadRequest,
			gin.H{"status": "username is empty"},
			"username is empty",
		)
	}

	if len(hu.upload.User.Name) == 0 {
		hu.upload.SetUsername(userName)
	}

	hu.logger.Info("uploaded user", "name", userName)
	return nil
}

func (hu *UploadHandler) HandleFile(ginCtx *gin.Context, part *multipart.Part) *util.ErrorResponse {
	fileName, err := hu.GetFile(part)
	if err != nil {
		part.Close()
		return util.NewErrResponse(http.StatusBadRequest, gin.H{"status": "empty filename"}, "empty filename")
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
		return util.NewErrResponse(
			http.StatusInternalServerError,
			gin.H{"status": "failed to upload file chunk"},
			util.ErrToString("error uploading file chunk", err),
		)
	}

	hu.upload.CurrFile.Size += writeSize
	hu.logger.Info("saved file", "size", writeSize, "path", fullPath)
	return nil
}

func (hu *UploadHandler) HandleAuth() {
	// TODO: Still need to think how a simple auth need to be implemented
}
