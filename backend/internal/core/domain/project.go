package domain

import "time"

type Status int

const (
	NotStarted Status = iota
	Building
	Deploying
	Deployed
	Failed
)

type Project struct {
	ID               uint      `gorm:"primaryKey;column:id" json:"id"`
	Name             string    `gorm:"not null;column:name" json:"name"`
	Status           Status    `gorm:"colum:status" json:"status"`
	ProjectDirectory string    `gorm:"column:project_directory" json:"project_directory"`
	BuildURL         string    `gorm:"column:build_url" json:"build_url"`
	RepoURL          string    `gorm:"not null;column:repo_url" json:"repo_url"`
	URL              string    `gorm:"column:url" json:"url"`
	CreatedAt        time.Time `gorm:"created_at" json:"created_at"`
	UpdatedAt        time.Time `gorm:"updated_at" json:"updated_at"`
	DeployedURL      string    `gorm:"deployed_url" json:"deployed_url"`
}
