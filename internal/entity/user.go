package entity

import "errors"

type User struct {
	ID       string
	Name     string
	TeamName string
	IsActive bool
}

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user alredy exists")
)
