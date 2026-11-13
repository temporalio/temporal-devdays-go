// Package claimcheck backs 21-claim-check.md.
package claimcheck

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"

	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/claimcheck/aws"
	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/claimcheck/s3"
)

const bucket = "claim-checks"

// s3Client is the in-memory stand-in; a real Worker builds an S3 client once
// at startup.
var s3Client = s3.NewClient()

// fetchDocument stands in for an HTTP download: a 3 MB document, over the
// 2 MB payload limit.
func fetchDocument(ctx context.Context, url string) ([]byte, error) {
	return bytes.Repeat([]byte("x"), 3*1024*1024), nil
}

// slide: 21-claim-check.md "heading: Store the bytes, return the key"
func uploadToStorage(ctx context.Context, data []byte) (string, error) {
	key := uuid.NewString()
	_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	return key, err
}

// end slide

// slide: 21-claim-check.md "heading: Fetch by key"
func fetchFromStorage(ctx context.Context, key string) ([]byte, error) {
	out, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer out.Body.Close()
	return io.ReadAll(out.Body)
}

// end slide

func DownloadDocument(ctx context.Context, url string) (string, error) {
	data, err := fetchDocument(ctx, url)
	if err != nil {
		return "", err
	}
	return uploadToStorage(ctx, data)
}

func SummarizeDocument(ctx context.Context, key string) (string, error) {
	data, err := fetchFromStorage(ctx, key)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d bytes", len(data)), nil
}

func DeleteDocument(ctx context.Context, key string) error {
	_, err := s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}

// BuildReport is for the External Storage slide: a result far too big to
// inline comfortably.
func BuildReport(ctx context.Context) (string, error) {
	return strings.Repeat("r", 1024*1024), nil
}
