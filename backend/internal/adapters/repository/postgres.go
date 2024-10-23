package repository

import (
	"codeflare/internal/core/domain"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PGStore struct {
	db *gorm.DB
}

func NewPGStore(dsn string) (*PGStore, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &PGStore{
		db: db,
	}, nil
}

func (s *PGStore) AutoMigrate() error {
	err := s.db.AutoMigrate(&domain.Project{})
	if err != nil {
		return err
	}
	return nil
}

func (s *PGStore) FindRepo(url string) (bool, error) {
	var project domain.Project
	err := s.db.Where("repo_url = ?", url).First(&project).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (s *PGStore) CreateProject(proj *domain.Project) (uint, error) {
	result := s.db.Create(proj)
	if result.Error != nil {
		return 0, result.Error
	}
	return proj.ID, nil
}

func (s *PGStore) UpdateStatus(id uint, status domain.Status) error {
	return s.db.Model(&domain.Project{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}).Error
}

func (s *PGStore) GetProject(id uint) (*domain.Project, error) {
	var proj domain.Project
	if err := s.db.Find(&proj, id).Error; err != nil {
		return nil, err
	}
	return &proj, nil
}

func (s *PGStore) UpdateURL(id uint, url string) error {
	return s.db.Model(&domain.Project{}).Where("id = ?", id).Update("url", url).Error
}

func (s *PGStore) UpdateBuildURL(id uint, url string) error {
	return s.db.Model(&domain.Project{}).Where("id = ?", id).Update("build_url", url).Error
}

func (s *PGStore) DeleteProject(projectID uint) error {
	return s.db.Model(&domain.Project{}).Delete("id = ?", projectID).Error
}

func (s *PGStore) UpdateDeployedURL(name string, deployed_url string) error {
	return s.db.Model(&domain.Project{}).Where("name = ?", name).Update("deployed_url", deployed_url).Error
}
func (s *PGStore) GetProjectByName(name string) (*domain.Project, error) {
	var proj domain.Project
	if err := s.db.Where("name = ?", name).First(&proj).Error; err != nil {
		return nil, err
	}
	return &proj, nil
}
