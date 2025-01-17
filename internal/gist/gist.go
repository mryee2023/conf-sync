package gist

import (
	"context"
	"time"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

type Client struct {
	client *github.Client
	gistID string
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
	}
}

// NewReadOnlyClient creates a new client with read-only access (no token required)
func NewReadOnlyClient(gistID string) *Client {
	return &Client{
		client: github.NewClient(nil),
		gistID: gistID,
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
