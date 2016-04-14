package octokat

import (
	"encoding/json"
	"errors"
	"strings"
)

type IssueHook struct {
	Action  string      `json:"action"`
	Sender  *User       `json:"sender"`
	Repo    *Repository `json:"repository"`
	Issue   *Issue      `json:"issue"`
	Comment *Comment    `json:"comment, omitempty"`
}

type PullRequestHook struct {
	Action      string       `json:"action"`
	Number      int          `json:"number"`
	Sender      *User        `json:"sender"`
	Repo        *Repository  `json:"repository"`
	PullRequest *PullRequest `json:"pull_request"`
}

func ParseIssueHook(raw []byte) (*IssueHook, error) {
	hook := IssueHook{}
	if err := json.Unmarshal(raw, &hook); err != nil {
		return nil, err
	}

	if hook.Issue == nil {
		return nil, ErrInvalidPostReceiveHook
	}

	return &hook, nil
}

func (h *IssueHook) IsOpened() bool {
	return h.Action == "opened"
}

// we will know if it is a comment if the Action is "created"
// since with NSQ we can't parse the headers
func (h *IssueHook) IsComment() bool {
	return h.Action == "created"
}

func ParsePullRequestHook(raw []byte) (*PullRequestHook, error) {
	hook := PullRequestHook{}
	if err := json.Unmarshal(raw, &hook); err != nil {
		return nil, err
	}

	// it is possible the JSON was parsed, however,
	// was not from Github (maybe was from Bitbucket)
	// So we'll check to be sure certain key fields
	// were populated
	if hook.PullRequest == nil {
		return nil, ErrInvalidPostReceiveHook
	}

	return &hook, nil
}

func (h *PullRequestHook) IsOpened() bool {
	return h.Action == "opened"
}

func (h *PullRequestHook) IsSynchronize() bool {
	return h.Action == "synchronize"
}

var ErrInvalidPostReceiveHook = errors.New("Invalid Post Receive Hook")

type PostReceiveHook struct {
	Before  string      `json:"before"`
	After   string      `json:"after"`
	Ref     string      `json:"ref"`
	Repo    *Repository `json:"repository"`
	Commits []*Commit   `json:"commits"`
	Head    *Commit     `json:"head_commit"`
	Deleted bool        `json:"deleted"`
}

func ParseHook(raw []byte) (*PostReceiveHook, error) {
	hook := PostReceiveHook{}
	if err := json.Unmarshal(raw, &hook); err != nil {
		return nil, err
	}

	// it is possible the JSON was parsed, however,
	// was not from Github (maybe was from Bitbucket)
	// So we'll check to be sure certain key fields
	// were populated
	switch {
	case hook.Repo == nil:
		return nil, ErrInvalidPostReceiveHook
	case len(hook.Ref) == 0:
		return nil, ErrInvalidPostReceiveHook
	}

	return &hook, nil
}

func (h *PostReceiveHook) IsGithubPages() bool {
	return strings.HasSuffix(h.Ref, "/gh-pages")
}

func (h *PostReceiveHook) IsTag() bool {
	return strings.HasPrefix(h.Ref, "refs/tags/")
}

func (h *PostReceiveHook) IsHead() bool {
	return strings.HasPrefix(h.Ref, "refs/heads/")
}

func (h *PostReceiveHook) Branch() string {
	return strings.Replace(h.Ref, "refs/heads/", "", -1)
}

func (h *PostReceiveHook) IsDeleted() bool {
	return h.Deleted || h.After == "0000000000000000000000000000000000000000"
}
