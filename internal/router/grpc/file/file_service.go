package filegrpc

import (
	"context"
	"github.com/Petr09Mitin/xrust-beze-back/internal/services/file"
	pb "github.com/Petr09Mitin/xrust-beze-back/proto/file"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

type FileGRPCService struct {
	pb.UnimplementedFileServiceServer
	fileService file.FileService
	logger      zerolog.Logger
}

func NewFileGRPCService(fileService file.FileService, logger zerolog.Logger) *FileGRPCService {
	return &FileGRPCService{
		fileService: fileService,
		logger:      logger,
	}
}

func (f *FileGRPCService) MoveTempFileToAvatars(ctx context.Context, req *pb.MoveTempFileToAvatarsRequest) (*pb.MoveTempFileToAvatarsResponse, error) {
	filename := strings.TrimSpace(req.GetFilename())
	if filename == "" {
		return nil, status.Error(codes.InvalidArgument, "filename is empty")
	}

	err := f.fileService.MoveTempFileToAvatars(ctx, filename)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.MoveTempFileToAvatarsResponse{
		Filename: filename,
	}, nil
}

func (f *FileGRPCService) DeleteAvatar(ctx context.Context, req *pb.DeleteAvatarRequest) (*pb.DeleteAvatarResponse, error) {
	filename := strings.TrimSpace(req.GetFilename())
	if filename == "" {
		return nil, status.Error(codes.InvalidArgument, "filename is empty")
	}

	err := f.fileService.DeleteAvatar(ctx, filename)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteAvatarResponse{}, nil
}
