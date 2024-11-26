package services

import (
    "context"
    "fmt"
    "io"
    "mime/multipart"
    "path/filepath"
    "time"
    "onboarding-app/models"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/credentials"
    "github.com/aws/aws-sdk-go-v2/service/s3"
    // "github.com/aws/aws-sdk-go-v2/service/s3/types"
    // presign "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
    "github.com/google/uuid"
)

type S3Config struct {
    AccessKeyID     string
    SecretAccessKey string
    Region          string
    Bucket          string
    Endpoint        string
}

type S3FileService struct {
    client *s3.Client
    bucket string
    config S3Config
}

func NewS3FileService(config S3Config) (*S3FileService, error) {
    // Create custom AWS configuration
    customResolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
        if config.Endpoint != "" {
            return aws.Endpoint{
                URL:           config.Endpoint,
                SigningRegion: config.Region,
            }, nil
        }
        return aws.Endpoint{}, &aws.EndpointNotFoundError{}
    })

    cfg := aws.Config{
        Region:           config.Region,
        Credentials:      credentials.NewStaticCredentialsProvider(config.AccessKeyID, config.SecretAccessKey, ""),
        EndpointResolver: customResolver,
    }

    client := s3.NewFromConfig(cfg)

    return &S3FileService{
        client: client,
        bucket: config.Bucket,
        config: config,
    }, nil
}

func (s *S3FileService) UploadFile(ctx context.Context, file *multipart.FileHeader) (*models.UploadResult, error) {
    src, err := file.Open()
    if err != nil {
        return nil, fmt.Errorf("unable to open file: %v", err)
    }
    defer src.Close()

    // Generate unique filename
    ext := filepath.Ext(file.Filename)
    key := fmt.Sprintf("documents/%s/%s%s", 
        time.Now().Format("2006/01/02"),
        uuid.New().String(),
        ext,
    )

    // Upload to S3
    _, err = s.client.PutObject(ctx, &s3.PutObjectInput{
        Bucket:      aws.String(s.bucket),
        Key:         aws.String(key),
        Body:        src,
        ContentType: aws.String(file.Header.Get("Content-Type")),
    })

    if err != nil {
        return nil, fmt.Errorf("unable to upload file to S3: %v", err)
    }

    return &models.UploadResult{
        Location:  fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.config.Region, key),
        Filename:  file.Filename,
        Key:       key,
        Size:      file.Size,
        MimeType:  file.Header.Get("Content-Type"),
    }, nil
}

func (s *S3FileService) DeleteFile(ctx context.Context, key string) error {
    _, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
        Bucket: aws.String(s.bucket),
        Key:    aws.String(key),
    })
    return err
}

func (s *S3FileService) GetPresignedURL(ctx context.Context, key string) (string, error) {
    presignClient := s3.NewPresignClient(s.client)
    
    request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
        Bucket: aws.String(s.bucket),
        Key:    aws.String(key),
    }, func(opts *s3.PresignOptions) {
        opts.Expires = time.Duration(15 * time.Minute)
    })
    
    if err != nil {
        return "", fmt.Errorf("couldn't get presigned URL: %v", err)
    }

    return request.URL, nil
}

// DownloadFile gets the file directly from S3
func (s *S3FileService) DownloadFile(ctx context.Context, key string) (io.ReadCloser, string, error) {
    result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
        Bucket: aws.String(s.bucket),
        Key:    aws.String(key),
    })
    
    if err != nil {
        return nil, "", fmt.Errorf("couldn't download file: %v", err)
    }

    contentType := "application/octet-stream"
    if result.ContentType != nil {
        contentType = *result.ContentType
    }

    return result.Body, contentType, nil
}

// CheckIfFileExists verifies if a file exists in S3
func (s *S3FileService) CheckIfFileExists(ctx context.Context, key string) (bool, error) {
    _, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
        Bucket: aws.String(s.bucket),
        Key:    aws.String(key),
    })
    
    if err != nil {
        // Just check the error string since we're not using the error type
        if err.Error() == "NotFound" || err.Error() == "NoSuchKey" {
            return false, nil
        }
        return false, err
    }
    
    return true, nil
}