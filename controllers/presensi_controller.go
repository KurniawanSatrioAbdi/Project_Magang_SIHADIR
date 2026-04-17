package controllers

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"math"
	"presensi-kominfo/config"
	"presensi-kominfo/models"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// ABSEN MASUK

func AbsenMasuk(c *gin.Context) {

	cookie, err := c.Cookie("user_id")
	if err != nil {
		c.String(400, "User tidak ditemukan")
		return
	}

	userID, _ := strconv.Atoi(cookie)

	latitude := c.PostForm("latitude")
	longitude := c.PostForm("longitude")
	foto := c.PostForm("foto")

	now := time.Now()

	tanggal := now.Format("2006-01-02")
	jamMasuk := now.Format("15:04:05")

	status := "Hadir"

	var setting models.Setting
	config.DB.First(&setting)

	// Cek apakah hari ini Jumat (weekday == 5)
	isJumat := now.Weekday() == time.Friday

	// Pilih jam masuk sesuai hari
	jamMasukStr := setting.JamMasuk
	toleransiStr := setting.Toleransi
	if isJumat && setting.JamMasukJumat != "" {
		jamMasukStr = setting.JamMasukJumat
	}
	if isJumat && setting.ToleransiJumat != "" {
		toleransiStr = setting.ToleransiJumat
	}

	jamMasukDB := now
	if jamMasukStr != "" {
		jamSetting, err := time.Parse("15:04", jamMasukStr)
		if err == nil {
			jamMasukDB = time.Date(now.Year(), now.Month(), now.Day(),
				jamSetting.Hour(), jamSetting.Minute(), 0, 0, now.Location())
		}
	}

	jamToleransi := jamMasukDB
	if toleransiStr != "" {
		toleransiSetting, err := time.Parse("15:04", toleransiStr)
		if err == nil {
			jamToleransi = time.Date(now.Year(), now.Month(), now.Day(),
				toleransiSetting.Hour(), toleransiSetting.Minute(), 0, 0, now.Location())
		}
	}

	if now.After(jamToleransi) {
		status = "Terlambat"
	}

	// Validasi lokasi dari setting
	var kantorLat, kantorLng float64
	var radiusMeter float64 = 300
	if setting.KantorLat != "" && setting.KantorLng != "" {
		fmt.Sscanf(setting.KantorLat, "%f", &kantorLat)
		fmt.Sscanf(setting.KantorLng, "%f", &kantorLng)
	} else {
		kantorLat = -6.901107073253839
		kantorLng = 110.62897617699826
	}
	if setting.RadiusMeter > 0 {
		radiusMeter = float64(setting.RadiusMeter)
	}

	// Hitung jarak
	if latitude != "" && longitude != "" {
		var userLat, userLng float64
		fmt.Sscanf(latitude, "%f", &userLat)
		fmt.Sscanf(longitude, "%f", &userLng)
		jarak := hitungJarakPresensi(userLat, userLng, kantorLat, kantorLng)
		if jarak > radiusMeter {
			c.String(400, fmt.Sprintf("Anda berada di luar zona presensi (%.0f meter dari kantor)", jarak))
			return
		}
	}

	var cek models.Presensi
	config.DB.Where("user_id=? AND tanggal=?", userID, tanggal).First(&cek)

	if cek.ID != 0 {
		c.String(400, "Sudah absen hari ini")
		return
	}

	namaFile := ""
	if foto != "" {
		parts := strings.Split(foto, ",")
		if len(parts) == 2 {
			data, _ := base64.StdEncoding.DecodeString(parts[1])
			namaFile = strconv.FormatInt(time.Now().Unix(), 10) + ".png"
			os.MkdirAll("uploads", os.ModePerm)
			os.WriteFile("uploads/"+namaFile, data, 0644)
		}
	}

	presensi := models.Presensi{
		UserID:    userID,
		Tanggal:   tanggal,
		JamMasuk:  jamMasuk,
		Status:    status,
		Latitude:  latitude,
		Longitude: longitude,
		Foto:      namaFile,
	}

	config.DB.Create(&presensi)

	c.String(200, "Presensi masuk berhasil")
}

// ABSEN PULANG

func AbsenPulang(c *gin.Context) {

	cookie, _ := c.Cookie("user_id")
	userID, _ := strconv.Atoi(cookie)

	now := time.Now()

	tanggal := now.Format("2006-01-02")
	jamPulang := now.Format("15:04:05")

	var setting models.Setting
	config.DB.First(&setting)

	// Cek apakah hari ini Jumat
	isJumatPulang := now.Weekday() == time.Friday
	jamPulangStr := setting.JamPulang
	if isJumatPulang && setting.JamPulangJumat != "" {
		jamPulangStr = setting.JamPulangJumat
	}

	jamPulangDB := now
	if jamPulangStr != "" {
		jamSetting, err := time.Parse("15:04", jamPulangStr)
		if err == nil {
			jamPulangDB = time.Date(now.Year(), now.Month(), now.Day(),
				jamSetting.Hour(), jamSetting.Minute(), 0, 0, now.Location())
		}
	}

	if now.Before(jamPulangDB) {
		c.String(400, "Belum waktunya pulang")
		return
	}

	var presensi models.Presensi
	config.DB.Where("user_id=? AND tanggal=?", userID, tanggal).First(&presensi)

	if presensi.ID == 0 {
		c.String(400, "Belum absen masuk")
		return
	}

	if presensi.JamPulang != "" {
		c.String(400, "Sudah absen pulang")
		return
	}

	presensi.JamPulang = jamPulang
	config.DB.Save(&presensi)

	c.String(200, "Absen pulang berhasil")
}

// RIWAYAT PEGAWAI

func RiwayatPresensi(c *gin.Context) {

	cookie, _ := c.Cookie("user_id")
	userID, _ := strconv.Atoi(cookie)

	fmt.Println("=== RiwayatPresensi ===")
	fmt.Println("cookie user_id:", cookie, "| parsed:", userID)

	var presensi []models.Presensi
	config.DB.Where("user_id=?", userID).
		Order("tanggal DESC").
		Find(&presensi)

	var izin []models.Izin
	config.DB.Where("user_id=?", userID).
		Order("tanggal DESC").
		Find(&izin)

	fmt.Println("Presensi count:", len(presensi), "| Izin count:", len(izin))
	fmt.Println("======================")

	c.HTML(200, "riwayat_presensi.html", gin.H{
		"data": presensi,
		"izin": izin,
		"role": "pegawai",
	})
}

// RIWAYAT MAGANG

func RiwayatPresensiMagang(c *gin.Context) {

	cookie, _ := c.Cookie("user_id")
	userID, _ := strconv.Atoi(cookie)

	fmt.Println("=== RiwayatPresensiMagang ===")
	fmt.Println("cookie user_id:", cookie, "| parsed:", userID)

	var presensi []models.Presensi
	config.DB.Where("user_id=?", userID).
		Order("tanggal DESC").
		Find(&presensi)

	var izin []models.Izin
	config.DB.Where("user_id=?", userID).
		Order("tanggal DESC").
		Find(&izin)

	fmt.Println("Presensi count:", len(presensi), "| Izin count:", len(izin))
	fmt.Println("=============================")

	c.HTML(200, "riwayat_presensi.html", gin.H{
		"data": presensi,
		"izin": izin,
		"role": "magang",
	})
}

// DASHBOARD ADMIN

func DashboardAdmin(c *gin.Context) {

	type Lokasi struct {
		Nama      string
		Latitude  string
		Longitude string
		Status    string
	}

	today := time.Now().Format("2006-01-02")

	var lokasi []Lokasi
	config.DB.Table("presensi").
		Select("users.nama, presensi.latitude, presensi.longitude, presensi.status").
		Joins("LEFT JOIN users ON users.id = presensi.user_id").
		Where("presensi.tanggal LIKE ?", today+"%").
		Scan(&lokasi)

	var total, hadir, terlambat int64
	config.DB.Model(&models.Presensi{}).Where("tanggal LIKE ?", today+"%").Count(&total)
	config.DB.Model(&models.Presensi{}).Where("tanggal LIKE ? AND status = ?", today+"%", "Hadir").Count(&hadir)
	config.DB.Model(&models.Presensi{}).Where("tanggal LIKE ? AND status = ?", today+"%", "Terlambat").Count(&terlambat)

	c.HTML(200, "dashboard_admin.html", gin.H{
		"total":     total,
		"hadir":     hadir,
		"terlambat": terlambat,
		"lokasi":    lokasi,
	})
}

// DATA PRESENSI ADMIN

func DataPresensiAdmin(c *gin.Context) {

	type Result struct {
		Nama      string
		Tanggal   string
		JamMasuk  string
		JamPulang string
		Status    string
		Latitude  string
		Longitude string
		Foto      string
	}

	// filter tanggal — kalau kosong tampilkan semua
	date := c.Query("date")

	var data []Result
	db := config.DB.Table("presensi").
		Select("users.nama, presensi.tanggal, presensi.jam_masuk, presensi.jam_pulang, presensi.status, presensi.latitude, presensi.longitude, presensi.foto").
		Joins("LEFT JOIN users ON users.id = presensi.user_id").
		Order("presensi.tanggal DESC, presensi.jam_masuk ASC")

	if date != "" {
		db = db.Where("presensi.tanggal LIKE ?", date+"%")
	}
	db.Scan(&data)

	var total, hadir, terlambat, izin int64
	if date != "" {
		config.DB.Model(&models.Presensi{}).Where("tanggal LIKE ?", date+"%").Count(&total)
		config.DB.Model(&models.Presensi{}).Where("tanggal LIKE ? AND status = ?", date+"%", "Hadir").Count(&hadir)
		config.DB.Model(&models.Presensi{}).Where("tanggal LIKE ? AND status = ?", date+"%", "Terlambat").Count(&terlambat)
		config.DB.Model(&models.Izin{}).Where("tanggal LIKE ? AND status = ?", date+"%", "Disetujui").Count(&izin)
	} else {
		config.DB.Model(&models.Presensi{}).Count(&total)
		config.DB.Model(&models.Presensi{}).Where("status = ?", "Hadir").Count(&hadir)
		config.DB.Model(&models.Presensi{}).Where("status = ?", "Terlambat").Count(&terlambat)
		config.DB.Model(&models.Izin{}).Where("status = ?", "Disetujui").Count(&izin)
	}

	c.HTML(200, "presensi_admin.html", gin.H{
		"data":      data,
		"date":      date,
		"total":     total,
		"hadir":     hadir,
		"terlambat": terlambat,
		"izin":      izin,
	})
}

////////////////////////////////////////////////////
// EXPORT EXCEL
////////////////////////////////////////////////////

func ExportPresensiExcel(c *gin.Context) {

	type Result struct {
		Nama      string
		Tanggal   string
		JamMasuk  string
		JamPulang string
		Status    string
	}

	var data []Result

	config.DB.Table("presensi").
		Select("users.nama, presensi.tanggal, presensi.jam_masuk, presensi.jam_pulang, presensi.status").
		Joins("LEFT JOIN users ON users.id = presensi.user_id").
		Scan(&data)

	file := excelize.NewFile()
	sheet := "Sheet1"

	file.SetCellValue(sheet, "A1", "Nama")
	file.SetCellValue(sheet, "B1", "Tanggal")
	file.SetCellValue(sheet, "C1", "Jam Masuk")
	file.SetCellValue(sheet, "D1", "Jam Pulang")
	file.SetCellValue(sheet, "E1", "Status")

	for i, v := range data {
		row := i + 2
		file.SetCellValue(sheet, "A"+strconv.Itoa(row), v.Nama)
		file.SetCellValue(sheet, "B"+strconv.Itoa(row), v.Tanggal)
		file.SetCellValue(sheet, "C"+strconv.Itoa(row), v.JamMasuk)
		file.SetCellValue(sheet, "D"+strconv.Itoa(row), v.JamPulang)
		file.SetCellValue(sheet, "E"+strconv.Itoa(row), v.Status)
	}

	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename=laporan_presensi.xlsx")

	file.Write(c.Writer)
}

func hitungJarakPresensi(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371000
	dLat := (lat2 - lat1) * 3.14159265358979 / 180
	dLng := (lng2 - lng1) * 3.14159265358979 / 180
	a := sinHalf(dLat)*sinHalf(dLat) +
		cosRad(lat1)*cosRad(lat2)*sinHalf(dLng)*sinHalf(dLng)
	return R * 2 * atanSqrt(a)
}

func sinHalf(x float64) float64 {
	return math.Sin(x / 2)
}

func cosRad(deg float64) float64 {
	return math.Cos(deg * 3.14159265358979 / 180)
}

func atanSqrt(a float64) float64 {
	return math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}
