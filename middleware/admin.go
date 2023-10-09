package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spo-iitk/ras-backend/constants"
)

func EnsureAdmin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		role := GetRoleID(ctx)

		if role != constants.OPC && role != constants.GOD {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		ctx.Next()
	}
}

func EnsurePsuedoAdmin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		role := GetRoleID(ctx)

		if role != constants.OPC && role != constants.GOD && role != constants.APC && role != constants.CHAIR {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		ctx.Next()
	}
}
