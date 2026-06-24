package util

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/bytedance/gopkg/util/xxhash3"
	"github.com/gin-gonic/gin"
)

const MAX_REQUEST_SIZE = 10 << 30 // 10 GB
const BASE_PATH = "test_dir"
const MAX_BYTE_READ = 4096 // 4KB
const MAX_DB_PING = 3
const MAX_COOKIE_AGE = 3600 * 24 // 24 hrs

// user
const MIN_USER_LEN = 4
const MAX_USER_LEN = 16

// password
const MIN_PASS_LEN = 8
const MAX_PASS_LEN = 24

type FileData struct {
	Id          int64   `db:"id"`
	Filename    string  `db:"filename"`
	Extension   *string `db:"extension"`
	Path        string  `db:"path"`
	Size        int64   `db:"size"`
	Hash        int64   `db:"hash"`
	UploadedAt  string  `db:"uploaded_at"`
	Group       *string `db:"file_group"`
	Description *string `db:"file_desc"`
	UserId      int32   `db:"user_id"`
}

type User struct {
	Id   int32  `json:"user_id"`
	Name string `json:"username"`
}

type File struct {
	Name   string `json:"filename"`
	Size   int64  `json:"filesize"`
	Ext    string `json:"-"`
	IsErr  bool   `json:"-"`
	Status bool   `json:"status"`
}

type FileStorage struct {
	Filename string
	Ext      string
	Path     string
	Size     int64
	Header   []byte
}

// this is use for client as well as server side for hashing password
type AuthPayload struct {
	Username string
	Password string // can be hash or not depending on its usage
}

func NewAuthPayload(username, password string) AuthPayload {
	return AuthPayload{
		Username: username,
		Password: password,
	}
}

// compute xxhash3
func GetXhHash(b []byte) uint64 {
	return xxhash3.Hash(b)
}

// compare two xxhash3 hashes
func CompareXhHash(a, b uint64) bool {
	return a == b
}

func GetExtension(fileName string) (string, bool) {
	ext := filepath.Ext(fileName)
	if ext == "" {
		return "", false
	} else {
		return ext, true
	}
}

func ErrToString(msg string, err error) string {
	return fmt.Errorf("%s: %w", msg, err).Error()
}

type ErrorResponse struct {
	errStatusCode int            // the response status code type
	errStatusMsg  map[string]any // for sending out for response
	errLogMsg     string         // for logging
}

func NewErrResponse(code int, errStatusMsg map[string]any, logMsg string) *ErrorResponse {
	return &ErrorResponse{
		errStatusCode: code,
		errStatusMsg:  errStatusMsg,
		errLogMsg:     logMsg,
	}
}

func (er *ErrorResponse) Do(ginCtx *gin.Context, logger *slog.Logger) {
	logger.Error(er.errLogMsg)
	ginCtx.JSON(er.errStatusCode, er.errStatusMsg)
}
