package service

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

type CounterRepository interface {
	Increment(ctx context.Context) (int64, error)
	Get(ctx context.Context) (int64, error)
}

type СounterService struct {
	repo CounterRepository
}

func NewCounterService(repo CounterRepository) *СounterService {
	v := validator.New()
	if err := v.Var(repo, "required"); err != nil {
		panic("СounterService: repo is required")
	}
	return &СounterService{repo: repo}
}

func (s *СounterService) Increment(ctx context.Context) (int64, error) {
	val, err := s.repo.Increment(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "service increment")
	}
	return val, nil
}

func (s *СounterService) Get(ctx context.Context) (int64, error) {
	val, err := s.repo.Get(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "service get")
	}
	return val, nil
}
