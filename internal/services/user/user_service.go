package user_service

import (
	"context"
	"strings"
	"time"

	custom_errors "github.com/Petr09Mitin/xrust-beze-back/internal/models/error"
	authpb "github.com/Petr09Mitin/xrust-beze-back/proto/auth"
	filepb "github.com/Petr09Mitin/xrust-beze-back/proto/file"
	"github.com/rs/zerolog"

	auth_model "github.com/Petr09Mitin/xrust-beze-back/internal/models/auth"
	user_model "github.com/Petr09Mitin/xrust-beze-back/internal/models/user"
	moderation_repo "github.com/Petr09Mitin/xrust-beze-back/internal/repository/moderation"
	user_repo "github.com/Petr09Mitin/xrust-beze-back/internal/repository/user"
)

type UserService interface {
	Create(ctx context.Context, user *user_model.User, hashedPassword string) error
	GetByID(ctx context.Context, id string) (*user_model.User, error)
	GetByEmail(ctx context.Context, email string) (*user_model.User, error)
	GetByEmailWithPassword(ctx context.Context, email string) (*auth_model.RegisterRequest, error)
	GetByUsername(ctx context.Context, username string) (*user_model.User, error)
	GetByUsernameWithPassword(ctx context.Context, username string) (*auth_model.RegisterRequest, error)
	Update(ctx context.Context, user *user_model.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, page, limit int) ([]*user_model.User, error)
	FindMatchingUsers(ctx context.Context, userID string) ([]*user_model.User, error)
	FindUsersByUsername(ctx context.Context, userID, username string, limit, offset int64) ([]*user_model.User, error)
}

type userService struct {
	userRepo       user_repo.UserRepo
	moderationRepo moderation_repo.ModerationRepo
	fileGRPC       filepb.FileServiceClient
	authGRPC       authpb.AuthServiceClient
	timeout        time.Duration
	logger         zerolog.Logger
}

func NewUserService(userRepo user_repo.UserRepo, moderationRepo moderation_repo.ModerationRepo, fileGRPC filepb.FileServiceClient, authGRPC authpb.AuthServiceClient, timeout time.Duration, logger zerolog.Logger) UserService {
	return &userService{
		userRepo:       userRepo,
		moderationRepo: moderationRepo,
		fileGRPC:       fileGRPC,
		authGRPC:       authGRPC,
		timeout:        timeout,
		logger:         logger,
	}
}

func (s *userService) Create(ctx context.Context, user *user_model.User, hashedPassword string) error {
	if err := s.checkUserForProfanity(ctx, user); err != nil {
		return err
	}

	existingUser, err := s.userRepo.GetByEmail(ctx, user.Email)
	if err == nil && existingUser != nil {
		return custom_errors.ErrEmailAlreadyExists
	}
	existingUser, err = s.userRepo.GetByUsername(ctx, user.Username)
	if err == nil && existingUser != nil {
		return custom_errors.ErrUsernameAlreadyExists
	}
	if user.Avatar != "" {
		_, err = s.fileGRPC.MoveTempFileToAvatars(ctx, &filepb.MoveTempFileToAvatarsRequest{
			Filename: user.Avatar,
		})
		if err != nil {
			return err
		}
	}
	err = s.userRepo.Create(ctx, user, hashedPassword)
	if err != nil {
		return err
	}
	return nil
}

func (s *userService) GetByID(ctx context.Context, id string) (*user_model.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *userService) GetByEmail(ctx context.Context, email string) (*user_model.User, error) {
	return s.userRepo.GetByEmail(ctx, email)
}

func (s *userService) GetByEmailWithPassword(ctx context.Context, email string) (*auth_model.RegisterRequest, error) {
	return s.userRepo.GetByEmailWithPassword(ctx, email)
}

func (s *userService) GetByUsername(ctx context.Context, username string) (*user_model.User, error) {
	return s.userRepo.GetByUsername(ctx, username)
}

func (s *userService) GetByUsernameWithPassword(ctx context.Context, username string) (*auth_model.RegisterRequest, error) {
	return s.userRepo.GetByUsernameWithPassword(ctx, username)
}

func (s *userService) Update(ctx context.Context, user *user_model.User) error {
	if err := s.checkUserForProfanity(ctx, user); err != nil {
		return err
	}
	// Проверяем существование пользователя
	existingUser, err := s.userRepo.GetByID(ctx, user.ID.Hex())
	if err != nil {
		return err
	}
	// Чтобы эти поля не обновлялись
	user.CreatedAt = existingUser.CreatedAt
	user.LastActiveAt = existingUser.LastActiveAt
	// Проверка на уникальность email и username, если они изменились
	if existingUser.Email != user.Email {
		userWithEmail, err := s.userRepo.GetByEmail(ctx, user.Email)
		if err == nil && userWithEmail != nil && userWithEmail.ID != user.ID {
			return custom_errors.ErrEmailAlreadyExists
		}
	}
	if existingUser.Username != user.Username {
		userWithUsername, err := s.userRepo.GetByUsername(ctx, user.Username)
		if err == nil && userWithUsername != nil && userWithUsername.ID != user.ID {
			return custom_errors.ErrUsernameAlreadyExists
		}
	}

	if existingUser.Avatar != user.Avatar {
		_, err = s.fileGRPC.MoveTempFileToAvatars(ctx, &filepb.MoveTempFileToAvatarsRequest{
			Filename: user.Avatar,
		})
		if err != nil {
			return err
		}

		_, err = s.fileGRPC.DeleteAvatar(ctx, &filepb.DeleteAvatarRequest{
			Filename: existingUser.Avatar,
		})
		if err != nil {
			s.logger.Error().Err(err).Msg("failed to delete avatar") // not crit, still can return 200
		}
	}

	return s.userRepo.Update(ctx, user)
}

func (s *userService) Delete(ctx context.Context, id string) error {
	// Проверяем существование пользователя
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if user.Avatar != "" {
		_, err = s.fileGRPC.DeleteAvatar(ctx, &filepb.DeleteAvatarRequest{
			Filename: user.Avatar,
		})
		if err != nil {
			return err
		}
	}

	return s.userRepo.Delete(ctx, id)
}

func (s *userService) List(ctx context.Context, page, limit int) ([]*user_model.User, error) {
	return s.userRepo.List(ctx, page, limit)
}

func (s *userService) FindMatchingUsers(ctx context.Context, userID string) ([]*user_model.User, error) {

	// мб добавить проверку, что уровень навыка учащегося <= уровню навыка учителя
	// но это очень субъективно, поэтому не факт, что нужно

	currentUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var skillsToLearn []string

	for _, skill := range currentUser.SkillsToLearn {
		skillsToLearn = append(skillsToLearn, skill.Name)
	}

	// Находим пользователей с подходящими навыками
	matchingUsers, err := s.userRepo.FindBySkills(ctx, skillsToLearn)
	if err != nil {
		return nil, err
	}

	// Фильтруем текущего пользователя из результатов
	filteredUsers := make([]*user_model.User, 0)
	for _, u := range matchingUsers {
		if u.ID != currentUser.ID {
			filteredUsers = append(filteredUsers, u)
		}
	}

	return filteredUsers, nil
}

func (s *userService) FindUsersByUsername(ctx context.Context, userID, username string, limit, offset int64) ([]*user_model.User, error) {
	if limit > 1000 || limit <= 0 {
		limit = 1000
	}
	if offset < 0 {
		offset = 0
	}
	return s.userRepo.FindByUsername(ctx, userID, username, limit, offset)
}

// func (s *userService) checkForProfanity(ctx context.Context, fieldName, text string) error {
// 	if text == "" {
// 		return nil
// 	}
// 	hasProfanity, err := s.moderationRepo.CheckSwearing(ctx, text)
// 	if err != nil {
// 		s.logger.Error().Err(err).Str("field", fieldName).Msg("failed to check field for profanity")
// 		return custom_errors.ErrModerationUnavailable
// 	}
// 	if hasProfanity {
// 		return &custom_errors.ProfanityError{FieldName: fieldName}
// 	}
// 	return nil
// }

func (s *userService) checkForProfanity(ctx context.Context, text string) (bool, error) {
	if text == "" {
		return false, nil
	}
	hasProfanity, err := s.moderationRepo.CheckSwearing(ctx, text)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to check field for profanity")
		return false, nil // moderation недоступен — игнорируем
	}
	return hasProfanity, nil
}

func (s *userService) checkUserForProfanity(ctx context.Context, user *user_model.User) error {
	profanityErr := &custom_errors.ProfanityAggregateError{}
	if has, _ := s.checkForProfanity(ctx, user.Username); has {
		profanityErr.Add("Username")
	}
	if has, _ := s.checkForProfanity(ctx, user.Bio); has {
		profanityErr.Add("Bio")
	}
	joinedHrefs := strings.Join(user.Hrefs, " ")
	if has, _ := s.checkForProfanity(ctx, joinedHrefs); has {
		profanityErr.Add("Hrefs")
	}
	if !profanityErr.IsEmpty() {
		return profanityErr
	}
	return nil
}
