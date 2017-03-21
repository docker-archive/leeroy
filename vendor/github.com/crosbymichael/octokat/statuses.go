package octokat

import (
	"fmt"
	"time"
)

type Status struct {
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
	State       string    `json:"state,omitempty"`
	TargetURL   string    `json:"target_url,omitempty"`
	Description string    `json:"description,omitempty"`
	ID          int       `json:"id,omitempty"`
	URL         string    `json:"url,omitempty"`
	Creator     User      `json:"creator,omitempty"`
	Context     string    `json:"context,omitempty"`
}

type StatusOptions struct {
	State       string `json:"state"`
	Description string `json:"description"`
	URL         string `json:"target_url"`
	Context     string `json:"context"`
}

type CombinedStatus struct {
	State      string   `json:"state"`
	Sha        string   `json:"sha"`
	TotalCount int      `json:"total_count"`
	Statuses   []Status `json:"statuses"`
}

// List all statuses for a given commit
//
// See http://developer.github.com/v3/repos/statuses
func (c *Client) Statuses(repo Repo, sha string, options *Options) (statuses []Status, err error) {
	path := fmt.Sprintf("repos/%s/statuses/%s", repo, sha)
	err = c.jsonGet(path, options, &statuses)
	return
}

func (c *Client) CombinedStatus(repo Repo, sha string, options *Options) (status CombinedStatus, err error) {
	path := fmt.Sprintf("repos/%s/commits/%s/status", repo, sha)
	err = c.jsonGet(path, options, &status)
	return status, err
}

// Set a status for a given sha
//
// See https://developer.github.com/v3/repos/statuses/#create-a-status
func (c *Client) SetStatus(repo Repo, sha string, options *StatusOptions) (status *Status, err error) {
	path := fmt.Sprintf("repos/%s/statuses/%s", repo, sha)
	mutOptions := &Options{
		Params: map[string]string{
			"state":       options.State,
			"description": options.Description,
			"target_url":  options.URL,
			"context":     options.Context,
		},
	}
	err = c.jsonPost(path, mutOptions, &status)
	return
}
