package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func GetRepoInfo(url string) (bool, error) {
	parts := strings.Split(url, "/")
	owner := parts[len(parts)-2]
	repo := parts[len(parts)-1]

	resp, err := http.Get("https://api.github.com/repos/" + owner + "/" + repo)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return false, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return false, fmt.Errorf("repository not found")
	}
	private, _ := response["private"].(bool)
	return private, nil
}
