package github

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/crosbymichael/octokat"
)

type PullRequestReviewCommentHook struct {
	Action      string
	PullRequest *octokat.PullRequest
	Comment     *octokat.Comment
	Repo        *octokat.Repository
}

func (pr PullRequestReviewCommentHook) IsOpen() bool {
	return pr.PullRequest.State == "open"
}

func (g GitHub) MoveTriageForward(repo *octokat.Repository, number int, comment *octokat.Comment) error {
	if !isMaintainer(comment.User) {
		return nil
	}

	nwo := nameWithOwner(repo)

	triage, err := g.labelExist(nwo, number, triageLabel)
	if err != nil {
		return err
	}

	if triage {
		newLabel := designReviewLabel
		if strings.TrimSpace(comment.Body) == "LGTM" {
			newLabel = codeReviewLabel
		}

		if err := g.toggleLabels(nwo, number, triageLabel, newLabel); err != nil {
			return err
		}
	}

	return nil
}

func ParsePullRequestReviewCommentHook(body io.Reader) (PullRequestReviewCommentHook, error) {
	h := PullRequestReviewCommentHook{}
	if err := json.NewDecoder(body).Decode(&h); err != nil {
		return h, err
	}

	return h, nil
}

func isMaintainer(user octokat.User) bool {
	return (user.Type == "Owner" || user.Type == "Collaborator") && !bot(user)
}
