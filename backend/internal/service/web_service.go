package service

import (
	"context"
	"time"

	"github.com/user/video-downloader-backend/internal/delivery/helpers"
	"github.com/user/video-downloader-backend/internal/dto"
	"github.com/user/video-downloader-backend/internal/infrastructure/contextpool"
)

type WebService interface {
	Contact(ctx context.Context, req *dto.ContactRequest) error
}

type webService struct {
	mailHelper helpers.MailHelper
}

func NewWebService(mailHelper helpers.MailHelper) WebService {
	return &webService{
		mailHelper: mailHelper,
	}
}

func (s *webService) Contact(ctx context.Context, req *dto.ContactRequest) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	return s.mailHelper.SendContactEmail(subCtx, req)
}
