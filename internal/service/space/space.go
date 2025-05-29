package space

import (
	"context"
	"errors"
	"webserver/internal/model"
	"webserver/internal/model/rabbit"

	api_errors "webserver/internal/errors"

	"github.com/google/uuid"
)

func (s *Space) GetSpaceByID(ctx context.Context, id uuid.UUID) (model.Space, error) {
	space, err := s.cache.GetSpaceByID(ctx, id)
	if err != nil {
		if !errors.Is(err, api_errors.ErrSpaceNotExists) {
			return model.Space{}, err
		}
	}

	if err == nil {
		return space, nil
	}

	return s.repo.GetSpaceByID(ctx, id)
}

// IsUserInSpace проверяет, состоит ли пользователь в пространстве
func (s *Space) IsUserInSpace(ctx context.Context, userID int64, spaceID uuid.UUID) error {
	return s.repo.CheckParticipant(ctx, userID, spaceID)
}

func (s *Space) CreateSpace(ctx context.Context, req rabbit.CreateSpaceRequest) error {
	return s.worker.CreateSpace(ctx, req)
}
