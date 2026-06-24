package server

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func SaveSession(ginCtx *gin.Context, userId int32) {
	session := sessions.Default(ginCtx)
	session.Options(sessions.Options{
		Path:     "/",
		Domain:   "",
		MaxAge:   86400, // 24 hrs
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	session.Set("userID", userId)
	session.Save()
}

func AuthMiddleware(ginCtx *gin.Context) {
	session := sessions.Default(ginCtx)
	userId := session.Get("userID")
	if userId == nil {
		ginCtx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
		return
	}

	ginCtx.Set("userID", userId)
	ginCtx.Next()
}
