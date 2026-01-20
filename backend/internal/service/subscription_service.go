package service

import "github.com/user/video-downloader-backend/internal/repository"

type SubscriptionService interface{}

type subscriptionService struct {
	repo repository.SubscriptionRepository
}

func NewSubscriptionService(repo repository.SubscriptionRepository) SubscriptionService {
	return &subscriptionService{
		repo: repo,
	}
}
