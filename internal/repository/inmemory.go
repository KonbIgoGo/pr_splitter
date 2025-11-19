package repository

import (
	"context"
	"errors"
	"math/rand/v2"
	"slices"
	"sync"
	"time"

	"github.com/KonbIgoGo/pr_splitter/internal/entity"
)

var _ PRRepository = (*inmemoryImpl)(nil)
var _ UserRepository = (*inmemoryImpl)(nil)
var _ TeamRepository = (*inmemoryImpl)(nil)

type inmemoryImpl struct {
	prRepoMx *sync.RWMutex
	prRepo   map[string]*entity.PR

	userRepoMx *sync.RWMutex
	userRepo   map[string]*entity.User

	teamRepoMx *sync.RWMutex
	teamRepo   map[string]*entity.Team
}

func (i *inmemoryImpl) getUserTeamIDs(userId string) ([]string, error) {
	i.userRepoMx.RLock()
	defer i.userRepoMx.RUnlock()

	i.teamRepoMx.RLock()
	defer i.teamRepoMx.RUnlock()

	user, ok := i.userRepo[userId]
	if !ok {
		return nil, entity.ErrUserNotFound
	}

	userTeam, ok := i.teamRepo[user.TeamName]
	if !ok {
		return nil, entity.ErrTeamNotFound
	}

	res := make([]string, 0, len(userTeam.Members))

	for _, m := range userTeam.Members {
		res = append(res, m.UserID)
	}

	return res, nil
}

func (i *inmemoryImpl) CreatePullRequest(_ context.Context, id string, name string, authorID string) (entity.PR, error) {
	i.prRepoMx.Lock()
	defer i.prRepoMx.Unlock()

	if _, ok := i.prRepo[id]; ok {
		return entity.PR{}, entity.ErrPRAlreadyExists
	}

	team, err := i.getUserTeamIDs(authorID)

	if err != nil {
		return entity.PR{}, err
	}

	return entity.PR{
		ID:                  id,
		Name:                name,
		AuthorID:            authorID,
		Status:              entity.OPEN,
		CreatedAt:           time.Now(),
		AssignedReviewersID: pickTwoWithExclusion(team, authorID),
	}, nil
}

func (i *inmemoryImpl) MergePullRequest(_ context.Context, id string) (entity.PR, error) {
	i.prRepoMx.Lock()
	defer i.prRepoMx.Unlock()

	pr, ok := i.prRepo[id]
	if !ok {
		return entity.PR{}, entity.ErrPRNotFound
	}

	if pr.Status == entity.MERGED {
		return *pr, nil
	}

	pr.Status = entity.MERGED
	pr.MergedAt = time.Now()

	return *pr, nil
}

func (i *inmemoryImpl) ReassignPullRequest(_ context.Context, id string, oldUserID string) (entity.PR, error) {
	i.prRepoMx.Lock()
	defer i.prRepoMx.Unlock()

	pr, ok := i.prRepo[id]
	if !ok {
		return entity.PR{}, entity.ErrPRNotFound
	}

	team, err := i.getUserTeamIDs(oldUserID)
	if err != nil {
		return entity.PR{}, err
	}

	replaced, err := pickWithExclusion(team, []string{pr.AuthorID, oldUserID})
	if err != nil {
		return entity.PR{}, err
	}

	for i, id := range pr.AssignedReviewersID {
		if id == oldUserID {
			pr.AssignedReviewersID[i] = replaced
			return *pr, nil
		}
	}

	return entity.PR{}, entity.ErrUserNotFound
}

func (i *inmemoryImpl) SetIsActiveUser(_ context.Context, id string, isActive bool) (entity.User, error) {
	i.userRepoMx.Lock()
	defer i.userRepoMx.Unlock()

	if _, ok := i.userRepo[id]; !ok {
		return entity.User{}, entity.ErrUserNotFound
	}

	i.userRepo[id].IsActive = isActive

	return *i.userRepo[id], nil
}

func (i *inmemoryImpl) GetReviewUser(_ context.Context, id string) ([]entity.PR, error) {
	res := make([]entity.PR, 0)
	for _, v := range i.prRepo {
		if slices.Contains(v.AssignedReviewersID, id) {
			res = append(res, *v)
		}
	}
	return res, nil
}

func (i *inmemoryImpl) AddTeam(_ context.Context, team entity.Team) (entity.Team, error) {
	i.teamRepoMx.Lock()
	defer i.teamRepoMx.Unlock()

	if _, ok := i.teamRepo[team.TeamName]; ok {
		return entity.Team{}, entity.ErrTeamAlreadyExists
	}
	i.teamRepo[team.TeamName] = &team

	for _, u := range team.Members {
		i.addUser(entity.User{
			ID:       u.UserID,
			Name:     u.Username,
			IsActive: u.IsActive,
			TeamName: team.TeamName,
		})
	}

	return *i.teamRepo[team.TeamName], nil
}

func (i *inmemoryImpl) addUser(user entity.User) {
	i.userRepoMx.Lock()
	defer i.userRepoMx.Unlock()

	u, ok := i.userRepo[user.ID]

	if !ok {
		i.userRepo[user.ID] = &entity.User{}
	}

	i.userRepo[user.ID].ID = u.ID
	i.userRepo[user.ID].IsActive = u.IsActive
	i.userRepo[user.ID].Name = u.Name
	i.userRepo[user.ID].TeamName = u.TeamName
}

func (i *inmemoryImpl) GetTeam(ctx context.Context, name string) (entity.Team, error) {
	i.teamRepoMx.RLock()
	defer i.teamRepoMx.RUnlock()

	team, ok := i.teamRepo[name]
	if !ok {
		return entity.Team{}, entity.ErrTeamNotFound
	}

	return *team, nil
}

func pickWithExclusion(ids []string, exclude []string) (string, error) {
	if len(ids)-len(exclude) <= 0 {
		return "", errors.New("nothing to pick")
	}

	candidates := make([]string, 0, len(ids)-len(exclude))

	for _, m := range ids {
		if !slices.Contains(exclude, m) {
			candidates = append(candidates, m)
		}
	}

	return candidates[rand.N(len(candidates))], nil
}

func pickTwoWithExclusion(ids []string, exclude string) []string {
	excludes := make([]string, 1)
	excludes[0] = exclude

	if len(ids) < 2 {
		return ids
	}

	res := make([]string, 2)
	res[0], _ = pickWithExclusion(ids, excludes)
	excludes = append(excludes, res[0])
	res[1], _ = pickWithExclusion(ids, excludes)

	return res
}

func NewInmemoryRepository() *inmemoryImpl {
	return &inmemoryImpl{
		prRepoMx:   new(sync.RWMutex),
		userRepoMx: new(sync.RWMutex),
		teamRepoMx: new(sync.RWMutex),

		prRepo:   make(map[string]*entity.PR),
		teamRepo: make(map[string]*entity.Team),
		userRepo: make(map[string]*entity.User),
	}
}
