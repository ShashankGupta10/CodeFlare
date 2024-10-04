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

func (s *deployService) ValidateURL(url string) error {
	return fmt.Errorf("")
}

func (s *deployService) CloneRepo(url string) (string, error) {
	return "", nil
}

func (s *deployService) BuildRepo(url string) (string, error) {
	return "", nil
}

func (s *deployService) UploadToS3(url string) (string, error) {
	return "", nil
}

func (s *deployService) AddDNSRecord(url string) error {
	return nil
}
