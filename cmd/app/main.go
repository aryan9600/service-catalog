package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aryan9600/service-catalog/internal/api"
	"github.com/aryan9600/service-catalog/internal/auth"
	"github.com/aryan9600/service-catalog/internal/models"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("failed to read .env")
	}

	if err := models.SetDBConfiguration(); err != nil {
		panic(err)
	}
	if err := models.InitDB(); err != nil {
		panic(err)
	}

	if os.Getenv("AUTO_MIGRATE") == "true" {
		log.Println("running migrations...")
		if err := models.Migrate("", false); err != nil {
			panic(err)
		}
	}

	if err := auth.SetTokenGenerationConfig(); err != nil {
		panic(err)
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	router := api.NewRouter()
	router.Run(fmt.Sprintf(":%s", port))
}
