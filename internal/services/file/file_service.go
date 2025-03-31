package file

import (
	"context"
	custom_errors "github.com/Petr09Mitin/xrust-beze-back/internal/models/error"
	filerepo "github.com/Petr09Mitin/xrust-beze-back/internal/repository/file"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"path"
)

type FileService interface {
	UploadTempFile(ctx context.Context, filepath string) (filename string, err error)
	MoveTempFileToAvatars(ctx context.Context, filename string) (err error)
	DeleteAvatar(ctx context.Context, filename string) (err error)
}

type FileServiceImpl struct {
	fileRepo filerepo.FileRepo
	logger   zerolog.Logger
}

func NewFileService(fileRepo filerepo.FileRepo, logger zerolog.Logger) FileService {
	return &FileServiceImpl{
		fileRepo: fileRepo,
		logger:   logger,
	}
}

func (f *FileServiceImpl) UploadTempFile(ctx context.Context, filepath string) (string, error) {
	filename := uuid.New().String() + path.Ext(filepath)
	return filename, f.fileRepo.UploadTemp(ctx, filepath, filename)
}

func (f *FileServiceImpl) MoveTempFileToAvatars(ctx context.Context, filename string) (err error) {
	exists, err := f.fileRepo.CheckIfTempExists(ctx, filename)
	if err != nil {
		return err
	}
	if !exists {
		return custom_errors.ErrFileNotFound
	}
	err = f.fileRepo.CopyFromTempToAvatars(ctx, filename)
	if err != nil {
		return err
	}
	err = f.fileRepo.DeleteTemp(ctx, filename)
	if err != nil {
		return err
	}
	return nil
}

func (f *FileServiceImpl) DeleteAvatar(ctx context.Context, filename string) (err error) {
	exists, err := f.fileRepo.CheckIfAvatarExists(ctx, filename)
	if err != nil {
		return err
	}
	if !exists {
		return custom_errors.ErrFileNotFound
	}
	err = f.fileRepo.DeleteAvatar(ctx, filename)
	if err != nil {
		return err
	}
	return nil
}
