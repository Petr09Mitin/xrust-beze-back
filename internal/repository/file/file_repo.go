package filerepo

import (
	"context"
	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/lifecycle"
	"github.com/rs/zerolog"
)

type FileRepo interface {
	UploadTemp(ctx context.Context, filepath, filename string) error
	CopyFromTempToAvatars(ctx context.Context, filename string) (err error)
	CopyFromTempToVoiceMessages(ctx context.Context, filename string) (err error)
	DeleteAvatar(ctx context.Context, filename string) (err error)
	DeleteVoiceMessage(ctx context.Context, filename string) (err error)
	DeleteTemp(ctx context.Context, filename string) (err error)
	CheckIfAvatarExists(ctx context.Context, filename string) (exist bool, err error)
	CheckIfVoiceMessageExists(ctx context.Context, filename string) (exist bool, err error)
	CheckIfTempExists(ctx context.Context, filename string) (exist bool, err error)
	CopyFromTempToAttachments(ctx context.Context, filenames []string) (err error)
	DeleteAttachments(ctx context.Context, filenames []string) (err error)
	CheckIfAttachmentExists(ctx context.Context, filename string) (exist bool, err error)
}

type FileRepoImpl struct {
	minioClient *minio.Client
	logger      zerolog.Logger
}

func NewFileRepo(minioClient *minio.Client, logger zerolog.Logger) (FileRepo, error) {
	fr := &FileRepoImpl{
		minioClient: minioClient,
		logger:      logger,
	}
	err := fr.initTempBucket()
	if err != nil {
		return nil, err
	}
	err = fr.initAvatarsBucket()
	if err != nil {
		return nil, err
	}
	err = fr.initVoiceMessagesBucket()
	if err != nil {
		return nil, err
	}
	err = fr.initAttachmentsBucket()
	if err != nil {
		return nil, err
	}
	return fr, nil
}

func (f *FileRepoImpl) UploadTemp(ctx context.Context, filepath, filename string) error {
	_, err := f.minioClient.FPutObject(ctx, config.TempMinioBucket, filename, filepath, minio.PutObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (f *FileRepoImpl) CopyFromTempToAvatars(ctx context.Context, filename string) error {
	_, err := f.minioClient.CopyObject(ctx, minio.CopyDestOptions{
		Bucket: config.AvatarsMinioBucket,
		Object: filename,
	}, minio.CopySrcOptions{
		Bucket: config.TempMinioBucket,
		Object: filename,
	})
	if err != nil {
		return err
	}

	return nil
}

func (f *FileRepoImpl) CopyFromTempToVoiceMessages(ctx context.Context, filename string) error {
	_, err := f.minioClient.CopyObject(ctx, minio.CopyDestOptions{
		Bucket: config.VoiceMessagesMinioBucket,
		Object: filename,
	}, minio.CopySrcOptions{
		Bucket: config.TempMinioBucket,
		Object: filename,
	})
	if err != nil {
		return err
	}

	return nil
}

func (f *FileRepoImpl) DeleteAvatar(ctx context.Context, filename string) error {
	return f.minioClient.RemoveObject(ctx, config.AvatarsMinioBucket, filename, minio.RemoveObjectOptions{})
}

func (f *FileRepoImpl) DeleteVoiceMessage(ctx context.Context, filename string) error {
	return f.minioClient.RemoveObject(ctx, config.VoiceMessagesMinioBucket, filename, minio.RemoveObjectOptions{})
}

func (f *FileRepoImpl) DeleteTemp(ctx context.Context, filename string) error {
	return f.minioClient.RemoveObject(ctx, config.TempMinioBucket, filename, minio.RemoveObjectOptions{})
}

func (f *FileRepoImpl) CheckIfTempExists(ctx context.Context, filename string) (exist bool, err error) {
	return f.checkIfObjectExists(ctx, config.TempMinioBucket, filename)
}

func (f *FileRepoImpl) CheckIfAvatarExists(ctx context.Context, filename string) (exist bool, err error) {
	return f.checkIfObjectExists(ctx, config.AvatarsMinioBucket, filename)
}

func (f *FileRepoImpl) CheckIfVoiceMessageExists(ctx context.Context, filename string) (exist bool, err error) {
	return f.checkIfObjectExists(ctx, config.VoiceMessagesMinioBucket, filename)
}

func (f *FileRepoImpl) checkIfObjectExists(ctx context.Context, bucket, filename string) (exist bool, err error) {
	_, err = f.minioClient.StatObject(ctx, bucket, filename, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (f *FileRepoImpl) initTempBucket() error {
	exists, err := f.minioClient.BucketExists(context.Background(), config.TempMinioBucket)
	if err != nil {
		return err
	}

	if !exists {
		err = f.minioClient.MakeBucket(context.Background(), config.TempMinioBucket, minio.MakeBucketOptions{
			Region: "ru-central1",
		})
		if err != nil {
			return err
		}

		// store temp files for 1 day then delete
		err = f.minioClient.SetBucketLifecycle(context.Background(), config.TempMinioBucket, &lifecycle.Configuration{
			Rules: []lifecycle.Rule{
				{
					Expiration: lifecycle.Expiration{
						Days: 1,
					},
				},
			},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *FileRepoImpl) initAvatarsBucket() error {
	exists, err := f.minioClient.BucketExists(context.Background(), config.AvatarsMinioBucket)
	if err != nil {
		return err
	}
	if !exists {
		err = f.minioClient.MakeBucket(context.Background(), config.AvatarsMinioBucket, minio.MakeBucketOptions{
			Region: "ru-central1",
		})
		if err != nil {
			return err
		}

		err := f.minioClient.SetBucketPolicy(context.Background(), config.AvatarsMinioBucket, f.getPublicReadPolicy(config.AvatarsMinioBucket))
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *FileRepoImpl) initVoiceMessagesBucket() error {
	exists, err := f.minioClient.BucketExists(context.Background(), config.VoiceMessagesMinioBucket)
	if err != nil {
		return err
	}
	if !exists {
		err = f.minioClient.MakeBucket(context.Background(), config.VoiceMessagesMinioBucket, minio.MakeBucketOptions{
			Region: "ru-central1",
		})
		if err != nil {
			return err
		}

		err := f.minioClient.SetBucketPolicy(context.Background(), config.VoiceMessagesMinioBucket, f.getPublicReadPolicy(config.VoiceMessagesMinioBucket))
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *FileRepoImpl) initAttachmentsBucket() error {
	exists, err := f.minioClient.BucketExists(context.Background(), config.AttachmentsMinioBucket)
	if err != nil {
		return err
	}
	if !exists {
		err = f.minioClient.MakeBucket(context.Background(), config.AttachmentsMinioBucket, minio.MakeBucketOptions{
			Region: "ru-central1",
		})
		if err != nil {
			return err
		}

		err := f.minioClient.SetBucketPolicy(context.Background(), config.AttachmentsMinioBucket, f.getPublicReadPolicy(config.AttachmentsMinioBucket))
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *FileRepoImpl) getPublicReadPolicy(bucketName string) string {
	return `{
        "Version": "2012-10-17",
        "Statement": [
            {
                "Sid": "PublicRead",
                "Effect": "Allow",
                "Principal": "*",
                "Action": ["s3:GetObject"],
                "Resource": ["arn:aws:s3:::` + bucketName + `/*"]
            }
        ]
    }`
}

func (f *FileRepoImpl) CopyFromTempToAttachments(ctx context.Context, filenames []string) (err error) {
	for _, filename := range filenames {
		_, err := f.minioClient.CopyObject(ctx, minio.CopyDestOptions{
			Bucket: config.AttachmentsMinioBucket,
			Object: filename,
		}, minio.CopySrcOptions{
			Bucket: config.TempMinioBucket,
			Object: filename,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
func (f *FileRepoImpl) DeleteAttachments(ctx context.Context, filenames []string) (err error) {
	for _, filename := range filenames {
		err := f.minioClient.RemoveObject(ctx, config.AttachmentsMinioBucket, filename, minio.RemoveObjectOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}
func (f *FileRepoImpl) CheckIfAttachmentExists(ctx context.Context, filename string) (exist bool, err error) {
	return f.checkIfObjectExists(ctx, config.AttachmentsMinioBucket, filename)
}
