package domain

type Project struct {
	ID       uint   `gorm:"primaryKey;column:id"`
	Name     string `gorm:"not null;column:name"`
	State    string `gorm:"not null;column:state"`
	BuildURL string `gorm:"not null;column:build_url"`
	RepoURL  string `gorm:"not null;column:repo_url"`
	URL      string `gorm:"not null;column:url"`
}
