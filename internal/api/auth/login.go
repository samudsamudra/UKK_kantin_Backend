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

//
// =========================
// Payload & Response
// =========================
//

type loginPayload struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type loginResp struct {
	Token              string `json:"token"`
	UserID             string `json:"user_id"` // public id
	Role               string `json:"role"`
	MustChangePassword bool   `json:"must_change_password"`
	ExpiresAt          int64  `json:"expires_at"`
	Email              string `json:"email"`
}

type loginClaims struct {
	UserID   uint   `json:"user_id"`
	PublicID string `json:"public_id"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

//
// =========================
// JWT Helper
// =========================
//

func getJWTSecret() []byte {
	sec := os.Getenv("JWT_SECRET")
	if sec == "" {
		sec = "dev_jwt_secret_change_me"
	}
	return []byte(sec)
}

//
// =========================
// Login Handler
// =========================
// POST /api/auth/login
// =========================
//

func Login(c *gin.Context) {
	var p loginPayload
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Cari user berdasarkan email
	var u app.User
	if err := app.DB.
		Where("email = ?", p.Email).
		First(&u).Error; err != nil {

		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword(
		[]byte(u.PasswordHash),
		[]byte(p.Password),
	); err != nil {

		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Build JWT
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

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(getJWTSecret())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create token"})
		return
	}

	// Response
	c.JSON(http.StatusOK, loginResp{
		Token:              signed,
		UserID:             u.PublicID,
		Role:               string(u.Role),
		MustChangePassword: u.MustChangePassword,
		ExpiresAt:          exp.Unix(),
		Email:              u.Email,
	})
}
