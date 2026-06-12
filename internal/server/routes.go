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

	var username string
	var totalFileSize int64
	isUpload := false

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}

		if err != nil {
			r.logger.Error("error", "reading part", err)
			ginCtx.JSON(http.StatusInternalServerError, gin.H{"status": "multipart error"})
			return
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

			username = string(userByte)

			if username == "" {
				r.logger.Error("username is empty")
				ginCtx.JSON(http.StatusBadRequest, gin.H{"status": "username is empty"})
				return
			}

			r.logger.Info("uploaded user", "name", username)
		case "file":
			fileName := part.FileName()
			if fileName == "" {
				part.Close()
				continue
			}

			writeSize, fullPath, err := storage.SaveFile(part, fileName)
			part.Close()
			if err != nil {
				r.logger.Error("error", "upload file chunk", err)
				ginCtx.JSON(http.StatusInternalServerError, gin.H{"status": "failed to upload file chunk"})
				return
			}

			totalFileSize += writeSize
			isUpload = true
			r.logger.Info("saved file chunk", "size", writeSize, "path", fullPath)
		default:
			part.Close()
		}
	}

	if !isUpload {
		r.logger.Error("error to upload file")
		ginCtx.JSON(http.StatusBadRequest, gin.H{"status": "failed to upload file"})
		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{
		"status":   "ok",
		"username": username,
	})
}
