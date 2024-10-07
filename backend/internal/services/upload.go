package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"codeflare/internal/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
)

const GIT_API_URL = "https://api.github.com/repos/"

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
				if item.Name() != ".git" && item.Name() != "bin" {
					// fmt.Println("dir", item.Name())
					q = append(q, fullpath+"/")
				}
			} else {
				filePaths = append(filePaths, fullpath)
			}
		}
	}
	return filePaths, nil
}

func GetRepoContent(url, projectName, userId string) (string, error) {
	err := ValidateRepoUrl(url)
	if err != nil {
		return "", err
	}
	project := models.NewProject(url, userId, projectName)

	cmd := exec.Command("git", "clone", url, "./projects/"+project.ID)
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



// projectName, bucketName string, filePaths []string
func UploadToS3(projectId, bucketName string, filePaths []string) error {
	if err := godotenv.Load(); err != nil {
		fmt.Println("load env")
		return err
	}
	fmt.Println("HERE AFTER ENV")
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-south-1"))
	if err != nil {
		fmt.Println("load from cfg")
		return err
	}
	fmt.Println("HERE AFTER CFG")
	client := s3.NewFromConfig(cfg)
	fmt.Println("HERE AFTER CLIENT")
	for _, filepath := range filePaths {
		file, err := os.Open(filepath)
		if err != nil {
			return err
		} else {
			defer file.Close()
			_, err := client.PutObject(context.TODO(), &s3.PutObjectInput{
				Bucket: aws.String("codeflare6969"),
				Key:    aws.String(filepath),
				Body:   file,
			})

			if err != nil {
				return err
			}
		}

	}
	return nil

}

func main() {

	pid, err := GetRepoContent("https://github.com/sarthak0714/ttt", "temp", "zzz")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("HERE BRO")

	files, err := GetFilePaths("./projects/" + pid)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("HERE AGAIN BRO")
	er := UploadToS3(pid, "codeflare6969", files)
	if er != nil {
		fmt.Println(er)
	}
	fmt.Println("Done")

}
