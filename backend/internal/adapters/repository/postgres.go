package repository

import (
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
