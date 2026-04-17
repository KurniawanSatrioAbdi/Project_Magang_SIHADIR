package controllers

import (
	"net/http"
	"strconv"

	"presensi-kominfo/config"
	"presensi-kominfo/models"

	"github.com/gin-gonic/gin"
)

func GetSetting(c *gin.Context) {

	var setting models.Setting
	config.DB.First(&setting)

	if setting.RadiusMeter == 0 {
		setting.RadiusMeter = 300
	}
	if setting.KantorLat == "" {
		setting.KantorLat = "-6.901107073253839"
	}
	if setting.KantorLng == "" {
		setting.KantorLng = "110.62897617699826"
	}

	c.HTML(http.StatusOK, "setting.html", gin.H{
		"data": setting,
	})
}

func UpdateSetting(c *gin.Context) {

	jamMasuk := c.PostForm("jam_masuk")
	jamPulang := c.PostForm("jam_pulang")
	toleransi := c.PostForm("toleransi")
	jamMasukJumat := c.PostForm("jam_masuk_jumat")
	jamPulangJumat := c.PostForm("jam_pulang_jumat")
	toleransiJumat := c.PostForm("toleransi_jumat")
	kantorLat := c.PostForm("kantor_lat")
	kantorLng := c.PostForm("kantor_lng")
	radiusStr := c.PostForm("radius_meter")

	radius, err := strconv.Atoi(radiusStr)
	if err != nil || radius <= 0 {
		radius = 300
	}

	var setting models.Setting
	config.DB.First(&setting)

	setting.JamMasuk = jamMasuk
	setting.JamPulang = jamPulang
	setting.Toleransi = toleransi
	setting.JamMasukJumat = jamMasukJumat
	setting.JamPulangJumat = jamPulangJumat
	setting.ToleransiJumat = toleransiJumat
	setting.KantorLat = kantorLat
	setting.KantorLng = kantorLng
	setting.RadiusMeter = radius

	config.DB.Save(&setting)

	c.Redirect(302, "/admin/setting")
}
