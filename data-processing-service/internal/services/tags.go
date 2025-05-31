package services

import (
	"context" //nolint:gosec
	"fmt"

	"data-processing-service/internal/models"
	"data-processing-service/internal/repo"
)

type Tags interface {
	Create(ctx context.Context, params models.Tag) (models.Tag, error)
	Read(ctx context.Context) (ReadResult, error)
	Update(ctx context.Context, params UpdateParams) error
	Delete(ctx context.Context, deviceID int32) error
}

type TagsService struct {
	repo            repo.Tags
	messagesService Messages
}

func NewTagsService(r repo.Tags, messagesService Messages) Tags {
	return &TagsService{
		repo:            r,
		messagesService: messagesService,
	}
}

func (s *TagsService) Create(ctx context.Context, params models.Tag) (models.Tag, error) {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return models.Tag{}, fmt.Errorf("s.repo.BeginTx: %w", err)
	}
	defer tx.Rollback()

	ret, err := tx.Create(params)
	if err != nil {
		return models.Tag{}, fmt.Errorf("tx.Create: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return models.Tag{}, fmt.Errorf("tx.Commit: %w", err)
	}

	s.messagesService.UpdateTags()

	return ret, nil
}

type (
	ReadResult struct {
		Tags []models.Tag
	}
)

func (s *TagsService) Read(ctx context.Context) (ReadResult, error) {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return ReadResult{}, fmt.Errorf("s.repo.BeginTx: %w", err)
	}
	defer tx.Rollback()

	ret, err := tx.Read(ctx)
	if err != nil {
		return ReadResult{}, fmt.Errorf("tx.Read: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return ReadResult{}, fmt.Errorf("tx.Commit: %w", err)
	}

	return ReadResult{
		Tags: ret.Tags,
	}, nil
}

type (
	UpdateParams struct {
		ID             int32
		Name           *string
		DeviceId       *int32
		Regexp         *string
		CompareType    *string
		Value          *string
		ArrayIndex     *int32
		Subject        *string
		ServinityLevel *string
	}
)

func (s *TagsService) Update(ctx context.Context, params UpdateParams) error {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("s.repo.BeginTx: %w", err)
	}
	defer tx.Rollback()

	err = tx.Update(ctx, repo.UpdateTagsOpts{
		ID:             params.ID,
		Name:           params.Name,
		DeviceId:       params.DeviceId,
		Regexp:         params.Regexp,
		CompareType:    params.CompareType,
		Value:          params.Value,
		ArrayIndex:     params.ArrayIndex,
		Subject:        params.Subject,
		ServinityLevel: params.ServinityLevel,
	})
	if err != nil {
		return fmt.Errorf("s.repo.Update: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	s.messagesService.UpdateTags()

	return nil
}

func (s *TagsService) Delete(ctx context.Context, tagID int32) error {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("s.repo.BeginTx: %w", err)
	}
	defer tx.Rollback()

	err = tx.Delete(ctx, tagID)
	if err != nil {
		return fmt.Errorf("tx.Delete: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	s.messagesService.UpdateTags()

	return nil
}
