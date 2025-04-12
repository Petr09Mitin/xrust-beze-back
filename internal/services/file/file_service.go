package file

import (
	"context"
	custom_errors "github.com/Petr09Mitin/xrust-beze-back/internal/models/error"
	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/validation"
	filerepo "github.com/Petr09Mitin/xrust-beze-back/internal/repository/file"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"path"
)

type FileService interface {
	UploadTempFile(ctx context.Context, filepath string) (filename string, err error)
	MoveTempFileToAvatars(ctx context.Context, filename string) (err error)
	DeleteAvatar(ctx context.Context, filename string) (err error)
	MoveTempFileToVoiceMessages(ctx context.Context, filename string) (err error)
	DeleteVoiceMessage(ctx context.Context, filename string) (err error)
	MoveTempFilesToAttachments(ctx context.Context, filenames []string) (err error)
	DeleteAttachments(ctx context.Context, filenames []string) (err error)
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

func (f *FileServiceImpl) MoveTempFileToAvatars(ctx context.Context, filename string) error {
	valid := validation.IsValidImageFilepath(filename)
	if !valid {
		return custom_errors.ErrInvalidFileFormat
	}
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

func (f *FileServiceImpl) DeleteAvatar(ctx context.Context, filename string) error {
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

func (f *FileServiceImpl) MoveTempFileToVoiceMessages(ctx context.Context, filename string) error {
	valid := validation.IsValidVoiceMessageExt(filename)
	if !valid {
		return custom_errors.ErrInvalidFileFormat
	}
	exists, err := f.fileRepo.CheckIfTempExists(ctx, filename)
	if err != nil {
		return err
	}
	if !exists {
		return custom_errors.ErrFileNotFound
	}
	err = f.fileRepo.CopyFromTempToVoiceMessages(ctx, filename)
	if err != nil {
		return err
	}
	err = f.fileRepo.DeleteTemp(ctx, filename)
	if err != nil {
		return err
	}
	return nil
}

func (f *FileServiceImpl) DeleteVoiceMessage(ctx context.Context, filename string) error {
	exists, err := f.fileRepo.CheckIfVoiceMessageExists(ctx, filename)
	if err != nil {
		return err
	}
	if !exists {
		f.logger.Error().Err(err).Msg("file does not exist")
		return custom_errors.ErrFileNotFound
	}
	err = f.fileRepo.DeleteVoiceMessage(ctx, filename)
	if err != nil {
		f.logger.Error().Err(err).Msg("failed to delete file")
		return err
	}
	return nil
}

func (f *FileServiceImpl) MoveTempFilesToAttachments(ctx context.Context, filenames []string) error {
	for _, filename := range filenames {
		exists, err := f.fileRepo.CheckIfTempExists(ctx, filename)
		if err != nil {
			return err
		}
		if !exists {
			return custom_errors.ErrFileNotFound
		}
	}

	err := f.fileRepo.CopyFromTempToAttachments(ctx, filenames)
	if err != nil {
		return err
	}

	for _, filename := range filenames {
		err = f.fileRepo.DeleteTemp(ctx, filename)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *FileServiceImpl) DeleteAttachments(ctx context.Context, filenames []string) error {
	for _, filename := range filenames {
		exists, err := f.fileRepo.CheckIfAttachmentExists(ctx, filename)
		if err != nil {
			return err
		}
		if !exists {
			f.logger.Error().Err(err).Msg("files do not exist")
			return custom_errors.ErrFileNotFound
		}
	}
	err := f.fileRepo.DeleteAttachments(ctx, filenames)
	if err != nil {
		f.logger.Error().Err(err).Msg("failed to delete files")
		return err
	}
	return nil
}
