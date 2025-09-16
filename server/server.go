package server

import (
	"log"
	"rliterate-octo-waddle/db"
	"rliterate-octo-waddle/server/handlers"
	"rliterate-octo-waddle/server/middleware"

	"github.com/gin-gonic/gin"
)

func StartAuthenticationServer() {
	go hub.Run()

	// Connect to PostgreSQL
	postgres, msg := db.ConnectPSQL()
	err := postgres.Ping()
	if err != nil {
		log.Fatal("Error connecting to the database:", err)
	}
	defer postgres.Close()
	handlers.CreateUsersTable(postgres)

	// Set up Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/ws", serveWs)
	protected := router.Group("/api")
	protected.Use(middleware.JWTMiddleware())
	addOpenRoutes(router, postgres, msg)
	addProtectedRoutes(protected, postgres)
	log.Println("[CONNECTED] server listenting on nginx proxy http://localhost/")
	log.Println(msg)
	router.Run(":8081")
}
