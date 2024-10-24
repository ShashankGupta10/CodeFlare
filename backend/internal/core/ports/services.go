package ports

import "codeflare/internal/core/domain"

type DeployService interface {
	BuildRepo()
	Deploy()
	StartCleanupTicker()
	AlreadyDeployed(string) (bool, error)
	AddDNSRecord(string, string) error
	DeleteProject(uint) error
	CreateProject(*domain.Project) (uint, error)
	QueueBuild(uint)
	GetProject(uint) (*domain.Project, error)
}
