package study_material_service

import (
	"context"

	custom_errors "github.com/Petr09Mitin/xrust-beze-back/internal/models/error"
	study_material_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/study_material"
	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/defaults"
	study_material_repo "github.com/Petr09Mitin/xrust-beze-back/internal/repository/study_material"
	filepb "github.com/Petr09Mitin/xrust-beze-back/proto/file"
	userpb "github.com/Petr09Mitin/xrust-beze-back/proto/user"
	"github.com/rs/zerolog"
)

type StudyMaterialAPIService interface {
	GetByID(ctx context.Context, id string) (*study_material_models.StudyMaterial, error)
	GetByTags(ctx context.Context, tag []string) ([]*study_material_models.StudyMaterial, error)
	GetByName(ctx context.Context, name string) ([]*study_material_models.StudyMaterial, error)
	GetByAuthorID(ctx context.Context, authorID string) ([]*study_material_models.StudyMaterial, error)
	Delete(ctx context.Context, materialID string, userID string) error
}

type studyMaterialAPIService struct {
	studyMaterialRepo study_material_repo.StudyMaterialAPIRepository
	userGRPC          userpb.UserServiceClient
	fileGRPC          filepb.FileServiceClient
	logger            zerolog.Logger
}

func NewStudyMaterialAPIService(studyMaterialRepo study_material_repo.StudyMaterialAPIRepository, userGRPC userpb.UserServiceClient,
	fileGRPC filepb.FileServiceClient, logger zerolog.Logger) StudyMaterialAPIService {
	return &studyMaterialAPIService{
		studyMaterialRepo: studyMaterialRepo,
		userGRPC:          userGRPC,
		fileGRPC:          fileGRPC,
		logger:            logger,
	}
}

func (s *studyMaterialAPIService) GetByID(ctx context.Context, id string) (*study_material_models.StudyMaterial, error) {
	material, err := s.studyMaterialRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	materials, err := s.addAuthorsInfo(ctx, []*study_material_models.StudyMaterial{material})
	if err != nil {
		return nil, err
	}
	return materials[0], nil
}

func (s *studyMaterialAPIService) GetByTags(ctx context.Context, tags []string) ([]*study_material_models.StudyMaterial, error) {
	materials, err := s.studyMaterialRepo.GetByTags(ctx, tags)
	if err != nil {
		return nil, err
	}
	return s.addAuthorsInfo(ctx, materials)
}

func (s *studyMaterialAPIService) GetByName(ctx context.Context, name string) ([]*study_material_models.StudyMaterial, error) {
	materials, err := s.studyMaterialRepo.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return s.addAuthorsInfo(ctx, materials)
}

func (s *studyMaterialAPIService) GetByAuthorID(ctx context.Context, authorID string) ([]*study_material_models.StudyMaterial, error) {
	materials, err := s.studyMaterialRepo.GetByAuthorID(ctx, authorID)
	if err != nil {
		return nil, err
	}
	return s.addAuthorsInfo(ctx, materials)
}

func (s *studyMaterialAPIService) Delete(ctx context.Context, materialID string, userID string) error {
	// Получаем материал, чтобы проверить, имеет ли пользователь права на удаление
	material, err := s.studyMaterialRepo.GetByID(ctx, materialID)
	if err != nil {
		return err
	}
	if material.AuthorID != userID {
		return custom_errors.ErrUserIDMismatch
	}

	_, err = s.fileGRPC.DeleteAttachments(ctx, &filepb.DeleteAttachmentsRequest{
		Filenames: []string{material.Filename},
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to delete attachment file")
		// Продолжаем удаление из БД, даже если не удалось удалить файл
	}

	return s.studyMaterialRepo.Delete(ctx, materialID)
}

func (s *studyMaterialAPIService) fillAuthorInfo(ctx context.Context, mat *study_material_models.StudyMaterial) {
	userResp, err := s.userGRPC.GetUserByID(ctx, &userpb.GetUserByIDRequest{Id: mat.AuthorID})
	if err == nil && userResp.User != nil {
		mat.Author.Username = userResp.User.Username
		mat.Author.Avatar = userResp.User.AvatarUrl
		mat.Author.Avatar = defaults.ApplyDefaultIfEmptyAvatar(mat.Author.Avatar)
	}

	// Код ниже не реализован, чтобы не возвращать ошибку "not found", если пользователь не найден
	// Если пользователь не найден, то поле Author будет пустым

	// if err != nil {
	// 	s.logger.Error().Err(err).Msg("failed to get user by id")
	// 	return err
	// }
	// if userResp.User != nil {
	// 	mat.Author.Username = userResp.User.Username
	// 	mat.Author.Avatar = userResp.User.AvatarUrl
	// }
	// return nil
}

func (s *studyMaterialAPIService) addAuthorsInfo(ctx context.Context, materials []*study_material_models.StudyMaterial) ([]*study_material_models.StudyMaterial, error) {
	if len(materials) == 0 {
		return materials, custom_errors.ErrNotFound
	}
	// Для первого материала
	result := []*study_material_models.StudyMaterial{materials[0]}
	s.fillAuthorInfo(ctx, materials[0])
	// Для остальных материалов
	for i := 1; i < len(materials); i++ {
		if materials[i].AuthorID == materials[i-1].AuthorID {
			materials[i].Author = materials[i-1].Author
		} else {
			s.fillAuthorInfo(ctx, materials[i])
		}
		result = append(result, materials[i])
	}
	return result, nil
}
