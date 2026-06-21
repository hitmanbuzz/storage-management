package storage

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"storage-management/internal/auth"
	"storage-management/internal/database"
	"storage-management/internal/util"

	"github.com/gin-gonic/gin"
)

func NewFile(name string, ext string) *util.File {
	return &util.File{
		Name:   name,
		Size:   0,
		Ext:    ext,
		IsErr:  false,
		Status: true,
	}
}

type Upload struct {
	User     util.User    `json:"user"`
	Files    []*util.File `json:"files"`
	CurrFile *util.File   `json:"-"`
}

func NewUpload() *Upload {
	return &Upload{
		Files: make([]*util.File, 0),
	}
}

func (u *Upload) SetUsername(name string) {
	u.User.Name = name
}

func (u *Upload) AddFile(file *util.File) {
	u.Files = append(u.Files, file)
}

type UploadHandler struct {
	reader *multipart.Reader
	upload *Upload
	db     *database.DatabaseHandler
	logger *slog.Logger
}

func NewUploadHandler(mr *multipart.Reader, db *database.DatabaseHandler, logger *slog.Logger) *UploadHandler {
	return &UploadHandler{
		reader: mr,
		upload: NewUpload(),
		db:     db,
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
	case "password":
		// TODO: handle password
	case "file":
		err := hu.HandleFile(ginCtx, part)
		return err
	}
	return nil
}

func (hu *UploadHandler) HandleUser(ginCtx *gin.Context, part *multipart.Part) *util.ErrorResponse {
	username, err := hu.GetData(part)
	part.Close()
	if err != nil {
		return util.NewErrResponse(
			http.StatusTeapot,
			gin.H{"status": "failed reading username"},
			err.Error(),
		)
	}

	if len(username) < util.MIN_USER_LEN || len(username) > util.MAX_USER_LEN {
		return util.NewErrResponse(
			http.StatusBadRequest,
			gin.H{"status": fmt.Sprintf("username should be %d - %d in length", util.MIN_USER_LEN, util.MAX_USER_LEN)},
			"username length requirement is not met",
		)
	}

	if len(hu.upload.User.Name) == 0 {
		hu.upload.SetUsername(username)
	}

	hu.logger.Info("uploaded user", "name", username)
	return nil
}

func (hu *UploadHandler) HandlePassword(ginCtx *gin.Context, part *multipart.Part) *util.ErrorResponse {
	password, err := hu.GetData(part)
	part.Close()
	if err != nil {
		return util.NewErrResponse(
			http.StatusTeapot,
			gin.H{"status": "failed reading password"},
			err.Error(),
		)
	}

	if len(password) < util.MIN_PASS_LEN || len(password) > util.MAX_PASS_LEN {
		return util.NewErrResponse(
			http.StatusBadRequest,
			gin.H{"status": fmt.Sprintf("password should be %d - %d in length", util.MIN_PASS_LEN, util.MAX_PASS_LEN)},
			"password length requirement not met",
		)
	}

	_, _, err = hu.db.IsUserExist(ginCtx.Request.Context(), hu.upload.User.Name)
	if err != nil {
		return util.NewErrResponse(
			http.StatusNotFound,
			gin.H{"status": "user not found"},
			"user not found",
		)
	}

	// TODO

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

	headerBuf, err := hu.ReadHeader(part)
	if err != nil {
		hu.upload.CurrFile.IsErr = true
		hu.upload.CurrFile.Status = false
		return util.NewErrResponse(
			http.StatusInternalServerError,
			gin.H{"status": "failed to read file header"},
			util.ErrToString("error reading file header", err),
		)
	}

	// TODO: find the `headerBuf` xxhash value within the database

	f, err := SaveFile(headerBuf, part, fileName)
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

	hu.logger.Debug("total file size", "size", f.Size)

	hu.upload.CurrFile.Size = f.Size
	hu.logger.Info("saved file", "size", f.Size, "path", f.Path)
	return nil
}

func (hu *UploadHandler) HandleAuth() {
	// TODO: Still need to think how a simple auth need to be implemented
}
