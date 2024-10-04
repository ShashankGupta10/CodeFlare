package port

import (
	"codeflare/internal/adapters/repository"
)

type Repository interface {
	NewPGStore(connectionString string) (*repository.PGStore, error)
}
