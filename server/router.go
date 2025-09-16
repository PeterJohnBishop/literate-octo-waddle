package server

import (
	"database/sql"
	"rliterate-octo-waddle/server/handlers"

	"github.com/gin-gonic/gin"
)

func addOpenRoutes(r *gin.Engine, db *sql.DB, dbStatus string) {
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"api": "[CONNECTED] api is running on nginx proxy http://localhost/",
			"db":  dbStatus,
		})
	})
	r.POST("auth/login", func(c *gin.Context) {
		handlers.Login(db, c)
	})
	r.POST("auth/register", func(c *gin.Context) {
		handlers.RegisterUser(db, c)
	})
	r.GET("auth/refresh", func(c *gin.Context) {
		handlers.Refresh(c)
	})
	r.POST("auth/logout", func(c *gin.Context) {
		handlers.Logout(c)
	})
}

func addProtectedRoutes(r *gin.RouterGroup, db *sql.DB) {

	r.GET("/users", func(c *gin.Context) {
		handlers.GetUsers(db, c)
	})
	r.GET("/users/:id", func(c *gin.Context) {
		handlers.GetUserByID(db, c)
	})
	r.PUT("/users", func(c *gin.Context) {
		handlers.UpdateUser(db, c)
	})
	r.DELETE("/users/:id", func(c *gin.Context) {
		handlers.DeleteUserByID(db, c)
	})
}
