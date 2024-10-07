package main

import (
	"codeflare/internal/adapters/handlers"
	"codeflare/internal/adapters/repository"
	"codeflare/internal/config"
	"codeflare/internal/core/services"
	"github.com/labstack/echo/v4"
	"fmt"
)

func main() {
	cfg := config.LoadConfig()
	db, err := repository.NewPGStore(cfg.DatabaseURL)
	if err != nil {
		fmt.Println("gg")
		return
	}

	migrateErr := db.AutoMigrate()
	if migrateErr != nil {
		fmt.Println("migrate err")
		return 
	}
	deployService := services.NewDeployService(db)
	h := handlers.NewApiHandler(deployService)
	
	e := echo.New()
	e.GET("/", h.HomeHandler)
	e.POST("/deploy", h.DeployHandler)
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", cfg.ServerPort)))
}
