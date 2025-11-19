package entity

import (
	"errors"
	"time"
)

type Status int

const (
	OPEN Status = iota
	MERGED
)

type PR struct {
	ID                  string
	Name                string
	AuthorID            string
	AssignedReviewersID []string
	Status              Status
	CreatedAt           time.Time
	MergedAt            time.Time
}

var (
	ErrPRNotFound          = errors.New("pr not found")
	ErrPRAlreadyExists     = errors.New("pr already exists")
	ErrPRNothingToReassign = errors.New("nothing to reassign in pr")
)
