package server

import (
	"io"
	"log"
	"net/http"
	"storage-management/internal/util"

	"github.com/gin-gonic/gin"
)

type route struct {
	engine *gin.Engine
}

func newRoute(engine *gin.Engine) *route {
	return &route{
		engine: engine,
	}
}

func (r *route) Upload(ginCtx *gin.Context) {
	ginCtx.Request.Body = http.MaxBytesReader(ginCtx.Writer, ginCtx.Request.Body, util.MAX_FILE_SIZE)

	mr, err := ginCtx.Request.MultipartReader()
	if err != nil {
		log.Println("error reading upload file:", err)
		ginCtx.JSON(http.StatusBadRequest, gin.H{"status": "not multipart"})
		return
	}

	for {
		_, err := mr.NextPart()
		if err == io.EOF {
			break
		}

		if err != nil {
			ginCtx.JSON(http.StatusInternalServerError, gin.H{"status": "multipart error"})
			return
		}
	}
}
