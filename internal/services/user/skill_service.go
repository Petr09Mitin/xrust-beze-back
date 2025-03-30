package user_service

import (
	"context"
	"github.com/rs/zerolog"
	"time"

	user_model "github.com/Petr09Mitin/xrust-beze-back/internal/models/user"
	user_repo "github.com/Petr09Mitin/xrust-beze-back/internal/repository/user"
)

type SkillService interface {
	GetAll(ctx context.Context) ([]user_model.SkillsByCategory, error)
	GetByCategory(ctx context.Context, category string) ([]string, error)
	GetAllCategories(ctx context.Context) ([]string, error)
}

type skillService struct {
	skillRepo user_repo.SkillRepo
	timeout   time.Duration
	logger    zerolog.Logger
}

func NewSkillService(skillRepo user_repo.SkillRepo, timeout time.Duration, logger zerolog.Logger) SkillService {
	return &skillService{
		skillRepo: skillRepo,
		timeout:   timeout,
		logger:    logger,
	}
}

func (s *skillService) GetAll(ctx context.Context) ([]user_model.SkillsByCategory, error) {
	return s.skillRepo.GetAllSkills(ctx)
}

func (s *skillService) GetByCategory(ctx context.Context, category string) ([]string, error) {
	return s.skillRepo.GetSkillsByCategory(ctx, category)
}

func (s *skillService) GetAllCategories(ctx context.Context) ([]string, error) {
	return s.skillRepo.GetAllCategories(ctx)
}
