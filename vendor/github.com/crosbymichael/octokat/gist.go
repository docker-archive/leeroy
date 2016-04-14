package octokat

import (
	"fmt"
	"time"
)

type File struct {
	Filename string `json:"filename,omitempty"`
	RawUrl   string `json:"raw_url,omitempty"`
	Type     string `json:"type,omitempty"`
	Language string `json:"language,omitempty"`
	Size     int64  `json:"size,omitempty"`
	Content  string `json:"content,omitempty"`
}

type Gist struct {
	Id          string    `json:"id,omitempty"`
	Public      bool      `json:"public,omitempty"`
	Description string    `json:"description,omitempty"`
	HtmlUrl     string    `json:"html_url,omitempty"`
	Url         string    `json:"url,omitempty"`
	Files       File      `json:"files,omitempty"`
	User        User      `json:"user,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}

func (c *Client) CreateGist(desc string, public bool, files map[string]File) (gist Gist, err error) {
	path := fmt.Sprintf("gists")
	m := map[string]interface{}{
		"public": public,
		"files":  files,
	}

	if desc != "" {
		m["description"] = desc
	}
	options := &Options{Params: m}

	err = c.jsonPost(path, options, &gist)
	return
}
