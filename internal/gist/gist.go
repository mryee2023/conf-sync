package gist

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v57/github"
	"github.com/mryee2023/conf-sync/internal/logger"
	"golang.org/x/oauth2"
)

const (
	// Rate limit constants
	MinIntervalAuth    = time.Second      // For authenticated requests (5000/hour)
	MinIntervalUnauth  = 60 * time.Second // For unauthenticated requests (60/hour)
	BackoffMultiplier  = 2                // Multiply interval by this when rate limited
	MaxBackoffInterval = 15 * time.Minute // Maximum backoff interval
)

var (
	// MinTime is used as a starting point for file modification times
	MinTime = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
)

// GistFile represents a file in a Gist
type GistFile struct {
	Content   []byte
	UpdatedAt time.Time
	IsDeleted bool
}

// Client represents a GitHub Gist client
type Client struct {
	client *github.Client
	gistID string
	isAuth bool
}

// NewClient creates a new client with write access (requires token)
func NewClient(token, gistID string) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &Client{
		client: github.NewClient(tc),
		gistID: gistID,
		isAuth: true,
	}
}

// NewReadOnlyClient creates a new client with read-only access (no token required)
func NewReadOnlyClient(gistID string) *Client {
	return &Client{
		client: github.NewClient(nil),
		gistID: gistID,
		isAuth: false,
	}
}

// GetMinInterval returns the minimum interval based on authentication status
func (c *Client) GetMinInterval() time.Duration {
	if c.isAuth {
		return MinIntervalAuth
	}
	return MinIntervalUnauth
}

// IsRateLimited checks if an error is due to rate limiting
func (c *Client) IsRateLimited(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "rate limit")
}

// DownloadFiles downloads all files from the Gist
func (c *Client) DownloadFiles() (map[string]*GistFile, error) {
	ctx := context.Background()
	gist, _, err := c.client.Gists.Get(ctx, c.gistID)
	if err != nil {
		return nil, fmt.Errorf("failed to get gist: %v", err)
	}

	files := make(map[string]*GistFile)
	for filename, file := range gist.Files {
		name := string(filename)
		if file.Content == nil {
			logger.Debug("File %s is marked as deleted", name)
			files[name] = &GistFile{IsDeleted: true}
			continue
		}

		files[name] = &GistFile{
			Content:   []byte(*file.Content),
			UpdatedAt: gist.UpdatedAt.Time,
		}
	}

	return files, nil
}

// UploadFiles uploads multiple files to the Gist
func (c *Client) UploadFiles(files map[string][]byte) error {
	if !c.isAuth {
		return errors.New("write operations require authentication")
	}

	gistFiles := make(map[github.GistFilename]github.GistFile)
	for name, content := range files {
		str := string(content)
		gistFiles[github.GistFilename(name)] = github.GistFile{
			Content: github.String(str),
		}
	}

	ctx := context.Background()
	gist := &github.Gist{
		Files: gistFiles,
	}

	_, _, err := c.client.Gists.Edit(ctx, c.gistID, gist)
	return err
}

// DeleteFile marks a file as deleted in the Gist
func (c *Client) DeleteFile(name string) error {
	if !c.isAuth {
		return errors.New("write operations require authentication")
	}

	ctx := context.Background()
	gist := &github.Gist{
		Files: map[github.GistFilename]github.GistFile{
			github.GistFilename(name): {},
		},
	}

	_, _, err := c.client.Gists.Edit(ctx, c.gistID, gist)
	return err
}
