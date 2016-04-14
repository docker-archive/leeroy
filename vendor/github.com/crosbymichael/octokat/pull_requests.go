package octokat

import (
	"fmt"
	"time"
)

type PullRequest struct {
	URL               string     `json:"url,omitempty"`
	ID                int        `json:"id,omitempty"`
	HTMLURL           string     `json:"html_url,omitempty"`
	DiffURL           string     `json:"diff_url,omitempty"`
	PatchURL          string     `json:"patch_url,omitempty"`
	IssueURL          string     `json:"issue_url,omitempty"`
	Number            int        `json:"number,omitempty"`
	State             string     `json:"state,omitempty"`
	Title             string     `json:"title,omitempty"`
	User              User       `json:"user,omitempty"`
	Body              string     `json:"body,omitempty"`
	CreatedAt         time.Time  `json:"created_at,omitempty"`
	UpdatedAt         time.Time  `json:"updated_at,omitempty"`
	ClosedAt          *time.Time `json:"closed_at,omitempty"`
	MergedAt          *time.Time `json:"merged_at,omitempty"`
	MergeCommitSha    string     `json:"merge_commit_sha,omitempty"`
	Assignee          *User      `json:"assignee,omitempty"`
	CommitsURL        string     `json:"commits_url,omitempty"`
	ReviewCommentsURL string     `json:"review_comments_url,omitempty"`
	ReviewCommentURL  string     `json:"review_comment_url,omitempty"`
	CommentsURL       string     `json:"comments_url,omitempty"`
	Head              Commit     `json:"head,omitempty"`
	Base              Commit     `json:"base,omitempty"`
	Merged            bool       `json:"merged,omitempty"`
	MergedBy          User       `json:"merged_by,omitempty"`
	Comments          int        `json:"comments,omitempty"`
	ReviewComments    int        `json:"review_comments,omitempty"`
	Commits           int        `json:"commits,omitempty"`
	Additions         int        `json:"additions,omitempty"`
	Deletions         int        `json:"deletions,omitempty"`
	ChangedFiles      int        `json:"changed_files,omitempty"`
	Mergeable         *bool      `json:"mergeable,omitempty"`
	CommentsBody      []Comment  `json:"-"`
}

type PullRequestFile struct {
	FileName    string `json:"filename,omitempty"`
	Sha         string `json:"sha,omitempty"`
	Status      string `json:"status,omitempty"`
	Additions   int    `json:"additions,omitempty"`
	Deletions   int    `json:"deletions,omitempty"`
	Changes     int    `json:"changes,omitempty"`
	BlobUrl     string `json:"blob_url,omitempty"`
	RawUrl      string `json:"raw_url,omitempty"`
	ContentsUrl string `json:"contents_url,omitempty"`
	Patch       string `json:"patch,omitempty"`
}

// Get a pull request
//
// See http://developer.github.com/v3/pulls/#get-a-single-pull-request
func (c *Client) PullRequest(repo Repo, number string, options *Options) (pr *PullRequest, err error) {
	path := fmt.Sprintf("repos/%s/pulls/%s", repo, number)
	err = c.jsonGet(path, options, &pr)
	return
}

// Get all pull requests
//
// See http://developer.github.com/v3/pulls/#list-pull-requests
func (c *Client) PullRequests(repo Repo, options *Options) (prs []*PullRequest, err error) {
	path := fmt.Sprintf("repos/%s/pulls", repo)
	err = c.jsonGet(path, options, &prs)
	return
}

// Get all pull request files
//
// See http://developer.github.com/v3/pulls/#list-pull-requests-files
func (c *Client) PullRequestFiles(repo Repo, number string, options *Options) (prfs []*PullRequestFile, err error) {
	path := fmt.Sprintf("repos/%s/pulls/%s/files", repo, number)
	err = c.jsonGet(path, options, &prfs)
	return
}

type PullRequestParams struct {
	Base  string `json:"base,omitempty"`
	Head  string `json:"head,omitempty"`
	Title string `json:"title,omitempty"`
	Body  string `json:"body,omitempty"`
}

type PullRequestForIssueParams struct {
	Base  string `json:"base,omitempty"`
	Head  string `json:"head,omitempty"`
	Issue string `json:"issue,omitempty"`
}

// Create a pull request
//
// See http://developer.github.com/v3/pulls/#create-a-pull-request
// See http://developer.github.com/v3/pulls/#alternative-input
func (c *Client) CreatePullRequest(repo Repo, options *Options) (pr *PullRequest, err error) {
	path := fmt.Sprintf("repos/%s/pulls", repo)
	err = c.jsonPost(path, options, &pr)
	return
}
