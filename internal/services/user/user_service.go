package user_service

import (
	"context"
	"errors"
	"strings"
	"time"

	review_repo "github.com/Petr09Mitin/xrust-beze-back/internal/repository/review"

	custom_errors "github.com/Petr09Mitin/xrust-beze-back/internal/models/error"
	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/defaults"
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
	CreateReview(ctx context.Context, review *user_model.Review) (*user_model.Review, error)
	UpdateReview(ctx context.Context, userID string, review *user_model.Review) (*user_model.Review, error)
	DeleteReview(ctx context.Context, userID string, reviewID string) error
	FindBySkillsToShare(ctx context.Context, skills []string, limit, offset int64) ([]*user_model.User, error)
	FindBySkillsToLearn(ctx context.Context, skills []string, limit, offset int64) ([]*user_model.User, error)
}

type userService struct {
	userRepo       user_repo.UserRepo
	moderationRepo moderation_repo.ModerationRepo
	reviewRepo     review_repo.ReviewRepo
	fileGRPC       filepb.FileServiceClient
	authGRPC       authpb.AuthServiceClient
	timeout        time.Duration
	logger         zerolog.Logger
}

func NewUserService(userRepo user_repo.UserRepo, moderationRepo moderation_repo.ModerationRepo, reviewRepo review_repo.ReviewRepo, fileGRPC filepb.FileServiceClient, authGRPC authpb.AuthServiceClient, timeout time.Duration, logger zerolog.Logger) UserService {
	return &userService{
		userRepo:       userRepo,
		moderationRepo: moderationRepo,
		reviewRepo:     reviewRepo,
		fileGRPC:       fileGRPC,
		authGRPC:       authGRPC,
		timeout:        timeout,
		logger:         logger,
	}
}

func (s *userService) Create(ctx context.Context, user *user_model.User, hashedPassword string) error {
	//if err := s.checkUserForProfanity(ctx, user); err != nil {
	//	return err
	//}
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
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	user, err = s.fillUserReviews(ctx, user)
	if err != nil {
		return nil, err
	}
	user.Avatar = defaults.ApplyDefaultIfEmptyAvatar(user.Avatar)
	return user, nil
}

func (s *userService) GetByEmail(ctx context.Context, email string) (*user_model.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	user, err = s.fillUserReviews(ctx, user)
	if err != nil {
		return nil, err
	}
	user.Avatar = defaults.ApplyDefaultIfEmptyAvatar(user.Avatar)
	return user, nil
}

func (s *userService) GetByEmailWithPassword(ctx context.Context, email string) (*auth_model.RegisterRequest, error) {
	return s.userRepo.GetByEmailWithPassword(ctx, email)
}

func (s *userService) GetByUsername(ctx context.Context, username string) (*user_model.User, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	user, err = s.fillUserReviews(ctx, user)
	if err != nil {
		return nil, err
	}
	user.Avatar = defaults.ApplyDefaultIfEmptyAvatar(user.Avatar)
	return user, nil
}

func (s *userService) GetByUsernameWithPassword(ctx context.Context, username string) (*auth_model.RegisterRequest, error) {
	return s.userRepo.GetByUsernameWithPassword(ctx, username)
}

func (s *userService) Update(ctx context.Context, user *user_model.User) error {
	//if err := s.checkUserForProfanity(ctx, user); err != nil {
	//	return err
	//}
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

	if existingUser.Avatar != user.Avatar && user.Avatar != defaults.DefaultAvatarPath {
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

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return err
	}
	user, err = s.fillUserReviews(ctx, user)
	if err != nil {
		return err
	}
	return nil
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
	users, err := s.userRepo.List(ctx, page, limit)
	if err != nil {
		return nil, err
	}
	err = s.fillUsersRatings(ctx, users)
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		user.Avatar = defaults.ApplyDefaultIfEmptyAvatar(user.Avatar)
	}
	return users, nil
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
			u.Avatar = defaults.ApplyDefaultIfEmptyAvatar(u.Avatar)
			filteredUsers = append(filteredUsers, u)
		}
	}

	err = s.fillUsersRatings(ctx, filteredUsers)
	if err != nil {
		return nil, err
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
	users, err := s.userRepo.FindByUsername(ctx, userID, username, limit, offset)
	if err != nil {
		return nil, err
	}
	err = s.fillUsersRatings(ctx, users)
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		user.Avatar = defaults.ApplyDefaultIfEmptyAvatar(user.Avatar)
	}
	return users, nil
}

//func (s *userService) checkForProfanity(ctx context.Context, fieldName, text string) error {
//	if text == "" {
//		return nil
//	}
//	hasProfanity, err := s.moderationRepo.CheckSwearing(ctx, text)
//	if err != nil {
//		s.logger.Error().Err(err).Str("field", fieldName).Msg("failed to check field for profanity")
//		return custom_errors.ErrModerationUnavailable
//	}
//	if hasProfanity {
//		return &custom_errors.ProfanityError{FieldName: fieldName}
//	}
//	return nil
//}

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

func (s *userService) CreateReview(ctx context.Context, review *user_model.Review) (*user_model.Review, error) {
	_, err := s.userRepo.GetByID(ctx, review.UserIDBy)
	if err != nil {
		return nil, custom_errors.ErrUserNotExists
	}
	_, err = s.userRepo.GetByID(ctx, review.UserIDTo)
	if err != nil {
		return nil, custom_errors.ErrUserNotExists
	}
	_, err = s.reviewRepo.GetByUserIDByAndUserIDTo(ctx, review.UserIDBy, review.UserIDTo)
	if !errors.Is(err, custom_errors.ErrNotFound) {
		return nil, custom_errors.ErrDuplicateReview
	}
	created := time.Now().Unix()
	review.Created = created
	review.Updated = created
	newReview, err := s.reviewRepo.Create(ctx, review)
	if err != nil {
		return nil, err
	}
	newReview.UserBy, err = s.userRepo.GetByID(ctx, newReview.UserIDBy)
	if err != nil {
		return nil, err
	}
	return newReview, nil
}

func (s *userService) fillUserReviews(ctx context.Context, user *user_model.User) (*user_model.User, error) {
	reviews, err := s.reviewRepo.GetReviewsByUserIDTo(ctx, user.ID.Hex())
	if err != nil {
		return nil, err
	}
	if len(reviews) == 0 {
		reviews = make([]*user_model.Review, 0)
	}
	for _, review := range reviews {
		review.UserBy, err = s.userRepo.GetByID(ctx, review.UserIDBy)
		if err != nil {
			s.logger.Error().Err(err).Msg("failed to get user by id")
			return nil, err
		}
	}
	user.Reviews = reviews
	ratings, err := s.reviewRepo.GetAvgRatingsByUserIDs(ctx, []string{user.ID.Hex()})
	if err != nil {
		return nil, err
	}
	// если у юзера нет отзывов, и, соответственно, нет значения в мапе ratings - проставится null value = 0
	user.Rating = ratings[user.ID.Hex()]
	return user, nil
}

func (s *userService) fillUsersRatings(ctx context.Context, users []*user_model.User) error {
	if len(users) == 0 {
		return nil
	}
	userIDs := make([]string, len(users))
	for i, user := range users {
		userIDs[i] = user.ID.Hex()
	}
	ratings, err := s.reviewRepo.GetAvgRatingsByUserIDs(ctx, userIDs)
	if err != nil {
		return err
	}
	for _, user := range users {
		user.Rating = ratings[user.ID.Hex()]
	}
	return nil
}

func (s *userService) UpdateReview(ctx context.Context, userID string, review *user_model.Review) (*user_model.Review, error) {
	oldReview, err := s.reviewRepo.GetByID(ctx, review.ID)
	if err != nil {
		return nil, err
	}
	if oldReview.UserIDBy != userID {
		return nil, custom_errors.ErrUnauthorized
	}
	oldReview.Text = review.Text
	oldReview.Rating = review.Rating
	oldReview.Updated = time.Now().Unix()
	err = s.reviewRepo.Update(ctx, oldReview)
	if err != nil {
		return nil, err
	}
	oldReview.UserBy, err = s.userRepo.GetByID(ctx, oldReview.UserIDBy)
	if err != nil {
		return nil, err
	}
	return oldReview, nil
}

func (s *userService) DeleteReview(ctx context.Context, userID string, reviewID string) error {
	oldReview, err := s.reviewRepo.GetByID(ctx, reviewID)
	if err != nil {
		return err
	}
	if oldReview.UserIDBy != userID {
		return custom_errors.ErrUnauthorized
	}
	err = s.reviewRepo.DeleteByID(ctx, reviewID)
	if err != nil {
		return err
	}
	return nil
}

func (s *userService) FindBySkillsToShare(ctx context.Context, skills []string, limit, offset int64) ([]*user_model.User, error) {
	users, err := s.userRepo.FindBySkillsToShare(ctx, skills, limit, offset)
	if err != nil {
		return nil, err
	}
	for i := range users {
		users[i], err = s.fillUserReviews(ctx, users[i])
		if err != nil {
			return nil, err
		}
		users[i].Avatar = defaults.ApplyDefaultIfEmptyAvatar(users[i].Avatar)
	}
	return users, nil
}

func (s *userService) FindBySkillsToLearn(ctx context.Context, skills []string, limit, offset int64) ([]*user_model.User, error) {
	users, err := s.userRepo.FindBySkillsToLearn(ctx, skills, limit, offset)
	if err != nil {
		return nil, err
	}
	for i := range users {
		users[i], err = s.fillUserReviews(ctx, users[i])
		if err != nil {
			return nil, err
		}
		users[i].Avatar = defaults.ApplyDefaultIfEmptyAvatar(users[i].Avatar)
	}
	return users, nil
}
