package database

import (
	"context"
	"fmt"
	"storage-management/internal/auth"
	"storage-management/internal/util"
	"time"

	"github.com/jackc/pgx/v5"
)

// `payload` contains the hash password (different from the payload sent from client)
func (db *DatabaseHandler) InsertUser(pctx context.Context, payload auth.AuthPayload) (util.User, error) {
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
	if err != nil {
		if err == pgx.ErrNoRows {
			return user, fmt.Errorf("username already exist")
		}
		return user, err
	}

	return user, nil
}

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
	if err != nil {
		return -1, err
	}

	return fileId, nil
}
