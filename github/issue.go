package github

import (
	"strings"

	"github.com/crosbymichael/octokat"
)

func (g GitHub) IssueInfoCheck(issueHook *octokat.IssueHook) error {
	body := strings.ToLower(issueHook.Issue.Body)
	// we don't care about proposals or features
	if strings.ContainsAny(body, "proposal & feature") {
		return nil
	}

	// parse if they gave us
	// docker info, docker version, uname -a
	if !strings.Contains(body, "docker version") || !strings.Contains(body, "docker info") || !strings.Contains(body, "uname -a") {
		// get content
		repo := getRepo(issueHook.Repo)
		content, err := g.getContent(repo, issueHook.Issue.Number, false)
		if err != nil {
			return err
		}
		// comment on the issue

		if err := g.addNeedMoreInfoComment(repo, issueHook.Issue.Number, content); err != nil {
			return err
		}
	}

	return nil
}

func (g GitHub) LabelIssueComment(issueHook *octokat.IssueHook) error {
	var labelmap map[string]string = map[string]string{
		"#dibs":    "status/claimed",
		"#claimed": "status/claimed",
		"#mine":    "status/claimed",
		"windows":  "os/windows",
	}

	repo := getRepo(issueHook.Repo)

	for token, label := range labelmap {
		// if comment matches predefined actions AND author is not bot
		if strings.Contains(strings.ToLower(issueHook.Comment.Body), token) && g.User != issueHook.Sender.Login {
			if err := g.addLabel(repo, issueHook.Issue.Number, label); err != nil {
				return err
			}
		}
	}

	return nil
}
