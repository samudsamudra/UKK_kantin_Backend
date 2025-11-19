package admin

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/samudsamudra/UKK_kantin/internal/app"
	"net/http"
)

// ErrNoStanOwner returned when the user has no stan linked.
var ErrNoStanOwner = errors.New("no stan found for this admin")

// getUserIDFromContext extracts uint user_id set by JWT middleware.
func getUserIDFromContext(c *gin.Context) (uint, bool) {
	v, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	uid, ok := v.(uint)
	return uid, ok
}

// GetStanByCurrentUser finds the Stan associated with the current authenticated user.
// Returns ErrNoStanOwner if not found.
func GetStanByCurrentUser(c *gin.Context) (*app.Stan, error) {
	uid, ok := getUserIDFromContext(c)
	if !ok {
		return nil, errors.New("user_id not found in context")
	}

	var stan app.Stan
	if err := app.DB.Where("user_id = ?", uid).First(&stan).Error; err != nil {
		return nil, ErrNoStanOwner
	}
	return &stan, nil
}

// requireStanOrAbort middleware helper: if no stan, returns 403.
func requireStanOrAbort(c *gin.Context) (*app.Stan, bool) {
	stan, err := GetStanByCurrentUser(c)
	if err != nil {
		if errors.Is(err, ErrNoStanOwner) {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin has no stan"})
			return nil, false
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return nil, false
	}
	return stan, true
}
