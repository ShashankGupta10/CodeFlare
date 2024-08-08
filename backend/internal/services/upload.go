package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"codeflare/internal/models"
)

const GIT_API_URL = "https://api.github.com/repos/"

func GetRepoContent(url, projectName, userId string) (string, error) {
	err := ValidateRepoUrl(url)
	if err != nil {
		return "", err
	}
	project := models.NewProject(url, userId, projectName)

	cmd := exec.Command("git", "clone", url, "./projects/"+project.Name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error cloning repository: %v", err)

	}

	return project.ID, nil

}

func ValidateRepoUrl(url string) error {
	parts := strings.Split(url, "/")
	owner := parts[len(parts)-2]
	repo := parts[len(parts)-1]

	resp, err := http.Get("https://api.github.com/repos/" + owner + "/" + repo)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("repository not found")
	}

	return nil
}
