package github

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/crosbymichael/octokat"
)

const (
	pollKey      = "USER POLL"
	pollTemplate = `*USER POLL*

*The best way to get notified of updates is to use the _Subscribe_ button on this page.*

Please don't use "+1" or "I have this too" comments on issues. We automatically
collect those comments to keep the thread short.

The people listed below have upvoted this issue by leaving a +1 comment:
`
)

// LabelIssueComment checks if someone has claimed dibs on this issue
func (g GitHub) LabelIssueComment(issueHook *octokat.IssueHook) error {
	if err := g.maybeClaimIssue(issueHook); err != nil {
		return err
	}

	return g.maybeOpinion(issueHook)
}

func (g GitHub) maybeClaimIssue(issueHook *octokat.IssueHook) error {
	labelmap := map[string]string{
		"#dibs":    "status/claimed",
		"#claimed": "status/claimed",
		"#mine":    "status/claimed",
	}

	repo := nameWithOwner(issueHook.Repo)

	for token, label := range labelmap {
		// if comment matches predefined actions AND author is not bot
		if strings.Contains(strings.ToLower(issueHook.Comment.Body), token) && g.User != issueHook.Sender.Login {
			logrus.Debugf("Adding label %#v to issue %d", label, issueHook.Issue.Number)
			if err := g.addLabel(repo, issueHook.Issue.Number, label); err != nil {
				return err
			}
			logrus.Infof("Added label %#v to issue %d", label, issueHook.Issue.Number)
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
		issueID := strconv.Itoa(issueHook.Issue.Number)
		comments, err := g.Client().Comments(repo, issueID, options)
		if err != nil {
			return err
		}

		var poll octokat.Comment

		for _, c := range comments {
			if strings.ToLower(c.User.Login) == g.User && strings.Contains(c.Body, pollKey) {
				poll = c
			} else if strings.TrimSpace(c.Body) == "+1" || strings.TrimSpace(c.Body) == ":+1:" {
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
			if _, err := g.Client().AddComment(repo, issueID, tmpl); err != nil {
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

func labelFromVersion(version, suffix string) string {
	switch {
	// Dev suffix is associated with a master build.
	case suffix == "dev":
		return "version/master"
	// For a version `X.Y.Z`, add a label of the form `version/X.Y`.
	case strings.HasPrefix(suffix, "cs"):
		fallthrough
	case strings.HasPrefix(suffix, "rc"):
		fallthrough
	case strings.HasPrefix(suffix, "ce"):
		fallthrough
	case strings.HasPrefix(suffix, "ee"):
		fallthrough
	case suffix == "":
		return "version/" + version[0:strings.LastIndex(version, ".")]
	// The default for unknown suffix is to consider the version unsupported.
	default:
		return "version/unsupported"
	}
}

// IssueAddVersionLabel adds a version label to an issue if it matches the regex
func (g GitHub) IssueAddVersionLabel(issueHook *octokat.IssueHook) error {
	serverVersion := regexp.MustCompile(`Server:\s+Version:\s+(\d+\.\d+\.\d+)-?(\S*)`)
	versionSubmatch := serverVersion.FindStringSubmatch(issueHook.Issue.Body)
	if len(versionSubmatch) < 3 {
		return nil
	}

	label := labelFromVersion(versionSubmatch[1], versionSubmatch[2])
	return g.addLabel(nameWithOwner(issueHook.Repo), issueHook.Issue.Number, label)
}
