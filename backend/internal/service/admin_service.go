package service

import (
	"context"
	"time"

	"github.com/user/video-downloader-backend/internal/dto"
	"github.com/user/video-downloader-backend/internal/infrastructure/contextpool"
	"github.com/user/video-downloader-backend/internal/model"
	"github.com/user/video-downloader-backend/internal/repository"
)

type AdminService interface {
	GetDashboardData(ctx context.Context, params model.QueryParamsRequest) (*dto.DashboardResponse, error)
}

type adminService struct {
	adminRepository repository.AdminRepository
}

func NewAdminService(adminRepository repository.AdminRepository) AdminService {
	return &adminService{
		adminRepository: adminRepository,
	}
}

func (s *adminService) GetDashboardData(ctx context.Context, params model.QueryParamsRequest) (*dto.DashboardResponse, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	data, err := s.adminRepository.GetDashboardData(subCtx, params)
	if err != nil {
		return nil, err
	}
	return data, nil
}
