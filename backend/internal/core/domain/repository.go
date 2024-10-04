package domain

type Project struct {
	ID       uint   `gorm:"primaryKey"`
	name     string `gorm:"not null"`
	state    string `gorm:"not null"`
	buildURL string `gorm:"not null"`
	repoURL  string `gorm:"not null"`
	URL      string `gorm:"not null"`
}
