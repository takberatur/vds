package repository

import "github.com/jackc/pgx/v5/pgxpool"

type SubscriptionRepository interface {
	BaseRepository
}

type subscriptionRepository struct {
	*baseRepository
}

func NewSubscriptionRepository(db *pgxpool.Pool) SubscriptionRepository {
	return &subscriptionRepository{
		baseRepository: NewBaseRepository(db).(*baseRepository),
	}
}
