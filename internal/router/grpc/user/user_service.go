package user_grpc

import (
	"context"
	"fmt"
	user_model "github.com/Petr09Mitin/xrust-beze-back/internal/models/user"
	user_service "github.com/Petr09Mitin/xrust-beze-back/internal/services/user"
	pb "github.com/Petr09Mitin/xrust-beze-back/proto/user"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UserService представляет gRPC сервис для пользователей
type UserService struct {
	pb.UnimplementedUserServiceServer
	userService user_service.UserService
	logger      zerolog.Logger
}

// NewUserService создает новый gRPC сервис для пользователей
func NewUserService(userService user_service.UserService, logger zerolog.Logger) *UserService {
	return &UserService{
		userService: userService,
		logger:      logger,
	}
}

// CreateUser создает нового пользователя
func (s *UserService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	var skillsToLearn []user_model.Skill
	var skillsToShare []user_model.Skill

	for _, skill := range req.SkillsToLearn {
		skillsToLearn = append(skillsToLearn, user_model.Skill{
			Name:        skill.Name,
			Level:       skill.Level,
			Description: skill.Description,
		})
	}

	for _, skill := range req.SkillsToShare {
		skillsToShare = append(skillsToShare, user_model.Skill{
			Name:        skill.Name,
			Level:       skill.Level,
			Description: skill.Description,
		})
	}

	u := &user_model.User{
		Username:        req.Username,
		Email:           req.Email,
		SkillsToLearn:   skillsToLearn,
		SkillsToShare:   skillsToShare,
		Bio:             req.Bio,
		Avatar:          req.AvatarUrl,
		PreferredFormat: req.PreferredFormat,
	}

	if err := s.userService.Create(ctx, u); err != nil {
		// Проверяем, является ли ошибка ошибкой валидации
		if _, ok := err.(validator.ValidationErrors); ok {
			s.logger.Error().Err(err).Msg("validation err")
			return nil, status.Errorf(codes.InvalidArgument, "validation error: %v", err)
		}
		s.logger.Error().Err(err).Msg("create user err")
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &pb.UserResponse{
		User: convertDomainToProto(u),
	}, nil
}

// GetUser получает пользователя по ID
func (s *UserService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
	u, err := s.userService.GetByID(ctx, req.Id)
	if err != nil {
		s.logger.Error().Err(err).Msg("get user err")
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	return &pb.UserResponse{
		User: convertDomainToProto(u),
	}, nil
}

// UpdateUser обновляет пользователя
func (s *UserService) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	objectID, err := primitive.ObjectIDFromHex(req.Id)
	if err != nil {
		s.logger.Error().Err(err).Msg("invalid ID format")
		return nil, status.Errorf(codes.InvalidArgument, "invalid ID format: %v", err)
	}

	var skillsToLearn []user_model.Skill
	var skillsToShare []user_model.Skill

	for _, skill := range req.SkillsToLearn {
		skillsToLearn = append(skillsToLearn, user_model.Skill{
			Name:        skill.Name,
			Level:       skill.Level,
			Description: skill.Description,
		})
	}

	for _, skill := range req.SkillsToShare {
		skillsToShare = append(skillsToShare, user_model.Skill{
			Name:        skill.Name,
			Level:       skill.Level,
			Description: skill.Description,
		})
	}

	u := &user_model.User{
		ID:              objectID,
		Username:        req.Username,
		Email:           req.Email,
		SkillsToLearn:   skillsToLearn,
		SkillsToShare:   skillsToShare,
		Bio:             req.Bio,
		Avatar:          req.AvatarUrl,
		PreferredFormat: req.PreferredFormat,
	}

	if err := s.userService.Update(ctx, u); err != nil {
		// Проверяем, является ли ошибка ошибкой валидации
		if _, ok := err.(validator.ValidationErrors); ok {
			s.logger.Error().Err(err).Msg("validation err")
			return nil, status.Errorf(codes.InvalidArgument, "validation error: %v", err)
		}
		s.logger.Error().Err(err).Msg("update user err")
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	return &pb.UserResponse{
		User: convertDomainToProto(u),
	}, nil
}

// DeleteUser удаляет пользователя
func (s *UserService) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	if err := s.userService.Delete(ctx, req.Id); err != nil {
		s.logger.Error().Err(err).Msg("delete user err")
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}

	return &pb.DeleteUserResponse{
		Success: true,
	}, nil
}

// ListUsers возвращает список пользователей
func (s *UserService) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	users, err := s.userService.List(ctx, int(req.Page), int(req.Limit))
	if err != nil {
		s.logger.Error().Err(err).Msg("list users err")
		return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}

	var pbUsers []*pb.User
	for _, u := range users {
		pbUsers = append(pbUsers, convertDomainToProto(u))
	}

	return &pb.ListUsersResponse{
		Users: pbUsers,
	}, nil
}

// FindMatchingUsers находит подходящих пользователей
func (s *UserService) FindMatchingUsers(ctx context.Context, req *pb.FindMatchingUsersRequest) (*pb.ListUsersResponse, error) {
	users, err := s.userService.FindMatchingUsers(ctx, req.UserId)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to find matching users")
		return nil, status.Errorf(codes.Internal, "failed to find matching users: %v", err)
	}

	var pbUsers []*pb.User
	for _, u := range users {
		pbUsers = append(pbUsers, convertDomainToProto(u))
	}

	return &pb.ListUsersResponse{
		Users: pbUsers,
	}, nil
}

// convertDomainToProto конвертирует доменную модель в protobuf
func convertDomainToProto(u *user_model.User) *pb.User {
	var skillsToLearn []*pb.Skill
	var skillsToShare []*pb.Skill

	for _, skill := range u.SkillsToLearn {
		skillsToLearn = append(skillsToLearn, &pb.Skill{
			Name:        skill.Name,
			Level:       skill.Level,
			Description: skill.Description,
		})
	}

	for _, skill := range u.SkillsToShare {
		skillsToShare = append(skillsToShare, &pb.Skill{
			Name:        skill.Name,
			Level:       skill.Level,
			Description: skill.Description,
		})
	}

	return &pb.User{
		Id:              u.ID.Hex(),
		Username:        u.Username,
		Email:           u.Email,
		SkillsToLearn:   skillsToLearn,
		SkillsToShare:   skillsToShare,
		Bio:             u.Bio,
		AvatarUrl:       u.Avatar,
		CreatedAt:       timestamppb.New(u.CreatedAt),
		UpdatedAt:       timestamppb.New(u.UpdatedAt),
		LastActiveAt:    timestamppb.New(u.LastActiveAt),
		PreferredFormat: u.PreferredFormat,
	}
}

func ConvertProtoToDomain(u *pb.User) (*user_model.User, error) {
	skillsToLearn := make([]user_model.Skill, 0)
	skillsToShare := make([]user_model.Skill, 0)

	for _, skill := range u.GetSkillsToLearn() {
		skillsToLearn = append(skillsToLearn, user_model.Skill{
			Name:        skill.Name,
			Level:       skill.Level,
			Description: skill.Description,
		})
	}

	for _, skill := range u.GetSkillsToShare() {
		skillsToShare = append(skillsToShare, user_model.Skill{
			Name:        skill.Name,
			Level:       skill.Level,
			Description: skill.Description,
		})
	}

	id, err := primitive.ObjectIDFromHex(u.Id)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &user_model.User{
		ID:              id,
		Username:        u.GetUsername(),
		Email:           u.GetEmail(),
		SkillsToLearn:   skillsToLearn,
		SkillsToShare:   skillsToShare,
		Bio:             u.GetBio(),
		Avatar:          u.GetAvatarUrl(),
		CreatedAt:       u.GetCreatedAt().AsTime(),
		UpdatedAt:       u.GetUpdatedAt().AsTime(),
		LastActiveAt:    u.GetLastActiveAt().AsTime(),
		PreferredFormat: u.GetPreferredFormat(),
	}, nil
}
