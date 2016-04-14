package octokat

import (
	"fmt"
)

type Merge struct {
	Sha     string `json:"sha,omitempty"`
	Merged  bool   `json:"merged,omitempty"`
	Message string `json:"message,omitempty"`
}

// Merge a pull request
//
// See http://developer.github.com/v3/pulls/#merge-a-pull-request-merge-buttontrade
func (c *Client) MergePullRequest(repo Repo, number string, options *Options) (merge Merge, err error) {
	path := fmt.Sprintf("repos/%s/pulls/%s/merge", repo, number)
	err = c.jsonPut(path, options, &merge)
	return
}
