package main

import (
	"github.com/gin-gonic/gin"

	"ewallet-monolith/config"
	"ewallet-monolith/controllers"
	"ewallet-monolith/middleware"
)

func main() {
	db := config.InitDB()

	authController := controllers.NewAuthController(db)
	txController := controllers.NewTransactionController(db)

	r := gin.Default()

	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/register", authController.Register)
		auth.POST("/login", authController.Login)
	}

	protected := r.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.POST("/transactions/topup", txController.TopUp)
		protected.POST("/transactions/pay", txController.InitiatePayment)
		protected.POST("/transactions/pay/:id/confirm", txController.ConfirmPayment)
	}

	r.Run(":8080")
}