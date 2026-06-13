package storage

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var _ Storage = (*GarageClient)(nil)

type GarageClient struct {
	S3Client   *s3.Client
	BucketName string
}

func MustInitGarageClient() *GarageClient {
	endpoint := os.Getenv("GARAGE_ENDPOINT")
	accessKey := os.Getenv("GARAGE_ACCESS_KEY")
	secretKey := os.Getenv("GARAGE_SECRET_KEY")
	bucketName := os.Getenv("GARAGE_BUCKET_NAME")

	credProvider := credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("garage"),
		config.WithCredentialsProvider(credProvider),
	)
	if err != nil {
		panic(fmt.Sprintf("unable to load AWS SDK config: %v", err))
	}

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String("http://" + endpoint)
		o.UsePathStyle = true
	})

	return &GarageClient{
		S3Client:   s3Client,
		BucketName: bucketName,
	}
}

func (g *GarageClient) Upload(ctx context.Context, key string, body io.Reader, contentType string) error {
	_, err := g.S3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(g.BucketName),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	return err
}

func (g *GarageClient) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	out, err := g.S3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(g.BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return out.Body, nil
}

func (g *GarageClient) Delete(ctx context.Context, key string) error {
	_, err := g.S3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(g.BucketName),
		Key:    aws.String(key),
	})
	return err
}

func (g *GarageClient) Stat(ctx context.Context, key string) (*FileInfo, error) {
	out, err := g.S3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(g.BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return &FileInfo{
		Key:          key,
		Size:         aws.ToInt64(out.ContentLength),
		ContentType:  aws.ToString(out.ContentType),
		LastModified: aws.ToTime(out.LastModified),
	}, nil
}

func (g *GarageClient) List(ctx context.Context, prefix string) ([]FileInfo, error) {
	var items []FileInfo
	var token *string
	for {
		out, err := g.S3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:            aws.String(g.BucketName),
			Prefix:            aws.String(prefix),
			ContinuationToken: token,
		})
		if err != nil {
			return nil, err
		}
		for _, o := range out.Contents {
			if o.Key == nil {
				continue
			}
			items = append(items, FileInfo{
				Key:          aws.ToString(o.Key),
				Size:         aws.ToInt64(o.Size),
				LastModified: aws.ToTime(o.LastModified),
			})
		}
		if out.NextContinuationToken == nil {
			break
		}
		token = out.NextContinuationToken
	}
	return items, nil
}
