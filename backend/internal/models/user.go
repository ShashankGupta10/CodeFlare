package models

type User struct {
	Id       string `json:"user_id"`
	Name     string `json:"name"`
	Password string `json:"password"`
}
