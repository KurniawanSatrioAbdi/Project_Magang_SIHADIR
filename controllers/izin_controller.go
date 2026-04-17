package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"presensi-kominfo/config"
	"presensi-kominfo/models"

	"github.com/gin-gonic/gin"
)

// Helper potong tanggal
func potongTanggalIzin(t string) string {
	if len(t) > 10 {
		return t[:10]
	}
	return t
}

// HALAMAN IZIN PEGAWAI

func HalamanIzinPegawai(c *gin.Context) {
	cookie, _ := c.Cookie("user_id")
	userID, _ := strconv.Atoi(cookie)

	var izin []models.Izin
	config.DB.Where("user_id = ?", userID).Find(&izin)

	for i := range izin {
		izin[i].Tanggal = potongTanggalIzin(izin[i].Tanggal)
	}

	c.HTML(http.StatusOK, "izin_pegawai.html", gin.H{
		"data": izin,
	})
}

// HALAMAN IZIN MAGANG

func HalamanIzinMagang(c *gin.Context) {
	cookie, _ := c.Cookie("user_id")
	userID, _ := strconv.Atoi(cookie)

	var izin []models.Izin
	config.DB.Where("user_id = ?", userID).Find(&izin)

	for i := range izin {
		izin[i].Tanggal = potongTanggalIzin(izin[i].Tanggal)
	}

	var user models.User
	config.DB.First(&user, userID)

	c.HTML(http.StatusOK, "izin_magang.html", gin.H{
		"data": izin,
		"user": user,
	})
}

// AJUKAN IZIN

func AjukanIzin(c *gin.Context) {
	cookie, _ := c.Cookie("user_id")
	userID, _ := strconv.Atoi(cookie)

	if userID == 0 {
		c.String(http.StatusBadRequest, "User tidak ditemukan")
		return
	}

	tanggal := c.PostForm("tanggal")
	jenis := c.PostForm("jenis")
	keterangan := c.PostForm("keterangan")

	// ── Upload surat (opsional) ──
	filePath := ""
	file, err := c.FormFile("file_surat")
	if err == nil && file != nil {
		ext := strings.ToLower(filepath.Ext(file.Filename))
		allowed := map[string]bool{".pdf": true, ".jpg": true, ".jpeg": true, ".png": true}

		if !allowed[ext] {
			c.String(http.StatusBadRequest, "Format file tidak didukung. Gunakan PDF, JPG, atau PNG.")
			return
		}
		if file.Size > 5*1024*1024 {
			c.String(http.StatusBadRequest, "Ukuran file maksimal 5 MB.")
			return
		}

		uploadDir := "./uploads/surat"
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			c.String(http.StatusInternalServerError, "Gagal membuat folder upload.")
			return
		}

		fileName := fmt.Sprintf("surat_%d_%d%s", userID, time.Now().UnixNano(), ext)
		filePath = "uploads/surat/" + fileName

		if err := c.SaveUploadedFile(file, "./"+filePath); err != nil {
			c.String(http.StatusInternalServerError, "Gagal menyimpan file.")
			return
		}
	}

	data := models.Izin{
		UserID:     uint(userID),
		Tanggal:    tanggal,
		Jenis:      jenis,
		Keterangan: keterangan,
		FileSurat:  filePath,
		Status:     "Pending",
	}

	config.DB.Create(&data)

	referer := c.Request.Referer()
	if referer == "" {
		referer = "/izin"
	}
	c.Redirect(http.StatusFound, referer)
}

// ADMIN SETUJUI IZIN

func SetujuiIzin(c *gin.Context) {
	id := c.Param("id")

	var izin models.Izin
	config.DB.First(&izin, id)
	izin.Status = "Disetujui"
	config.DB.Save(&izin)

	c.Redirect(http.StatusFound, "/admin/izin")
}

// ADMIN TOLAK IZIN

func TolakIzin(c *gin.Context) {
	id := c.Param("id")

	var izin models.Izin
	config.DB.First(&izin, id)
	izin.Status = "Ditolak"
	config.DB.Save(&izin)

	c.Redirect(http.StatusFound, "/admin/izin")
}

// DATA IZIN UNTUK ADMIN

func DataIzinAdmin(c *gin.Context) {
	type Result struct {
		ID         uint
		Nama       string
		Jenis      string
		Tanggal    string
		Keterangan string
		FileSurat  string // ← tambahan
		Status     string
	}

	var data []Result

	config.DB.Table("izin").
		Select("izin.id, users.nama, izin.jenis, izin.tanggal, izin.keterangan, izin.file_surat, izin.status").
		Joins("left join users on users.id = izin.user_id").
		Order("izin.id DESC").
		Scan(&data)

	for i := range data {
		data[i].Tanggal = potongTanggalIzin(data[i].Tanggal)
	}

	var total, pending, disetujui, ditolak int64
	config.DB.Model(&models.Izin{}).Count(&total)
	config.DB.Model(&models.Izin{}).Where("status = ?", "Pending").Count(&pending)
	config.DB.Model(&models.Izin{}).Where("status = ?", "Disetujui").Count(&disetujui)
	config.DB.Model(&models.Izin{}).Where("status = ?", "Ditolak").Count(&ditolak)

	c.HTML(http.StatusOK, "izin_admin.html", gin.H{
		"data":      data,
		"total":     total,
		"pending":   pending,
		"disetujui": disetujui,
		"ditolak":   ditolak,
	})
}

// LIHAT SURAT (tab baru)

func LihatSurat(c *gin.Context) {
	id := c.Param("id")

	var izin models.Izin
	if err := config.DB.First(&izin, id).Error; err != nil {
		c.String(http.StatusNotFound, "Data izin tidak ditemukan.")
		return
	}

	if izin.FileSurat == "" {
		c.String(http.StatusNotFound, "Tidak ada surat yang dilampirkan.")
		return
	}

	absPath := "./" + izin.FileSurat
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		c.String(http.StatusNotFound, "File surat tidak ditemukan di server.")
		return
	}

	ext := strings.ToLower(filepath.Ext(izin.FileSurat))
	contentType := "application/octet-stream"
	switch ext {
	case ".pdf":
		contentType = "application/pdf"
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	}

	// inline = terbuka langsung di browser, bukan download
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"surat_%s%s\"", id, ext))
	c.Header("Content-Type", contentType)
	c.File(absPath)
}
