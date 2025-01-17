package gist

import (
	"context"
	"time"
	"net/http"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

// RateLimits defines the minimum intervals between requests
const (
	MinIntervalAuth    = 1 * time.Second  // For authenticated requests (5000/hour)
	MinIntervalUnauth  = 60 * time.Second // For unauthenticated requests (60/hour)
	BackoffMultiplier  = 2                // Multiply interval by this when rate limited
	MaxBackoffInterval = 15 * time.Minute // Maximum backoff interval
)

type Client struct {
	client *github.Client
	gistID string
	authenticated bool
}

// NewClient creates a new client with write access (requires token)
func NewClient(token, gistID string) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return &Client{
		client: client,
		gistID: gistID,
		authenticated: true,
	}
}

// NewReadOnlyClient creates a new client with read-only access (no token required)
func NewReadOnlyClient(gistID string) *Client {
	return &Client{
		client: github.NewClient(nil),
		gistID: gistID,
		authenticated: false,
	}
}

// FileContent represents a file's content and its last modification time
type FileContent struct {
	Content    []byte
	UpdatedAt  time.Time
	IsDeleted  bool
}

// UploadFiles uploads multiple files to the gist
func (c *Client) UploadFiles(files map[string][]byte) error {
	gistFiles := make(map[github.GistFilename]github.GistFile)
	for name, content := range files {
		gistFiles[github.GistFilename(name)] = github.GistFile{
			Content: github.String(string(content)),
		}
	}

	gist := &github.Gist{Files: gistFiles}
	_, _, err := c.client.Gists.Edit(context.Background(), c.gistID, gist)
	return err
}

// DownloadFiles downloads all files from the gist
func (c *Client) DownloadFiles() (map[string]FileContent, error) {
	gist, _, err := c.client.Gists.Get(context.Background(), c.gistID)
	if err != nil {
		return nil, err
	}

	files := make(map[string]FileContent)
	for filename, file := range gist.Files {
		if file.Content == nil {
			// File was deleted
			files[string(filename)] = FileContent{
				IsDeleted: true,
				UpdatedAt: gist.UpdatedAt.Time,
			}
			continue
		}
		// 使用 UTC 时间以确保时间比较的一致性
		updateTime := gist.UpdatedAt.Time.UTC()
		files[string(filename)] = FileContent{
			Content:   []byte(*file.Content),
			UpdatedAt: updateTime,
		}
	}

	return files, nil
}

// GetLastModified returns the last modification time of the gist
func (c *Client) GetLastModified() (time.Time, error) {
	gist, _, err := c.client.Gists.Get(context.Background(), c.gistID)
	if err != nil {
		return time.Time{}, err
	}
	return gist.UpdatedAt.Time.UTC(), nil
}

// DeleteFile deletes a file from the gist
func (c *Client) DeleteFile(filename string) error {
	gist := &github.Gist{
		Files: map[github.GistFilename]github.GistFile{
			github.GistFilename(filename): {
				Content: nil,
			},
		},
	}
	_, _, err := c.client.Gists.Edit(context.Background(), c.gistID, gist)
	return err
}

// GetMinInterval returns the minimum recommended interval between requests
func (c *Client) GetMinInterval() time.Duration {
	if c.authenticated {
		return MinIntervalAuth
	}
	return MinIntervalUnauth
}

// IsRateLimited checks if the error is a rate limit error
func (c *Client) IsRateLimited(err error) bool {
	if err == nil {
		return false
	}
	if _, ok := err.(*github.RateLimitError); ok {
		return true
	}
	if errResp, ok := err.(*github.ErrorResponse); ok {
		return errResp.Response.StatusCode == http.StatusTooManyRequests
	}
	return false
}
