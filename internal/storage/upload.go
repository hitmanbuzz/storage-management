package storage

import (
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"storage-management/internal/auth"
	"storage-management/internal/database"
	"storage-management/internal/util"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
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
				hu.logger.Debug("done uploading file", "filename", hu.upload.CurrFile.Name)
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
	case "file":
		err := hu.HandleFile(ginCtx, part)
		return err
	default:
		return util.NewErrResponse(http.StatusBadRequest, gin.H{"status": "invalid form"}, fmt.Sprintf("invalid form: %s", formName))
	}
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

	fileId, filePath, err := hu.db.IsHashExist(
		ginCtx.Request.Context(),
		util.GetXhHash(headerBuf),
	)

	switch err {
	case pgx.ErrNoRows:
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

		userId, exist := auth.GetUserId(ginCtx)
		if !exist {
			hu.upload.CurrFile.IsErr = true
			hu.upload.CurrFile.Status = false
			return util.NewErrResponse(
				http.StatusUnauthorized,
				gin.H{"status": "unauthorized"},
				"failed to upload due to uanuthorized",
			)
		}

		fileId, err := hu.db.InsertFile(ginCtx, f, userId)
		if err == pgx.ErrNoRows {
			os.Remove(f.Path)
			return util.NewErrResponse(
				http.StatusInternalServerError,
				gin.H{"status": "failed to store file metadata"},
				util.ErrToString("error storing file metadata", err),
			)
		}
		hu.upload.CurrFile.Size += f.Size
		hu.logger.Info("saved file", "file id", fileId, "filename", f.Filename, "extension", f.Ext, "size", f.Size, "path", f.Path)
		return nil
	default:
		hu.upload.CurrFile.IsErr = true
		hu.upload.CurrFile.Status = false
		return util.NewErrResponse(
			http.StatusConflict,
			gin.H{
				"status":    "file already exist",
				"file id":   fileId,
				"file path": filePath, // will remove later
			},
			fmt.Sprintf("file already exist: file id = %d | filepath = %s", fileId, filePath),
		)
	}
}
