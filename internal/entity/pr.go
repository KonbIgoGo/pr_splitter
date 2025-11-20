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
	ErrPRNotFound                = errors.New("pr not found")
	ErrPRAlreadyExists           = errors.New("pr already exists")
	ErrPRNoCandidates            = errors.New("no active replacement candidate in team")
	ErrReviewerIsNotAssignedToPR = errors.New("reviewer is not assigned to the PR")
	ErrPRMerged                  = errors.New(" cannot reassign on merged PR")
)
