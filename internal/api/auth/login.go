package auth

import (
	"net/http"
	"os"
	"strings"
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
	Email    string `json:"email" binding:"required"` // ⬅️ FIX DI SINI
	Password string `json:"password" binding:"required"`
}

type loginResp struct {
	Token              string `json:"token"`
	UserID             string `json:"user_id"`
	Role               string `json:"role"`
	MustChangePassword bool   `json:"must_change_password"`
	ExpiresAt          int64  `json:"expires_at"`
	Email              string `json:"email"`
}

type loginClaims struct {
	UserID   uint   `json:"user_id"`
	PublicID string `json:"public_id"`
	Role     string `json:"role"` // info only
	jwt.RegisteredClaims
}

//
// =========================
// Helpers
// =========================
//

func getJWTSecret() []byte {
	sec := os.Getenv("JWT_SECRET")
	if sec == "" {
		sec = "dev_jwt_secret_change_me"
	}
	return []byte(sec)
}

// validasi email sekolah (khusus siswa)
func isSchoolEmail(email string) bool {
	email = strings.ToLower(strings.TrimSpace(email))
	return strings.HasSuffix(email, "@smk_tlkm-mlg.com")
}

//
// =========================
// Login Handler
// =========================
//

func Login(c *gin.Context) {
	var p loginPayload
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "data tidak valid",
			"message": "email dan password wajib diisi",
		})
		return
	}

	// Cari user
	var u app.User
	if err := app.DB.
		Where("email = ?", p.Email).
		First(&u).Error; err != nil {

		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "login gagal",
			"message": "email atau password salah",
		})
		return
	}

	// =========================
	// VALIDASI EMAIL SISWA
	// =========================
	if u.Role == app.RoleSiswa {
		if !isSchoolEmail(u.Email) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "email tidak valid",
				"message": "gunakan email resmi sekolah",
			})
			return
		}
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword(
		[]byte(u.PasswordHash),
		[]byte(p.Password),
	); err != nil {

		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "login gagal",
			"message": "email atau password salah",
		})
		return
	}

	// =========================
	// BUILD JWT
	// =========================
	exp := time.Now().Add(24 * time.Hour)

	claims := &loginClaims{
		UserID:   u.ID,
		PublicID: u.PublicID,
		Role:     string(u.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   u.PublicID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(getJWTSecret())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server error",
			"message": "gagal membuat token",
		})
		return
	}

	c.JSON(http.StatusOK, loginResp{
		Token:              signed,
		UserID:             u.PublicID,
		Role:               string(u.Role),
		MustChangePassword: u.MustChangePassword,
		ExpiresAt:          exp.Unix(),
		Email:              u.Email,
	})
}