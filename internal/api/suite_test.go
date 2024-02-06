package api

import (
	"os"
	"testing"

	"github.com/aryan9600/service-catalog/internal/auth"
	"github.com/aryan9600/service-catalog/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/lib/pq"
)

var router *gin.Engine

func TestMain(m *testing.M) {
	if err := godotenv.Load("../../.env.test"); err != nil {
		panic(err)
	}
	if err := models.SetDBConfiguration(); err != nil {
		panic(err)
	}
	if err := models.InitDB(); err != nil {
		panic(err)
	}
	if err := models.Migrate("file://../models/migrations", true); err != nil {
		panic(err)
	}

	if err := models.Migrate("file://../models/migrations", false); err != nil {
		panic(err)
	}
	if err := auth.SetTokenGenerationConfig(); err != nil {
		panic(err)
	}

	populateUsers()
	populateServicesAndVersions()

	router = NewRouter()
	code := m.Run()
	os.Exit(code)
}

func populateUsers() {
	pwd1, err := models.GetPasswordHash("pwd1")
	if err != nil {
		panic(err)
	}
	pwd2, err := models.GetPasswordHash("pwd2")
	if err != nil {
		panic(err)
	}
	users := []models.User{
		{
			Username: "user1",
			Password: pwd1,
		},
		{
			Username: "user2",
			Password: pwd2,
		},
	}
	if err := models.DB.Table(models.UserTableName).Create(&users).Error; err != nil {
		panic(err)
	}
}

func populateServicesAndVersions() {
	services := []models.Service{
		{
			Name:        "auth",
			Description: "authentication and authrorization",
			Versions:    pq.StringArray{"1.0", "1.1"},
			UserID:      1,
		},
		{
			Name:        "storage",
			Description: "durable kv store",
			Versions:    pq.StringArray{"0.1", "0.2"},
			UserID:      1,
		},
		{
			Name:        "dns",
			Description: "domains, subdomains and wildcard domains",
			Versions:    pq.StringArray{"1", "2"},
			UserID:      1,
		},
		{
			Name:        "observability",
			Description: "logging, metrics, profiling",
			Versions:    pq.StringArray{"v1", "v2"},
			UserID:      2,
		},
		{
			Name:        "service mesh",
			Description: "networking, mTLS",
			Versions:    pq.StringArray{"alpha", "beta"},
			UserID:      2,
		},
	}

	if err := models.DB.Table(models.ServiceTableName).Create(&services).Error; err != nil {
		panic(err)
	}
	var versions []models.Version
	for _, svc := range services {
		for _, v := range svc.Versions {
			versions = append(versions, models.Version{
				Version:   v,
				ServiceID: int(svc.ID),
			})
		}
	}

	if err := models.DB.Table(models.VersionTableName).Create(&versions).Error; err != nil {
		panic(err)
	}
}
