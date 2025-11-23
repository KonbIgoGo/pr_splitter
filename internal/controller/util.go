package controller

import (
	"errors"
	"net/http"

	"github.com/KonbIgoGo/pr_splitter/generated"
	"github.com/KonbIgoGo/pr_splitter/internal/entity"
)

type Error struct {
	Code    generated.ErrorResponseErrorCode `json:"code"`
	Message string                           `json:"message"`
}

func convertErrors(err error) (int, generated.ErrorResponse) {
	if errors.Is(err, entity.ErrPRNotFound) ||
		errors.Is(err, entity.ErrUserNotFound) ||
		errors.Is(err, entity.ErrPRNotFound) {
		return http.StatusNotFound, generated.ErrorResponse{
			Error: Error{
				Message: err.Error(),
			},
		}
	} else if errors.Is(err, entity.ErrPRAlreadyExists) {
		return http.StatusConflict, generated.ErrorResponse{
			Error: Error{
				Code:    generated.PREXISTS,
				Message: err.Error(),
			},
		}
	} else if errors.Is(err, entity.ErrTeamAlreadyExists) {
		return http.StatusConflict, generated.ErrorResponse{
			Error: Error{
				Code:    generated.TEAMEXISTS,
				Message: err.Error(),
			},
		}
	} else if errors.Is(err, entity.ErrPRNoCandidates) {
		return http.StatusConflict, generated.ErrorResponse{
			Error: Error{
				Code:    generated.NOCANDIDATE,
				Message: err.Error(),
			},
		}
	} else if errors.Is(err, entity.ErrReviewerIsNotAssignedToPR) {
		return http.StatusConflict, generated.ErrorResponse{
			Error: Error{
				Code:    generated.NOTASSIGNED,
				Message: err.Error(),
			},
		}
	} else if errors.Is(err, entity.ErrPRMerged) {
		return http.StatusConflict, generated.ErrorResponse{
			Error: Error{
				Code:    generated.PRMERGED,
				Message: err.Error(),
			},
		}
	} else {
		return http.StatusInternalServerError, generated.ErrorResponse{
			Error: Error{
				Message: err.Error(),
			},
		}
	}
}
