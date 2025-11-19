package auth

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/samudsamudra/UKK_kantin/internal/app"
)

type loginPayload struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type loginResp struct {
	Token     string `json:"token"`
	UserID    string `json:"user_id"` // public id
	Role      string `json:"role"`
	ExpiresAt int64  `json:"expires_at"`
	Username	string `json:"username"`
}

type loginClaims struct {
	UserID   uint   `json:"user_id"`
	PublicID string `json:"public_id"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func getJWTSecret() []byte {
	sec := os.Getenv("JWT_SECRET")
	if sec == "" {
		sec = "dev_jwt_secret_change_me"
	}
	return []byte(sec)
}

// Login handler: POST /api/auth/login
func Login(c *gin.Context) {
	var p loginPayload
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var u app.User
	if err := app.DB.Where("username = ?", p.Username).First(&u).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(p.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// build claims
	exp := time.Now().Add(24 * time.Hour)
	claims := &loginClaims{
		UserID:   u.ID,
		PublicID: u.PublicID,
		Role:     string(u.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString(getJWTSecret())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create token"})
		return
	}

c.JSON(http.StatusOK, loginResp{
	Token:     signed,
	UserID:    u.PublicID,
	Role:      string(u.Role),
	ExpiresAt: exp.Unix(),
	Username:  u.Username,
	})
}
