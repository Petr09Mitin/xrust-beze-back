package authgrpc

import (
	"context"

	"github.com/Petr09Mitin/xrust-beze-back/internal/services/auth"
	authpb "github.com/Petr09Mitin/xrust-beze-back/proto/auth"
	"github.com/rs/zerolog"
)

type AuthService struct {
	authpb.UnimplementedAuthServiceServer
	authService auth.AuthService
	logger      zerolog.Logger
}

func NewAuthService(authService auth.AuthService, logger zerolog.Logger) *AuthService {
	return &AuthService{
		authService: authService,
		logger:      logger,
	}
}

func (s *AuthService) ValidateSession(ctx context.Context, req *authpb.SessionRequest) (*authpb.SessionResponse, error) {
	session, _, err := s.authService.ValidateSession(ctx, req.SessionId)
	if err != nil || session == nil {
		return &authpb.SessionResponse{
			UserId: "",
			Valid:  false,
		}, nil
	}
	return &authpb.SessionResponse{
		UserId: session.UserID,
		Valid:  true,
	}, nil
}
