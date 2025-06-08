package space

import (
	"context"
	"errors"
	"webserver/internal/model"
	"webserver/internal/model/rabbit"

	api_errors "webserver/internal/errors"

	"github.com/google/uuid"
)

func (s *Service) GetSpaceByID(ctx context.Context, id uuid.UUID) (model.Space, error) {
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
func (s *Service) IsUserInSpace(ctx context.Context, userID int64, spaceID uuid.UUID) (bool, error) {
	return s.repo.CheckParticipant(ctx, userID, spaceID)
}

func (s *Service) CreateSpace(ctx context.Context, req rabbit.CreateSpaceRequest) error {
	return s.worker.CreateSpace(ctx, req)
}

func (s *Service) AddParticipant(ctx context.Context, req rabbit.AddParticipantRequest) error {
	return s.worker.AddParticipant(ctx, req)
}

func (s *Service) IsSpacePersonal(ctx context.Context, spaceID uuid.UUID) (bool, error) {
	space, err := s.cache.GetSpaceByID(ctx, spaceID)
	if err != nil {
		if !errors.Is(err, api_errors.ErrSpaceNotExists) {
			return false, err
		}
	}

	if err == nil {
		return space.Personal, nil
	}

	return s.repo.IsSpacePersonal(ctx, spaceID)
}

func (s *Service) IsSpaceExists(ctx context.Context, spaceID uuid.UUID) (bool, error) {
	_, err := s.cache.GetSpaceByID(ctx, spaceID)
	if err != nil {
		if !errors.Is(err, api_errors.ErrSpaceNotExists) {
			return false, err
		}
	}

	if err == nil {
		return true, nil
	}

	return s.repo.IsSpaceExists(ctx, spaceID)
}

func (s *Service) CheckInvitation(ctx context.Context, from, to int64, spaceID uuid.UUID) (bool, error) {
	return s.repo.CheckInvitation(ctx, from, to, spaceID)
}
