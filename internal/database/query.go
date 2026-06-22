package database

import (
	"context"
	"storage-management/internal/util"
	"time"
)

// `payload` contains the hash password (different from the payload sent from client)
//
// return (user, error)
func (db *DatabaseHandler) InsertUser(pctx context.Context, payload util.AuthPayload) (util.User, error) {
	ctx, cancel := context.WithTimeout(pctx, 3*time.Second)
	defer cancel()

	query := `
		INSERT into users (username, password_hash)
		VALUES ($1, $2)
		ON CONFLICT (username) DO NOTHING
		RETURNING id
	`

	var user util.User
	user.Name = payload.Username

	err := db.pool.QueryRow(ctx, query, payload.Username, payload.Password).Scan(&user.Id)
	return user, err
}

// return (fileId, error)
func (db *DatabaseHandler) InsertFile(pctx context.Context, f util.FileStorage, userId int32) (int64, error) {
	ctx, cancel := context.WithTimeout(pctx, 3*time.Second)
	defer cancel()

	initHash := util.GetXhHash(f.Header)

	query := `
		INSERT into files (filename, extension, path, size, hash, file_group, file_desc, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	var fileId int64
	err := db.pool.QueryRow(ctx, query, f.Filename, f.Ext, f.Path, f.Size, initHash, nil, nil, userId).Scan(&fileId)
	return fileId, err
}

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

// return (fileId, filePath, error)
func (db *DatabaseHandler) IsHashExist(pctx context.Context, targetHash int64, targetSize int64) (int64, string, error) {
	ctx, cancel := context.WithTimeout(pctx, 3*time.Second)
	defer cancel()

	query := `SELECT id, path FROM files WHERE hash = $1 AND size >= $2`

	var fileId int64
	var filePath string

	err := db.pool.QueryRow(ctx, query, targetHash, targetSize).Scan(&fileId, &filePath)
	return fileId, filePath, err
}
