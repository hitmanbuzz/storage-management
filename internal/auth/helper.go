package auth

import (
	"storage-management/internal/util"

	"github.com/matthewhartstonge/argon2"
)

func Encrypt(payload util.AuthPayload) util.AuthPayload {
	argon := argon2.DefaultConfig()
	encoded, _ := argon.HashEncoded([]byte(payload.Password))
	result := util.NewAuthPayload(payload.Username, string(encoded))
	return result
}
