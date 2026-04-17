package main

import (
	"html/template"

	"presensi-kominfo/config"
	"presensi-kominfo/controllers"
	"presensi-kominfo/models"

	"github.com/gin-gonic/gin"
)

func main() {

	// INIT GIN
	r := gin.Default()

	// KONEKSI DATABASE
	config.ConnectDatabase()

	// Auto migrate tambah kolom file_surat
	config.DB.AutoMigrate(&models.Izin{})

	// LOAD HTML
	r.SetFuncMap(template.FuncMap{
		"addOne": func(i int) int {
			return i + 1
		},
	})

	r.LoadHTMLGlob("templates/*")

	// static file
	r.Static("/uploads", "./uploads")
	r.Static("/assets", "./assets")

	// LOGIN
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "login.html", nil)
	})

	r.POST("/login", controllers.Login)

	// LOGOUT
	r.GET("/logout", func(c *gin.Context) {
		c.SetCookie("user_id", "", -1, "/", "", false, true)
		c.Redirect(302, "/")
	})

	// ADMIN
	admin := r.Group("/admin")
	{
		admin.GET("", controllers.DashboardAdmin)
		admin.GET("/presensi", controllers.DataPresensiAdmin)
		admin.GET("/export-presensi", controllers.ExportPresensiExcel)

		admin.GET("/logbook", controllers.HalamanLogbookAdmin)
		admin.GET("/sertifikat/:id", controllers.DownloadSertifikat)
		admin.GET("/izin", controllers.DataIzinAdmin)

		admin.GET("/setting", controllers.GetSetting)
		admin.POST("/setting", controllers.UpdateSetting)

		admin.GET("/users", controllers.DataUser)
	}

	// aksi admin
	r.GET("/setujui-izin/:id", controllers.SetujuiIzin)
	r.GET("/tolak-izin/:id", controllers.TolakIzin)
	r.GET("/lihat-surat/:id", controllers.LihatSurat) // ← BARU
	r.POST("/tambah-user", controllers.TambahUser)
	r.GET("/hapus-user/:id", controllers.HapusUser)

	// PEGAWAI
	r.GET("/pegawai", func(c *gin.Context) {
		var setting models.Setting
		config.DB.First(&setting)
		if setting.KantorLat == "" {
			setting.KantorLat = "-6.901107073253839"
		}
		if setting.KantorLng == "" {
			setting.KantorLng = "110.62897617699826"
		}
		if setting.RadiusMeter == 0 {
			setting.RadiusMeter = 300
		}
		var user models.User
		userID, _ := c.Cookie("user_id")
		config.DB.First(&user, userID)
		c.HTML(200, "dashboard_pegawai.html", gin.H{"setting": setting, "user": user})
	})

	r.POST("/absen-masuk", controllers.AbsenMasuk)
	r.POST("/absen-pulang", controllers.AbsenPulang)

	r.GET("/riwayat-presensi", controllers.RiwayatPresensi)

	r.GET("/izin", controllers.HalamanIzinPegawai)
	r.POST("/ajukan-izin", controllers.AjukanIzin)

	// MAGANG
	r.GET("/magang", func(c *gin.Context) {
		var setting models.Setting
		config.DB.First(&setting)
		if setting.KantorLat == "" {
			setting.KantorLat = "-6.901107073253839"
		}
		if setting.KantorLng == "" {
			setting.KantorLng = "110.62897617699826"
		}
		if setting.RadiusMeter == 0 {
			setting.RadiusMeter = 300
		}
		var user models.User
		userID, _ := c.Cookie("user_id")
		config.DB.First(&user, userID)
		c.HTML(200, "dashboard_magang.html", gin.H{"setting": setting, "user": user})
	})

	r.GET("/riwayat-presensi-magang", controllers.RiwayatPresensiMagang)

	r.GET("/izin-magang", controllers.HalamanIzinMagang)

	r.GET("/logbook", controllers.HalamanLogbook)
	r.POST("/tambah-logbook", controllers.TambahLogbook)
	r.GET("/edit-logbook/:id", controllers.EditLogbook)
	r.POST("/update-logbook/:id", controllers.UpdateLogbook)
	r.GET("/hapus-logbook/:id", controllers.HapusLogbook)

	r.GET("/download-logbook", controllers.DownloadLogbook)

	// RUN SERVER
	r.Run(":8080")
}
