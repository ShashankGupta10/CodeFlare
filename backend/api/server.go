package api

import (
	"codeflare/utils"
	"fmt"
	"net/http"
	"os"
	"os/exec"

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
	e.POST("/get_repo", s.getRepo)
	return e.Start(s.listenAddr)
}

// func (s *Server) HandleLink(c echo.Context){

// }
func (s *Server) getRepo(c echo.Context) error {

	url := c.QueryParam("url")
	projectName := c.QueryParam("projectName")

	if url == "" || projectName == "" {
		return c.JSON(http.StatusBadRequest, "URL and projectName are required")
	}

	private, err := utils.GetRepoInfo(url)
	if err != nil {
		fmt.Println("Error getting repository information:", err)
		return c.JSON(http.StatusInternalServerError, "Error getting repository information")
	}

	if private {
		return c.JSON(http.StatusBadRequest, "Repository is private")
	}

	cmd := exec.Command("git", "clone", url, "./projects/"+projectName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error cloning repository:", err)
		return c.JSON(http.StatusInternalServerError, "Error cloning repository")
	}

	return c.JSON(http.StatusOK, "Repository cloned successfully!")
}
