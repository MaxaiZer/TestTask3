package models

type User struct {
	ID       int64
	Name     string
	Password []byte
	Email    string
}
