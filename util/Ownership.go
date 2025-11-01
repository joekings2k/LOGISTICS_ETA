package util

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func EnsureOwnership(ctx *gin.Context, resourceOwner, resourceUser uuid.UUID) bool {
	if resourceOwner != resourceUser {
		err := errors.New("you do not own this resource")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return false
	}
	return true
}