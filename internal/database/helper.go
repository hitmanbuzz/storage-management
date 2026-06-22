package database

import (
	"context"
	"time"
)

// return (userId, hashPass, error)
func (db *DatabaseHandler) IsUserExist(pctx context.Context, username string) (int32, string, error) {
	ctx, cancel := context.WithTimeout(pctx, 3*time.Second)
	defer cancel()

	query := `SELECT id, password_hash FROM users WHERE username = $1`

	var userId int32
	var hashPass string
	err := db.pool.QueryRow(ctx, query, username).Scan(&userId, &hashPass)
	return userId, hashPass, err
}

func (db *DatabaseHandler) IsHashExist(pctx context.Context, targetHash int64, targetSize int64) (int64, string, error) {
	ctx, cancel := context.WithTimeout(pctx, 3*time.Second)
	defer cancel()

	query := `SELECT id, path FROM files WHERE hash = $1 AND size >= $2`

	var fileId int64
	var filePath string

	err := db.pool.QueryRow(ctx, query, targetHash, targetSize).Scan(&fileId, &filePath)
	return fileId, filePath, err
}
