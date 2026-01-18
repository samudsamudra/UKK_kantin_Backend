package admin

import (
	"bufio"
	"encoding/csv"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/samudsamudra/UKK_kantin/internal/app"
)

const (
	siswaDomain      = "@smk_tlkm-mlg.com"
	defaultPassword  = "password123"
)

func AdminImportSiswa(c *gin.Context) {
	// defense-in-depth
	roleAny, ok := c.Get("role")
	if !ok || roleAny.(string) != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "super admin only"})
		return
	}

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	// detect CSV / TSV
	reader := csv.NewReader(bufio.NewReader(file))
	reader.TrimLeadingSpace = true

	// default CSV delimiter
	reader.Comma = ','

	// peek first line to detect TSV
	peek, err := reader.Read()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file"})
		return
	}
	if len(peek) == 1 && strings.Contains(peek[0], "\t") {
		reader.Comma = '\t'
	}

	// reset reader
	file.Seek(0, io.SeekStart)
	reader = csv.NewReader(bufio.NewReader(file))
	if len(peek) == 1 && strings.Contains(peek[0], "\t") {
		reader.Comma = '\t'
	}

	header, err := reader.Read()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid header"})
		return
	}

	colIndex := -1
	for i, h := range header {
		if strings.ToLower(strings.TrimSpace(h)) == "nama_lengkap" {
			colIndex = i
			break
		}
	}

	if colIndex == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "column 'nama_lengkap' not found"})
		return
	}

	success := 0
	skipped := 0
	errors := []string{}

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil || colIndex >= len(row) {
			skipped++
			continue
		}

		nama := strings.TrimSpace(row[colIndex])
		if nama == "" {
			skipped++
			continue
		}

		parts := strings.Fields(strings.ToLower(nama))
		if len(parts) < 2 {
			skipped++
			continue
		}

		email := parts[0] + "_" + parts[1] + siswaDomain

		// check existing
		var ex app.User
		if err := app.DB.Where("email = ?", email).First(&ex).Error; err == nil {
			skipped++
			continue
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
		if err != nil {
			errors = append(errors, email)
			continue
		}

		user := app.User{
			Email:              email,
			PasswordHash:       string(hash),
			Role:               "siswa",
			MustChangePassword: true,
		}

		if err := app.DB.Create(&user).Error; err != nil {
			errors = append(errors, email)
			continue
		}

		// optional: create siswa profile
		siswa := app.Siswa{
			Nama: nama,
			UserID:      user.ID,
		}
		_ = app.DB.Create(&siswa).Error

		success++
	}

	c.JSON(http.StatusOK, gin.H{
		"created": success,
		"skipped": skipped,
		"errors":  errors,
		"default_password": defaultPassword,
	})
}
