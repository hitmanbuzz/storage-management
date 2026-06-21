package server

import (
	"log/slog"
	"net/http"
	"storage-management/internal/database"
	"storage-management/internal/storage"
	"storage-management/internal/util"

	"github.com/gin-gonic/gin"
)

type route struct {
	engine *gin.Engine
	logger *slog.Logger
	db     *database.DatabaseHandler
}

func newRoutes(engine *gin.Engine, db *database.DatabaseHandler, logger *slog.Logger) *route {
	return &route{
		engine: engine,
		logger: logger,
		db:     db,
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

	upload := storage.NewUploadHandler(mr, r.db, r.logger)
	status := upload.Do(ginCtx)
	if status {
		ginCtx.JSON(http.StatusOK, upload.GetUpload())
	}
}
