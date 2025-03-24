package user

import (
	"context"
	"errors"
	"time"

	user_model "github.com/Petr09Mitin/xrust-beze-back/internal/models/user"
	user_repo "github.com/Petr09Mitin/xrust-beze-back/internal/repository"
)

type UserService interface {
	Create(ctx context.Context, user *user_model.User) error
	GetByID(ctx context.Context, id string) (*user_model.User, error)
	GetByEmail(ctx context.Context, email string) (*user_model.User, error)
	GetByUsername(ctx context.Context, username string) (*user_model.User, error)
	Update(ctx context.Context, user *user_model.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, page, limit int) ([]*user_model.User, error)
	FindMatchingUsers(ctx context.Context, userID string) ([]*user_model.User, error)
}

type userService struct {
	userRepo user_repo.UserRepo
	timeout  time.Duration
}

func NewUserService(userRepo user_repo.UserRepo, timeout time.Duration) UserService {
	return &userService{
		userRepo: userRepo,
		timeout:  timeout,
	}
}

func (s *userService) Create(ctx context.Context, user *user_model.User) error {
	// Валидация пользователя
	if err := user.Validate(); err != nil {
		return err
	}

	// Проверка на уникальность email и username
	existingUser, err := s.userRepo.GetByEmail(ctx, user.Email)
	if err == nil && existingUser != nil {
		return errors.New("email already exists")
	}

	existingUser, err = s.userRepo.GetByUsername(ctx, user.Username)
	if err == nil && existingUser != nil {
		return errors.New("username already exists")
	}

	return s.userRepo.Create(ctx, user)
}

func (s *userService) GetByID(ctx context.Context, id string) (*user_model.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *userService) GetByEmail(ctx context.Context, email string) (*user_model.User, error) {
	return s.userRepo.GetByEmail(ctx, email)
}

func (s *userService) GetByUsername(ctx context.Context, username string) (*user_model.User, error) {
	return s.userRepo.GetByUsername(ctx, username)
}

func (s *userService) Update(ctx context.Context, user *user_model.User) error {

	// добавить проверку соответствия id авторизованного пользователя и того, что хотим удалить

	if err := user.Validate(); err != nil {
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
			return errors.New("email already exists")
		}
	}

	if existingUser.Username != user.Username {
		userWithUsername, err := s.userRepo.GetByUsername(ctx, user.Username)
		if err == nil && userWithUsername != nil && userWithUsername.ID != user.ID {
			return errors.New("username already exists")
		}
	}

	user.UpdatedAt = time.Now()

	return s.userRepo.Update(ctx, user)
}

func (s *userService) Delete(ctx context.Context, id string) error {

	// добавить проверку соответствия id авторизованного пользователя и того, что хотим удалить

	// Проверяем существование пользователя
	_, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
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
	var skillsToShare []string

	for _, skill := range currentUser.SkillsToLearn {
		skillsToLearn = append(skillsToLearn, skill.Name)
	}

	for _, skill := range currentUser.SkillsToShare {
		skillsToShare = append(skillsToShare, skill.Name)
	}

	// Находим пользователей с подходящими навыками
	matchingUsers, err := s.userRepo.FindBySkills(ctx, skillsToLearn, skillsToShare)
	if err != nil {
		return nil, err
	}

	// Фильтруем текущего пользователя из результатов
	var filteredUsers []*user_model.User
	for _, u := range matchingUsers {
		if u.ID != currentUser.ID {
			filteredUsers = append(filteredUsers, u)
		}
	}

	return filteredUsers, nil
}
