package main

import (
	"codeflare/internal/adapters/handlers"
	"codeflare/internal/adapters/repository"
	"codeflare/internal/config"
	"codeflare/internal/core/services"
	"codeflare/pkg/utils"
	"fmt"
	"log"

	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/echo/v4"
)

func main() {
	cfg := config.LoadConfig()
	db, err := repository.NewPGStore(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("error initializing database:", err)
	}

	if err := db.AutoMigrate(); err != nil {
		log.Fatal("migration error:", err)
	}

	deployService := services.NewDeployService(db, 5)
	h := handlers.NewApiHandler(deployService)

	// Start the build and deploy goroutines
	go deployService.BuildRepo()
	go deployService.Deploy()

	// Start the cleanup ticker
	deployService.StartCleanupTicker()

	e := echo.New()
	e.Use(middleware.CORS())
	e.HideBanner = true
	e.Use(utils.CustomLogger())

	e.GET("/", h.HomeHandler)
	e.POST("/deploy", h.DeployHandler)
	e.GET("/project/:id", h.ProjectStatusHandler)
	e.DELETE("/project/:id", h.DeleteProjectHandler)

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", cfg.ServerPort)))
}
