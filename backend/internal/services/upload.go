package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
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

func GetFilePaths(repoPath string) ([]string, error) {

	// DFS in development??? 😱
	var filePaths []string
	q := []string{repoPath}

	for len(q) > 0 {
		curr := q[0]
		q = q[1:]

		data, err := os.ReadDir(curr)
		if err != nil {
			return nil, err
		}

		for _, item := range data {
			fullpath := filepath.Join(curr, item.Name())
			// relpath, err := filepath.Rel(repoPath, fullpath)
			// if err != nil {
			// 	return nil, err
			// }
			// relpath = strings.ReplaceAll(relpath, "\\", "/")
			if item.IsDir() {
				q = append(q, fullpath)
				filePaths = append(filePaths, fullpath+"/")
			} else {
				filePaths = append(filePaths, fullpath)
			}
		}
	}
	return filePaths, nil
}

// func UploadToS3(projectName, bucketName string, filePaths []string) error {
// 	client:=s3.S3
// }

// func main() {
// 	vals, err := GetFilePaths("./projects/novanity")
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	fmt.Println(vals)
// }
