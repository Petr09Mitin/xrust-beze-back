package study_material_repo

import (
	"context"
	pb "github.com/Petr09Mitin/xrust-beze-back/proto/file"
	"github.com/rs/zerolog"
)

type FileRepo interface {
	CopyAttachmentToStudyFiles(ctx context.Context, filename string) (string, error)
}

type FileRepoImpl struct {
	fileGRPCClient pb.FileServiceClient
	logger         zerolog.Logger
}

func NewFileRepo(fileGRPCClient pb.FileServiceClient, logger zerolog.Logger) FileRepo {
	return &FileRepoImpl{
		fileGRPCClient: fileGRPCClient,
		logger:         logger,
	}
}

func (f *FileRepoImpl) CopyAttachmentToStudyFiles(ctx context.Context, filename string) (string, error) {
	res, err := f.fileGRPCClient.CopyAttachmentToStudyMaterials(ctx, &pb.CopyAttachmentToStudyMaterialsRequest{
		Filename: filename,
	})
	if err != nil {
		return "", err
	}

	return res.GetFilename(), nil
}
