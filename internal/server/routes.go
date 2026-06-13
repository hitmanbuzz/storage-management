package server

import (
	"io"
	"log/slog"
	"net/http"
	"storage-management/internal/storage"
	"storage-management/internal/util"

	"github.com/gin-gonic/gin"
)

type route struct {
	engine *gin.Engine
	logger *slog.Logger
}

func newRoutes(engine *gin.Engine, logger *slog.Logger) *route {
	return &route{
		engine: engine,
		logger: logger,
	}
}

func (r *route) Upload(ginCtx *gin.Context) {
	ginCtx.Request.Body = http.MaxBytesReader(ginCtx.Writer, ginCtx.Request.Body, util.MAX_REQUEST_SIZE)

	mr, err := ginCtx.Request.MultipartReader()
	if err != nil {
		r.logger.Error("error reading", "upload file", err)
		ginCtx.JSON(http.StatusBadRequest, gin.H{"status": "expect multipart"})
		return
	}

	upload := storage.NewUpload()

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			if upload.CurrFile != nil {
				r.logger.Debug("eof reached", "add file", upload.CurrFile.Name)
				upload.AddFile(upload.CurrFile)
			}
			break
		}

		if err != nil {
			r.logger.Error("error", "reading part", err)
			ginCtx.JSON(http.StatusInternalServerError, gin.H{"status": "multipart error"})
			return
		}

		if upload.CurrFile != nil {
			if upload.CurrFile.IsErr {
				upload.AddFile(upload.CurrFile)
				upload.CurrFile = nil
			}
		}

		formName := part.FormName()

		if formName == "" {
			part.Close()
			continue
		}

		switch formName {
		case "user":
			userByte, err := io.ReadAll(part)
			part.Close()
			if err != nil {
				r.logger.Error("error", "failed to read username", err)
				return
			}

			username := string(userByte)

			if username == "" {
				r.logger.Error("username is empty")
				ginCtx.JSON(http.StatusBadRequest, gin.H{"status": "username is empty"})
				return
			}

			if upload.User.Name == "" {
				upload.SetUsername(username)
			}

			r.logger.Info("uploaded user", "name", username)
		case "token":
			// TODO: handle token
		case "file":
			fileName := part.FileName()
			if fileName == "" {
				part.Close()
				continue
			}

			ext, _ := util.GetExtension(fileName)

			if upload.CurrFile == nil {
				upload.CurrFile = storage.NewFile(fileName, ext)
				r.logger.Debug("update current file from nil", "file", fileName)
			} else if upload.CurrFile.Name != fileName {
				upload.AddFile(upload.CurrFile)
				upload.CurrFile = storage.NewFile(fileName, ext)
			}

			writeSize, fullPath, err := storage.SaveFile(part, fileName)
			part.Close()
			if err != nil {
				upload.CurrFile.IsErr = true
				upload.CurrFile.Status = false
				r.logger.Error("error", "upload file chunk", err)
				ginCtx.JSON(http.StatusInternalServerError, gin.H{"status": "failed to upload file chunk"})
				continue
			}

			upload.CurrFile.Size += writeSize
			r.logger.Info("saved file chunk", "size", writeSize, "path", fullPath)
		default:
			part.Close()
		}
	}

	ginCtx.JSON(http.StatusOK, upload)
}
