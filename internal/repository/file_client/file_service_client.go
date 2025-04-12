package file_client

import (
	"context"
	filepb "github.com/Petr09Mitin/xrust-beze-back/proto/file"
	"github.com/rs/zerolog"
)

type FileServiceClient interface {
	MoveTempFileToAvatars(ctx context.Context, filename string) (string, error)
	DeleteAvatar(ctx context.Context, filename string) error
	MoveTempFileToVoiceMessages(ctx context.Context, filename string) (string, error)
	DeleteVoiceMessage(ctx context.Context, filename string) error
}

type FileServiceClientImpl struct {
	fileGRPC filepb.FileServiceClient
	logger   zerolog.Logger
}

func NewFileServiceClient(fileGRPC filepb.FileServiceClient, logger zerolog.Logger) FileServiceClient {
	return &FileServiceClientImpl{
		fileGRPC: fileGRPC,
		logger:   logger,
	}
}

func (f *FileServiceClientImpl) MoveTempFileToAvatars(ctx context.Context, filename string) (string, error) {
	res, err := f.fileGRPC.MoveTempFileToAvatars(ctx, &filepb.MoveTempFileToAvatarsRequest{
		Filename: filename,
	})
	if err != nil {
		return "", err
	}

	return res.GetFilename(), nil
}

func (f *FileServiceClientImpl) MoveTempFileToVoiceMessages(ctx context.Context, filename string) (string, error) {
	res, err := f.fileGRPC.MoveTempFileToVoiceMessages(ctx, &filepb.MoveTempFileToVoiceMessagesRequest{
		Filename: filename,
	})
	if err != nil {
		return "", err
	}

	return res.GetFilename(), nil
}

func (f *FileServiceClientImpl) DeleteAvatar(ctx context.Context, filename string) error {
	_, err := f.fileGRPC.DeleteAvatar(ctx, &filepb.DeleteAvatarRequest{
		Filename: filename,
	})
	if err != nil {
		return err
	}

	return nil
}

func (f *FileServiceClientImpl) DeleteVoiceMessage(ctx context.Context, filename string) error {
	_, err := f.fileGRPC.DeleteVoiceMessage(ctx, &filepb.DeleteVoiceMessageRequest{
		Filename: filename,
	})
	if err != nil {
		return err
	}

	return nil
}
