package usecase

import (
	"context"

	"github.com/KonbIgoGo/pr_splitter/generated"
	"github.com/KonbIgoGo/pr_splitter/internal/entity"
)

func parseTeamMembers(members []entity.TeamMember) []generated.TeamMember {
	resMembers := make([]generated.TeamMember, 0, len(members))
	for _, m := range members {
		resMembers = append(resMembers, generated.TeamMember{
			IsActive: m.IsActive,
			UserId:   m.UserID,
			Username: m.Username,
		})
	}
	return resMembers
}

func (i *useCaseImpl) TeamAdd(ctx context.Context, name string, members []generated.TeamMember) (generated.Team, error) {
	membersRes := make([]entity.TeamMember, 0, len(members))
	for _, m := range members {
		membersRes = append(membersRes, entity.TeamMember{
			UserID:   m.UserId,
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}

	team, err := i.teamRepository.AddTeam(ctx, entity.Team{
		TeamName: name,
		Members:  membersRes,
	})
	if err != nil {
		return generated.Team{}, err
	}

	return generated.Team{
		Members:  parseTeamMembers(team.Members),
		TeamName: team.TeamName,
	}, nil
}
func (i *useCaseImpl) TeamGet(ctx context.Context, name string) (generated.Team, error) {
	team, err := i.teamRepository.GetTeam(ctx, name)
	if err != nil {
		return generated.Team{}, err
	}

	return generated.Team{
		Members:  parseTeamMembers(team.Members),
		TeamName: team.TeamName,
	}, nil
}
