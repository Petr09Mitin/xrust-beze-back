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

func (f *FileGRPCService) MoveTempFileToVoiceMessages(ctx context.Context, req *pb.MoveTempFileToVoiceMessagesRequest) (*pb.MoveTempFileToVoiceMessagesResponse, error) {
	filename := strings.TrimSpace(req.GetFilename())
	if filename == "" {
		return nil, status.Error(codes.InvalidArgument, "filename is empty")
	}

	err := f.fileService.MoveTempFileToVoiceMessages(ctx, filename)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.MoveTempFileToVoiceMessagesResponse{
		Filename: filename,
	}, nil
}

func (f *FileGRPCService) DeleteVoiceMessage(ctx context.Context, req *pb.DeleteVoiceMessageRequest) (*pb.DeleteVoiceMessageResponse, error) {
	filename := strings.TrimSpace(req.GetFilename())
	if filename == "" {
		return nil, status.Error(codes.InvalidArgument, "filename is empty")
	}

	err := f.fileService.DeleteVoiceMessage(ctx, filename)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteVoiceMessageResponse{}, nil
}

func (f *FileGRPCService) MoveTempFilesToAttachments(ctx context.Context, req *pb.MoveTempFilesToAttachmentsRequest) (*pb.MoveTempFilesToAttachmentsResponse, error) {
	filenames := req.GetFilenames()
	if len(filenames) == 0 {
		return nil, status.Error(codes.InvalidArgument, "filenames are empty")
	}

	err := f.fileService.MoveTempFilesToAttachments(ctx, filenames)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.MoveTempFilesToAttachmentsResponse{
		Filenames: filenames,
	}, nil
}

func (f *FileGRPCService) DeleteAttachments(ctx context.Context, req *pb.DeleteAttachmentsRequest) (*pb.DeleteAttachmentsResponse, error) {
	filenames := req.GetFilenames()
	if len(filenames) == 0 {
		return nil, status.Error(codes.InvalidArgument, "filenames are empty")
	}

	err := f.fileService.DeleteAttachments(ctx, filenames)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteAttachmentsResponse{}, nil
}

func (f *FileGRPCService) CopyAttachmentToStudyMaterials(ctx context.Context, req *pb.CopyAttachmentToStudyMaterialsRequest) (*pb.CopyAttachmentToStudyMaterialsResponse, error) {
	filename := strings.TrimSpace(req.GetFilename())
	if len(filename) == 0 {
		return nil, status.Error(codes.InvalidArgument, "filename is empty")
	}

	filename, err := f.fileService.CopyAttachmentToStudyMaterials(ctx, filename)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CopyAttachmentToStudyMaterialsResponse{
		Filename: filename,
	}, nil
}
