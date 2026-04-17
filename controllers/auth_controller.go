package controllers

import (
	"fmt"
	"net/http"
	"presensi-kominfo/config"
	"presensi-kominfo/models"

	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {

	email := c.PostForm("email")
	password := c.PostForm("password")

	var user models.User

	config.DB.Where("email = ? AND password = ?", email, password).First(&user)

	// LOGIN GAGAL
	if user.ID == 0 {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"error": true,
		})
		return
	}

	// SIMPAN COOKIE USER LOGIN
	c.SetCookie("user_id", fmt.Sprintf("%d", user.ID), 3600, "/", "", false, true)
	c.SetCookie("role", user.Role, 3600, "/", "", false, true)

	// LOGIN ADMIN
	if user.Role == "admin" {
		c.Redirect(http.StatusFound, "/admin")
		return
	}

	// LOGIN PEGAWAI
	if user.Role == "pegawai" {
		c.Redirect(http.StatusFound, "/pegawai")
		return
	}

	// LOGIN MAGANG
	if user.Role == "magang" {
		c.Redirect(http.StatusFound, "/magang")
		return
	}
}
