package controllers

import (
	"net/http"
	"strconv"

	"fmt"
	"presensi-kominfo/config"
	"presensi-kominfo/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
)

// Helper potong tanggal jadi YYYY-MM-DD
func potongTanggal(t string) string {
	if len(t) > 10 {
		return t[:10]
	}
	return t
}

func HalamanLogbook(c *gin.Context) {

	cookie, _ := c.Cookie("user_id")
	userID, _ := strconv.Atoi(cookie)

	var logbook []models.Logbook

	config.DB.Where("user_id = ?", userID).Find(&logbook)

	// Potong format tanggal jadi YYYY-MM-DD
	for i := range logbook {
		logbook[i].Tanggal = potongTanggal(logbook[i].Tanggal)
	}

	var user models.User
	config.DB.First(&user, userID)

	c.HTML(http.StatusOK, "logbook.html", gin.H{
		"data": logbook,
		"user": user,
	})
}

func HalamanLogbookAdmin(c *gin.Context) {

	type Result struct {
		ID        uint
		Nama      string
		Tanggal   string
		Kegiatan  string
		Deskripsi string
	}

	var data []Result

	config.DB.Table("logbook").
		Select("logbook.id, users.nama, logbook.tanggal, logbook.kegiatan, logbook.deskripsi").
		Joins("left join users on users.id = logbook.user_id").
		Scan(&data)

	// Potong format tanggal jadi YYYY-MM-DD
	for i := range data {
		data[i].Tanggal = potongTanggal(data[i].Tanggal)
	}

	// Hitung total peserta magang dari database
	var totalMagang int64
	config.DB.Model(&models.User{}).Where("role = ?", "magang").Count(&totalMagang)

	// Ambil list peserta magang untuk dropdown sertifikat
	var pesertaList []models.User
	config.DB.Where("role = ?", "magang").Find(&pesertaList)
	fmt.Println("=== DEBUG pesertaList ===")
	fmt.Println("Count:", len(pesertaList))
	for _, p := range pesertaList {
		fmt.Println("  -", p.ID, p.Nama, p.Role)
	}
	fmt.Println("========================")

	c.HTML(http.StatusOK, "logbook_admin.html", gin.H{
		"data":        data,
		"totalMagang": totalMagang,
		"pesertaList": pesertaList,
	})
}

func TambahLogbook(c *gin.Context) {

	cookie, _ := c.Cookie("user_id")
	userID, _ := strconv.Atoi(cookie)

	tanggal := c.PostForm("tanggal")
	kegiatan := c.PostForm("kegiatan")
	deskripsi := c.PostForm("deskripsi")

	log := models.Logbook{
		UserID:    uint(userID),
		Tanggal:   tanggal,
		Kegiatan:  kegiatan,
		Deskripsi: deskripsi,
	}

	config.DB.Create(&log)

	c.Redirect(http.StatusFound, "/logbook")
}

func EditLogbook(c *gin.Context) {

	id := c.Param("id")

	var logbook models.Logbook

	config.DB.First(&logbook, id)

	// Potong format tanggal jadi YYYY-MM-DD
	logbook.Tanggal = potongTanggal(logbook.Tanggal)

	c.HTML(http.StatusOK, "edit_logbook.html", gin.H{
		"logbook": logbook,
	})
}

func UpdateLogbook(c *gin.Context) {

	id := c.Param("id")

	var logbook models.Logbook

	config.DB.First(&logbook, id)

	logbook.Tanggal = c.PostForm("tanggal")
	logbook.Kegiatan = c.PostForm("kegiatan")
	logbook.Deskripsi = c.PostForm("deskripsi")

	config.DB.Save(&logbook)

	c.Redirect(http.StatusFound, "/logbook")
}

func HapusLogbook(c *gin.Context) {

	id := c.Param("id")

	config.DB.Delete(&models.Logbook{}, id)

	c.Redirect(http.StatusFound, "/admin/logbook")
}

func DownloadLogbook(c *gin.Context) {

	cookie, _ := c.Cookie("user_id")
	userID, _ := strconv.Atoi(cookie)

	var logbook []models.Logbook
	config.DB.Where("user_id = ?", userID).Find(&logbook)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// ── JUDUL ──
	pdf.SetFont("Arial", "B", 16)
	pdf.SetTextColor(26, 32, 53)
	pdf.CellFormat(0, 10, "Logbook Kegiatan Magang", "", 1, "C", false, 0, "")
	pdf.Ln(2)

	// Garis bawah judul
	pdf.SetDrawColor(59, 111, 240)
	pdf.SetLineWidth(0.8)
	pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
	pdf.Ln(6)

	// ── HEADER TABEL ──
	colNo := 10.0
	colTanggal := 28.0
	colKegiatan := 52.0
	colDeskripsi := 90.0
	rowH := 8.0

	pdf.SetFillColor(59, 111, 240)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 9)
	pdf.SetDrawColor(200, 200, 200)
	pdf.SetLineWidth(0.3)

	pdf.CellFormat(colNo, rowH, "No", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colTanggal, rowH, "Tanggal", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colKegiatan, rowH, "Kegiatan", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colDeskripsi, rowH, "Deskripsi", "1", 1, "C", true, 0, "")

	// ── ISI TABEL ──
	pdf.SetTextColor(26, 32, 53)
	pdf.SetFont("Arial", "", 8.5)

	for i, log := range logbook {
		tanggal := potongTanggal(log.Tanggal)
		kegiatan := log.Kegiatan
		deskripsi := log.Deskripsi

		if i%2 == 0 {
			pdf.SetFillColor(245, 248, 252)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}

		x := pdf.GetX()
		y := pdf.GetY()

		lines := pdf.SplitLines([]byte(deskripsi), colDeskripsi-2)
		lineCount := len(lines)
		if lineCount < 1 {
			lineCount = 1
		}
		cellH := float64(lineCount) * 5.5
		if cellH < rowH {
			cellH = rowH
		}

		pdf.SetXY(x, y)
		pdf.CellFormat(colNo, cellH, strconv.Itoa(i+1), "1", 0, "C", true, 0, "")
		pdf.CellFormat(colTanggal, cellH, tanggal, "1", 0, "C", true, 0, "")
		pdf.MultiCell(colKegiatan, 5.5, kegiatan, "1", "L", true)
		pdf.SetXY(x+colNo+colTanggal+colKegiatan, y)
		pdf.MultiCell(colDeskripsi, 5.5, deskripsi, "1", "L", true)

		newY := pdf.GetY()
		expectedY := y + cellH
		if newY < expectedY {
			pdf.SetY(expectedY)
		}
	}

	// ── FOOTER ──
	pdf.Ln(8)
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(107, 122, 153)
	pdf.CellFormat(0, 6, "Dokumen ini digenerate otomatis oleh Sistem Presensi", "", 1, "C", false, 0, "")

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=logbook.pdf")

	err := pdf.Output(c.Writer)
	if err != nil {
		c.String(500, err.Error())
		return
	}
}

func DownloadSertifikat(c *gin.Context) {

	userID := c.Param("id")
	mentor := c.Query("mentor")
	kabid := c.Query("kabid")

	var user models.User
	config.DB.First(&user, userID)

	type LogInfo struct {
		UserID  uint
		Tanggal string
	}
	var earliestLog LogInfo
	config.DB.Table("logbook").
		Select("user_id, tanggal").
		Where("user_id = ?", userID).
		Order("tanggal ASC").
		Limit(1).
		Scan(&earliestLog)

	tanggalMulai := potongTanggal(earliestLog.Tanggal)
	tanggalSelesai := time.Now().Format("2006-01-02")

	bulanIndo := map[string]string{
		"01": "Januari", "02": "Februari", "03": "Maret", "04": "April",
		"05": "Mei", "06": "Juni", "07": "Juli", "08": "Agustus",
		"09": "September", "10": "Oktober", "11": "November", "12": "Desember",
	}
	formatTanggal := func(t string) string {
		if len(t) < 10 {
			return t
		}
		dd := t[8:10]
		mm := t[5:7]
		yyyy := t[0:4]
		return dd + " " + bulanIndo[mm] + " " + yyyy
	}

	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetMargins(0, 0, 0)
	pdf.AddPage()

	w := 297.0
	h := 210.0

	pdf.SetFillColor(224, 229, 255)
	pdf.Rect(0, 0, w, h, "F")

	pdf.SetFillColor(59, 111, 240)
	pdf.Rect(w-40, 0, 40, 40, "F")
	pdf.SetFillColor(100, 140, 255)
	pdf.Rect(w-28, 8, 20, 20, "F")

	pdf.SetFillColor(59, 111, 240)
	pdf.Rect(0, h-40, 40, 40, "F")
	pdf.SetFillColor(100, 140, 255)
	pdf.Rect(8, h-28, 20, 20, "F")

	pdf.SetFillColor(255, 255, 255)
	pdf.RoundedRect(18, 14, w-36, h-28, 6, "1234", "F")

	pdf.SetDrawColor(59, 111, 240)
	pdf.SetLineWidth(0.8)
	pdf.RoundedRect(22, 18, w-44, h-36, 4, "1234", "D")

	pdf.SetFont("Arial", "B", 38)
	pdf.SetTextColor(59, 111, 240)
	pdf.SetXY(0, 32)
	pdf.CellFormat(w, 14, "SERTIFIKAT", "", 1, "C", false, 0, "")

	titleW := 100.0
	titleX := (w - titleW) / 2
	pdf.SetDrawColor(100, 80, 200)
	pdf.SetLineWidth(1.2)
	pdf.Rect(titleX, 28, titleW, 18, "D")

	pdf.SetFont("Arial", "B", 13)
	pdf.SetTextColor(40, 40, 100)
	pdf.SetXY(0, 48)
	pdf.CellFormat(w, 8, "KELULUSAN MAGANG", "", 1, "C", false, 0, "")

	pdf.SetFont("Arial", "", 11)
	pdf.SetTextColor(80, 80, 120)
	pdf.SetXY(0, 64)
	pdf.CellFormat(w, 7, "Sertifikat ini diberikan kepada", "", 1, "C", false, 0, "")

	pdf.SetFont("Arial", "BI", 30)
	pdf.SetTextColor(59, 111, 240)
	pdf.SetXY(0, 74)
	pdf.CellFormat(w, 16, user.Nama, "", 1, "C", false, 0, "")

	pdf.SetFont("Arial", "", 10.5)
	pdf.SetTextColor(40, 40, 80)
	keterangan := "Telah menyelesaikan program magang di Kantor Dinas Komunikasi dan Informatika"
	keterangan2 := "Kabupaten Demak pada tanggal " + formatTanggal(tanggalMulai) + " - " + formatTanggal(tanggalSelesai)
	pdf.SetXY(0, 93)
	pdf.CellFormat(w, 6, keterangan, "", 1, "C", false, 0, "")
	pdf.SetXY(0, 100)
	pdf.CellFormat(w, 6, keterangan2, "", 1, "C", false, 0, "")

	lineY := 118.0
	mentorX := w/2 - 80.0
	pdf.SetDrawColor(59, 111, 240)
	pdf.SetLineWidth(0.5)
	pdf.Line(mentorX, lineY, mentorX+60, lineY)
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(26, 32, 53)
	pdf.SetXY(mentorX-10, lineY+3)
	pdf.CellFormat(80, 6, "Mentor", "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(mentorX-10, lineY+10)
	pdf.CellFormat(80, 6, mentor, "", 1, "C", false, 0, "")

	kabidX := w/2 + 20.0
	pdf.Line(kabidX, lineY, kabidX+60, lineY)
	pdf.SetFont("Arial", "B", 10)
	pdf.SetXY(kabidX-10, lineY+3)
	pdf.CellFormat(80, 6, "Kepala Bidang", "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(kabidX-10, lineY+10)
	pdf.CellFormat(80, 6, kabid, "", 1, "C", false, 0, "")

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=sertifikat_"+user.Nama+".pdf")

	err := pdf.Output(c.Writer)
	if err != nil {
		c.String(500, err.Error())
	}
}
