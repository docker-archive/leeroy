package github

import (
	"strings"

	"github.com/crosbymichael/octokat"
)

func (g GitHub) toggleLabels(repo octokat.Repo, issueNum int, labelToRemove, labelToAdd string) error {
	if err := g.removeLabel(repo, issueNum, labelToRemove); err != nil {
		return err
	}
	if err := g.addLabel(repo, issueNum, labelToAdd); err != nil {
		return err
	}
	return nil
}

func (g GitHub) addLabel(repo octokat.Repo, issueNum int, labels ...string) error {
	issue := octokat.Issue{
		Number: issueNum,
	}

	err := g.Client().ApplyLabel(repo, &issue, labels)
	if err == nil || strings.Contains(err.Error(), "Label does not exist") {
		return nil
	}
	return err
}

func (g GitHub) removeLabel(repo octokat.Repo, issueNum int, labels ...string) error {
	issue := octokat.Issue{
		Number: issueNum,
	}

	for _, label := range labels {
		if err := g.Client().RemoveLabel(repo, &issue, label); err != nil && !strings.Contains(err.Error(), "Label does not exist") {
			return err
		}
	}

	return nil
}
