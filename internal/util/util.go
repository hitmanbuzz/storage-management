package util

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/bytedance/gopkg/util/xxhash3"
	"github.com/gin-gonic/gin"
)

const MAX_REQUEST_SIZE = 10 << 30 // 10 GB
const BASE_PATH = "test_dir"
const MAX_BYTE_READ = 4096 // 4KB

// sha512
func GetShaHash(b []byte) string {
	hashBytes := sha512.Sum512(b)
	return hex.EncodeToString(hashBytes[:])
}

// xxhash3
func GetXhHash(b []byte) uint64 {
	return xxhash3.Hash(b)
}

// compare two xxhash3 hashes
func CompareXhHash(a, b uint64) bool {
	return a == b
}

// compare two sha hash (same as string comparison)
func CompareShaHash(a, b string) bool {
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
