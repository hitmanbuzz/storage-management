package server

import (
	"io"
	"log"
	"net/http"
	"storage-management/internal/storage"
	"storage-management/internal/util"

	"github.com/gin-gonic/gin"
)

type route struct {
	engine *gin.Engine
}

func newRoutes(engine *gin.Engine) *route {
	return &route{
		engine: engine,
	}
}

func (r *route) Upload(ginCtx *gin.Context) {
	ginCtx.Request.Body = http.MaxBytesReader(ginCtx.Writer, ginCtx.Request.Body, util.MAX_REQUEST_SIZE)

	mr, err := ginCtx.Request.MultipartReader()
	if err != nil {
		log.Println("error reading upload file:", err)
		ginCtx.JSON(http.StatusBadRequest, gin.H{"status": "not multipart"})
		return
	}

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}

		if err != nil {
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
				break
			}

			username := string(userByte)
			log.Println("Upload User:", username)
		case "file":
			fileName := part.FileName()
			if fileName == "" {
				part.Close()
				continue
			}

			writeSize, subPath, err := storage.SaveFile(part, fileName)
			part.Close()
			if err != nil {
				log.Println(err)
				ginCtx.String(http.StatusInternalServerError, "write failed")
				return
			}

			log.Printf("FILE SAVED: %d | %s/%s\n", writeSize, util.BASE_PATH, subPath)
		default:
			part.Close()
		}
	}

	ginCtx.Status(200)
}
