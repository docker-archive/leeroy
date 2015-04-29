package github

import (
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/crosbymichael/octokat"
)

func (g GitHub) IssueInfoCheck(issueHook *octokat.IssueHook) error {
	body := strings.ToLower(issueHook.Issue.Body)
	title := strings.ToLower(issueHook.Issue.Title)

	// we don't care about proposals or features
	if strings.Contains(title, "proposal") || strings.Contains(title, "feature") {
		log.Debugf("Issue is talking about a proposal or feature so ignoring.")
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
		log.Debugf("commenting on issue %d about needing more info", issueHook.Issue.Number)
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
			log.Debugf("Adding label %#v to issue %d", label, issueHook.Issue.Number)
			if err := g.addLabel(repo, issueHook.Issue.Number, label); err != nil {
				return err
			}
			log.Infof("Added label %#v to issue %d", label, issueHook.Issue.Number)
		}
	}

	return nil
}
