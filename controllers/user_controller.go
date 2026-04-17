package controllers

import (
	"net/http"
	"presensi-kominfo/config"
	"presensi-kominfo/models"

	"github.com/gin-gonic/gin"
)

func DataUser(c *gin.Context) {

	var users []models.User

	config.DB.Find(&users)

	c.HTML(http.StatusOK, "users.html", gin.H{
		"data": users,
	})
}

func TambahUser(c *gin.Context) {

	nama := c.PostForm("nama")
	email := c.PostForm("email")
	password := c.PostForm("password")
	role := c.PostForm("role")

	user := models.User{
		Nama:     nama,
		Email:    email,
		Password: password,
		Role:     role,
	}

	config.DB.Create(&user)

	c.Redirect(http.StatusMovedPermanently, "/admin/users")
}

func HapusUser(c *gin.Context) {

	id := c.Param("id")

	config.DB.Delete(&models.User{}, id)

	c.Redirect(http.StatusMovedPermanently, "/admin/users")
}
