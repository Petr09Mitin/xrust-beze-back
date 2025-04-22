package study_material_grpc

import (
	"context"

	custom_errors "github.com/Petr09Mitin/xrust-beze-back/internal/models/error"
	study_material_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/study_material"
	study_material_service "github.com/Petr09Mitin/xrust-beze-back/internal/services/study_material"
	pb "github.com/Petr09Mitin/xrust-beze-back/proto/study_material"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StudyMaterialServer struct {
	pb.UnimplementedStudyMaterialServiceServer
	service study_material_service.StudyMaterialService
	logger  zerolog.Logger
}

func NewStudyMaterialServer(service study_material_service.StudyMaterialService, logger zerolog.Logger) *StudyMaterialServer {
	return &StudyMaterialServer{
		service: service,
		logger:  logger,
	}
}

func (s *StudyMaterialServer) GetStudyMaterialByID(ctx context.Context, req *pb.GetStudyMaterialByIDRequest) (*pb.StudyMaterialResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	material, err := s.service.GetByID(ctx, req.Id)
	if err != nil {
		if err == custom_errors.ErrNotFound {
			return nil, status.Error(codes.NotFound, "study material not found")
		}
		s.logger.Error().Err(err).Str("id", req.Id).Msg("Failed to get study material by ID")
		return nil, status.Error(codes.Internal, "failed to get study material")
	}

	return &pb.StudyMaterialResponse{
		StudyMaterial: convertToProtoStudyMaterial(material),
	}, nil
}

func (s *StudyMaterialServer) GetStudyMaterialsByTags(ctx context.Context, req *pb.GetStudyMaterialsByTagsRequest) (*pb.StudyMaterialListResponse, error) {
	if len(req.Tags) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one tag is required")
	}

	materials, err := s.service.GetByTags(ctx, req.Tags)
	if err != nil {
		if err == custom_errors.ErrNotFound {
			return &pb.StudyMaterialListResponse{
				StudyMaterials: []*pb.StudyMaterial{},
			}, nil
		}
		s.logger.Error().Err(err).Interface("tags", req.Tags).Msg("Failed to get study materials by tags")
		return nil, status.Error(codes.Internal, "failed to get study materials by tags")
	}

	return &pb.StudyMaterialListResponse{
		StudyMaterials: convertToProtoStudyMaterialList(materials),
	}, nil
}

func (s *StudyMaterialServer) GetStudyMaterialsByName(ctx context.Context, req *pb.GetStudyMaterialsByNameRequest) (*pb.StudyMaterialListResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	materials, err := s.service.GetByName(ctx, req.Name)
	if err != nil {
		if err == custom_errors.ErrNotFound {
			return &pb.StudyMaterialListResponse{
				StudyMaterials: []*pb.StudyMaterial{},
			}, nil
		}
		s.logger.Error().Err(err).Str("name", req.Name).Msg("Failed to get study materials by name")
		return nil, status.Error(codes.Internal, "failed to get study materials by name")
	}

	return &pb.StudyMaterialListResponse{
		StudyMaterials: convertToProtoStudyMaterialList(materials),
	}, nil
}

func (s *StudyMaterialServer) GetStudyMaterialsByAuthorID(ctx context.Context, req *pb.GetStudyMaterialsByAuthorIDRequest) (*pb.StudyMaterialListResponse, error) {
	if req.AuthorId == "" {
		return nil, status.Error(codes.InvalidArgument, "author_id is required")
	}

	materials, err := s.service.GetByAuthorID(ctx, req.AuthorId)
	if err != nil {
		if err == custom_errors.ErrNotFound {
			return &pb.StudyMaterialListResponse{
				StudyMaterials: []*pb.StudyMaterial{},
			}, nil
		}
		s.logger.Error().Err(err).Str("author_id", req.AuthorId).Msg("Failed to get study materials by author ID")
		return nil, status.Error(codes.Internal, "failed to get study materials by author ID")
	}

	return &pb.StudyMaterialListResponse{
		StudyMaterials: convertToProtoStudyMaterialList(materials),
	}, nil
}

func (s *StudyMaterialServer) DeleteStudyMaterial(ctx context.Context, req *pb.DeleteStudyMaterialRequest) (*pb.DeleteStudyMaterialResponse, error) {
	if req.Id == "" || req.AuthorId == "" {
		return nil, status.Error(codes.InvalidArgument, "id and author_id are required")
	}

	err := s.service.Delete(ctx, req.Id, req.AuthorId)
	if err != nil {
		if err == custom_errors.ErrUserIDMismatch {
			return nil, status.Error(codes.PermissionDenied, "not authorized to delete this study material")
		}
		if err == custom_errors.ErrNotFound {
			return nil, status.Error(codes.NotFound, "study material not found")
		}

		s.logger.Error().Err(err).Str("id", req.Id).Str("author_id", req.AuthorId).Msg("Failed to delete study material")
		return nil, status.Error(codes.Internal, "failed to delete study material")
	}

	return &pb.DeleteStudyMaterialResponse{
		Success: true,
	}, nil
}

// Вспомогательные функции для конвертации моделей
func convertToProtoStudyMaterial(material *study_material_models.StudyMaterial) *pb.StudyMaterial {
	if material == nil {
		return nil
	}

	return &pb.StudyMaterial{
		Id:       material.ID,
		Name:     material.Name,
		Filename: material.Filename,
		Tags:     material.Tags,
		AuthorId: material.AuthorID,
		Author: &pb.User{
			Id:       material.Author.ID.Hex(),
			Username: material.Author.Username,
			Email:    material.Author.Email,
			Avatar:   material.Author.Avatar,
		},
		Created: material.Created,
		Updated: material.Updated,
	}
}

func convertToProtoStudyMaterialList(materials []*study_material_models.StudyMaterial) []*pb.StudyMaterial {
	if materials == nil {
		return nil
	}

	result := make([]*pb.StudyMaterial, len(materials))
	for i, material := range materials {
		result[i] = convertToProtoStudyMaterial(material)
	}

	return result
}
