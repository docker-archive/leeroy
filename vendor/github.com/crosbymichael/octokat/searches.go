package octokat

import (
	"fmt"
	"time"
)

type SearchItem struct {
	URL         string `json:"url,omitempty,omitempty"`
	LABELSURL   string `json:"labels_url,omitempty,omitempty"`
	COMMENTSURL string `json:"comments_url,omitempty,omitempty"`
	EVENTSURL   string `json:"events_url,omitempty,omitempty"`
	HTMLURL     string `json:"html_url,omitempty,omitempty"`
	Id          int    `json:"id,omitempty"`
	Number      int    `json:"number,omitempty"`
	Title       string `json:"title,omitempty"`
	User        User   `json:"user,omitempty"`
	Labels      []struct {
		URL   string `json:"url,omitempty"`
		Name  string `json:"name,omitempty"`
		Color string `json:"color,omitempty"`
	} `json:"labels,omitempty"`
	State     string `json:"state,omitempty"`
	Assignee  User   `json:"assignee,omitempty"`
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
	} `json:"milestone,omitempty"`
	Comments    int        `json:"comments,omitempty"`
	CreatedAt   time.Time  `json:"created_at,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at,omitempty"`
	ClosedAt    *time.Time `json:"closed_at,omitempty"`
	PullRequest struct {
		HTMLURL  string `json:"html_url,omitempty"`
		DiffURL  string `json:"diff_url,omitempty"`
		PatchURL string `json:"patch_url,omitempty"`
	} `json:"pull_request,omitempty"`
	Body  string  `json:"body,omitempty"`
	Score float64 `json:"score,omitempty"`
}

type SearchIssue struct {
	TotalCount int           `json:"total_count,omitempty"`
	Items      []*SearchItem `json:"items,omitempty"`
}

// Search issues
//
// See http://developer.github.com/v3/search/#search-issues
func (c *Client) SearchIssues(query string, options *Options) (issues []*SearchItem, err error) {
	var (
		path        = fmt.Sprintf("search/issues?%s", query)
		issuesFound = SearchIssue{}
	)
	err = c.jsonGet(path, options, &issuesFound)
	issues = issuesFound.Items
	return
}
