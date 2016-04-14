package octokat

import (
	"fmt"
)

type Label struct {
	URL   string `json:"url,omitempty"`
	Name  string `json:"name,omitempty"`
	Color string `json:"color,omitempty"`
}

func (c *Client) Labels(repo Repo) (labels []*Label, err error) {
	path := fmt.Sprintf("repos/%s/labels", repo)
	err = c.jsonGet(path, &Options{}, &labels)
	return
}

func (c *Client) ApplyLabel(repo Repo, issue *Issue, labels []string) error {
	path := fmt.Sprintf("repos/%s/issues/%d/labels", repo, issue.Number)
	out := []*Label{}
	return c.jsonPost(path, &Options{Params: labels}, &out)
}

func (c *Client) RemoveLabel(repo Repo, issue *Issue, label string) error {
	path := fmt.Sprintf("repos/%s/issues/%d/labels/%s", repo, issue.Number, label)
	_, err := c.request("DELETE", path, nil, nil)
	return err
}
