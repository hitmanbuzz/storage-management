package auth

import (
	"context"
	"fmt"
	"net/http"
	"storage-management/internal/database"
	"storage-management/internal/util"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/matthewhartstonge/argon2"
)

type AuthHandler struct {
	Payload util.AuthPayload
}

// return (nil, err) if username or password length requirement is not met
func NewAuthHandler(payload util.AuthPayload) (*AuthHandler, error) {
	if len(payload.Username) > util.MAX_USER_LEN || len(payload.Username) < util.MIN_USER_LEN {
		return nil, fmt.Errorf("username length should be within %d - %d: %d", util.MIN_USER_LEN, util.MAX_USER_LEN, len(payload.Username))
	} else if len(payload.Password) > util.MAX_PASS_LEN || len(payload.Password) < util.MIN_PASS_LEN {
		return nil, fmt.Errorf("password length should be within %d - %d: %d", util.MIN_PASS_LEN, util.MAX_PASS_LEN, len(payload.Password))
	}

	return &AuthHandler{
		Payload: payload,
	}, nil
}

func (ah *AuthHandler) Register(ctx context.Context, db *database.DatabaseHandler) *util.ErrorResponse {
	payload := Encrypt(ah.Payload)
	user, err := db.InsertUser(ctx, payload)
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			return util.NewErrResponse(
				http.StatusNotFound,
				gin.H{"status": "user already exist"},
				fmt.Sprintf("user already exist: username = %s", ah.Payload.Username),
			)
		default:
			return util.NewErrResponse(
				http.StatusInternalServerError,
				gin.H{"status": "internal server error"},
				fmt.Sprintf("error inserting user: username = %s | password = %s | error = %v", ah.Payload.Username, ah.Payload.Password, err),
			)
		}
	}

	return util.NewErrResponse(
		http.StatusCreated,
		gin.H{"status": "user created"},
		fmt.Sprintf("user created: id = %d | username = %s", user.Id, user.Name),
	)
}

func (ah *AuthHandler) Login(ctx context.Context, db *database.DatabaseHandler) *util.ErrorResponse {
	userID, hashPass, err := db.IsUserExist(ctx, ah.Payload.Username)
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			return util.NewErrResponse(
				http.StatusNotFound,
				gin.H{"status": "user not found"},
				fmt.Sprintf("user not found: username = %s | password = %s", ah.Payload.Username, ah.Payload.Password),
			)
		default:
			return util.NewErrResponse(
				http.StatusInternalServerError,
				gin.H{"status": "internal server error"},
				fmt.Sprintf("error checking user exist: username = %s | password = %s | error = %v", ah.Payload.Username, ah.Payload.Password, err),
			)
		}
	}

	ok, err := argon2.VerifyEncoded([]byte(ah.Payload.Password), []byte(hashPass))
	if err != nil {
		return util.NewErrResponse(
			http.StatusInternalServerError,
			gin.H{"status": "internal server error"},
			fmt.Sprintf("invalid argon2 hash, error: %v", err),
		)
	}

	if !ok {
		return util.NewErrResponse(
			http.StatusUnauthorized,
			gin.H{"status": "incorrect password"},
			fmt.Sprintf("incorrect password: username = %s | password = %s", ah.Payload.Username, ah.Payload.Password),
		)
	}

	return util.NewErrResponse(
		http.StatusOK,
		gin.H{"status": "login successful"},
		fmt.Sprintf("user login successful: id = %d | username = %s", userID, ah.Payload.Username),
	)
}
