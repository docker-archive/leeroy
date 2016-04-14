package octokat

import (
	"fmt"
	"time"
)

type Comment struct {
	Url       string    `json:"url,omitempty,omitempty"`
	Id        int       `json:"id,omitempty"`
	Body      string    `json:"body,omitempty"`
	Path      string    `json:"path,omitempty"`
	Position  int       `json:"position,omitempty"`
	User      User      `json:"user,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// Get comments for an issue for pull request
//
// See http://developer.github.com/v3/pulls/comments/
func (c *Client) Comments(repo Repo, number string, options *Options) (comments []Comment, err error) {
	path := fmt.Sprintf("repos/%s/issues/%s/comments", repo, number)
	err = c.jsonGet(path, options, &comments)
	return
}

// Add a comment to an issue or pull request
//
// See http://developer.github.com/v3/issues/comments/#create-a-comment
func (c *Client) AddComment(repo Repo, number, body string) (comment Comment, err error) {
	path := fmt.Sprintf("repos/%s/issues/%s/comments", repo, number)
	options := &Options{Params: map[string]string{"body": body}}

	err = c.jsonPost(path, options, &comment)
	return
}

// Add a comment to an issue or pull request
//
// See https://developer.github.com/v3/issues/comments/#edit-a-comment
func (c *Client) PatchComment(repo Repo, number, body string) (comment Comment, err error) {
	path := fmt.Sprintf("repos/%s/issues/%s/comments", repo, number)
	options := &Options{Params: map[string]string{"body": body}}

	err = c.jsonPatch(path, options, &comment)
	return
}

func (c *Client) RemoveComment(repo Repo, commentId int) error {
	path := fmt.Sprintf("repos/%s/issues/comments/%d", repo, commentId)
	return c.delete(path, Headers{})
}
