// Package s3 is an in-memory stand-in for the slice of the S3 client the
// claim-check slides use, with the same call shapes:
// client.PutObject(ctx, &s3.PutObjectInput{Bucket, Key, Body}) and so on. The
// Activities run without a bucket.
package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
)

type Client struct {
	mu      sync.Mutex
	objects map[string][]byte
}

func NewClient() *Client { return &Client{objects: map[string][]byte{}} }

type PutObjectInput struct {
	Bucket, Key *string
	Body        io.Reader
}

type PutObjectOutput struct{}

type GetObjectInput struct{ Bucket, Key *string }

type GetObjectOutput struct{ Body io.ReadCloser }

type DeleteObjectInput struct{ Bucket, Key *string }

type DeleteObjectOutput struct{}

func path(bucket, key *string) string { return *bucket + "/" + *key }

func (c *Client) PutObject(ctx context.Context, in *PutObjectInput) (*PutObjectOutput, error) {
	data, err := io.ReadAll(in.Body)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.objects[path(in.Bucket, in.Key)] = data
	return &PutObjectOutput{}, nil
}

func (c *Client) GetObject(ctx context.Context, in *GetObjectInput) (*GetObjectOutput, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	data, ok := c.objects[path(in.Bucket, in.Key)]
	if !ok {
		return nil, fmt.Errorf("NoSuchKey: %s", path(in.Bucket, in.Key))
	}
	return &GetObjectOutput{Body: io.NopCloser(bytes.NewReader(data))}, nil
}

func (c *Client) DeleteObject(ctx context.Context, in *DeleteObjectInput) (*DeleteObjectOutput, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.objects, path(in.Bucket, in.Key))
	return &DeleteObjectOutput{}, nil
}

// Len is the number of objects stored, for the tests.
func (c *Client) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.objects)
}
