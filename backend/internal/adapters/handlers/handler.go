package handlers

import (
	"codeflare/internal/core/ports"
	"github.com/labstack/echo/v4"
)

type ApiHandler struct {
	DeployService port.DeployService
}

func NewApiHandler(deployService port.DeployService) *ApiHandler {
	return &ApiHandler{
		DeployService: deployService,
	}
}

func (s *ApiHandler) HomeHandler(c echo.Context) error {
	return c.String(200, "Hello bro")
}

