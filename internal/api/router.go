package api

import (
	"os"

	"github.com/aryan9600/service-catalog/docs"
	"github.com/aryan9600/service-catalog/internal/middleware"
	"github.com/gin-gonic/gin"

	"github.com/natefinch/lumberjack"
	"github.com/rs/zerolog"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// NewRouter returns a Gin router configured with all endpoints and middleware.
func NewRouter() *gin.Engine {
	fileName := os.Getenv("LOG_FILE")
	if fileName == "" {
		fileName = "file.log"
	}
	z := zerolog.New(&lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    100,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	})

	docs.SwaggerInfo.Title = "Service Catalog"

	router := gin.Default()
	router.Use(middleware.StructuredLogger(&z))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	auth := router.Group("auth")
	auth.POST("/register", Register)
	auth.POST("/login", Login)

	services := router.Group("services")
	services.Use(middleware.JwtAuthMiddleware())

	services.GET("", ListServices)
	services.POST("", CreateService)
	services.GET(":id", GetService)
	services.PATCH(":id", UpdateService)

	services.POST(":id/version", CreateVersion)

	return router
}
