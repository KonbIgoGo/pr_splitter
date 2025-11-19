package entity

import "errors"

type Team struct {
	Members  []TeamMember
	TeamName string
}

type TeamMember struct {
	UserID   string
	Username string
	IsActive bool
}

var (
	ErrTeamNotFound      = errors.New("team not found")
	ErrTeamAlreadyExists = errors.New("team alredy exists")
)
