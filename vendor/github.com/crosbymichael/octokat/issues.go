package octokat

import (
	"fmt"
	"time"
)

type Issue struct {
	URL       string   `json:"url,omitempty,omitempty"`
	HTMLURL   string   `json:"html_url,omitempty,omitempty"`
	Number    int      `json:"number,omitempty"`
	State     string   `json:"state,omitempty"`
	Title     string   `json:"title,omitempty"`
	Body      string   `json:"body,omitempty"`
	User      User     `json:"user,omitempty"`
	Labels    []*Label `json:"labels,omitempty"`
	Assignee  User     `json:"assignee,omitempty"`
	Milestone struct {
		URL          string     `json:"url,omitempty"`
		Number       int        `json:"number,omitempty"`
		State        string     `json:"state,omitempty"`
		Title        string     `json:"title,omitempty"`
		Description  string     `json:"description,omitempty"`
		Creator      User       `json:"creator,omitempty"`
		OpenIssues   int        `json:"open_issues,omitempty"`
		ClosedIssues int        `json:"closed_issues,omitempty"`
		CreatedAt    time.Time  `json:"created_at,omitempty"`
		DueOn        *time.Time `json:"due_on,omitempty"`
	}
	Comments    int `json:"comments,omitempty"`
	PullRequest struct {
		HTMLURL  string `json:"html_url,omitempty"`
		DiffURL  string `json:"diff_url,omitempty"`
		PatchURL string `json:"patch_url,omitempty"`
	} `json:"pull_request,omitempty"`
	CreatedAt time.Time  `json:"created_at,omitempty"`
	ClosedAt  *time.Time `json:"closed_at,omitempty"`
	UpdatedAt time.Time  `json:"updated_at,omitempty"`
}

// List issues
//
// See http://developer.github.com/v3/issues/#list-issues-for-a-repository
func (c *Client) Issues(repo Repo, options *Options) (issues []*Issue, err error) {
	path := fmt.Sprintf("repos/%s/issues", repo)
	err = c.jsonGet(path, options, &issues)
	return
}

// Fetch a single issue
//
// See http://developer.github.com/v3/issues/#get-a-single-issue
func (c *Client) Issue(repo Repo, number int, options *Options) (issue *Issue, err error) {
	path := fmt.Sprintf("repos/%s/issues/%d", repo, number)
	err = c.jsonGet(path, options, &issue)
	return
}

// Edit an issue
//
// See http://developer.github.com/v3/issues/#edit-an-issue
func (c *Client) PatchIssue(repo Repo, number string, options *Options) (patchedIssue *Issue, err error) {
	path := fmt.Sprintf("repos/%s/issues/%s", repo, number)
	err = c.jsonPatch(path, options, &patchedIssue)
	return
}
