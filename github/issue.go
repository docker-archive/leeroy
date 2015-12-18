package github

import (
	"fmt"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/crosbymichael/octokat"
)

const (
	pollKey      = "USER POLL"
	pollTemplate = `*USER POLL*

*The best way to get notified when there are changes in this discussion is by clicking the Subscribe button in the top right.*

The people listed below have appreciated your meaningful discussion with a random +1:
`
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
		repo := nameWithOwner(issueHook.Repo)
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
	if err := g.maybeClaimIssue(issueHook); err != nil {
		return err
	}

	return g.maybeOpinion(issueHook)
}

func (g GitHub) maybeClaimIssue(issueHook *octokat.IssueHook) error {
	var labelmap map[string]string = map[string]string{
		"#dibs":    "status/claimed",
		"#claimed": "status/claimed",
		"#mine":    "status/claimed",
	}

	repo := nameWithOwner(issueHook.Repo)

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

func (g GitHub) maybeOpinion(issueHook *octokat.IssueHook) error {
	body := strings.TrimSpace(issueHook.Comment.Body)

	if body == "+1" {
		login := issueHook.Comment.User.Login
		commenters := map[string]int{login: issueHook.Comment.Id}

		options := &octokat.Options{
			QueryParams: map[string]string{"per_page": "100"},
		}

		repo := getRepo(issueHook.Repo)
		issueId := strconv.Itoa(issueHook.Issue.Number)
		comments, err := g.Client().Comments(repo, issueId, options)
		if err != nil {
			return err
		}

		var poll octokat.Comment

		for _, c := range comments {
			if strings.ToLower(c.User.Login) == g.User && strings.Contains(c.Body, pollKey) {
				poll = c
			} else if strings.TrimSpace(c.Body) == "+1" {
				commenters[c.User.Login] = c.Id
			}
		}

		if poll.Body != "" {
			if !strings.Contains(poll.Body, login) {
				for k := range commenters {
					poll.Body += fmt.Sprintf("\n@%s", k)
				}
				if _, err := g.Client().PatchComment(repo, strconv.Itoa(poll.Id), poll.Body); err != nil {
					return err
				}
			}
		} else {
			tmpl := pollTemplate
			for k := range commenters {
				tmpl += fmt.Sprintf("\n@%s", k)
			}
			if _, err := g.Client().AddComment(repo, issueId, tmpl); err != nil {
				return err
			}
		}

		for _, v := range commenters {
			if err := g.Client().RemoveComment(repo, v); err != nil {
				return err
			}
		}
	}

	return nil
}
