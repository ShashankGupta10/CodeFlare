package handlers

import (
	"codeflare/internal/config"
	"codeflare/internal/core/domain"
	"codeflare/internal/core/ports"
	"codeflare/pkg/utils"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

type ApiHandler struct {
	DeployService ports.DeployService
}

func NewApiHandler(deployService ports.DeployService) *ApiHandler {
	return &ApiHandler{
		DeployService: deployService,
	}
}

func (s *ApiHandler) HomeHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"msg":  "works",
		"time": time.Now().Format("01-01-2006 15:04:00"),
	})
}

func (s *ApiHandler) DeployHandler(c echo.Context) error {
	var requestBody struct {
		RepoURL          string `json:"repo_url"`
		ProjectDirectory string `json:"project_directory"`
	}

	if err := c.Bind(&requestBody); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
	}

	if requestBody.RepoURL == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid repo_url"})
	}

	alreadyDeployed, err := s.DeployService.AlreadyDeployed(requestBody.RepoURL)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if alreadyDeployed {
		return c.JSON(http.StatusConflict, map[string]string{"error": "Repo already deployed"})
	}

	if err := utils.ValidateURL(requestBody.RepoURL); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Extract project name from the URL
	urlParts := strings.Split(requestBody.RepoURL, "/")
	if len(urlParts) < 5 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid repo URL format"})
	}
	projectName := urlParts[4]

	// Clone the repository
	err = utils.CloneRepo(requestBody.RepoURL)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to clone repository"})
	}

	smallcaseName := strings.ToLower(projectName)
	project := &domain.Project{
		Name:             smallcaseName,
		RepoURL:          requestBody.RepoURL,
		Status:           domain.NotStarted,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		BuildURL:         "",
		URL:              "",
		ProjectDirectory: requestBody.ProjectDirectory,
		ErrorMessage: "",
	}

	projectID, err := s.DeployService.CreateProject(project)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create project"})
	}

	s.DeployService.QueueBuild(projectID)
	return c.JSON(http.StatusAccepted, map[string]interface{}{
		"message": "Deployment queued",
		"id":      projectID,
	})
}

func (s *ApiHandler) ProjectStatusHandler(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid project ID"})
	}

	project, err := s.DeployService.GetProject(uint(id))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Project not found"})
	}

	return c.JSON(http.StatusOK, project)
}

func (s *ApiHandler) DeleteProjectHandler(c echo.Context) error {
	// Get the secret phrase from the environment or configuration
	expectedSecret := config.LoadConfig().DeleteSecretPhrase

	// Get the secret phrase from the request header
	providedSecret := c.Request().Header.Get("X-Delete-Secret")

	// Verify the secret phrase
	if providedSecret != expectedSecret {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized delete request"})
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid project ID"})
	}

	if err := s.DeployService.DeleteProject(uint(id)); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete project"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Project deleted successfully"})
}
