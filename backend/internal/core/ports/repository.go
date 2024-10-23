package ports

import (
	"codeflare/internal/core/domain"
)

type Repository interface {
	CreateProject(proj *domain.Project) (uint, error)
	AutoMigrate() error
	FindRepo(url string) (bool, error)
	UpdateStatus(id uint, status domain.Status) error
	UpdateDeployedURL(name string, deployed_url string) error
	GetProject(id uint) (*domain.Project, error)
	UpdateURL(id uint, url string) error
	UpdateBuildURL(id uint, url string) error
	DeleteProject(projectID uint) error
	GetProjectByName(name string) (*domain.Project, error)
}
