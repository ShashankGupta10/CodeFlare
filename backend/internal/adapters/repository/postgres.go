package repository

import (
	"codeflare/internal/core/domain"
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

func (s *PGStore) AutoMigrate () error {
	err := s.db.AutoMigrate(&domain.Project{})
	if err != nil {
		return err
	}
	return nil
}

func (s *PGStore) CreateUser(username string) error {
	// err:= s.db.Create()
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