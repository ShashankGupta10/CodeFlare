package models

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

type Project struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	RepoURL   string    `json:"repo_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Status    string    `json:"status"`
}

func NewProject(repoURL, userId, repoName string) *Project {
	now := time.Now()
	return &Project{
		ID:        GenerateProjectID(),
		RepoURL:   repoURL,
		Name:      repoName,
		UserID:    userId,
		CreatedAt: now,
		UpdatedAt: now,
		Status:    "pending",
	}
}

func (p *Project) UpdateStatus(status string) {
	p.Status = status
	p.UpdatedAt = time.Now()
}

func GenerateProjectID() string {
	timestamp := time.Now().Unix()
	randomBytes := make([]byte, 4)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}
	randomString := hex.EncodeToString(randomBytes)
	return fmt.Sprintf("%d-%s", timestamp, randomString)
}
