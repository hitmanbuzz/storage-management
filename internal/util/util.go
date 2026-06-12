package util

import (
	"crypto/sha512"
	"encoding/hex"
	"path/filepath"

	"github.com/bytedance/gopkg/util/xxhash3"
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
		return "unknown", false
	} else {
		return ext, true
	}
}
