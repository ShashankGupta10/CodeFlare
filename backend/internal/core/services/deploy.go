package services

import (
	"codeflare/internal/adapters/repository"
	"codeflare/internal/core/ports"
	"fmt"
)

type deployService struct {
	db *repository.PGStore
}

func NewDeployService(db *repository.PGStore) port.DeployService {
	return &deployService{db: db}
}

func (s *deployService) Deploy() {
	fmt.Println("hello i am here in deploy")
}
