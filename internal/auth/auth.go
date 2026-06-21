package auth

import (
	"github.com/matthewhartstonge/argon2"
)

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

type AuthHandler struct {
	payload AuthPayload
}

func NewAuthHandler(payload AuthPayload) *AuthHandler {
	return &AuthHandler{
		payload: payload,
	}
}

// encrypt the password (for register)
func (ah *AuthHandler) Encrypt() (AuthPayload, error) {
	argon := argon2.DefaultConfig()
	encoded, _ := argon.HashEncoded([]byte(ah.payload.Password))
	result := NewAuthPayload(ah.payload.Username, string(encoded))
	return result, nil
}
