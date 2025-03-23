package user

import (
	"errors"
	"time"

	user_model "github.com/Petr09Mitin/xrust-beze-back/internal/models/user"
)

type userService struct {
	userRepo user_model.Repository
	timeout  time.Duration
}

// NewUserService создает новый сервис для пользователей
func NewUserService(userRepo user_model.Repository, timeout time.Duration) user_model.Service {
	return &userService{
		userRepo: userRepo,
		timeout:  timeout,
	}
}

// Create создает нового пользователя
func (s *userService) Create(user *user_model.User) error {
	// Проверка на уникальность email и username
	existingUser, err := s.userRepo.GetByEmail(user.Email)
	if err == nil && existingUser != nil {
		return errors.New("email already exists")
	}

	existingUser, err = s.userRepo.GetByUsername(user.Username)
	if err == nil && existingUser != nil {
		return errors.New("username already exists")
	}

	// Валидация полей
	if user.Username == "" {
		return errors.New("username is required")
	}
	if user.Email == "" {
		return errors.New("email is required")
	}

	return s.userRepo.Create(user)
}

// GetByID получает пользователя по ID
func (s *userService) GetByID(id string) (*user_model.User, error) {
	return s.userRepo.GetByID(id)
}

// GetByEmail получает пользователя по email
func (s *userService) GetByEmail(email string) (*user_model.User, error) {
	return s.userRepo.GetByEmail(email)
}

// GetByUsername получает пользователя по имени пользователя
func (s *userService) GetByUsername(username string) (*user_model.User, error) {
	return s.userRepo.GetByUsername(username)
}

// Update обновляет пользователя
func (s *userService) Update(user *user_model.User) error {
	// Проверяем существование пользователя
	existingUser, err := s.userRepo.GetByID(user.ID.Hex())
	if err != nil {
		return err
	}

	// Проверка на уникальность email и username, если они изменились
	if user.Email != existingUser.Email {
		checkUser, err := s.userRepo.GetByEmail(user.Email)
		if err == nil && checkUser != nil {
			return errors.New("email already exists")
		}
	}

	if user.Username != existingUser.Username {
		checkUser, err := s.userRepo.GetByUsername(user.Username)
		if err == nil && checkUser != nil {
			return errors.New("username already exists")
		}
	}

	return s.userRepo.Update(user)
}

// Delete удаляет пользователя
func (s *userService) Delete(id string) error {
	// Проверяем существование пользователя
	_, err := s.userRepo.GetByID(id)
	if err != nil {
		return err
	}

	return s.userRepo.Delete(id)
}

// List возвращает список пользователей с пагинацией
func (s *userService) List(page, limit int) ([]*user_model.User, error) {
	return s.userRepo.List(page, limit)
}

// FindMatchingUsers находит подходящих пользователей для обмена знаниями
func (s *userService) FindMatchingUsers(userID string) ([]*user_model.User, error) {
	// Получаем пользователя
	currentUser, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// Извлекаем навыки пользователя
	var skillsToLearn []string
	var skillsToShare []string

	for _, skill := range currentUser.SkillsToLearn {
		skillsToLearn = append(skillsToLearn, skill.Name)
	}

	for _, skill := range currentUser.SkillsToShare {
		skillsToShare = append(skillsToShare, skill.Name)
	}

	// Находим пользователей с подходящими навыками
	matchingUsers, err := s.userRepo.FindBySkills(skillsToLearn, skillsToShare)
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