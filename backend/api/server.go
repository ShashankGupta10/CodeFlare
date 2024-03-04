package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Server struct {
	listenAddr string
}

func NewServer(addr string) *Server {
	return &Server{
		listenAddr: addr,
	}
}

// func (s *Server)

func (s *Server) Run() error {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "works")
	})
	return e.Start(s.listenAddr)
}

func (s *Server) HandleLink(c echo.Context){
	
}


