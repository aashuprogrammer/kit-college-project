package utils

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type R2Client struct {
	s3Client   *s3.Client
	bucketName string
	publicURL  string
	localDir   string
}

func NewR2Client(accountID, accessKeyID, secretAccessKey, bucketName, publicURL, localDir string) (*R2Client, error) {
	// If credentials are not set, act as a local storage client
	if accessKeyID == "" || secretAccessKey == "" || bucketName == "" || accountID == "" {
		fmt.Printf("R2 Credentials not fully configured. Activating Local File storage fallback in directory: %s\n", localDir)
		// Ensure local directory exists
		err := os.MkdirAll(localDir, 0755)
		if err != nil {
			return nil, fmt.Errorf("failed to create local storage directory: %v", err)
		}
		return &R2Client{
			localDir: localDir,
		}, nil
	}

	// Clean public URL
	resolvedURL := publicURL
	if !strings.HasPrefix(resolvedURL, "http://") && !strings.HasPrefix(resolvedURL, "https://") {
		resolvedURL = "https://" + resolvedURL
	}
	resolvedURL = strings.TrimSuffix(resolvedURL, "/")

	// R2 Endpoint: https://<account_id>.r2.cloudflarestorage.com
	r2Endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)

	// Use AWS SDK v2 config with custom endpoint resolver
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load S3 SDK config: %v", err)
	}

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = &r2Endpoint
	})

	return &R2Client{
		s3Client:   s3Client,
		bucketName: bucketName,
		publicURL:  resolvedURL,
	}, nil
}

func (r *R2Client) UploadFile(ctx context.Context, key string, file io.Reader, contentType string) (string, error) {
	// Local fallback mode
	if r.s3Client == nil {
		localPath := filepath.Join(r.localDir, key)
		// Ensure target subdirectory exists
		dir := filepath.Dir(localPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", fmt.Errorf("failed to create local file directory: %v", err)
		}

		out, err := os.Create(localPath)
		if err != nil {
			return "", fmt.Errorf("failed to create local file: %v", err)
		}
		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			return "", fmt.Errorf("failed to write local file: %v", err)
		}

		// Return serving path relative to static route /uploads
		// Use forward slashes for URLs
		urlPath := filepath.ToSlash(key)
		return fmt.Sprintf("/uploads/%s", urlPath), nil
	}

	// Cloudflare R2 Upload
	_, err := r.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &r.bucketName,
		Key:         &key,
		Body:        file,
		ContentType: &contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload object to R2: %v", err)
	}

	return fmt.Sprintf("%s/%s", r.publicURL, key), nil
}
