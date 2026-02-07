package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/user/video-downloader-backend/internal/model"
	"github.com/user/video-downloader-backend/internal/repository"
)

type SubscriptionService interface {
	Upsert(ctx context.Context, sub *model.Subscription) (*model.Subscription, error)
	FindCurrentByUserAndApp(ctx context.Context, userID uuid.UUID, appID uuid.UUID, now time.Time) (*model.Subscription, error)
	FindByID(ctx context.Context, subID uuid.UUID) (*model.Subscription, error)
	FindAll(ctx context.Context, params model.QueryParamsRequest) ([]model.Subscription, model.Pagination, error)
	Delete(ctx context.Context, subID uuid.UUID) error
	BulkDelete(ctx context.Context, subIDs []uuid.UUID) error
}

type subscriptionService struct {
	repo repository.SubscriptionRepository
}

func NewSubscriptionService(repo repository.SubscriptionRepository) SubscriptionService {
	return &subscriptionService{
		repo: repo,
	}
}

func (s *subscriptionService) Upsert(ctx context.Context, sub *model.Subscription) (*model.Subscription, error) {
	return s.repo.Upsert(ctx, sub)
}

func (s *subscriptionService) FindCurrentByUserAndApp(ctx context.Context, userID uuid.UUID, appID uuid.UUID, now time.Time) (*model.Subscription, error) {
	return s.repo.FindCurrentByUserAndApp(ctx, userID, appID, now)
}

func (s *subscriptionService) FindAll(ctx context.Context, params model.QueryParamsRequest) ([]model.Subscription, model.Pagination, error) {
	return s.repo.FindAll(ctx, params)
}

func (s *subscriptionService) Delete(ctx context.Context, subID uuid.UUID) error {
	return s.repo.Delete(ctx, subID)
}

func (s *subscriptionService) BulkDelete(ctx context.Context, subIDs []uuid.UUID) error {
	return s.repo.BulkDelete(ctx, subIDs)
}

func (s *subscriptionService) FindByID(ctx context.Context, subID uuid.UUID) (*model.Subscription, error) {
	return s.repo.FindByID(ctx, subID)
}
