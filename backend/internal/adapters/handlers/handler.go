package handlers

import (
	"codeflare/internal/core/ports"
	// "codeflare/internal/core/services"
	"encoding/json"
	"fmt"

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

func (s *ApiHandler) DeployHandler(c echo.Context) error {
	jsonBody := make(map[string]interface{})
	err := json.NewDecoder(c.Request().Body).Decode(&jsonBody)
	if err != nil {
		return nil
	}

	repoUrl := jsonBody["repo_url"]
	repoUrlStr := fmt.Sprint(repoUrl)

	alreadyDeployed, alreadyDeployedErr := s.DeployService.AlreadyDeployed(repoUrlStr)
	if alreadyDeployedErr != nil {
		return alreadyDeployedErr
	}
	if alreadyDeployed {
		return fmt.Errorf("repo already deployed")
	}

	validateErr := s.DeployService.ValidateURL(repoUrlStr)
	if validateErr != nil {
		return validateErr
	}

	fmt.Println("Validated URL")

	dir, cloneErr := s.DeployService.CloneRepo(repoUrlStr)
	if cloneErr != nil {
		return cloneErr
	}

	fmt.Println("Cloned repo")

	_, buildErr := s.DeployService.BuildRepo(dir)
	if buildErr != nil {
		return buildErr
	}

	fmt.Println("Repo built successfully")

	url, uploadErr := s.DeployService.UploadToS3(dir)
	if uploadErr != nil {
		return uploadErr
	}
	fmt.Print("UPLOAD TO S3 success", url)

	addDNSRecordErr := s.DeployService.AddDNSRecord(url)
	if addDNSRecordErr != nil {
		return addDNSRecordErr
	}

	fmt.Println("ADDED SUCCESS")
	return c.String(200, "hello")
}
