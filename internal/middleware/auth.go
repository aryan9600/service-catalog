package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/aryan9600/service-catalog/internal/auth"
	"github.com/aryan9600/service-catalog/internal/models"
	"github.com/gin-gonic/gin"
)

// JwtAuthMiddleware returns a middleware that checks if the request originates
// from an authenticated user. If it does, it sets the user's ID in the request's
// context under the 'userID' key.
func JwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		userID, err := auth.ExtractUserIDFromToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthenticated"})
			c.Abort()
			return
		}
		_, err = models.GetUserByID(userID)
		if err != nil {
			if errors.Is(err, models.ErrRecordNotFound) {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthenticated"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("unable to fetch user")})
			}
			c.Abort()
			return
		}
		c.Set("userID", userID)
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	bearerToken := c.Request.Header.Get("Authorization")
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1]
	}
	return ""
}
