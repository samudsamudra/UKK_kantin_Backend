package siswa

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-pdf/fpdf"

	"github.com/samudsamudra/UKK_kantin/internal/app"
)

// GET /api/siswa/orders/:id/receipt/pdf
// Generate struk / nota dalam bentuk PDF
func SiswaGetOrderReceiptPDF(c *gin.Context) {
	user, ok := getUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	transaksiID := c.Param("id")
	if transaksiID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "transaksi id required"})
		return
	}

	// 🔑 Ambil siswa dari user_id
	var siswa app.Siswa
	if err := app.DB.
		Where("user_id = ?", user.ID).
		First(&siswa).Error; err != nil {

		c.JSON(http.StatusForbidden, gin.H{"error": "user is not siswa"})
		return
	}

	// Ambil transaksi + detail + menu (pastikan milik siswa ini)
	var trx app.Transaksi
	if err := app.DB.
		Preload("Details.Menu").
		Where("public_id = ? AND siswa_id = ?", transaksiID, siswa.ID).
		First(&trx).Error; err != nil {

		c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
		return
	}

	// Ambil stan
	var stan app.Stan
	_ = app.DB.Where("id = ?", trx.StanID).First(&stan)

	// =========================
	// PDF SETUP
	// =========================

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// Judul
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "STRUK PEMBELIAN")
	pdf.Ln(12)

	// Info transaksi
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(0, 7, "Nama Siswa : "+siswa.Nama)
	pdf.Ln(6)
	pdf.Cell(0, 7, "Stan       : "+stan.NamaStan)
	pdf.Ln(6)
	pdf.Cell(0, 7, "Tanggal    : "+formatTanggalStruk(trx.CreatedAt))
	pdf.Ln(6)
	pdf.Cell(0, 7, "Status     : "+string(trx.Status))
	pdf.Ln(10)

	// Header tabel
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(80, 8, "Menu", "1", 0, "", false, 0, "")
	pdf.CellFormat(20, 8, "Qty", "1", 0, "C", false, 0, "")
	pdf.CellFormat(40, 8, "Harga", "1", 0, "R", false, 0, "")
	pdf.CellFormat(40, 8, "Subtotal", "1", 1, "R", false, 0, "")

	pdf.SetFont("Arial", "", 11)

	var total float64

	for _, d := range trx.Details {
		sub := float64(d.Qty) * d.HargaBeli
		total += sub

		pdf.CellFormat(80, 8, d.Menu.NamaMakanan, "1", 0, "", false, 0, "")
		pdf.CellFormat(20, 8, strconv.Itoa(d.Qty), "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 8, formatRupiah(d.HargaBeli), "1", 0, "R", false, 0, "")
		pdf.CellFormat(40, 8, formatRupiah(sub), "1", 1, "R", false, 0, "")
	}

	// Total
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(140, 8, "TOTAL", "1", 0, "R", false, 0, "")
	pdf.CellFormat(40, 8, formatRupiah(total), "1", 1, "R", false, 0, "")

	// Output PDF ke buffer
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate pdf"})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header(
		"Content-Disposition",
		"inline; filename=struk-"+trx.PublicID+".pdf",
	)
	c.Data(http.StatusOK, "application/pdf", buf.Bytes())
}

// =========================
// HELPER (KHUSUS STRUK)
// =========================

// formatTanggalStruk mengembalikan tanggal absolut (BUKAN relative)
// Contoh: "13 Jan 2026 20:39 WIB"
func formatTanggalStruk(t time.Time) string {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err == nil {
		t = t.In(loc)
	}
	return t.Format("02 Jan 2006 15:04") + " WIB"
}

// formatRupiah format angka jadi mata uang sederhana
func formatRupiah(v float64) string {
	return fmt.Sprintf("Rp %.2f", app.Round2(v))
}
