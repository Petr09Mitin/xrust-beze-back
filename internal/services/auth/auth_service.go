package auth

import (
	"context"
	"time"

	auth_model "github.com/Petr09Mitin/xrust-beze-back/internal/models/auth"
	custom_errors "github.com/Petr09Mitin/xrust-beze-back/internal/models/error"
	user_model "github.com/Petr09Mitin/xrust-beze-back/internal/models/user"
	session_repo "github.com/Petr09Mitin/xrust-beze-back/internal/repository/auth"
	user_grpc "github.com/Petr09Mitin/xrust-beze-back/internal/router/grpc/user"
	user_pb "github.com/Petr09Mitin/xrust-beze-back/proto/user"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	CreateSession(ctx context.Context, userID string) (*auth_model.Session, error)
	ValidateSession(ctx context.Context, sessionID string) (*auth_model.Session, error)
	DeleteSession(ctx context.Context, sessionID string) error
	// Register(ctx context.Context, req *auth_model.RegisterRequest) error
	Register(ctx context.Context, req *auth_model.RegisterRequest) (*user_model.User, error)
	Login(ctx context.Context, req *auth_model.LoginRequest) (*auth_model.Session, error)

	TestUserConnection(ctx context.Context) ([]*user_model.User, error)
}

type authService struct {
	sessionRepo session_repo.SessionRepository
	userGRPC    user_pb.UserServiceClient
	logger      zerolog.Logger
	sessionTTL  time.Duration
}

func NewAuthService(sessionRepo session_repo.SessionRepository, userClient user_pb.UserServiceClient,
	logger zerolog.Logger, sessionTTL time.Duration) AuthService {
	return &authService{
		sessionRepo: sessionRepo,
		userGRPC:    userClient,
		logger:      logger,
		sessionTTL:  sessionTTL,
	}
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func isPasswordCorrect(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (s *authService) CreateSession(ctx context.Context, userID string) (*auth_model.Session, error) {
	now := time.Now()
	sess := &auth_model.Session{
		ID:        uuid.New().String(),
		UserID:    userID,
		CreatedAt: now,
		ExpiresAt: now.Add(s.sessionTTL),
	}

	if err := s.sessionRepo.Create(ctx, sess); err != nil {
		return nil, err
	}

	return sess, nil
}

func (s *authService) ValidateSession(ctx context.Context, sessionID string) (*auth_model.Session, error) {
	sess, err := s.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if sess == nil || time.Now().After(sess.ExpiresAt) {
		return nil, nil
	}

	return sess, nil
}

func (s *authService) DeleteSession(ctx context.Context, sessionID string) error {
	return s.sessionRepo.Delete(ctx, sessionID)
}

// func (s *authService) Register(ctx context.Context, req *auth.RegisterRequest) error {
// 	hashedPassword, err := hashPassword(req.Password)
// 	if err != nil {
// 		return err
// 	}

// 	return s.userGRPC.CreateUser(ctx, req.Username, req.Email, hashedPassword)
// }

func (s *authService) Register(ctx context.Context, req *auth_model.RegisterRequest) (*user_model.User, error) {
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	createReq := convertUserToCreateUserRequest(req.User, hashedPassword)

	resp, err := s.userGRPC.CreateUser(ctx, createReq)
	if err != nil {
		return nil, err
	}

	user, err := user_grpc.ConvertProtoToDomain(resp.User)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *authService) Login(ctx context.Context, req *auth_model.LoginRequest) (*auth_model.Session, error) {
	user, err := s.userGRPC.GetUserByEmailToLogin(ctx, &user_pb.GetUserByEmailRequest{
		Email: req.Email,
	})
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, custom_errors.ErrUserNotExists
	}
	if !isPasswordCorrect(req.Password, user.UserToLogin.Password) {
		return nil, custom_errors.ErrWrongPassword
	}
	return s.CreateSession(ctx, user.UserToLogin.Id)
}

func (s *authService) TestUserConnection(ctx context.Context) ([]*user_model.User, error) {
	resp, err := s.userGRPC.ListUsers(ctx, &user_pb.ListUsersRequest{
		Page:  1,
		Limit: 10,
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to list users")
		return nil, err
	}

	var users []*user_model.User
	for _, u := range resp.Users {
		user, err := user_grpc.ConvertProtoToDomain(u)
		if err != nil {
			s.logger.Error().Err(err).Msg("failed to convert user")
			continue
		}
		users = append(users, user)
	}

	return users, nil
}

func convertUserToCreateUserRequest(user user_model.User, hashedPassword string) *user_pb.CreateUserRequest {
	convertSkills := func(skills []user_model.Skill) []*user_pb.Skill {
		protoSkills := make([]*user_pb.Skill, 0, len(skills))
		for _, skill := range skills {
			protoSkills = append(protoSkills, &user_pb.Skill{
				Name:        skill.Name,
				Level:       skill.Level,
				Description: skill.Description,
			})
		}
		return protoSkills
	}

	return &user_pb.CreateUserRequest{
		Username:        user.Username,
		Email:           user.Email,
		HashedPassword:  hashedPassword,
		Bio:             user.Bio,
		AvatarUrl:       user.Avatar,
		PreferredFormat: user.PreferredFormat,
		SkillsToLearn:   convertSkills(user.SkillsToLearn),
		SkillsToShare:   convertSkills(user.SkillsToShare),
	}
}
